package compose

import (
	"context"
	"os"
	"testing"

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
	defer os.Chdir(origDir)

	suite.Run(t, &ComposeSuite{
		ctx: context.Background(),
	})
}

type ComposeSuite struct {
	suite.Suite

	ctx context.Context
}

func (s *ComposeSuite) Test_BadConfig_MissingImage() {
	cmp, cleanup, err := NewCompose(s.ctx, "docket.bad-config.missing-image", "no-mode")
	s.Error(err)
	s.Regexp("Compose file is invalid", err)
	s.Nil(cmp)
	s.NoError(cleanup())
}

func (s *ComposeSuite) Test_BadConfig_MultipleTestServices() {
	cmp, cleanup, err := NewCompose(s.ctx, "docket.bad-config.multiple-test-services", "no-mode")
	s.Error(err)
	s.Regexp("multiple test services found", err)
	s.Nil(cmp)
	s.NoError(cleanup())
}

func (s *ComposeSuite) Test_CannotFindDocketFiles() {
	cmp, cleanup, err := NewCompose(s.ctx, "docket.nonExistentPrefix", "no-mode")
	s.Error(err)
	s.Regexp("did not find any files", err)
	s.Nil(cmp)
	s.NoError(cleanup())
}

func (s *ComposeSuite) Test_PullWithoutImage() {
	cmp, cleanup, err := NewCompose(s.ctx, "docket.blank", "no-mode")
	defer func() { s.NoError(cleanup()) }()
	s.NoError(err)
	s.Require().NotNil(cmp)

	s.NoError(cmp.Pull(s.ctx))
}

func (s *ComposeSuite) Test_GetPort() {
	cmp, cleanup, err := NewCompose(s.ctx, "docket.published-ports", "full")
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
	cmp, cleanup, err := NewCompose(s.ctx, "docket.blank", "no-mode")
	defer func() { s.NoError(cleanup()) }()
	s.NoError(err)
	s.Require().NotNil(cmp)

	s.NoError(cmp.RunTestfuncOrExecGoTest(s.ctx, "testName", func() {}))
}

func (s *ComposeSuite) Test_RunTestfuncOrExecGoTest() {
	cmp, cleanup, err := NewCompose(s.ctx, "docket.test-service", "full")
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
	cmp, cleanup, err := NewCompose(s.ctx, "docket.test-service", "string-command")
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
	cmp, cleanup, err := NewCompose(s.ctx, "docket.test-service", "go-not-in-path")
	defer func() { s.NoError(cleanup()) }()
	s.NoError(err)
	s.Require().NotNil(cmp)

	s.Require().NoError(cmp.Up(s.ctx))
	defer func() { s.Require().NoError(cmp.Down(s.ctx)) }()

	err = cmp.RunTestfuncOrExecGoTest(s.ctx, "testName", func() {})
	s.Error(err)
	s.Regexp("failed to exec go test", err)
}
