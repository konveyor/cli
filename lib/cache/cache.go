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

package cache

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/konveyor/cli/lib/common"
	"github.com/konveyor/cli/lib/types"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// GetLocalCache returns the local cache.
// It creates the file if it doesn't exist.
func GetLocalCache() (types.LocalCache, error) {
	cache := types.LocalCache{
		ApiVersion: types.API_VERSION,
		Kind:       types.CACHE_FILE_KIND,
		Metadata:   types.MetadataInfo{Name: "cache"},
	}
	cachePath := filepath.Join(common.GetStorageDir(), types.CACHE_FILE)
	cacheBytes, err := ioutil.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Infof("The local cache doesn't exist. Creating it...")
			return cache, SaveLocalCache(cache)
		}
		return cache, fmt.Errorf("failed to read the local cache file at path %s . Error: %q", cachePath, err)
	}
	if err := yaml.Unmarshal(cacheBytes, &cache); err != nil {
		return cache, fmt.Errorf("failed to unmarshal the local cache from yaml. Error: %q", err)
	}
	return cache, nil
}

// SaveLocalCache saves the updated local cache to file.
func SaveLocalCache(cache types.LocalCache) error {
	storageDir := common.GetStorageDir()
	if err := os.MkdirAll(storageDir, types.DEFAULT_DIRECTORY_PERMISSIONS); err != nil {
		return fmt.Errorf("failed to create the storage directory %s . Error: %q", storageDir, err)
	}
	cachePath := filepath.Join(storageDir, types.CACHE_FILE)
	cacheYaml, err := yaml.Marshal(cache)
	if err != nil {
		return fmt.Errorf("failed to marshal the local cache to yaml. Error: %q", err)
	}
	if err := ioutil.WriteFile(cachePath, cacheYaml, types.DEFAULT_FILE_PERMISSIONS); err != nil {
		return fmt.Errorf("failed to write the local cache to a file at path %s . Error: %q", cachePath, err)
	}
	return nil
}
