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
	"database/sql"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const Version string = "0.1.0"

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "mibig-api",
	Short: "Manage the web API for the MIBiG database.",
	Long: `Run and manage the web API for the MIBiG database.

This tool will let you run the API and manage users in the database.`,
	Version: Version,
}

func SetBuildInfo(gitVer, buildTime string) {
	viper.Set("version", Version)
	viper.Set("buildTime", buildTime)
	viper.Set("gitVer", gitVer)

	rootCmd.Version = fmt.Sprintf("%s (%s, %s)", Version, gitVer, buildTime)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mibig-api.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
		viper.AddConfigPath("/etc/mibig-api")
		viper.AddConfigPath(".")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func InitDb() (*sql.DB, error) {
	db, err := sql.Open("postgres", viper.GetString("database.uri"))
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
