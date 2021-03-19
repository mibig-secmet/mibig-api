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

	"secondarymetabolites.org/mibig-api/pkg/models"
	"secondarymetabolites.org/mibig-api/pkg/models/postgres"
	"secondarymetabolites.org/mibig-api/pkg/utils"
)

// userAddRoleCmd represents the userAddRole command
var userAddRoleCmd = &cobra.Command{
	Use:   "add-role <email> <role> [<role>...]",
	Short: "Add role(s) to a user",
	Long: `Add role(s) to a user.

Only roles the user doesn't have yet will be added.`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		email := args[0]
		newRoleNames := args[1:]
		db, err := InitDb()
		if err != nil {
			panic(fmt.Errorf("Error opening database: %s", err))
		}

		submitterModel := postgres.NewSubmitterModel(db)

		user, err := submitterModel.Get(email, false)
		if err != nil {
			panic(fmt.Errorf("Error reading user for %s: %s", email, err))
		}

		oldRoleNames := models.RolesToStrings(user.Roles)

		roleNames := utils.UnionString(oldRoleNames, newRoleNames)

		user.Roles, err = submitterModel.GetRolesByName(roleNames)
		if err != nil {
			panic(fmt.Errorf("Error looking up roles for %v: %s", roleNames, err))
		}

		err = submitterModel.Update(user, "")
		if err != nil {
			panic(fmt.Errorf("Error updating user: %s", err))
		}

	},
}

func init() {
	userEditCmd.AddCommand(userAddRoleCmd)
}
