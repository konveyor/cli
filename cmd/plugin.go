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

package cmd

import (
	"strings"

	"github.com/konveyor/cli/lib"
	"github.com/konveyor/cli/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// GetPluginCommand returns the plugin command
func GetPluginCommand() *cobra.Command {
	pluginCmd := &cobra.Command{
		Use:   "plugin",
		Short: "Provides utilities for interacting with plugins.",
		Long: `Provides utilities for interacting with plugins.

	Plugins provide extended functionality that is not part of the major command-line distribution.
	Please refer to the documentation and examples for more information about how write your own plugins.
`,
	}
	pluginCmd.AddCommand(GetPluginListSubCommand())
	return pluginCmd
}

func GetPluginListSubCommand() *cobra.Command {
	nameOnly := false
	pluginListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all available plugin files on a user's PATH.",
		Long: `List all available plugin files on a user's PATH.

		Available plugin files are those that are: - executable - anywhere on the user's PATH - begin with "` + types.ValidPluginFilenamePrefix + `"
`,
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Debug("command plugin list called")
			plugins, err := lib.GetPluginsFromPath(nameOnly)
			if err != nil {
				logrus.Fatalf("failed to get the list of plugins from the PATH. Error: %q", err)
			}
			if len(plugins) == 0 {
				logrus.Info("No plugins were found in the PATH")
				return
			}
			logrus.Infof("The following compatible plugins are available:\n\n%s", strings.Join(plugins, "\n"))
		},
	}
	pluginListCmd.Flags().BoolVar(&nameOnly, "name-only", false, "If true, display only the binary name of each plugin, rather than its full path")
	return pluginListCmd
}
