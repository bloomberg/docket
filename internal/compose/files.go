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
		return nil, err
	}

	files := make([]string, len(infos))
	for i, info := range infos {
		files[i] = info.Name()
	}

	return filterAndSortFilenames(prefix, mode, files), nil
}
