package cve

import (
	"bytes"
	"fmt"
	"jamel/pkg/http"
	"log"
	"sync"
	"time"

	"github.com/anchore/clio"
	"github.com/anchore/grype/cmd/grype/cli/options"
	"github.com/anchore/grype/grype"
	"github.com/anchore/grype/grype/db/legacy/distribution"
	"github.com/anchore/grype/grype/matcher"
	"github.com/anchore/grype/grype/matcher/dotnet"
	"github.com/anchore/grype/grype/matcher/golang"
	"github.com/anchore/grype/grype/matcher/java"
	"github.com/anchore/grype/grype/matcher/javascript"
	"github.com/anchore/grype/grype/matcher/python"
	"github.com/anchore/grype/grype/matcher/ruby"
	"github.com/anchore/grype/grype/matcher/stock"
	"github.com/anchore/grype/grype/pkg"
	"github.com/anchore/grype/grype/presenter/models"
	"github.com/anchore/grype/grype/presenter/table"
	"github.com/anchore/grype/grype/vex"
	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/cataloging"
)

const (
	certfile = "/tmp/cert"
)

type Cve struct {
	opts *options.Grype
	m    sync.Mutex
}

func New() (*Cve, error) {
	_cve := &Cve{
		opts: options.DefaultGrype(clio.Identification{
			Name:    "cve",
			Version: "5.0",
		}),
	}
	go _cve.Update()
	return _cve, nil
}

func (c *Cve) UpdateTicker() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		if err := c.Update(); err != nil {
			log.Println(fmt.Errorf("failed to update cve db: %w", err))
		}
	}
}

func (c *Cve) getProviderConfig() pkg.ProviderConfig {
	config := syft.DefaultCreateSBOMConfig()
	config.Packages.JavaArchive.IncludeIndexedArchives = c.opts.Search.IncludeIndexedArchives
	config.Packages.JavaArchive.IncludeUnindexedArchives = c.opts.Search.IncludeUnindexedArchives
	config.Compliance.MissingVersion = cataloging.ComplianceActionDrop
	return pkg.ProviderConfig{
		SyftProviderConfig: pkg.SyftProviderConfig{
			RegistryOptions:        c.opts.Registry.ToOptions(),
			Exclusions:             c.opts.Exclusions,
			SBOMOptions:            config,
			Platform:               c.opts.Platform,
			Name:                   c.opts.Name,
			DefaultImagePullSource: c.opts.DefaultImagePullSource,
		},
		SynthesisConfig: pkg.SynthesisConfig{
			GenerateMissingCPEs: c.opts.GenerateMissingCPEs,
		},
	}
}

func (c *Cve) getMatchers() []matcher.Matcher {
	return matcher.NewDefaultMatchers(
		matcher.Config{
			Java: java.MatcherConfig{
				ExternalSearchConfig: c.opts.ExternalSources.ToJavaMatcherConfig(),
				UseCPEs:              c.opts.Match.Java.UseCPEs,
			},
			Ruby:       ruby.MatcherConfig(c.opts.Match.Ruby),
			Python:     python.MatcherConfig(c.opts.Match.Python),
			Dotnet:     dotnet.MatcherConfig(c.opts.Match.Dotnet),
			Javascript: javascript.MatcherConfig(c.opts.Match.Javascript),
			Golang: golang.MatcherConfig{
				UseCPEs:                                c.opts.Match.Golang.UseCPEs,
				AlwaysUseCPEForStdlib:                  c.opts.Match.Golang.AlwaysUseCPEForStdlib,
				AllowMainModulePseudoVersionComparison: c.opts.Match.Golang.AllowMainModulePseudoVersionComparison,
			},
			Stock: stock.MatcherConfig(c.opts.Match.Stock),
		},
	)
}

func (c *Cve) Update() error {
	c.m.Lock()
	defer c.m.Unlock()

	log.Println("start update task")
	defer log.Println("updated finished")

	if err := http.SaveCAs([]string{
		"https://toolbox-data.anchore.io",
		"https://grype.anchore.io",
	}, certfile); err != nil {
		return fmt.Errorf("ca error: %w", err)
	}
	config := c.opts.DB.ToCuratorConfig()
	config.RequireUpdateCheck = true
	config.CACert = certfile

	curator, err := distribution.NewCurator(config)
	if err != nil {
		return fmt.Errorf("failed to get curator: %w", err)
	}
	if _, err := curator.Update(); err != nil {
		return fmt.Errorf("unable to update vulnerability database: %w", err)
	}
	return nil
}

