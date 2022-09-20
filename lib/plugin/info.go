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

	"github.com/konveyor/cli/lib/cache"
	"github.com/konveyor/cli/lib/common"
	"github.com/konveyor/cli/lib/types"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func getAllSupportedPlatforms(pluginMeta types.PluginMetadata) []string {
	uniquePlatforms := map[string]bool{}
	for _, version := range pluginMeta.Spec.Versions {
		for _, platform := range version.Platforms {
			os := platform.Selector.MatchLabels.Os
			arch := platform.Selector.MatchLabels.Arch
			if os == "" && arch == "" {
				// if both OS and Arch are not mentioned, then simply leave out the supported platforms
				continue
			}
			if os == "" {
				os = "*"
			}
			if arch == "" {
				arch = "*"
			}
			uniquePlatforms[common.GetPlatformAsSingleString(os, arch)] = true
		}
	}
	return common.Keys(uniquePlatforms)
}

func GetPluginInfo(name string) (string, error) {
	// get plugin metadata
	pluginMeta := types.PluginMetadata{}
	pluginMeta, err := GetPluginMetadataFromLocalCache(name)
	if err != nil {
		logrus.Debugf("failed to get the plugin metadata from the local cache. Error: %w", err)
		pluginMeta, err = GetPluginMetadataFromGithub(name)
		if err != nil {
			if types.IsNotFoundError(err) {
				return "", fmt.Errorf("did not find a plugin named '%s' in the local cache or on Github", name)
			}
			return "", fmt.Errorf("failed to find any info for a plugin named '%s' on Github. Error: %w", name, err)
		}
	}
	// check if the plugin is installed
	localCache, err := cache.GetLocalCache()
	if err != nil {
		return "", fmt.Errorf("failed to get the local cache. Error: %w", err)
	}
	installed := false
	version := ""
	for _, inst := range localCache.Spec.Installed {
		if inst.Name == name {
			installed = true
			version = inst.Version
			break
		}
	}
	// format the plugin metadata for display
	pluginInfo := types.PluginInfo{
		Name:               pluginMeta.Metadata.Name,
		Description:        pluginMeta.Spec.Description,
		HomePage:           pluginMeta.Spec.HomePage,
		Documentation:      pluginMeta.Spec.Docs,
		Tutorials:          pluginMeta.Spec.Tutorials,
		Installed:          installed,
		InstalledVersion:   version,
		VersionsAvailable:  common.Apply(func(v types.PluginVersionMetadata) string { return v.Version }, pluginMeta.Spec.Versions),
		PlatformsSupported: getAllSupportedPlatforms(pluginMeta),
	}
	pluginInfoYaml, err := yaml.Marshal(pluginInfo)
	if err != nil {
		return "", fmt.Errorf("failed to marshal the plugin info to yaml. Error: %w", err)
	}
	return string(pluginInfoYaml), nil
}
