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

package compose_test

import (
	"context"
	"os"
	"testing"

	"github.com/bloomberg/docket/internal/compose"
	"github.com/stretchr/testify/suite"
)

func Test_Compose(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping docker-dependent tests in short mode")
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir("testdata"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatal(err)
		}
	}()

	suite.Run(t, &ComposeSuite{
		Suite: suite.Suite{},
		ctx:   context.Background(),
	})
}

type ComposeSuite struct {
	suite.Suite

	ctx context.Context
}

func (s *ComposeSuite) Test_BadConfig_MissingImage() {
	cmp, cleanup, err := compose.NewCompose(s.ctx, "docket.bad-config.missing-image", "no-mode")
	s.Error(err)
	s.Regexp("Compose file is invalid", err)
	s.Nil(cmp)
	s.NoError(cleanup())
}

func (s *ComposeSuite) Test_BadConfig_MultipleTestServices() {
	cmp, cleanup, err := compose.NewCompose(s.ctx, "docket.bad-config.multiple-test-services", "no-mode")
	s.Error(err)
	s.Regexp("multiple test services found", err)
	s.Nil(cmp)
	s.NoError(cleanup())
}

func (s *ComposeSuite) Test_CannotFindDocketFiles() {
	cmp, cleanup, err := compose.NewCompose(s.ctx, "docket.nonExistentPrefix", "no-mode")
	s.Error(err)
	s.Regexp("no matching docket files found", err)
	s.Nil(cmp)
	s.NoError(cleanup())
}

func (s *ComposeSuite) Test_PullWithoutImage() {
	cmp, cleanup, err := compose.NewCompose(s.ctx, "docket.blank", "no-mode")
	defer func() { s.NoError(cleanup()) }()
	s.NoError(err)
	s.Require().NotNil(cmp)

	s.NoError(cmp.Pull(s.ctx, nil))
}

func (s *ComposeSuite) Test_GetPort() {
	cmp, cleanup, err := compose.NewCompose(s.ctx, "docket.published-ports", "full")
	defer func() { s.NoError(cleanup()) }()
	s.NoError(err)
	s.Require().NotNil(cmp)

	s.Require().NoError(cmp.Up(s.ctx))
	defer func() { s.Require().NoError(cmp.Down(s.ctx)) }()

	port, err := cmp.GetPort(s.ctx, "alice", 80)
	s.NotZero(port)
	s.NoError(err)

	_, err = cmp.GetPort(s.ctx, "alice", 9999)
	s.Error(err)

	_, err = cmp.GetPort(s.ctx, "bob", 9999)
	s.Error(err)
}

func (s *ComposeSuite) Test_RunTestsLocally() {
	cmp, cleanup, err := compose.NewCompose(s.ctx, "docket.blank", "no-mode")
	defer func() { s.NoError(cleanup()) }()
	s.NoError(err)
	s.Require().NotNil(cmp)

	s.NoError(cmp.RunTestfuncOrExecGoTest(s.ctx, "testName", func() {}))
}

func (s *ComposeSuite) Test_RunTestfuncOrExecGoTest() {
	cmp, cleanup, err := compose.NewCompose(s.ctx, "docket.test-service", "full")
	defer func() { s.NoError(cleanup()) }()
	s.NoError(err)
	s.Require().NotNil(cmp)

	s.Require().NoError(cmp.Up(s.ctx))
	defer func() { s.NoError(cmp.Down(s.ctx)) }()

	// This should run the testName inside the container, not run the function locally.
	s.NoError(cmp.RunTestfuncOrExecGoTest(s.ctx, "TestHelloWorld", func() {
		s.Fail("This function should not have been called!")
	}))
}

func (s *ComposeSuite) Test_RunTestfuncOrExecGoTest_StringCommand() {
	cmp, cleanup, err := compose.NewCompose(s.ctx, "docket.test-service", "string-command")
	defer func() { s.NoError(cleanup()) }()
	s.NoError(err)
	s.Require().NotNil(cmp)

	s.Require().NoError(cmp.Up(s.ctx))
	defer func() { s.NoError(cmp.Down(s.ctx)) }()

	// This should run the testName inside the container, not run the function locally.
	s.NoError(cmp.RunTestfuncOrExecGoTest(s.ctx, "TestHelloWorld", func() {
		s.Fail("This function should not have been called!")
	}))
}

func (s *ComposeSuite) Test_RunTestfuncOrExecGoTest_FailsWithABadPath() {
	cmp, cleanup, err := compose.NewCompose(s.ctx, "docket.test-service", "go-not-in-path")
	defer func() { s.NoError(cleanup()) }()
	s.NoError(err)
	s.Require().NotNil(cmp)

	s.Require().NoError(cmp.Up(s.ctx))
	defer func() { s.Require().NoError(cmp.Down(s.ctx)) }()

	err = cmp.RunTestfuncOrExecGoTest(s.ctx, "testName", func() {})
	s.Error(err)
	s.Regexp("failed to exec go test", err)
}
