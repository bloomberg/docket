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
func TestDocketRunAtSuiteLevel(t *testing.T) {
	ctx := context.Background()

	var s EntireSuite

	docket.Run(ctx, &s.Context, t, func() { suite.Run(t, &s) })
}

type EntireSuite struct {
	suite.Suite
	docket.Context
}

// The following three redundant tests exist so that the docket test suite can test executing
// particular combinations of subtests of the suite.

func (s *EntireSuite) TestA() {
	s.T().Logf("docket mode = %q", s.Mode())
	s.Equal("something", os.Getenv("SECRET_VALUE"))
}

func (s *EntireSuite) TestB() {
	s.Equal("something", os.Getenv("SECRET_VALUE"))
}

func (s *EntireSuite) TestC() {
	s.Equal("something", os.Getenv("SECRET_VALUE"))
}

//----------------------------------------------------------

// Only one of the subtests of this suite should run inside docker.
func TestDocketRunForSingleSubtest(t *testing.T) {
	suite.Run(t, new(SubtestSuite))
}

type SubtestSuite struct {
	suite.Suite
}

func (s *SubtestSuite) TestOutsideDocker() {
	s.Equal("", os.Getenv("SECRET_VALUE"))
}

func (s *SubtestSuite) TestInsideDocker() {
	ctx := context.Background()

	docket.Run(ctx, nil, s.T(), func() {
		s.Equal("something", os.Getenv("SECRET_VALUE"))
	})
}
