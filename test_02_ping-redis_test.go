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

package docket

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

func Test_02_ping_redis(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping docker-dependent test suite in short mode")
	}

	suite.Run(t, &PingRedisSuite{
		dir: filepath.Join("testdata", "02_ping-redis"),
	})
}

type PingRedisSuite struct {
	suite.Suite

	dir string
}

func (s *PingRedisSuite) Test_DebugMode() {
	s.testMode(context.Background(), "debug")
}

func (s *PingRedisSuite) Test_FullMode() {
	s.testMode(context.Background(), "full")
}

//------------------------------------------------------------------------------

func (s *PingRedisSuite) testMode(ctx context.Context, mode string) {
	cmd := exec.CommandContext(ctx, "go", "test", "-v")
	cmd.Args = append(cmd.Args, goTestCoverageArgs(s.T().Name())...)
	cmd.Args = append(cmd.Args, goTestRaceDetectorArgs()...)
	cmd.Dir = s.dir
	cmd.Env = append(os.Environ(), "DOCKET_MODE="+mode, "DOCKET_DOWN=1")

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s", out)
	}
	s.NoError(err)
}
