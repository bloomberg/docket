package docket

/*
This file lists dependencies of the tests in testdata to make it easier to use
`go get` or `dep` to pull down those dependencies.
*/

import (
	_ "gopkg.in/redis.v5"
)
