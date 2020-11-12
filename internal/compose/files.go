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

// Ordering:
//   prefix.yaml
//   prefix.mode.yaml
//   prefix.mode.extra.yaml
func filterAndSortFilenames(prefix, mode string, files []string) []string {
	if prefix == "" || mode == "" {
		panic("prefix and mode must not be blank!")
	}

	prefixOnly := []string{}
	prefixAndMode := []string{}
	prefixModeAndMore := []string{}

	prefixOnlyPattern := regexp.MustCompile(
		fmt.Sprintf(`^%s\.ya?ml$`, regexp.QuoteMeta(prefix)))

	modeAndMaybeMorePattern := regexp.MustCompile(
		fmt.Sprintf(`^%s\.%s\.(.+\.)?ya?ml$`, regexp.QuoteMeta(prefix), regexp.QuoteMeta(mode)))

	for _, f := range files {
		if prefixOnlyPattern.MatchString(f) {
			prefixOnly = append(prefixOnly, f)

			continue
		}
		if mm := modeAndMaybeMorePattern.FindStringSubmatch(f); mm != nil {
			if mm[1] == "" {
				prefixAndMode = append(prefixAndMode, f)
			} else {
				prefixModeAndMore = append(prefixModeAndMore, f)
			}

			continue
		}
	}

	sort.Strings(prefixOnly)
	sort.Strings(prefixAndMode)
	sort.Strings(prefixModeAndMore)

	results := append(prefixOnly, prefixAndMode...)
	results = append(results, prefixModeAndMore...)

	return results
}

func findFiles(prefix, mode string) ([]string, error) {
	infos, err := ioutil.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("failed to read current dir: %w", err)
	}

	files := make([]string, len(infos))
	for i, info := range infos {
		files[i] = info.Name()
	}

	return filterAndSortFilenames(prefix, mode, files), nil
}
