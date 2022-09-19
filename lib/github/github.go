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

package github

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/google/go-github/v47/github"
	"github.com/sirupsen/logrus"
)

const (
	// REPO_OWNER is the username of the owner of the plugins Github repo.
	REPO_OWNER = "konveyor"
	// REPO_NAME is the name of the plugins Github repo.
	REPO_NAME = "cli"
	// REPO_BRANCH is the branch of the plugins Github repo where the metadata for the plugins are stored.
	REPO_BRANCH = "main"
	// REPO_PLUGINS_DIR is the directory on the Github repo where the metadata for the plugins are stored.
	REPO_PLUGINS_DIR = "plugins"
)

// GetPluginsListFromGithub returns the list of plugins from the Github repo.
func GetPluginsListFromGithub() ([]string, error) {
	client := github.NewClient(nil)
	_, dirContent, resp, err := client.Repositories.GetContents(
		context.Background(),
		REPO_OWNER,
		REPO_NAME,
		REPO_PLUGINS_DIR,
		&github.RepositoryContentGetOptions{Ref: REPO_BRANCH},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list the contents of the plugins folder on the Github repo. Error: %q", err)
	}
	logrus.Debugf("resp: %#v", resp)
	pluginNames := []string{}
	for i, pluginYaml := range dirContent {
		if pluginYaml == nil {
			logrus.Errorf("expected a file, but got a nil pointer")
			continue
		}
		logrus.Debugf("[%d] %#v", i, pluginYaml)
		if pluginYaml.Name == nil {
			logrus.Errorf("the file/directory name is nil")
			continue
		}
		name := strings.TrimSuffix(*pluginYaml.Name, ".yaml")
		logrus.Debugf("plugin name: %s", name)
		pluginNames = append(pluginNames, name)
	}
	return pluginNames, nil
}

// GetPluginYamlFromGithub gets the plugin yaml from the Github repo.
func GetPluginYamlFromGithub(name string) ([]byte, error) {
	return GetFileFromGithub(REPO_PLUGINS_DIR + "/" + name + ".yaml")
}

// GetFileFromGithub gets the contents of a file from the Github repo.
func GetFileFromGithub(path string) ([]byte, error) {
	client := github.NewClient(nil)
	fileContent, _, resp, err := client.Repositories.GetContents(
		context.Background(),
		REPO_OWNER,
		REPO_NAME,
		path,
		&github.RepositoryContentGetOptions{Ref: REPO_BRANCH},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list the contents of the plugins folder on the Github repo. Error: %q", err)
	}
	logrus.Debugf("resp: %#v", resp)
	if fileContent == nil {
		return nil, fmt.Errorf("expected a file, but got a nil pointer")
	}
	if fileContent.Content == nil {
		return nil, fmt.Errorf("the file content is nil")

	}
	contentb64 := *fileContent.Content
	if contentb64 == "" {
		return nil, nil
	}
	logrus.Debugf("contentb64: %s", contentb64)
	content, err := base64.StdEncoding.DecodeString(contentb64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode the file contents as base64. Error: %q", err)
	}
	return content, nil
}
