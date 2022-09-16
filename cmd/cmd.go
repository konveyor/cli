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
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/konveyor/cli/types"
	"github.com/sirupsen/logrus"
)

func Execute() {
	rootCmd := GetRootCommand()
	cmd, _, err := rootCmd.Find(os.Args[1:])
	if err == nil && cmd != nil {
		if err := rootCmd.Execute(); err != nil {
			logrus.Fatalf("Error: %q", err)
		}
		return
	}

	// default cmd if no cmd is given
	logrus.Debugf("Did not find a valid sub command given the args: %+v", os.Args)
	if len(os.Args) < 2 || strings.HasPrefix(os.Args[1], "-") {
		logrus.Fatalf("Invalid args. Try konveyor --help")
	}
	pluginCmd := exec.Command(types.ValidPluginFilenamePrefix + os.Args[1])
	logrus.Debugf("pluginCmd: %#v", pluginCmd)
	if pluginCmd.Path == "" {
		logrus.Debugf("the path to the plugin executable is empty")
		return
	}
	logrus.Infof("Executing the plugin: %s with the args: %+v", pluginCmd.Path, os.Args[2:])
	if err := ExecutePlugin(pluginCmd.Path, os.Args[2:], nil); err != nil {
		logrus.Fatalf("the plugin failed to run or did not exit properly. Error: %q", err)
	}
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
