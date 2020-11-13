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
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"

	"gopkg.in/yaml.v2"
)

// Compose represents a call to docker-compose.
type Compose struct {
	baseArgs []string

	testSvc string
}

// NewCompose returns a new Compose and cleanup function given a context, prefix, and mode.
func NewCompose(ctx context.Context, prefix, mode string) (
	cmp *Compose, cleanup func() error, err error,
) {
	cmp = &Compose{}
	cleanup = func() error { return nil }

	cmp.baseArgs, err = makeDocketFileArgs(prefix, mode)
	if err != nil {
		return nil, cleanup, err
	}

	cfg, err := cmp.getAndParseConfig(ctx)
	if err != nil {
		return nil, cleanup, err
	}

	goList, err := runGoList(ctx)
	if err != nil {
		return nil, cleanup, err
	}

	goPath, err := runGoEnvGOPATH(ctx)
	if err != nil {
		return nil, cleanup, err
	}

	mountsArgs, mountsCleanup, err := doSourceMounts(cfg, goList, goPath)
	if err != nil {
		return nil, cleanup, err
	}

	cmp.baseArgs = append(cmp.baseArgs, mountsArgs...)
	cleanup = chainCleanups(cleanup, mountsCleanup)

	cmp.testSvc, err = findSingleTestService(cfg)
	if err != nil {
		return nil, cleanup, err
	}

	return cmp, cleanup, nil
}

// Command makes an *exec.Cmd that calls `docker-compose` with the right arguments and environment.
//
// Command is intended to be a helper function. It is exported mainly so `dkt` can use it.
func (c Compose) Command(ctx context.Context, arg ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "docker-compose") // #nosec
	cmd.Args = append(cmd.Args, c.baseArgs...)
	cmd.Args = append(cmd.Args, arg...)
	cmd.Env = os.Environ()

	return cmd
}

// Down calls `docker-compose down`.
func (c Compose) Down(ctx context.Context) error {
	cmd := c.Command(ctx, "down")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	tracef("down %v\n", cmd.Args)
	defer tracef("down finished\n")

	return cmd.Run()
}

// GetConfig calls `docker-compose config` and returns the aggregated Compose file.
func (c Compose) GetConfig(ctx context.Context) ([]byte, error) {
	cmd := c.Command(ctx, "config")

	tracef("config %v\n", cmd.Args)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error getting config: err=%w out=%s", err, out)
	}

	return out, nil
}

var errPortNotFound = fmt.Errorf("port not found")

// GetPort runs `docker-compose port` and returns the public port for a service's port binding.
func (c Compose) GetPort(ctx context.Context, service string, port int) (int, error) {
	cmd := c.Command(ctx, "port", service, strconv.Itoa(port))

	tracef("port %v\n", cmd.Args)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("port error: err=%w out=%q", err, out)
	}

	re := regexp.MustCompile(":[[:digit:]]+$")
	match := re.Find(bytes.TrimSpace(out))
	if len(match) == 0 {
		return 0, fmt.Errorf("%w: %q", errPortNotFound, out)
	}

	return strconv.Atoi(string(match[1:])) // drop the leading colon
}

