package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cookiejar"

	"github.com/iocalebs/patrolbot/internal/config"
	"github.com/iocalebs/patrolbot/internal/mediawiki"
)

func mwTestData() error {
	generators := [](func(*mediawiki.Client, *capturingTransport) error){
		actionError,
		actionWarnings,
		tokensLogin,
		loginFailedWrongToken,
		loginSuccess,
	}

	for _, generator := range generators {
		mwclient, capturer, err := setup()
		if err != nil {
			return err
		}

		err = generator(mwclient, capturer)
		if err != nil {
			return err
		}
	}

	return nil
}

func setup() (*mediawiki.Client, *capturingTransport, error) {
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

	client := &http.Client{}
	client.Jar = jar
	client.Transport = transport

	mwclient := mediawiki.NewClient(wiki, client, *slog.Default())

	return mwclient, transport, nil
}

func actionError(mwclient *mediawiki.Client, capture *capturingTransport) error {
	capture.outPath = "testdata/mwaction_error.json"
	capture.requestMutator = func(req *http.Request) {
		q := req.URL.Query()
		q.Set("action", "foo")
		req.URL.RawQuery = q.Encode()
	}

	_, err := mwclient.LoginToken(context.Background())
	if err != nil && !errors.Is(err, mediawiki.ErrResponseError) {
		return err
	}

	return nil
}

func actionWarnings(mwclient *mediawiki.Client, capture *capturingTransport) error {
	capture.outPath = "testdata/mwaction_warnings.json"
	capture.requestMutator = func(req *http.Request) {
		q := req.URL.Query()
		q.Set("type", "foo")
		req.URL.RawQuery = q.Encode()
	}

	_, err := mwclient.LoginToken(context.Background())
	if err != nil && !errors.Is(err, mediawiki.ErrResponseWarnings) {
		return err
	}

	return nil
}

func tokensLogin(mwclient *mediawiki.Client, capture *capturingTransport) error {
	capture.skip = true // To avoid changing random token value and breaking test
	capture.outPath = "testdata/mwquerytokens_login.json"

	_, err := mwclient.LoginToken(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func loginFailedWrongToken(mwclient *mediawiki.Client, capture *capturingTransport) error {
	capture.outPath = "testdata/mwlogin_failed_wrongtoken.json"

	err := mwclient.Login(context.Background(), "foo")
	if err != nil && !errors.Is(err, mediawiki.ErrLoginFailed) {
		return err
	}

	return nil
}

func loginSuccess(mwclient *mediawiki.Client, capture *capturingTransport) error {
	token, err := mwclient.LoginToken(context.Background())
	if err != nil {
		return err
	}

	capture.outPath = "testdata/mwlogin_success.json"

	err = mwclient.Login(context.Background(), token)
	if err != nil {
		return err
	}

	return nil
}
