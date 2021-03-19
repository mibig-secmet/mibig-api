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
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	readpass "github.com/seehuhn/password"
	"github.com/spf13/cobra"

	"secondarymetabolites.org/mibig-api/pkg/models"
	"secondarymetabolites.org/mibig-api/pkg/models/postgres"
)

var (
	active       bool
	call_name    string
	email        string
	gdpr_consent bool
	institution  string
	name         string
	password     string
	public       bool
	role_list    []string
)

// userAddCmd represents the add command
var userAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add user for the MIBiG API server",
	Long: `Add user for the MIBiG API server.

Required parameters can be passed on the command line, or added in the interactive prompt.`,
	Run: func(cmd *cobra.Command, args []string) {
		user := models.Submitter{
			Email:       email,
			Name:        name,
			CallName:    call_name,
			Institution: institution,
			Public:      public,
			GDPRConsent: gdpr_consent,
			Active:      active,
		}

		db, err := InitDb()
		if err != nil {
			panic(fmt.Errorf("Error opening database: %s", err))
		}
		submitterModel := postgres.NewSubmitterModel(db)
		user.Roles, err = submitterModel.GetRolesByName(role_list)
		if err != nil {
			panic(fmt.Errorf("Error getting roles: %s", err))
		}
		roleModel := postgres.RoleModel{DB: db}

		if user.Email == "" || user.Name == "" || password == "" {
			for {
				password = InteractiveUserEdit(&user, &roleModel, submitterModel)
				if user.Email != "" && user.Name != "" && password != "" {
					break
				}
				fmt.Println("*** Invalid user data, please try again ***")
			}
		}

		err = submitterModel.Insert(&user, password)
		if err != nil {
			panic(fmt.Errorf("Error adding user: %s", err))
		}
	},
}

func init() {
	userCmd.AddCommand(userAddCmd)

	userAddCmd.Flags().BoolVarP(&active, "active", "a", true, "Added account is active")
	userAddCmd.Flags().StringVarP(&call_name, "call-name", "C", "", "How to address the user")
	userAddCmd.Flags().StringVarP(&email, "email", "e", "", "Email address of user")
	userAddCmd.Flags().BoolVarP(&gdpr_consent, "gdpr-consent", "g", false, "Added account is consentet to us using the data")
	userAddCmd.Flags().StringVarP(&institution, "institution", "i", "", "Name of user's institute/company")
	userAddCmd.Flags().StringVarP(&name, "name", "n", "", "Name of user")
	userAddCmd.Flags().StringVarP(&password, "password", "p", "", "Password of user")
	userAddCmd.Flags().BoolVarP(&public, "public", "P", false, "Added account is public")
	userAddCmd.Flags().StringSliceVarP(&role_list, "role", "r", []string{"guest"}, "Roles of the user")
}

func InteractiveUserEdit(user *models.Submitter, roleModel models.RoleModel, submitterModel models.SubmitterModel) string {
	reader := bufio.NewReader(os.Stdin)

	user.Email = readStringValue(reader, user.Email, "Email [%s]: ")
	user.Name = readStringValue(reader, user.Name, "Name [%s]: ")
	user.CallName = readStringValue(reader, user.CallName, "Call name [%s]: ")
	if user.CallName == "" {
		user.CallName = strings.Split(user.Name, " ")[0]
	}
	user.Institution = readStringValue(reader, user.Institution, "Organisation [%s]: ")
	new_password := readPassword()
	user.Public = readBool(reader, user.Public, "Public profile (true/false) [%s]: ")
	user.GDPRConsent = readBool(reader, user.GDPRConsent, "GDPR consent given (true/false) [%s]: ")
	user.Active = readBool(reader, user.Active, "Active (true/false) [%s]: ")
	user.Roles = readRoles(reader, roleModel, submitterModel, user.Roles)

	return new_password
}

func readStringValue(reader *bufio.Reader, old_value, template string) string {
	var newVal string
	for {
		fmt.Printf(template, old_value)
		tmp_string := readInput(reader)
		if tmp_string == "" {
			tmp_string = old_value
		}
		newVal = tmp_string
		if len(newVal) > 0 {
			break
		}
	}
	return newVal
}

func readPassword() string {
	var password string
	var password_repeat string

	for {
		pw_bytes, err := readpass.Read("Password (empty to keep old): ")
		if err != nil {
			panic(fmt.Errorf("Error reading password: %s", err))
		}
		password = string(pw_bytes)

		if password == "" {
			return password
		}

		pw_bytes, err = readpass.Read("Repeat password: ")
		if err != nil {
			panic(fmt.Errorf("Error reading password: %s", err))
		}
		password_repeat = string(pw_bytes)

		if strings.Compare(password, password_repeat) == 0 {
			break
		}
		fmt.Println("Password mismatch")
	}

	return password
}

func readBool(reader *bufio.Reader, oldVal bool, template string) bool {
	var newVal bool
	var err error

	for {
		fmt.Printf(template, strconv.FormatBool(oldVal))
		tmp_string := readInput(reader)
		if tmp_string == "" {
			return oldVal
		}
		newVal, err = strconv.ParseBool(tmp_string)
		if err == nil {
			break
		}
		fmt.Println("Invalid input: ", tmp_string)
	}

	return newVal
}

func readRoles(reader *bufio.Reader, roleModel models.RoleModel, submitterModel models.SubmitterModel, old_roles []models.Role) []models.Role {
	var new_roles []models.Role

	availableRoles, err := roleModel.List()
	if err != nil {
		panic(fmt.Errorf("Error reading roles: %s", err))
	}

	fmt.Println("Available roles:", strings.Join(models.RolesToStrings(availableRoles), ", "))

	for {
		stringRoles := strings.Join(models.RolesToStrings(old_roles), ", ")
		fmt.Printf("Roles [%s]: ", stringRoles)
		tmp_string := readInput(reader)
		if tmp_string == "" {
			return old_roles
		}
		parts := strings.Split(strings.Replace(tmp_string, " ", "", -1), ",")
		fmt.Fprintf(os.Stderr, "%v\n", parts)
		new_roles, err = submitterModel.GetRolesByName(parts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting roles: %s", err.Error())
			continue
		}

		fmt.Fprintf(os.Stderr, "%v\n", new_roles)
		// TODO: check if all roles are valid
		if len(new_roles) > 0 {
			break
		}
	}

	fmt.Fprintf(os.Stderr, "%v\n", new_roles)
	return new_roles
}

func readInput(reader *bufio.Reader) string {
	userInput, err := reader.ReadString('\n')
	if err != nil {
		panic(fmt.Errorf("Error reading user input: %s", err.Error()))
	}
	userInput = strings.Replace(userInput, "\n", "", -1)
	return userInput
}
