// Copyright 2019 Bloomberg Finance L.P.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
