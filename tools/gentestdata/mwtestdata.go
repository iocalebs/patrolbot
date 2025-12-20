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

func mwTestData(overwrite bool) error {
	mwclient, capturer, err := setup(overwrite)

	generators := [](func(context.Context, *mediawiki.Client, *capturingTransport) error){
		actionError,
		actionWarnings,
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
		if err != nil {
			return err
		}

		err = generator(context.TODO(), mwclient, capturer)
		if err != nil {
			return err
		}
	}

	return nil
}

func setup(overwrite bool) (*mediawiki.Client, *capturingTransport, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}

	wiki, ok := cfg.Wikis[testWiki]
	if !ok {
		return nil, nil, fmt.Errorf("wiki '%s' not found in config", testWiki)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, nil, err
	}

	transport := &capturingTransport{}
	transport.rt = http.DefaultTransport
	transport.overwrite = overwrite

	client := &http.Client{}
	client.Jar = jar
	client.Transport = transport

	mwclient := mediawiki.NewClient(wiki, client, *slog.Default())

	return mwclient, transport, nil
}

func actionError(ctx context.Context, mwclient *mediawiki.Client, transport *capturingTransport) error {
	transport.requestMutator = func(req *http.Request) {
		q := req.URL.Query()
		q.Set("action", "foo")
		req.URL.RawQuery = q.Encode()
	}

	_, err := mwclient.LoginToken(ctx)
	if err != nil && !errors.Is(err, mediawiki.ErrResponseError) {
		return err
	}

	return transport.writeCapture("testdata/mwaction_error.json")
}

func actionWarnings(ctx context.Context, mwclient *mediawiki.Client, transport *capturingTransport) error {
	transport.requestMutator = func(req *http.Request) {
		q := req.URL.Query()
		q.Set("type", "foo")
		req.URL.RawQuery = q.Encode()
	}

	_, err := mwclient.LoginToken(ctx)
	if err != nil && !errors.Is(err, mediawiki.ErrResponseWarnings) {
		return err
	}

	return transport.writeCapture("testdata/mwaction_warnings.json")
}

func tokensLogin(ctx context.Context, mwclient *mediawiki.Client, transport *capturingTransport) error {
	_, err := mwclient.LoginToken(ctx)
	if err != nil {
		return err
	}

	return transport.writeCapture("testdata/mwtokens_login.json")
}

func loginFailedWrongToken(ctx context.Context, mwclient *mediawiki.Client, transport *capturingTransport) error {
	err := mwclient.Login(ctx, "foo")
	if err != nil && !errors.Is(err, mediawiki.ErrLoginFailed) {
		return err
	}

	return transport.writeCapture("testdata/mwlogin_failed_wrongtoken.json")
}

func loginSuccess(ctx context.Context, mwclient *mediawiki.Client, transport *capturingTransport) error {
	token, err := mwclient.LoginToken(ctx)
	if err != nil {
		return err
	}

	err = mwclient.Login(ctx, token)
	if err != nil {
		return err
	}

	return transport.writeCapture("testdata/mwlogin_success.json")
}

func recentChanges(ctx context.Context, mwclient *mediawiki.Client, transport *capturingTransport) error {
	params := mediawiki.RecentChangesQueryParams{}
	params.RCStart = time.Now()
	params.RCEnd = time.Now().Add(-1 * time.Hour)
	params.RCShow = "!patrolled"

	paginator := mediawiki.NewRecentChangesPaginator(mwclient, params)

	errs := []error{}

	for curPage := 1; paginator.HasMorePages(); curPage++ {
		_, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}

		err = transport.writeCapture(fmt.Sprintf("testdata/mwrecentchanges_unpatrolled%d.json", curPage))
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func recentChangesError(ctx context.Context, mwclient *mediawiki.Client, transport *capturingTransport) error {
	params := mediawiki.RecentChangesQueryParams{}
	params.RCShow = "patrolled|!patrolled"
	paginator := mediawiki.NewRecentChangesPaginator(mwclient, params)

	_, err := paginator.NextPage(ctx)
	if err != nil && !errors.Is(err, mediawiki.ErrResponseError) {
		return err
	}

	return transport.writeCapture("testdata/mwrecentchanges_error.json")
}

func recentChangesWarnings(ctx context.Context, mwclient *mediawiki.Client, transport *capturingTransport) error {
	params := mediawiki.RecentChangesQueryParams{}
	params.RCShow = "foo"
	paginator := mediawiki.NewRecentChangesPaginator(mwclient, params)

	_, err := paginator.NextPage(ctx)
	if err != nil && !errors.Is(err, mediawiki.ErrResponseWarnings) {
		return err
	}

	err = transport.removePagination()
	if err != nil {
		return err
	}

	return transport.writeCapture("testdata/mwrecentchanges_warnings.json")
}
