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
		return fmt.Errorf("failed to make the directory %s for storing the plugins. Error: %w", outputDir, err)
	}
	outputPath := filepath.Join(outputDir, plugin.Metadata.Name+".tar.gz")
	logrus.Infof("Downloading the plugin from the URL: %s", platform.Uri)
	if err := github.Download(platform.Uri, outputPath, platform.Sha256); err != nil {
		return fmt.Errorf("failed to download the plugin named '%s'. Error: %w", plugin.Metadata.Name, err)
	}
	logrus.Info("Download complete.")
	logrus.Info("Expanding the plugin archive.")
	if err := github.ExtractTarGz(outputPath); err != nil {
		return fmt.Errorf("failed to extract the plugin archive at path %s . Error: %w", outputPath, err)
	}
	logrus.Info("Done expanding the archive.")
	pluginYaml, err := yaml.Marshal(plugin)
	if err != nil {
		return fmt.Errorf("failed to marshal the plugin metadata to yaml. Error: %w", err)
	}
	pluginYamlPath := filepath.Join(pluginDir, plugin.Metadata.Name+".yaml")
	if err := ioutil.WriteFile(pluginYamlPath, pluginYaml, types.DEFAULT_FILE_PERMISSIONS); err != nil {
		return fmt.Errorf("failed to write the plugin YAML to the path %s . Error: %w", pluginYamlPath, err)
	}
	localCache, err := cache.GetLocalCache()
	if err != nil {
		return fmt.Errorf("failed to get the local cache. Error: %w", err)
	}
	localCache.Spec.Installed = append(localCache.Spec.Installed, types.InstalledPlugin{
		Name:     plugin.Metadata.Name,
		Version:  version.Version,
		Platform: common.GetPlatformAsSingleString(runtime.GOOS, runtime.GOARCH),
		Bin:      platform.Bin,
	})
	if err := cache.SaveLocalCache(localCache); err != nil {
		return fmt.Errorf("failed to save the local cache. Error: %w", err)
	}
	return nil
}

// InstallPluginFromGithub downloads and installs a plugin from the Github repo.
func InstallPluginFromGithub(name string) error {
	if _, err := GetPluginFromLocalCache(name); err == nil {
		return types.ErrPluginAlreadyInstalled
	}
	plugin, err := GetPluginMetadataFromGithub(name)
	if err != nil {
		if types.IsNotFoundError(err) {
			return fmt.Errorf("did not find a plugin named '%s' on Github", name)
		}
		return fmt.Errorf("failed to get the plugin from the Github repo. Error: %w", err)
	}
	return InstallPlugin(plugin)
}

// UninstallPlugin uninstalls an installed plugin.
func UninstallPlugin(name string) error {
	if _, err := GetPluginFromLocalCache(name); err != nil {
		return err
	}
	localCache, err := cache.GetLocalCache()
	if err != nil {
		return fmt.Errorf("failed to get the local cache. Error: %w", err)
	}
	localCache.Spec.Installed = common.Filter(func(p types.InstalledPlugin) bool { return p.Name != name }, localCache.Spec.Installed)
	if err := cache.SaveLocalCache(localCache); err != nil {
		return fmt.Errorf("failed to save the local cache. Error: %w", err)
	}
	return os.RemoveAll(common.GetPluginDir(name))
}

// UninstallBrokenPlugins uninstalls any broken plugins not mentioned in the cache.
func UninstallBrokenPlugins() error {
	localCache, err := cache.GetLocalCache()
	if err != nil {
		return fmt.Errorf("failed to get the local cache. Error: %w", err)
	}
	pluginsDir := filepath.Join(common.GetStorageDir(), types.PLUGINS_DIR)
	fs, err := os.ReadDir(pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read the plugins directory %s . Error: %w", pluginsDir, err)
	}
	for _, f := range fs {
		fPath := filepath.Join(pluginsDir, f.Name())
		if !f.IsDir() {
			if err := os.RemoveAll(fPath); err != nil {
				logrus.Errorf("failed to remove the extraneous file in the plugins directory at path %s . Error: %w", fPath, err)
			}
			continue
		}
		idx := common.FindIndex(func(p types.InstalledPlugin) bool { return p.Name == f.Name() }, localCache.Spec.Installed)
		if idx == -1 {
			logrus.Infof("Found a broken plugin at %s . Removing...", fPath)
			if err := os.RemoveAll(fPath); err != nil {
				logrus.Errorf("failed to remove the broken plugin in the plugins directory at path %s . Error: %w", fPath, err)
			}
		}
	}
	if err := cache.SaveLocalCache(localCache); err != nil {
		return fmt.Errorf("failed to save the local cache. Error: %w", err)
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
