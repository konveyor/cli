/*
 *  Copyright IBM Corporation 2022
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package lib

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/konveyor/cli/types"
	"github.com/sirupsen/logrus"
)

func isExecutable(mode os.FileMode) bool {
	return mode&0111 != 0
}

func getKonveyorCommands() []string { return []string{"plugin", "version"} }

// getUniquePaths deduplicates the given paths.
func getUniquePaths(paths []string) []string {
	trimmedPaths := apply(strings.TrimSpace, paths)
	filteredPaths := filter(func(s string) bool { return len(s) > 0 }, trimmedPaths)
	realPaths := []string{}
	for _, filteredPath := range filteredPaths {
		realPath, err := filepath.EvalSymlinks(filteredPath)
		if err != nil {
			logrus.Errorf("failed to resolve the path %s . Error: %q", filteredPath, err)
			continue
		}
		realPaths = append(realPaths, realPath)
	}
	seen := map[string]bool{}
	uniquePaths := []string{}
	for _, realPath := range realPaths {
		if _, ok := seen[realPath]; ok {
			continue
		}
		seen[realPath] = true
		uniquePaths = append(uniquePaths, realPath)
	}
	return uniquePaths
}

// GetPluginsList looks for plugins in the given paths.
func GetPluginsList(paths []string, nameOnly bool) ([]string, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("the list of directories is empty")
	}

	pluginPaths := []string{}
	konveyorCmds := getKonveyorCommands()

	seen := map[string]bool{}
	for _, dir := range getUniquePaths(paths) {
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			logrus.Errorf("failed to read the directory %s . Error: %q . Skipping...\n", dir, err)
			continue
		}

		for _, f := range files {
			if f.IsDir() {
				continue
			}
			pluginName := f.Name()
			if !strings.HasPrefix(pluginName, types.ValidPluginFilenamePrefix) {
				continue
			}
			if _, ok := seen[pluginName]; ok {
				logrus.Warnf("The plugin named '%s' was found in multiple directories. Found again in %s", pluginName, dir)
				continue
			}
			seen[pluginName] = true
			if !isExecutable(f.Mode()) {
				logrus.Warnf("A file named '%s' was found in the directory %s but it is not executable", pluginName, dir)
			} else if contains(pluginName, konveyorCmds) {
				logrus.Warnf("The plugin '%s' has the same name as a built-in command of konveyor", pluginName)
			}
			if nameOnly {
				pluginPaths = append(pluginPaths, pluginName)
			} else {
				pluginPaths = append(pluginPaths, filepath.Join(dir, pluginName))
			}
		}
	}
	return pluginPaths, nil
}

func GetPluginsFromPath(nameOnly bool) ([]string, error) {
	envPath := os.Getenv("PATH")
	logrus.Debug("envPath", envPath)
	paths := filepath.SplitList(envPath)
	logrus.Debug("paths", paths)
	return GetPluginsList(paths, nameOnly)
}
