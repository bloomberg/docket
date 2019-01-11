package compose

import (
	"context"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
)

func runGoEnvGOPATH(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "go", "env", "GOPATH") // #nosec
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return filepath.SplitList(strings.TrimSpace(string(out))), nil
}

type goList struct {
	Dir        string
	ImportPath string
	Module     *struct {
		Path string
		Dir  string
	}
}

func runGoList(ctx context.Context) (goList, error) {
	cmd := exec.CommandContext(ctx, "go", "list", "-json") // #nosec
	out, err := cmd.Output()
	if err != nil {
		return goList{}, err
	}

	var gl goList
	if err := json.Unmarshal(out, &gl); err != nil {
		return goList{}, err
	}

	return gl, nil
}
