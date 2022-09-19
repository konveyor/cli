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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/konveyor/cli/lib/cache"
	"github.com/konveyor/cli/lib/common"
	"github.com/konveyor/cli/lib/plugin"
	"github.com/konveyor/cli/lib/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Execute is the start of the flow. It finds an executes the appropriate command based on the args.
func Execute() error {
	rootCmd := GetRootCommand()
	if cmd, _, err := rootCmd.Find(os.Args[1:]); err == nil && cmd != nil {
		return rootCmd.Execute()
	}

	cmdName := "" // first "non-flag" arguments
	rest := []string{}
	for i, arg := range os.Args[1:] {
		if !strings.HasPrefix(arg, "-") {
			cmdName = arg
			rest = os.Args[1+i+1:]
			break
		}
	}

	if cmdName == "" || cmdName == "help" || cmdName == "completion" || cmdName == cobra.ShellCompRequestCmd || cmdName == cobra.ShellCompNoDescRequestCmd {
		return rootCmd.Execute()
	}

	// Search for a plugin if no command is found.

	// Look in the local cache.
	localCache, err := cache.GetLocalCache()
	if err != nil {
		return fmt.Errorf("failed to get the local cache. Error: %q", err)
	}
	pluginPath := ""
	if len(localCache.Spec.Installed) > 0 {
		idx := common.FindIndex(func(p types.InstalledPlugin) bool { return p.Name == cmdName }, localCache.Spec.Installed)
		if idx != -1 {
			pluginPath = plugin.GetPluginBinPath(localCache.Spec.Installed[idx])
		} else {
			// Look in the PATH.
			pluginPaths, err := plugin.GetPluginsListFromPath(false)
			if err != nil {
				return fmt.Errorf("failed to get the list of plugins from the PATH. Error: %q", err)
			}
			idx := common.FindIndex(func(p string) bool { return filepath.Base(p) == types.VALID_PLUGIN_FILENAME_PREFIX+cmdName }, pluginPaths)
			if idx != -1 {
				pluginPath = pluginPaths[idx]
			}
		}
	}
	if pluginPath == "" {
		return fmt.Errorf("unknown command '%s'", cmdName)
	}
	logrus.Infof("Executing the plugin '%s' with the args: %+v", pluginPath, rest)
	if err := ExecutePlugin(pluginPath, rest, os.Environ()); err != nil {
		return fmt.Errorf("the plugin failed to run or did not exit properly. Error: %q", err)
	}
	return nil
}

// ExecutePlugin executes a plugin given the path to the binary, args and environment variables
func ExecutePlugin(executablePath string, cmdArgs, environment []string) error {
	// Windows does not support exec syscall.
	if runtime.GOOS == "windows" {
		cmd := Command(executablePath, cmdArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Env = environment
		return cmd.Run()
	}

	// invoke cmd binary relaying the environment and args given
	// append executablePath to cmdArgs, as execve will make first argument the "binary name".
	return syscall.Exec(executablePath, append([]string{executablePath}, cmdArgs...), environment)
}

// Command executes the given command on Windows
func Command(name string, arg ...string) *exec.Cmd {
	cmd := &exec.Cmd{
		Path: name,
		Args: append([]string{name}, arg...),
	}
	if filepath.Base(name) == name {
		lp, err := exec.LookPath(name)
		if lp != "" && !shouldSkipOnLookPathErr(err) {
			// Update cmd.Path even if err is non-nil.
			// If err is ErrDot (especially on Windows), lp may include a resolved
			// extension (like .exe or .bat) that should be preserved.
			cmd.Path = lp
		}
	}
	return cmd
}

func shouldSkipOnLookPathErr(err error) bool {
	return err != nil
}
