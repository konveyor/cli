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

package common

import (
	"os"
	"path/filepath"

	"github.com/konveyor/cli/lib/types"
	"github.com/sirupsen/logrus"
)

// Apply applies the given function to all the items in the list and returns a new list.
func Apply[T1 interface{}, T2 interface{}](f func(T1) T2, t1s []T1) []T2 {
	t2s := []T2{}
	for _, t1 := range t1s {
		t2s = append(t2s, f(t1))
	}
	return t2s
}

// Filter uses the given condition to filter the items in the list and returns a new list.
func Filter[T interface{}](f func(T) bool, ts []T) []T {
	t1s := []T{}
	for _, t := range ts {
		if f(t) {
			t1s = append(t1s, t)
		}
	}
	return t1s
}

// FindIndex returns the index of the first item in the list that satifies the condition.
// Returns -1 if no item satifies the condition.
func FindIndex[T interface{}](f func(T) bool, ts []T) int {
	for i, t := range ts {
		if f(t) {
			return i
		}
	}
	return -1
}

// Contains returns true iff the given item is present in the list.
func Contains[T comparable](t T, ts []T) bool {
	return FindIndex(func(t1 T) bool { return t1 == t }, ts) != -1
}

// GetStorageDir returns the directory where we store all the plugins.
func GetStorageDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		logrus.Warnf("Failed to get the user's home directory. Error: %q", err)
		return types.STORAGE_DIR
	}
	if home == "" {
		return types.STORAGE_DIR
	}
	return filepath.Join(home, types.STORAGE_DIR)
}

// GetPluginDir returns the path to the plugins directory.
func GetPluginDir(name string) string {
	return filepath.Join(GetStorageDir(), types.PLUGINS_DIR, name)
}

// GetPlatformAsSingleString returns the Os and Arch as a single string.
func GetPlatformAsSingleString(os, arch string) string {
	return os + "-" + arch
}

// Keys returns the keys of a map.
func Keys[K comparable, V interface{}](m map[K]V) []K {
	ks := []K{}
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}
