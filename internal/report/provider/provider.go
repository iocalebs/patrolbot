// Package provider sources report data from the MediaWiki [Action API].
//
// [Action API]: https://www.mediawiki.org/wiki/API:Action_API
package provider

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/iocalebs/patrolbot/internal/clock"
	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/mediawiki"
	"github.com/iocalebs/patrolbot/internal/report/data"
)

// Provider interacts with the MediaWiki [Action API] to obtain data for reports.
//
// [Action API]: https://www.mediawiki.org/wiki/API:Action_API
type Provider struct {
	clock    clock.Clock
	loggedIn bool
	logger   *slog.Logger
	mwclient *mediawiki.Client
	wiki     config.Wiki
}

// New creates a new [Source] instance.
func New(logger *slog.Logger, clock clock.Clock, mwclient *mediawiki.Client, wiki config.Wiki) *Provider {
	return &Provider{
		logger:   logger,
		clock:    clock,
		mwclient: mwclient,
		wiki:     wiki,
	}
}

// Data obtains MediaWiki data according to the configured report sources.
func (p *Provider) Data(ctx context.Context, dataConfig config.ReportData) data.ReportData {
	reportData := data.ReportData{}

	reportData.Config = p.wiki
	reportData.Config.Auth.Password = ""

	if dataConfig.PatrolExpiring == nil {
		return reportData
	}

	if !p.loggedIn {
		token, err := p.mwclient.LoginToken(ctx)
		if err != nil {
			p.logError(ctx, "Error obtaining login token from MediaWiki API:Tokens", err)
			reportData.Error = fmt.Sprintf("Error logging into MediaWiki: %v", err)

			return reportData
		}

		err = p.mwclient.Login(ctx, token)
		if err != nil {
			p.logError(ctx, "Error executing login via MediaWiki API:Login", err)
			reportData.Error = fmt.Sprintf("Error logging into MediaWiki: %v", err)

			return reportData
		}

		p.loggedIn = true
	}

	if dataConfig.PatrolExpiring != nil {
		expiring, err := p.patrolExpiring(ctx, *dataConfig.PatrolExpiring)
		if err != nil {
			msg := err.Error()
			msg = strings.ToUpper(msg[0:1]) + msg[1:]
			reportData.Error = msg
		}

		reportData.PatrolExpiring = expiring
	}

	return reportData
}

func (p *Provider) logError(ctx context.Context, msg string, err error) {
	var httpStatusError *mediawiki.HTTPStatusError
	if errors.As(err, &httpStatusError) {
		p.logger.ErrorContext(
			ctx,
			msg,
			"err",
			err,
			"statusCode",
			httpStatusError.StatusCode,
			"body",
			httpStatusError.Body,
		)
	} else {
		p.logger.ErrorContext(ctx, msg, "err", err)
	}
}
