package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/bloomberg/docket"
)

// TestService makes an http request to the redispinger service.
func TestService(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()

	docket.Run(ctx, nil, t, func() {
		makeServiceRequest(t)
	})
}

func makeServiceRequest(t *testing.T) {
	pingerURL := os.Getenv("REDISPINGER_URL")
	if pingerURL == "" {
		t.Fatalf("missing REDISPINGER_URL")
	}

	t.Logf("pingerURL = %q", pingerURL)

	resp, err := http.Get(pingerURL)

	if err != nil {
		t.Fatalf("failed http.Get to %v: %v", pingerURL, err)
	}

	t.Logf("response = %#v", resp)

	defer resp.Body.Close()
	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Logf("could not read body: %v", err)
	} else {
		t.Logf("body: %q", body)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("non-%v response code %v", http.StatusOK, resp.StatusCode)
	}
}
