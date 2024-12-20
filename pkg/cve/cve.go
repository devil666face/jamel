package cve

import (
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
	go _cve.UpdateTicker()
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

func (c *Cve) GetUnwrap(cvetype string, input string) (string, string, string, error) {
	report, err := c.Get(cvetype, input)
	_table, _json, _sbom := report.Get()
	return _table, _json, _sbom, err
}

func (c *Cve) Get(cvetype string, input string) (Report, error) {
	if lock := c.m.TryLock(); !lock {
		return Report{}, fmt.Errorf("cve database is updating, wait before update is over")
	}
	defer c.m.Unlock()

	store, status, closer, err := grype.LoadVulnerabilityDB(
		c.opts.DB.ToCuratorConfig(),
		c.opts.DB.AutoUpdate,
	)

	if err != nil {
		return Report{}, fmt.Errorf("failed to load vulnerability db: %w", err)
	}
	defer closer.Close()

	packages, pkgcontext, sbomdata, err := pkg.Provide(
		fmt.Sprintf("%s:%s", cvetype, input),
		c.getProviderConfig(),
	)
	if err != nil {
		return Report{}, fmt.Errorf("failed to load source data for analyze: %w", err)
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
		return Report{}, fmt.Errorf("failed to find matches: %w", err)
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
	return newReport(
		model,
		sbomdata,
	)
}
