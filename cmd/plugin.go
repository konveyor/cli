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
	"errors"
	"fmt"
	"strings"

	"github.com/konveyor/cli/lib/common"
	"github.com/konveyor/cli/lib/github"
	"github.com/konveyor/cli/lib/plugin"
	"github.com/konveyor/cli/lib/types"
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
	pluginCmd.AddCommand(GetPluginInstallCommand())
	pluginCmd.AddCommand(GetPluginUninstallCommand())
	pluginCmd.AddCommand(GetPluginTidyCommand())
	pluginCmd.AddCommand(GetPluginInfoCommand())
	return pluginCmd
}

// GetPluginListSubCommand returns a command to list all the installed plugins.
func GetPluginListSubCommand() *cobra.Command {
	nameOnly := false
	remote := false
	pluginListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all the installed plugins.",
		Long: `List all the installed plugins.

    Installed plugins are those that are: - executable - anywhere on the user's PATH - begin with "` + types.VALID_PLUGIN_FILENAME_PREFIX + `"
    Also includes any plugins in the ` + common.GetStorageDir() + ` directory
`,
		Run: func(*cobra.Command, []string) {
			if remote {
				logrus.Infof("Fetching the list of plugins from Github.")
			} else {
				logrus.Infof("Looking for installed plugins.")
			}
			if remote {
				plugins, err := github.GetPluginsListFromGithub()
				if err != nil {
					logrus.Fatalf("failed to get the list of plugins from Github. Error: %q", err)
				}
				logrus.Infof("The following plugins are available on Github:\n%s", strings.Join(plugins, "\n"))
				return
			}
			plugins, err := plugin.GetPluginsList(nameOnly)
			if err != nil {
				logrus.Fatalf("failed to get the list of plugins from the PATH. Error: %q", err)
			}
			if len(plugins) == 0 {
				logrus.Info("No plugins were found.")
				return
			}
			logrus.Infof("The following plugins are installed:\n%s", strings.Join(plugins, "\n"))
		},
	}
	pluginListCmd.Flags().BoolVar(&nameOnly, "name-only", false, "If true, display only the binary name of each plugin, rather than its full path")
	pluginListCmd.Flags().BoolVar(&remote, "remote", false, "If true, display only the list of plugins in the Github repo")
	return pluginListCmd
}

// GetPluginInstallCommand returns a command to install a plugin.
func GetPluginInstallCommand() *cobra.Command {
	pluginInstallCmd := &cobra.Command{
		Use:   "install",
		Args:  cobra.MinimumNArgs(1),
		Short: "Install a plugin",
		Long:  "Install a plugin",
		Run: func(_ *cobra.Command, args []string) {
			name := args[0]
			logrus.Infof("Looking for a plugin named '%s' on Github.", name)
			if err := plugin.InstallPluginFromGithub(name); err != nil {
				if errors.Is(err, types.ErrPluginAlreadyInstalled) {
					logrus.Fatal(err)
				}
				logrus.Fatalf("failed to find or install the plugin named '%s'. Error: %q", name, err)
			}
			logrus.Infof("The plugin named '%s' was installed!", name)
		},
	}
	return pluginInstallCmd
}

// GetPluginUninstallCommand returns a command to uninstall a plugin.
func GetPluginUninstallCommand() *cobra.Command {
	pluginUninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Args:  cobra.MinimumNArgs(1),
		Short: "Uninstall a plugin",
		Long:  "Uninstall a plugin",
		Run: func(_ *cobra.Command, args []string) {
			name := args[0]
			logrus.Infof("Looking for a plugin named '%s' among the installed plugins.", name)
			if err := plugin.UninstallPlugin(name); err != nil {
				logrus.Fatalf("failed to find or uninstall the plugin named '%s'. Error: %q", name, err)
			}
			logrus.Infof("The plugin named '%s' was uninstalled!", name)
		},
	}
	return pluginUninstallCmd
}

// GetPluginTidyCommand returns a command to tidy the plugins directory.
func GetPluginTidyCommand() *cobra.Command {
	pluginTidyCmd := &cobra.Command{
		Use:   "tidy",
		Args:  cobra.NoArgs,
		Short: "Cleans the plugins directory, removing any broken plugins to ensure consistency with the local cache",
		Long:  "Cleans the plugins directory, removing any broken plugins to ensure consistency with the local cache",
		Run: func(*cobra.Command, []string) {
			logrus.Infof("Looking for any inconsistencies between the local cache and the installed plugins.")
			if err := plugin.UninstallBrokenPlugins(); err != nil {
				logrus.Fatalf("failed to uninstall all the broken plugins. Error: %q", err)
			}
			logrus.Infof("Tidying done!")
		},
	}
	return pluginTidyCmd
}

// GetPluginInfoCommand returns a command to display info about a plugin.
func GetPluginInfoCommand() *cobra.Command {
	pluginInfoCmd := &cobra.Command{
		Use:   "info",
		Args:  cobra.MinimumNArgs(1),
		Short: "Displays info about a plugin - available versions, links to homepage, documentation, etc.",
		Long:  "Displays info about a plugin - available versions, links to homepage, documentation, etc.",
		Run: func(_ *cobra.Command, args []string) {
			logrus.Debugf("plugin info called")
			name := args[0]
			logrus.Infof("Looking for information on a plugin named '%s'", name)
			info, err := plugin.GetPluginInfo(name)
			if err != nil {
				logrus.Fatalf("failed to find a plugin named '%s'. Error: %q", name, err)
			}
			logrus.Infof("Found the following information about the '%s' plugin:", name)
			fmt.Println(info)
		},
	}
	return pluginInfoCmd
}
