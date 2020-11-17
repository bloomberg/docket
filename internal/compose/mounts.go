package compose

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func doSourceMounts(cfg cmpConfig, goList goList, goPath []string) (
	args []string, cleanup func() error, err error,
) {
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

type mountsFunc func(goList, []string) ([]cmpVolume, string, error)

// newMountsCfg makes a cmpConfig to bind mount Go sources for the services that need them.
func newMountsCfg(originalCfg cmpConfig, goList goList, goPath []string) (*cmpConfig, error) {
	mountsCfg := cmpConfig{
		Version:  "3.2",
		Services: map[string]cmpService{},
		Networks: nil,
	}

	if len(goPath) != 1 {
		return nil, errMultipleGOPATHs
	}

	var mountsFunc mountsFunc
	if goList.Module == nil {
		mountsFunc = mountsForModuleMode
	} else {
		mountsFunc = mountsForGOPATHMode
	}
	volumes, workingDir, err := mountsFunc(goList, goPath)
	if err != nil {
		return nil, err
	}

	for name, svc := range originalCfg.Services {
		if _, mountGoSources, err := parseDocketLabel(svc); err != nil {
			return nil, err
		} else if mountGoSources {
			mountsCfg.Services[name] = cmpService{
				Command:     nil,
				Environment: nil,
				Image:       "",
				Labels:      nil,
				Volumes:     volumes,
				WorkingDir:  workingDir,
			}
		}
	}

	if len(mountsCfg.Services) == 0 {
		return nil, nil
	}

	return &mountsCfg, nil
}

func mountsForModuleMode(goList goList, goPath []string) ([]cmpVolume, string, error) {
	const goPathTarget = "/go"

	pkgName, err := findPackageNameFromDirAndGOPATH(goList.Dir, goPath)
	if err != nil {
		return nil, "", err
	}

	volumes := []cmpVolume{
		{
			Type:   "bind",
			Source: goPath[0],
			Target: goPathTarget,
		},
	}

	workingDir := fmt.Sprintf("%s/src/%s", goPathTarget, pkgName)

	return volumes, workingDir, nil
}

func mountsForGOPATHMode(goList goList, goPath []string) ([]cmpVolume, string, error) {
	const goPathTarget = "/go"
	const goModuleDirTarget = "/go-module-dir"

	pathInsideModule, err := filepath.Rel(goList.Module.Dir, goList.Dir)
	if err != nil {
		return nil, "", fmt.Errorf("failed filepath.Rel: %w", err)
	}

	volumes := []cmpVolume{
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
	}

	workingDir := fmt.Sprintf("%s/%s", goModuleDirTarget, filepath.ToSlash(pathInsideModule))

	return volumes, workingDir, nil
}
