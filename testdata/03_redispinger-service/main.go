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
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	pingredis "github.com/bloomberg/docket/testdata/02_ping-redis"
)

func main() {
	listenAddr := "localhost:0"
	if len(os.Args) >= 2 {
		listenAddr = os.Args[1]
	}

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("err from net.Listen: %v", err)
	}

	fmt.Printf("Listening on %s\n", ln.Addr())

	http.HandleFunc("/", handler)

	if err := http.Serve(ln, nil); err != nil {
		log.Fatalf("err from http.Serve: %v", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	redisAddr := r.URL.Query().Get("redisAddr")
	if redisAddr == "" {
		http.Error(w, "missing redisAddr query parameter", http.StatusBadRequest)
		return
	}

	pong, err := pingredis.Ping(redisAddr)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to ping redis: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%s\n", pong)
}
