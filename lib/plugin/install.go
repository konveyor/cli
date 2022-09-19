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
	"runtime"

	"github.com/konveyor/cli/lib/cache"
	"github.com/konveyor/cli/lib/common"
	"github.com/konveyor/cli/lib/github"
	"github.com/konveyor/cli/lib/types"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// GetPluginMetadataFromLocalCache returns the plugin metadata from the storage directory.
func GetPluginMetadataFromLocalCache(name string) (types.PluginMetadata, error) {
	pluginDir := common.GetPluginDir(name)
	pluginYamlPath := filepath.Join(pluginDir, name+".yaml")
	plugin := types.PluginMetadata{}
	pluginYaml, err := ioutil.ReadFile(pluginYamlPath)
	if err != nil {
		return plugin, fmt.Errorf("failed to get the yaml for the plugin '%s' from the local cache. Error: %q", plugin, err)
	}
	if err := yaml.Unmarshal(pluginYaml, &plugin); err != nil {
		return plugin, fmt.Errorf("failed to parse the yaml for the plugin '%s'. Error: %q", plugin, err)
	}
	return plugin, nil
}

// GetPluginMetadataFromGithub returns the plugin metadata from the Github repo.
func GetPluginMetadataFromGithub(name string) (types.PluginMetadata, error) {
	plugin := types.PluginMetadata{}
	pluginYaml, err := github.GetPluginYamlFromGithub(name)
	if err != nil {
		return plugin, fmt.Errorf("failed to get the yaml for the plugin '%s' from the Github repo. Error: %q", plugin, err)
	}
	if err := yaml.Unmarshal(pluginYaml, &plugin); err != nil {
		return plugin, fmt.Errorf("failed to parse the yaml for the plugin '%s'. Error: %q", plugin, err)
	}
	return plugin, nil
}

// InstallPlugin installs a plugin given the the plugin metadata.
func InstallPlugin(plugin types.PluginMetadata) error {
	pluginDir := common.GetPluginDir(plugin.Metadata.Name)
	if len(plugin.Spec.Versions) == 0 {
		return fmt.Errorf("no versions are listed for the plugin")
	}
	version, platform, err := SelectProperVersionAndPlatform(plugin)
	if err != nil {
		return err
	}
	logrus.Infof("Found a version of the plugin that supports our current platform: %s", version.Version)
	outputDir := filepath.Join(pluginDir, version.Version, runtime.GOOS+"-"+runtime.GOARCH)
	if err := os.MkdirAll(outputDir, types.DEFAULT_DIRECTORY_PERMISSIONS); err != nil {
		return fmt.Errorf("failed to make the directory %s for storing the plugins. Error: %q", outputDir, err)
	}
	outputPath := filepath.Join(outputDir, plugin.Metadata.Name+".tar.gz")
	logrus.Infof("Downloading the plugin from the URL: %s", platform.Uri)
	if err := github.Download(platform.Uri, outputPath, platform.Sha256); err != nil {
		return fmt.Errorf("failed to download the plugin named '%s'. Error: %q", plugin.Metadata.Name, err)
	}
	logrus.Info("Download complete.")
	logrus.Info("Expanding the plugin archive.")
	if err := github.ExtractTarGz(outputPath); err != nil {
		return fmt.Errorf("failed to extract the plugin archive at path %s . Error: %q", outputPath, err)
	}
	logrus.Info("Done expanding the archive.")
	pluginYaml, err := yaml.Marshal(plugin)
	if err != nil {
		return fmt.Errorf("failed to marshal the plugin metadata to yaml. Error: %q", err)
	}
	pluginYamlPath := filepath.Join(pluginDir, plugin.Metadata.Name+".yaml")
	if err := ioutil.WriteFile(pluginYamlPath, pluginYaml, types.DEFAULT_FILE_PERMISSIONS); err != nil {
		return fmt.Errorf("failed to write the plugin YAML to the path %s . Error: %q", pluginYamlPath, err)
	}
	localCache, err := cache.GetLocalCache()
	if err != nil {
		return fmt.Errorf("failed to get the local cache. Error: %q", err)
	}
	localCache.Spec.Installed = append(localCache.Spec.Installed, types.InstalledPlugin{
		Name:     plugin.Metadata.Name,
		Version:  version.Version,
		Platform: runtime.GOOS + "-" + runtime.GOARCH,
		Bin:      platform.Bin,
	})
	if err := cache.SaveLocalCache(localCache); err != nil {
		return fmt.Errorf("failed to save the local cache. Error: %q", err)
	}
	return nil
}

