package cve

import (
	"bytes"
	"fmt"
	"io"

	"github.com/anchore/grype/grype/presenter/json"
	"github.com/anchore/grype/grype/presenter/models"
	"github.com/anchore/grype/grype/presenter/table"
	"github.com/anchore/syft/syft/format"
	"github.com/anchore/syft/syft/format/syftjson"
	"github.com/anchore/syft/syft/sbom"
)

type Report struct {
	table, json, sbom string
}

func (r Report) Get() (string, string, string) {
	return r.table, r.json, r.sbom
}

func newReport(
	model models.PresenterConfig,
	sbomdata *sbom.SBOM,
) (Report, error) {
	var err error
	_table, err := presToTable(model)
	if err != nil {
		err = fmt.Errorf("failed to write result in table: %w", err)
	}
	_json, err := presToJson(model)
	if err != nil {
		err = fmt.Errorf("failed to write result in json: %w", err)
	}
	_sbom, err := format.Encode(*sbomdata, syftjson.NewFormatEncoder())
	if err != nil {
		err = fmt.Errorf("failed to format sbom: %s", err)
	}
	return Report{
		table: string(_table),
		json:  string(_json),
		sbom:  string(_sbom),
	}, err
}

type Presenter interface {
	Present(io.Writer) error
}

func present(pres Presenter) ([]byte, error) {
	var buf bytes.Buffer
	if err := pres.Present(&buf); err != nil {
		return nil, fmt.Errorf("failed to write result in buf: %w", err)
	}
	return buf.Bytes(), nil
}

func presToTable(model models.PresenterConfig) ([]byte, error) {
	return present(
		table.NewPresenter(model, false),
	)
}

func presToJson(model models.PresenterConfig) ([]byte, error) {
	return present(
		json.NewPresenter(model),
	)
}
