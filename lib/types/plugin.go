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

package types

// PluginMetadata stores the plugin metadata from the YAML.
type PluginMetadata struct {
	ApiVersion string             `yaml:"apiVersion"`
	Kind       string             `yaml:"kind"`
	Metadata   MetadataInfo       `yaml:"metadata"`
	Spec       PluginMetadataSpec `yaml:"spec"`
}

// MetadataInfo contains the name.
type MetadataInfo struct {
	Name string `yaml:"name"`
}

// PluginMetadataSpec stores the specification of the plugin metadata.
type PluginMetadataSpec struct {
	HomePage         string                  `yaml:"homePage"`
	Docs             string                  `yaml:"docs"`
	Tutorials        string                  `yaml:"tutorials"`
	ShortDescription string                  `yaml:"shortDescription"`
	Description      string                  `yaml:"description"`
	Versions         []PluginVersionMetadata `yaml:"versions"`
}

// PluginVersionMetadata stores the metadata of a specific version of the plugin.
type PluginVersionMetadata struct {
	Version   string                     `yaml:"version"`
	Platforms []PluginVersionForPlatform `yaml:"platforms"`
}

// PluginVersionForPlatform contains the version and platform specific metadata.
type PluginVersionForPlatform struct {
	Selector Selector `yaml:"selector"`
	Uri      string   `yaml:"uri"`
	Sha256   string   `yaml:"sha256"`
	Bin      string   `yaml:"bin"`
}

// Selector contains the platform selector.
type Selector struct {
	MatchLabels MatchLabels `yaml:"matchLabels"`
}

// MatchLabels contains the platform selector.
type MatchLabels struct {
	Os   string `yaml:"os"`
	Arch string `yaml:"arch"`
}