func (c *Cve) Get(cvetype string, input string) ([]byte, error) {
	c.m.Lock()
	defer c.m.Unlock()

	store, status, closer, err := grype.LoadVulnerabilityDB(
		c.opts.DB.ToCuratorConfig(),
		c.opts.DB.AutoUpdate,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to load vulnerability db: %w", err)
	}
	defer closer.Close()

	packages, pkgcontext, sbomdata, err := pkg.Provide(
		fmt.Sprintf("%s:%s", cvetype, input),
		c.getProviderConfig(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load source data for analyze: %w", err)
	}

	mathcer := grype.VulnerabilityMatcher{
		Store:          *store,
		IgnoreRules:    c.opts.Ignore,
		NormalizeByCVE: c.opts.ByCVE,
		FailSeverity:   c.opts.FailOnSeverity(),
		Matchers:       c.getMatchers(),
		VexProcessor: vex.NewProcessor(vex.ProcessorOptions{
			Documents:   c.opts.VexDocuments,
			IgnoreRules: c.opts.Ignore,
		}),
	}

	remaining, ignored, err := mathcer.FindMatches(packages, pkgcontext)
	if err != nil {
		return nil, fmt.Errorf("failed to find matches: %w", err)
	}

	model := models.PresenterConfig{
		ID:               clio.Identification{},
		Matches:          *remaining,
		IgnoredMatches:   ignored,
		Packages:         packages,
		Context:          pkgcontext,
		MetadataProvider: store,
		SBOM:             sbomdata,
		AppConfig:        c.opts,
		DBStatus:         status,
	}
	var (
		buf  bytes.Buffer
		pres = table.NewPresenter(model, false)
		// pres = json.NewPresenter(model)
	)
	if err := pres.Present(&buf); err != nil {
		return nil, fmt.Errorf("failed to write result in buf: %w", err)
	}
	return buf.Bytes(), nil
}

// var (
// 	opts = options.DefaultGrype(clio.Identification{
// 		Name:    "cve",
// 		Version: "5.0",
// 	})
// )

// func getProviderConfig(opts *options.Grype) pkg.ProviderConfig {
// 	cfg := syft.DefaultCreateSBOMConfig()
// 	cfg.Packages.JavaArchive.IncludeIndexedArchives = opts.Search.IncludeIndexedArchives
// 	cfg.Packages.JavaArchive.IncludeUnindexedArchives = opts.Search.IncludeUnindexedArchives
// 	cfg.Compliance.MissingVersion = cataloging.ComplianceActionDrop

// 	return pkg.ProviderConfig{
// 		SyftProviderConfig: pkg.SyftProviderConfig{
// 			RegistryOptions:        opts.Registry.ToOptions(),
// 			Exclusions:             opts.Exclusions,
// 			SBOMOptions:            cfg,
// 			Platform:               opts.Platform,
// 			Name:                   opts.Name,
// 			DefaultImagePullSource: opts.DefaultImagePullSource,
// 		},
// 		SynthesisConfig: pkg.SynthesisConfig{
// 			GenerateMissingCPEs: opts.GenerateMissingCPEs,
// 		},
// 	}
// }

// func getMatchers(opts *options.Grype) []matcher.Matcher {
// 	return matcher.NewDefaultMatchers(
// 		matcher.Config{
// 			Java: java.MatcherConfig{
// 				ExternalSearchConfig: opts.ExternalSources.ToJavaMatcherConfig(),
// 				UseCPEs:              opts.Match.Java.UseCPEs,
// 			},
// 			Ruby:       ruby.MatcherConfig(opts.Match.Ruby),
// 			Python:     python.MatcherConfig(opts.Match.Python),
// 			Dotnet:     dotnet.MatcherConfig(opts.Match.Dotnet),
// 			Javascript: javascript.MatcherConfig(opts.Match.Javascript),
// 			Golang: golang.MatcherConfig{
// 				UseCPEs:                                opts.Match.Golang.UseCPEs,
// 				AlwaysUseCPEForStdlib:                  opts.Match.Golang.AlwaysUseCPEForStdlib,
// 				AllowMainModulePseudoVersionComparison: opts.Match.Golang.AllowMainModulePseudoVersionComparison,
// 			},
// 			Stock: stock.MatcherConfig(opts.Match.Stock),
// 		},
// 	)
// }
