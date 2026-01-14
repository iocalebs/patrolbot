package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type capturedResponse struct {
	statusCode int
	body       []byte
}

type capturingTransport struct {
	rt             http.RoundTripper
	overwrite      bool                    // if true, overwrite existing files when writing captures
	requestMutator func(req *http.Request) // to mutate request before it is sent
	captured       *capturedResponse       // captured response
}

func (transport *capturingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	defer func() {
		transport.requestMutator = nil
	}()

	if transport.requestMutator != nil {
		transport.requestMutator(req)
	}

	resp, err := transport.rt.RoundTrip(req)
	if err != nil {
		return nil, err
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
	transport.captured = &capturedResponse{
		statusCode: resp.StatusCode,
		body:       bodyBytes,
	}

	// Restore the response body so it can be read again
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return resp, nil
}

// Removes "continue" field from captured JSON response to generate
// a single page of test data when more than one is not desired.
func (transport *capturingTransport) removePagination() error {
	if transport.captured == nil {
		return errors.New("no response captured")
	}

	var respBody map[string]any

	err := json.Unmarshal(transport.captured.body, &respBody)
	if err != nil {
		return fmt.Errorf("error unmarshaling response body: %w", err)
	}

	delete(respBody, "continue")

	transport.captured.body, err = json.MarshalIndent(respBody, "", "  ")
	if err != nil {
		return err
	}

	transport.captured.body = append(transport.captured.body, '\n')

	return nil
}

func (transport *capturingTransport) writeCapture(outPath string) error {
	if transport.captured == nil {
		return errors.New("no response captured")
	}

	_, err := os.Stat(outPath)
	if err == nil {
		if !transport.overwrite {
			log.Printf("skipped %s", outPath)
			return nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot stat test file: %w", err)
	}

	err = os.WriteFile(outPath, transport.captured.body, 0600)
	if err != nil {
		return fmt.Errorf("error writing capture to file: %w", err)
	}

	log.Printf("%d %s: %s", transport.captured.statusCode, http.StatusText(transport.captured.statusCode), outPath)

	return nil
}
