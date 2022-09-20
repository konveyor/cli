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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
)

// GetRootCommand returns the root command that contains all the other commands.
func GetRootCommand() *cobra.Command {
	loglevel := string(logrus.InfoLevel.String())
	rootCmd := &cobra.Command{
		Use:   "konveyor",
		Short: "Konveyor provides a suite of tools that help migrate apps running on legacy platforms to new ones.",
		Long: `Konveyor provides a suite of tools that help migrate apps running on legacy platforms to new ones.
Each tool comes as a plugin that can be installed if necessary.

Try "konveyor plugin --help" for more info about plugins and their installation.
`,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			logl, err := logrus.ParseLevel(loglevel)
			if err != nil {
				logrus.Errorf("the log level '%s' is invalid, using 'info' log level instead. Error: %w", loglevel, err)
				logl = logrus.InfoLevel
			}
			logrus.SetLevel(logl)
			return nil
		},
	}
	rootCmd.PersistentFlags().StringVar(&loglevel, "log-level", logrus.InfoLevel.String(), "Set logging levels.")
	rootCmd.AddCommand(GetPluginCommand())
	rootCmd.AddCommand(GetVersionCommand())
	return rootCmd
}

// AvoidGoModWarnings does nothing but avoid warnings about unused packages in go.mod
func AvoidGoModWarnings() {
	// Just here because the scripts/detectgoversion/detect.go script uses `modfile`
	// and that script is ignored using a build tag.
	// So we need to use `modfile` in a file that's not ignored to avoid warnings.
	if _, err := modfile.Parse("go.mod", nil, nil); err != nil {
		logrus.Fatal(err)
	}
}
