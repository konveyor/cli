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
	"path"
	"runtime"
	"strings"
	"syscall"

	"github.com/konveyor/cli/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

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

	// search for a plugin if no command is found

	logrus.Debugf("Did not find a valid sub command given the args: %+v", os.Args)
	pluginCmd := exec.Command(types.ValidPluginFilenamePrefix + cmdName)
	logrus.Debugf("pluginCmd: %#v", pluginCmd)
	if !path.IsAbs(pluginCmd.Path) {
		return fmt.Errorf("unknown command '%s'", cmdName)
	}
	logrus.Infof("Executing the plugin '%s' with the args: %+v", pluginCmd.Path, rest)
	if err := ExecutePlugin(pluginCmd.Path, rest, os.Environ()); err != nil {
		return fmt.Errorf("the plugin failed to run or did not exit properly. Error: %q", err)
	}
	return nil
}

// ExecutePlugin executes a plugin given the path to the binary, args and environment variables
func ExecutePlugin(executablePath string, cmdArgs, environment []string) error {

	// Windows does not support exec syscall.
	if runtime.GOOS == "windows" {
		cmd := exec.Command(executablePath, cmdArgs...)
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
