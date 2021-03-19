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
	"strings"

	"github.com/spf13/cobra"

	"secondarymetabolites.org/mibig-api/pkg/models"
	"secondarymetabolites.org/mibig-api/pkg/models/postgres"
)

// userListCmd represents the list command
var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all MIBiG users",
	Long: `List all MIBiG  users.

List users and their roles.`,
	Run: func(cmd *cobra.Command, args []string) {
		listUsers()
	},
}

func init() {
	userCmd.AddCommand(userListCmd)
}

func listUsers() {
	db, err := InitDb()
	if err != nil {
		panic(fmt.Errorf("Error opening database: %s", err))
	}

	userModel := postgres.NewSubmitterModel(db)

	users, err := userModel.List()
	if err != nil {
		panic(fmt.Errorf("Error listing users: %s", err))
	}

	fmt.Printf("ID\tEmail\tName\tPublic\tGDPR\tActive\tRoles\n")
	for _, user := range users {
		role_string := strings.Join(models.RolesToStrings(user.Roles), ", ")
		fmt.Printf("%s\t%s\t%s\t%t\t%t\t%t\t%s\n", user.Id, user.Email, user.Name, user.Public, user.GDPRConsent, user.Active, role_string)
	}
}
