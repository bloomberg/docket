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

package pingredis

import (
	"bufio"
	"fmt"
	"net"
)

var ErrUnexpectedResponse = fmt.Errorf("unexpected response")

// Ping pings the Redis instance at redisAddr.
//
// This code is for demonstration purposes only.
// For production-worthy code, you should use an established redis client library.
//
func Ping(redisAddr string) (string, error) {
	conn, err := net.Dial("tcp", redisAddr)
	if err != nil {
		return "", fmt.Errorf("net.Dial failed: %w", err)
	}
	defer conn.Close()

	_, err = fmt.Fprintf(conn, "PING\r\n")
	if err != nil {
		return "", fmt.Errorf("failed to write: %w", err)
	}

	scanner := bufio.NewScanner(conn)
	var pong string
	for pong == "" && scanner.Scan() {
		line := scanner.Text()
		if line[0] != '+' {
			return "", fmt.Errorf("%w: %q", ErrUnexpectedResponse, line)
		}

		pong = line[1:]
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scanner error: %w", err)
	}

	return pong, nil
}
