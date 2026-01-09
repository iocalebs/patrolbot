// Package reporter provides functionality for generating reports with data sourced from the MediaWiki [Action API].
//
// [Action API]: https://www.mediawiki.org/wiki/API:Action_API
package reporter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"path/filepath"
	"slices"
	"text/template"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/report/provider"
)

var (
	// ErrReportTypeNotFound indicates that a report type was specified that does not exist in config.
	ErrReportTypeNotFound = errors.New("invalid report type")

	// ErrInvalidConfig indicates that the report generator cannot proceed due to invalid config.
	ErrInvalidConfig = errors.New("configuration error")
)

// Reporter generates reports with data sourced from the MediaWiki [Action API].
//
// [Action API]: https://www.mediawiki.org/wiki/API:Action_API
type Reporter struct {
	cfg      config.Reports
	provider *provider.Provider
}

// New creates a new [Reporter] instance.
func New(cfg config.Reports, provider *provider.Provider) *Reporter {
	return &Reporter{
		cfg:      cfg,
		provider: provider,
	}
}

// Report writes a report to the given [io.Writer].
func (r *Reporter) Report(ctx context.Context, reportType string, w io.Writer) error {
	reportCfg, ok := r.cfg.Types[reportType]
	if !ok {
		return fmt.Errorf("%w: %s", ErrReportTypeNotFound, reportType)
	}

	if reportCfg.Template == "" {
		return fmt.Errorf("%w: no template configured for report type %s", ErrInvalidConfig, reportType)
	}

	tmpl, err := template.ParseFiles(filepath.Join(r.cfg.TemplateDir, reportCfg.Template))
	if err != nil {
		return fmt.Errorf("error parsing report template: %w", err)
	}

	data := r.provider.Data(ctx, reportCfg.Data)

	err = tmpl.Execute(w, data)
	if err != nil {
		return fmt.Errorf("error generating report from template: %w", err)
	}

	return nil
}

// ListReports writes the available report types to the given [io.Writer].
func (r *Reporter) ListReports(w io.Writer) error {
	keys := slices.Collect(maps.Keys(r.cfg.Types))
	slices.Sort(keys)

	for _, key := range keys {
		_, err := w.Write([]byte(key + "\n"))
		if err != nil {
			return fmt.Errorf("error listing reports: %w", err)
		}
	}

	return nil
}
