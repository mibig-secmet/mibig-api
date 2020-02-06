/*
Copyright © 2020 Kai Blin <kblin@biosustain.dtu.dk>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"secondarymetabolites.org/mibig-api/pkg/web"
)

var debug bool

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the MIBiG API server",
	Long: `Run the MIBiG API server.

This service provides access to the MIBiG database.`,
	Run: func(cmd *cobra.Command, args []string) {
		if viper.GetString("server.name") == "" {
			address := viper.GetString("server.address")
			if address == "" {
				address = "localhost"
			}
			viper.Set("server.name", address)
		}

		web.Run(debug)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug info")
}
