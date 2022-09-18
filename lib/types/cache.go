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

// LocalCache contains the list of installed plugins and other app specific metadata.
type LocalCache struct {
	ApiVersion string         `yaml:"apiVersion"`
	Kind       string         `yaml:"kind"`
	Metadata   MetadataInfo   `yaml:"metadata"`
	Spec       LocalCacheSpec `yaml:"spec"`
}

// LocalCacheSpec contains the list of installed plugins and other app specific metadata.
type LocalCacheSpec struct {
	Installed []InstalledPlugin `yaml:"installed"`
}

// InstalledPlugin contains the metadata for an installed plugin.
type InstalledPlugin struct {
	Name     string `yaml:"name"`
	Version  string `yaml:"version"`
	Platform string `yaml:"platform"`
	Bin      string `yaml:"bin"`
}
