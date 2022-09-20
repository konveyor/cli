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

package plugin

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/konveyor/cli/lib/cache"
	"github.com/konveyor/cli/lib/common"
	"github.com/konveyor/cli/lib/github"
	"github.com/konveyor/cli/lib/types"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func isExecutable(mode os.FileMode) bool { return mode&0111 != 0 }

func getKonveyorCommands() []string { return []string{"plugin", "version"} }

// getUniquePaths deduplicates the given paths.
func getUniquePaths(paths []string) []string {
	trimmedPaths := common.Apply(strings.TrimSpace, paths)
	filteredPaths := common.Filter(func(s string) bool { return len(s) > 0 }, trimmedPaths)
	realPaths := []string{}
	for _, filteredPath := range filteredPaths {
		realPath, err := filepath.EvalSymlinks(filteredPath)
		if err != nil {
			logrus.Debugf("failed to resolve the path %s . Error: %w", filteredPath, err)
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
func GetPluginsList(nameOnly bool) ([]string, error) {
	pluginPaths1, err := GetPluginsListFromLocalCache(nameOnly)
	if err != nil {
		return nil, err
	}
	pluginPaths2, err := GetPluginsListFromPath(nameOnly)
	if err != nil {
		return nil, err
	}
	return append(pluginPaths1, pluginPaths2...), nil
}

// GetPluginsListFromLocalCache gets all the plugins in the storage directory.
func GetPluginsListFromLocalCache(nameOnly bool) ([]string, error) {
	localCache, err := cache.GetLocalCache()
	if err != nil {
		return nil, fmt.Errorf("failed to get the local cache. Error: %w", err)
	}
	if len(localCache.Spec.Installed) == 0 {
		return nil, nil
	}
	if nameOnly {
		pluginNames := common.Apply(func(p types.InstalledPlugin) string { return p.Name }, localCache.Spec.Installed)
		return pluginNames, nil
	}
	pluginPaths := []string{}
	for _, installed := range localCache.Spec.Installed {
		pluginPaths = append(pluginPaths, GetPluginBinPath(installed))
	}
	return pluginPaths, nil
}

// GetPluginBinPath returns the path to the plugin's entrypoint.
func GetPluginBinPath(installed types.InstalledPlugin) string {
	return filepath.Join(common.GetPluginDir(installed.Name), installed.Version, installed.Platform, installed.Bin)
}

// GetPluginsListFromPath get all the plugins with a valid prefix that are on the PATH.
func GetPluginsListFromPath(nameOnly bool) ([]string, error) {
	envPath := os.Getenv("PATH")
	logrus.Debug("envPath", envPath)
	paths := filepath.SplitList(envPath)
	logrus.Debug("paths", paths)
	if len(paths) == 0 {
		return nil, fmt.Errorf("the list of directories is empty")
	}
	pluginPaths := []string{}
	konveyorCmds := getKonveyorCommands()
	seen := map[string]bool{}
	for _, dir := range getUniquePaths(paths) {
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			logrus.Errorf("failed to read the directory %s . Error: %w . Skipping...\n", dir, err)
			continue
		}

		for _, f := range files {
			if f.IsDir() {
				continue
			}
			pluginName := f.Name()
			if !strings.HasPrefix(pluginName, types.VALID_PLUGIN_FILENAME_PREFIX) {
				continue
			}
			if _, ok := seen[pluginName]; ok {
				logrus.Warnf("The plugin named '%s' was found in multiple directories. Found again in %s", pluginName, dir)
				continue
			}
			seen[pluginName] = true
			if !isExecutable(f.Mode()) {
				logrus.Warnf("A file named '%s' was found in the directory %s but it is not executable", pluginName, dir)
			} else if common.Contains(pluginName, konveyorCmds) {
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

// GetPluginMetadataFromLocalCache returns the plugin metadata from the storage directory.
func GetPluginMetadataFromLocalCache(name string) (types.PluginMetadata, error) {
	pluginDir := common.GetPluginDir(name)
	pluginYamlPath := filepath.Join(pluginDir, name+".yaml")
	plugin := types.PluginMetadata{}
	pluginYaml, err := ioutil.ReadFile(pluginYamlPath)
	if err != nil {
		return plugin, fmt.Errorf("failed to get the yaml for the plugin '%s' from the local cache. Error: %w", plugin, err)
	}
	if err := yaml.Unmarshal(pluginYaml, &plugin); err != nil {
		return plugin, fmt.Errorf("failed to parse the yaml for the plugin '%s'. Error: %w", plugin, err)
	}
	return plugin, nil
}

// GetPluginMetadataFromGithub returns the plugin metadata from the Github repo.
func GetPluginMetadataFromGithub(name string) (types.PluginMetadata, error) {
	plugin := types.PluginMetadata{}
	pluginYaml, err := github.GetPluginYamlFromGithub(name)
	if err != nil {
		return plugin, fmt.Errorf("failed to get the yaml for the plugin '%s' from the Github repo. Error: %w", plugin, err)
	}
	if err := yaml.Unmarshal(pluginYaml, &plugin); err != nil {
		return plugin, fmt.Errorf("failed to parse the yaml for the plugin '%s'. Error: %w", plugin, err)
	}
	return plugin, nil
}

func GetPluginFromLocalCache(name string) (types.InstalledPlugin, error) {
	plugin := types.InstalledPlugin{}
	localCache, err := cache.GetLocalCache()
	if err != nil {
		return plugin, fmt.Errorf("failed to get the local cache. Error: %w", err)
	}
	idx := common.FindIndex(func(p types.InstalledPlugin) bool { return p.Name == name }, localCache.Spec.Installed)
	if idx == -1 {
		return plugin, types.ErrPluginNotInstalled
	}
	return localCache.Spec.Installed[idx], nil
}