// Pull calls `docker-compose pull`.
func (c Compose) Pull(ctx context.Context, args []string) error {
	cmd := c.Command(ctx, "pull")

	cmd.Args = append(cmd.Args, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	tracef("pull %v\n", cmd.Args)
	defer tracef("pull finished\n")

	return cmd.Run()
}

// RunTestfuncOrExecGoTest either calls testFunc directly or runs `docker-compose exec` to re-run
// `go test` inside the appropriate service (container).
func (c Compose) RunTestfuncOrExecGoTest(
	ctx context.Context, testName string, testFunc func(),
) error {
	if c.testSvc == "" {
		testFunc()

		return nil
	}

	var originalTestRunArg string
	if f := flag.Lookup("test.run"); f != nil {
		originalTestRunArg = f.Value.String()
	}

	args := []string{
		"exec",
		"-T", // disable pseudo-tty allocation
		c.testSvc,
		"go", "test",
		"-run", makeRunArgForTest(testName, originalTestRunArg),
	}

	if testing.Verbose() {
		args = append(args, "-v")
	}

	cmd := c.Command(ctx, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	tracef("exec %v\n", cmd.Args)
	defer tracef("exec finished\n")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to exec go test: %w", err)
	}

	return nil
}

// Up calls `docker-compose up`.
func (c Compose) Up(ctx context.Context, service ...string) error {
	cmd := c.Command(ctx, "up", "-d")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	tracef("up %v\n", cmd.Args)
	defer tracef("up finished\n")

	return cmd.Run()
}

//------------------------------------------------------------------------------

type cmpVolume struct {
	Type   string `yaml:"type"`
	Source string `yaml:"source"`
	Target string `yaml:"target"`
}

type cmpService struct {
	Command     interface{}       `yaml:"command,omitempty"` // []string or just a string
	Environment map[string]string `yaml:"environment,omitempty"`
	Image       string            `yaml:"image,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Volumes     []cmpVolume       `yaml:"volumes,omitempty"`
	WorkingDir  string            `yaml:"working_dir,omitempty"`
}

type cmpConfig struct {
	Version  string                `yaml:"version,omitempty"`
	Services map[string]cmpService `yaml:"services,omitempty"`
	Networks map[string]struct{}   `yaml:"networks,omitempty"`
}

func (c Compose) getAndParseConfig(ctx context.Context) (cmpConfig, error) {
	cfgBytes, err := c.GetConfig(ctx)
	if err != nil {
		return cmpConfig{}, err
	}

	var cfg cmpConfig
	if err := yaml.Unmarshal(cfgBytes, &cfg); err != nil {
		return cmpConfig{}, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	return cfg, nil
}

//------------------------------------------------------------------------------

var errNoMatchingDocketFiles = fmt.Errorf("no matching docket files found")

func makeDocketFileArgs(prefix, mode string) ([]string, error) {
	files, err := findFiles(prefix, mode)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("%w: prefix=%s, mode=%s", errNoMatchingDocketFiles, prefix, mode)
	}

	const sizeOfArgPair = 2
	args := make([]string, 0, len(files)*sizeOfArgPair)
	for _, f := range files {
		args = append(args, "--file", f)
	}

	return args, nil
}

var errMultipleTestServices = fmt.Errorf("multiple test services found")

func findSingleTestService(cfg cmpConfig) (string, error) {
	testSvc := ""
	for name, svc := range cfg.Services {
		if runGoTest, _, err := parseDocketLabel(svc); err != nil {
			return "", err
		} else if runGoTest {
			if testSvc != "" {
				return "",
					fmt.Errorf("%w (at least %q and %q)", errMultipleTestServices, testSvc, name)
			}
			testSvc = name
		}
	}

	return testSvc, nil
}

var errUnrecognizedDocketLabel = fmt.Errorf("unrecognized docket label")

func parseDocketLabel(svc cmpService) (runGoTest bool, mountGoSources bool, err error) {
	const docketLabelKey = "com.bloomberg.docket"

	labelData := svc.Labels[docketLabelKey]
	switch labelData {
	case "":
		return false, false, nil
	case "run go test":
		return true, true, nil
	case "mount go sources":
		return false, true, nil
	}

	return false, false,
		fmt.Errorf("%w: %q : %q", errUnrecognizedDocketLabel, docketLabelKey, labelData)
}

func doSourceMounts(cfg cmpConfig, goList goList, goPath []string) (args []string, cleanup func() error, err error) {
	noop := func() error { return nil }

	mountsCfg, err := newMountsCfg(cfg, goList, goPath)
	if err != nil {
		return nil, noop, err
	}
	if mountsCfg == nil {
		return nil, noop, nil
	}

	mountsFile, err := ioutil.TempFile(".", "docket-source-mounts.*.yaml")
	if err != nil {
		return nil, noop, fmt.Errorf("failed to create source mounts yaml: %w", err)
	}

	cleanup = func() error {
		if os.Getenv("DOCKET_KEEP_MOUNTS_FILE") != "" {
			tracef("Leaving %s alone\n", mountsFile.Name())

			return nil
		}

		return os.Remove(mountsFile.Name())
	}

	defer func() {
		if closeErr := mountsFile.Close(); closeErr != nil {
			args = nil
			err = closeErr
		}
	}()

	enc := yaml.NewEncoder(mountsFile)

	defer func() {
		if closeErr := enc.Close(); closeErr != nil {
			args = nil
			err = closeErr
		}
	}()

	if err := enc.Encode(mountsCfg); err != nil {
		return nil, noop, fmt.Errorf("failed to encode yaml: %w", err)
	}

	return []string{"--file", mountsFile.Name()}, cleanup, nil
}

var errMultipleGOPATHs = fmt.Errorf("docket doesn't support multipart GOPATHs")

// newMountsCfg makes a cmpConfig to bind mount Go sources.
func newMountsCfg(originalCfg cmpConfig, goList goList, goPath []string) (*cmpConfig, error) {
	mountsCfg := cmpConfig{
		Version:  "3.2",
		Services: map[string]cmpService{},
		Networks: nil,
	}

	needMounts := false

	if len(goPath) != 1 {
		return nil, errMultipleGOPATHs
	}

	var mountsSvc cmpService
	if goList.Module == nil {
		const goPathTarget = "/go"

		pkgName, err := findPackageNameFromDirAndGOPATH(goList.Dir, goPath)
		if err != nil {
			return nil, err
		}

		mountsSvc = cmpService{
			Command:     nil,
			Environment: nil,
			Image:       "",
			Labels:      nil,
			Volumes: []cmpVolume{
				{
					Type:   "bind",
					Source: goPath[0],
					Target: goPathTarget,
				},
			},
			WorkingDir: fmt.Sprintf("%s/src/%s", goPathTarget, pkgName),
		}
	} else {
		const goPathTarget = "/go"
		const goModuleDirTarget = "/go-module-dir"

		pathInsideModule, err := filepath.Rel(goList.Module.Dir, goList.Dir)
		if err != nil {
			return nil, fmt.Errorf("failed filepath.Rel: %w", err)
		}

		mountsSvc = cmpService{
			Command:     nil,
			Environment: nil,
			Image:       "",
			Labels:      nil,
			Volumes: []cmpVolume{
				{
					Type:   "bind",
					Source: filepath.Join(goPath[0], "pkg", "mod"),
					Target: fmt.Sprintf("%s/pkg/mod", goPathTarget),
				},
				{
					Type:   "bind",
					Source: goList.Module.Dir,
					Target: goModuleDirTarget,
				},
			},
			WorkingDir: fmt.Sprintf("%s/%s", goModuleDirTarget, filepath.ToSlash(pathInsideModule)),
		}
	}

	for name, svc := range originalCfg.Services {
		if _, mountGoSources, err := parseDocketLabel(svc); err != nil {
			return nil, err
		} else if mountGoSources {
			needMounts = true
			mountsCfg.Services[name] = mountsSvc
		}
	}

	if !needMounts {
		return nil, nil
	}

	return &mountsCfg, nil
}

func chainCleanups(a, b func() error) func() error {
	return func() error {
		if err := a(); err != nil {
			return err
		}
		if err := b(); err != nil {
			return err
		}

		return nil
	}
}
