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

// PluginInfo stores some info/summary about the plugin in a human readable format.
type PluginInfo struct {
	Name               string   `yaml:"name"`
	Description        string   `yaml:"description"`
	Installed          bool     `yaml:"installed"`
	InstalledVersion   string   `yaml:"installed-version,omitempty"`
	HomePage           string   `yaml:"home-page,omitempty"`
	Documentation      string   `yaml:"documentation,omitempty"`
	Tutorials          string   `yaml:"tutorials,omitempty"`
	VersionsAvailable  []string `yaml:"versions-available,omitempty"`
	PlatformsSupported []string `yaml:"platforms-supported,omitempty"`
}