// InstallPluginFromGithub downloads and installs a plugin from the Github repo.
func InstallPluginFromGithub(name string) error {
	plugin, err := GetPluginMetadataFromGithub(name)
	if err != nil {
		return fmt.Errorf("failed to get the plugin from the Github repo. Error: %q", err)
	}
	return InstallPlugin(plugin)
}

// UninstallPlugin uninstalls an installed plugin.
func UninstallPlugin(name string) error {
	localCache, err := cache.GetLocalCache()
	if err != nil {
		return fmt.Errorf("failed to get the local cache. Error: %q", err)
	}
	localCache.Spec.Installed = common.Filter(func(p types.InstalledPlugin) bool { return p.Name != name }, localCache.Spec.Installed)
	if err := cache.SaveLocalCache(localCache); err != nil {
		return fmt.Errorf("failed to save the local cache. Error: %q", err)
	}
	return os.RemoveAll(common.GetPluginDir(name))
}

// UninstallBrokenPlugins uninstalls any broken plugins not mentioned in the cache.
func UninstallBrokenPlugins() error {
	localCache, err := cache.GetLocalCache()
	if err != nil {
		return fmt.Errorf("failed to get the local cache. Error: %q", err)
	}
	pluginsDir := filepath.Join(common.GetStorageDir(), types.PLUGINS_DIR)
	fs, err := os.ReadDir(pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read the plugins directory %s . Error: %q", pluginsDir, err)
	}
	for _, f := range fs {
		fPath := filepath.Join(pluginsDir, f.Name())
		if !f.IsDir() {
			if err := os.RemoveAll(fPath); err != nil {
				logrus.Errorf("failed to remove the extraneous file in the plugins directory at path %s . Error: %q", fPath, err)
			}
			continue
		}
		idx := common.FindIndex(func(p types.InstalledPlugin) bool { return p.Name == f.Name() }, localCache.Spec.Installed)
		if idx == -1 {
			logrus.Infof("Found a broken plugin at %s . Removing...", fPath)
			if err := os.RemoveAll(fPath); err != nil {
				logrus.Errorf("failed to remove the broken plugin in the plugins directory at path %s . Error: %q", fPath, err)
			}
		}
	}
	if err := cache.SaveLocalCache(localCache); err != nil {
		return fmt.Errorf("failed to save the local cache. Error: %q", err)
	}
	return nil
}

// SelectProperVersionAndPlatform selects an appropriate version and platform for the plugin.
func SelectProperVersionAndPlatform(plugin types.PluginMetadata) (types.PluginVersionMetadata, types.PluginVersionForPlatform, error) {
	for _, version := range plugin.Spec.Versions {
		for _, platform := range version.Platforms {
			if (platform.Selector.MatchLabels.Os == "" || platform.Selector.MatchLabels.Os == runtime.GOOS) &&
				(platform.Selector.MatchLabels.Arch == "" || platform.Selector.MatchLabels.Arch == runtime.GOARCH) {
				return version, platform, nil
			}
		}
		logrus.Warnf("The version '%s' does not support our current platform. Trying next version.", version.Version)
	}
	return types.PluginVersionMetadata{}, types.PluginVersionForPlatform{}, fmt.Errorf("the plugin has no version that supports our current platform")
}
