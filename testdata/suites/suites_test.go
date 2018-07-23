package main

import (
	"context"
	"os"
	"testing"

	"github.com/bloomberg/docket"
	"github.com/stretchr/testify/suite"
)

//----------------------------------------------------------

// This suite is intended to run entirely inside docker.
func TestFullSuite(t *testing.T) {
	ctx := context.Background()

	cfgs := docket.ConfigMap{
		"full": {
			ComposeFiles: []string{
				"docker-compose.yml",
			},
			GoTestExec: &docket.GoTestExec{
				Service: "tester",
			},
		},
	}

	var s FullSuite

	docket.Run(ctx, cfgs, &s.Context, t, func() { suite.Run(t, &s) })
}

type FullSuite struct {
	suite.Suite
	docket.Context
}

// The following three redundant tests exist so that the docket test suite can test executing
// particular combinations of subtests of the suite.

func fullSuiteInsideDocker(s *FullSuite) {
	s.Equal("", s.ConfigName())
	s.Equal("something", os.Getenv("DOCKET_SUITES_TEST_SECRET_VALUE"))
}

func (s *FullSuite) TestA() {
	fullSuiteInsideDocker(s)
}

func (s *FullSuite) TestB() {
	fullSuiteInsideDocker(s)
}

func (s *FullSuite) TestC() {
	fullSuiteInsideDocker(s)
}

//----------------------------------------------------------

// Only one of the subtests of this suite should run inside docker.
func TestSubtestSuite(t *testing.T) {
	suite.Run(t, new(SubtestSuite))
}

type SubtestSuite struct {
	suite.Suite
}

func (s *SubtestSuite) TestOutsideDocker() {
	s.Equal("", os.Getenv("DOCKET_SUITES_TEST_SECRET_VALUE"))
}

func (s *SubtestSuite) TestInsideDocker() {
	ctx := context.Background()

	cfgs := docket.ConfigMap{
		"full": {
			ComposeFiles: []string{
				"docker-compose.yml",
			},
			GoTestExec: &docket.GoTestExec{
				Service: "tester",
			},
		},
	}

	var docketCtx docket.Context

	docket.Run(ctx, cfgs, &docketCtx, s.T(), func() {
		s.Equal("", docketCtx.ConfigName())
		s.Equal("something", os.Getenv("DOCKET_SUITES_TEST_SECRET_VALUE"))
	})
}
