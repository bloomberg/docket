// Copyright 2020 Bloomberg Finance L.P.
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

package main_test

import (
	"context"
	"os"
	"testing"

	"github.com/bloomberg/docket"
	"github.com/bloomberg/go-testgroup"
)

//----------------------------------------------------------

// This test group is intended to run entirely inside docker.
func TestDocketRunAtTopLevel(t *testing.T) {
	ctx := context.Background()

	var grp EntireGroup
	docket.Run(ctx, &grp.dctx, t, func() { testgroup.RunSerially(t, &grp) })
}

type EntireGroup struct {
	dctx docket.Context
}

// The following three redundant tests exist so that the docket test suite can test executing
// particular combinations of subtests of the suite.

func (grp *EntireGroup) A(t *testgroup.T) {
	t.Logf("docket mode = %q", grp.dctx.Mode())
	t.Equal("something", os.Getenv("SECRET_VALUE"))
}

func (grp *EntireGroup) B(t *testgroup.T) {
	t.Equal("something", os.Getenv("SECRET_VALUE"))
}

func (grp *EntireGroup) C(t *testgroup.T) {
	t.Equal("something", os.Getenv("SECRET_VALUE"))
}

//----------------------------------------------------------

// Only one of the subtests of this test group should run inside docker.
func TestDocketRunForSingleSubtest(t *testing.T) {
	var grp SubtestGroup
	testgroup.RunSerially(t, &grp)
}

type SubtestGroup struct{}

func (grp *SubtestGroup) OutsideDocker(t *testgroup.T) {
	t.Equal("", os.Getenv("SECRET_VALUE"))
}

func (grp *SubtestGroup) InsideDocker(t *testgroup.T) {
	ctx := context.Background()

	docket.Run(ctx, nil, t.T, func() {
		t.Equal("something", os.Getenv("SECRET_VALUE"))
	})
}
