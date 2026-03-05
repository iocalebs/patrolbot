package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/mediawiki"
)

const testWiki = "zwen"

func mwTestData(cfg config.Config, overwrite bool) error {
	wikiCfg, ok := cfg.Wikis[testWiki]
	if !ok {
		return fmt.Errorf("wiki '%s' not found in config", testWiki)
	}

	mwclient, capturer, err := setupMediaWiki(wikiCfg, overwrite)
	if err != nil {
		return err
	}

	generators := [](func(context.Context, *mediawiki.Client, *capturingTransport) error){
		tokensWarnings,
		tokensLogin,
		loginFailedWrongToken,
		loginSuccess,
		recentChanges,
		recentChangesError,
		recentChangesWarnings,
		logEvents,
		logEventsError,
		logEventsWarnings,
	}

	for _, generator := range generators {
		err = generator(context.TODO(), mwclient, capturer)
		if err != nil {
			return err
		}
	}

	return nil
}

func setupMediaWiki(cfg config.Wiki, overwrite bool) (*mediawiki.Client, *capturingTransport, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, nil, err
	}

	transport := &capturingTransport{
		rt:        http.DefaultTransport,
		overwrite: overwrite,
	}
	client := &http.Client{
		Jar:       jar,
		Transport: transport,
	}
	mwclient := mediawiki.NewClient(cfg, client, slog.Default(), "")

	return mwclient, transport, nil
}

func tokensLogin(ctx context.Context, mwclient *mediawiki.Client, transport *capturingTransport) error {
	_, err := mwclient.Tokens(ctx, "login")
	if err != nil {
		return err
	}

	return transport.writeCapture("testdata/mwtokens_login.json")
}

func tokensWarnings(ctx context.Context, mwclient *mediawiki.Client, transport *capturingTransport) error {
	transport.requestMutator = func(req *http.Request) {
		q := req.URL.Query()
		q.Set("type", "foo")
		req.URL.RawQuery = q.Encode()
	}

	_, err := mwclient.Tokens(ctx, "login")
	if err != nil && !errors.As(err, new(*mediawiki.APIError)) {
		return err
	}

	return transport.writeCapture("testdata/mwtokens_warnings.json")
}

func loginFailedWrongToken(ctx context.Context, mwclient *mediawiki.Client, transport *capturingTransport) error {
	err := mwclient.Login(ctx, "foo")
	if err != nil && !errors.Is(err, mediawiki.ErrLoginFailed) {
		return err
	}

	return transport.writeCapture("testdata/mwlogin_wrongtoken.json")
}

func loginSuccess(ctx context.Context, mwclient *mediawiki.Client, transport *capturingTransport) error {
	tokens, err := mwclient.Tokens(ctx, "login")
	if err != nil {
		return err
	}

	err = mwclient.Login(ctx, tokens.Login)
	if err != nil {
		return err
	}

	return transport.writeCapture("testdata/mwlogin_success.json")
}

func recentChanges(ctx context.Context, mwclient *mediawiki.Client, transport *capturingTransport) error {
	params := mediawiki.RecentChangesQueryParams{
		RCStart: time.Now(),
		RCEnd:   time.Now().Add(-7 * 24 * time.Hour),
		RCShow:  "!patrolled",
		RCProp:  []string{"ids", "loginfo", "title", "user"},
	}

	paginator := mediawiki.NewRecentChangesPaginator(mwclient, params)

	errs := []error{}

	maxPages := 3
	for curPage := 1; paginator.HasMorePages() && curPage <= maxPages; curPage++ {
		_, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}

		if curPage == maxPages {
			err = transport.removePagination()
			if err != nil {
				return err
			}
		}

		err = transport.writeCapture(fmt.Sprintf("testdata/mwrecentchanges_unpatrolled%d.json", curPage))
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func recentChangesError(ctx context.Context, mwclient *mediawiki.Client, transport *capturingTransport) error {
	params := mediawiki.RecentChangesQueryParams{
		RCShow: "patrolled|!patrolled",
	}
	paginator := mediawiki.NewRecentChangesPaginator(mwclient, params)

	_, err := paginator.NextPage(ctx)
	if err != nil && !errors.As(err, new(*mediawiki.APIError)) {
		return err
	}

	return transport.writeCapture("testdata/mwrecentchanges_errors.json")
}

func recentChangesWarnings(ctx context.Context, mwclient *mediawiki.Client, transport *capturingTransport) error {
	params := mediawiki.RecentChangesQueryParams{}
	params.RCShow = "foo"
	paginator := mediawiki.NewRecentChangesPaginator(mwclient, params)

	_, err := paginator.NextPage(ctx)
	if err != nil && !errors.As(err, new(*mediawiki.APIError)) {
		return err
	}

	err = transport.removePagination()
	if err != nil {
		return err
	}

	return transport.writeCapture("testdata/mwrecentchanges_warnings.json")
}

func logEvents(ctx context.Context, mwclient *mediawiki.Client, transport *capturingTransport) error {
	params := mediawiki.LogEventsQueryParams{
		LEStart: time.Now(),
		LEEnd:   time.Now().Add(-2 * time.Hour),
	}
	paginator := mediawiki.NewLogEventsPaginator(mwclient, params)

	errs := []error{}

	maxPages := 3
	for curPage := 1; paginator.HasMorePages() && curPage < maxPages; curPage++ {
		_, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}

		if curPage == maxPages {
			err = transport.removePagination()
			if err != nil {
				return err
			}
		}

		err = transport.writeCapture(fmt.Sprintf("testdata/mwlogevents%d.json", curPage))
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func logEventsError(ctx context.Context, mwclient *mediawiki.Client, transport *capturingTransport) error {
	params := mediawiki.LogEventsQueryParams{
		LEType: "foo",
	}
	paginator := mediawiki.NewLogEventsPaginator(mwclient, params)

	_, err := paginator.NextPage(ctx)
	if err != nil && !errors.As(err, new(*mediawiki.APIError)) {
		return err
	}

	return transport.writeCapture("testdata/mwlogevents_errors.json")
}

func logEventsWarnings(ctx context.Context, mwclient *mediawiki.Client, transport *capturingTransport) error {
	params := mediawiki.LogEventsQueryParams{
		LEProp: "foo",
	}
	paginator := mediawiki.NewLogEventsPaginator(mwclient, params)

	_, err := paginator.NextPage(ctx)
	if err != nil && !errors.As(err, new(*mediawiki.APIError)) {
		return err
	}

	err = transport.removePagination()
	if err != nil {
		return err
	}

	return transport.writeCapture("testdata/mwlogevents_warnings.json")
}
