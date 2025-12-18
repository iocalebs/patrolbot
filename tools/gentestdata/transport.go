package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

type capturingTransport struct {
	rt             http.RoundTripper
	outPath        string
	requestMutator func(req *http.Request)
	skip           bool
}

func (t *capturingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	defer func() {
		t.outPath = ""
		t.requestMutator = nil
		t.skip = false
	}()

	if t.requestMutator != nil {
		t.requestMutator(req)
	}

	resp, err := t.rt.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if t.outPath == "" || t.skip {
		return resp, nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var jsonData json.RawMessage

	err = json.Unmarshal(bodyBytes, &jsonData)
	if err != nil {
		return nil, err
	}

	bodyBytes, err = json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return nil, err
	}

	bodyBytes = append(bodyBytes, '\n')

	err = os.WriteFile(t.outPath, bodyBytes, 0600) //nolint:mnd
	if err != nil {
		return nil, err
	}

	// Restore the response body so it can be read again
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return resp, nil
}
