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

import "os"

const (
	// VALID_PLUGIN_FILENAME_PREFIX is the prefix used for plugin filenames.
	VALID_PLUGIN_FILENAME_PREFIX = "konveyor-"
	// DEFAULT_FILE_PERMISSIONS is the default permissions to use when creaing a new file.
	DEFAULT_FILE_PERMISSIONS os.FileMode = 0644
	// DEFAULT_DIRECTORY_PERMISSIONS is the default permissions to use when creaing a new directory.
	DEFAULT_DIRECTORY_PERMISSIONS os.FileMode = 0755
	// STORAGE_DIR is where all the app specific data is stored.
	STORAGE_DIR = ".konveyor"
	// PLUGINS_DIR is where all the plugins are stored.
	PLUGINS_DIR = "plugins"
	// CACHE_FILE contains the list of installed plugins and other app specific metadata.
	CACHE_FILE = "cache.yaml"
	// API_VERSION is the apiVersion (similar to K8s) used by our app specific files.
	API_VERSION = "cli.konveyor.io/v1alpha1"
	// KIND is the kind (similar to K8s) used by our app's local cache.
	CACHE_FILE_KIND = "Cache"
)
