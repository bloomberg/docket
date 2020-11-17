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

package compose

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"sort"
)

func findAndSortDocketFiles(prefix, mode string) ([]string, error) {
	infos, err := ioutil.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("failed to read current dir: %w", err)
	}

	files := make([]string, len(infos))
	for i, info := range infos {
		if info.IsDir() {
			continue
		}
		files[i] = info.Name()
	}

	fs := newFileSorter(prefix, mode)
	fs.AddFiles(files)

	return fs.Results(), nil
}

// Ordering:
//   prefix.yaml
//   prefix.mode.yaml
//   prefix.mode.extra.yaml
type fileSorter struct {
	prefixOnly        []string
	prefixAndMode     []string
	prefixModeAndMore []string

	prefixOnlyPattern       *regexp.Regexp
	modeAndMaybeMorePattern *regexp.Regexp
}

func newFileSorter(prefix, mode string) *fileSorter {
	if prefix == "" {
		panic("prefix must not be empty!")
	}
	if mode == "" {
		panic("mode must not be empty!")
	}

	return &fileSorter{
		prefixOnly:        nil,
		prefixAndMode:     nil,
		prefixModeAndMore: nil,

		prefixOnlyPattern: regexp.MustCompile(fmt.Sprintf(`^%s\.ya?ml$`,
			regexp.QuoteMeta(prefix))),

		modeAndMaybeMorePattern: regexp.MustCompile(fmt.Sprintf(`^%s\.%s\.(.+\.)?ya?ml$`,
			regexp.QuoteMeta(prefix), regexp.QuoteMeta(mode))),
	}
}

func (fs *fileSorter) AddFiles(files []string) {
	for _, f := range files {
		fs.AddFile(f)
	}
}

func (fs *fileSorter) AddFile(f string) {
	if fs.prefixOnlyPattern.MatchString(f) {
		fs.prefixOnly = append(fs.prefixOnly, f)

		return
	}

	mm := fs.modeAndMaybeMorePattern.FindStringSubmatch(f)
	if mm == nil {
		return
	}

	switch mm[1] {
	case "":
		fs.prefixAndMode = append(fs.prefixAndMode, f)
	default:
		fs.prefixModeAndMore = append(fs.prefixModeAndMore, f)
	}
}

func (fs *fileSorter) Results() []string {
	sort.Strings(fs.prefixOnly)
	sort.Strings(fs.prefixAndMode)
	sort.Strings(fs.prefixModeAndMore)

	results := make([]string, 0, len(fs.prefixOnly)+len(fs.prefixAndMode)+len(fs.prefixModeAndMore))
	results = append(results, fs.prefixOnly...)
	results = append(results, fs.prefixAndMode...)
	results = append(results, fs.prefixModeAndMore...)

	return results
}
