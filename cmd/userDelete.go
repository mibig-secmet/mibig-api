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
	"fmt"

	"github.com/spf13/cobra"

	"secondarymetabolites.org/mibig-api/pkg/models/postgres"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete <email>",
	Short: "Delete a user for the MIBiG API server",
	Long: `Delete a user for the MIBiG API server.

Also cleans up group memberships of the deleted user.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		email := args[0]
		db, err := InitDb()
		if err != nil {
			panic(fmt.Errorf("Error opening database: %s", err))
		}

		userModel := postgres.NewSubmitterModel(db)
		err = userModel.Delete(email)
		if err != nil {
			panic(fmt.Errorf("Error deleting user: %s", err))
		}
	},
}

func init() {
	userCmd.AddCommand(deleteCmd)
}
