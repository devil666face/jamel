package cve

import (
	"bytes"
	"fmt"
	"jamel/pkg/http"

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

var (
	opts = options.DefaultGrype(clio.Identification{
		Name:    "cve",
		Version: "5.0",
	})
)

func getProviderConfig(opts *options.Grype) pkg.ProviderConfig {
	cfg := syft.DefaultCreateSBOMConfig()
	cfg.Packages.JavaArchive.IncludeIndexedArchives = opts.Search.IncludeIndexedArchives
	cfg.Packages.JavaArchive.IncludeUnindexedArchives = opts.Search.IncludeUnindexedArchives
	cfg.Compliance.MissingVersion = cataloging.ComplianceActionDrop

	return pkg.ProviderConfig{
		SyftProviderConfig: pkg.SyftProviderConfig{
			RegistryOptions:        opts.Registry.ToOptions(),
			Exclusions:             opts.Exclusions,
			SBOMOptions:            cfg,
			Platform:               opts.Platform,
			Name:                   opts.Name,
			DefaultImagePullSource: opts.DefaultImagePullSource,
		},
		SynthesisConfig: pkg.SynthesisConfig{
			GenerateMissingCPEs: opts.GenerateMissingCPEs,
		},
	}
}

func getMatchers(opts *options.Grype) []matcher.Matcher {
	return matcher.NewDefaultMatchers(
		matcher.Config{
			Java: java.MatcherConfig{
				ExternalSearchConfig: opts.ExternalSources.ToJavaMatcherConfig(),
				UseCPEs:              opts.Match.Java.UseCPEs,
			},
			Ruby:       ruby.MatcherConfig(opts.Match.Ruby),
			Python:     python.MatcherConfig(opts.Match.Python),
			Dotnet:     dotnet.MatcherConfig(opts.Match.Dotnet),
			Javascript: javascript.MatcherConfig(opts.Match.Javascript),
			Golang: golang.MatcherConfig{
				UseCPEs:                                opts.Match.Golang.UseCPEs,
				AlwaysUseCPEForStdlib:                  opts.Match.Golang.AlwaysUseCPEForStdlib,
				AllowMainModulePseudoVersionComparison: opts.Match.Golang.AllowMainModulePseudoVersionComparison,
			},
			Stock: stock.MatcherConfig(opts.Match.Stock),
		},
	)
}

func Get(input string) ([]byte, error) {
	if err := Update(opts.DB); err != nil {
		return nil, err
	}

	store, status, closer, err := grype.LoadVulnerabilityDB(
		opts.DB.ToCuratorConfig(),
		opts.DB.AutoUpdate,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to load vulnerability db: %w", err)
	}
	defer closer.Close()

	packages, pkgcontext, sbomdata, err := pkg.Provide(input, getProviderConfig(opts))
	if err != nil {
		return nil, fmt.Errorf("failed to load source data for analyze: %w", err)
	}

	mathcer := grype.VulnerabilityMatcher{
		Store:          *store,
		IgnoreRules:    opts.Ignore,
		NormalizeByCVE: opts.ByCVE,
		FailSeverity:   opts.FailOnSeverity(),
		Matchers:       getMatchers(opts),
		VexProcessor: vex.NewProcessor(vex.ProcessorOptions{
			Documents:   opts.VexDocuments,
			IgnoreRules: opts.Ignore,
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
		AppConfig:        opts,
		DBStatus:         status,
	}
	var (
		buf  bytes.Buffer
		pres = table.NewPresenter(model, false)
	)
	if err := pres.Present(&buf); err != nil {
		return nil, fmt.Errorf("failed to write result in buf: %w", err)
	}
	return buf.Bytes(), nil
}

func Update(opts options.Database) error {
	if err := http.SaveCAs([]string{
		"https://toolbox-data.anchore.io",
		"https://grype.anchore.io",
	}, certfile); err != nil {
		return err
	}
	config := opts.ToCuratorConfig()
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
