/*

Copyright (C) 2017-2021  Ettore Di Giacinto <mudler@gentoo.org>
                         Daniele Rondina <geaaru@sabayonlinux.org>
Credits goes also to Gogs authors, some code portions and re-implemented design
are also coming from the Gogs project, which is using the go-macaron framework
and was really source of ispiration. Kudos to them!

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.

*/

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	s "github.com/MottainaiCI/mottainai-server/pkg/settings"
	utils "github.com/MottainaiCI/mottainai-server/pkg/utils"
	viper "github.com/spf13/viper"
)

const (
	srvName = `Mottainai CLI
Copyright (c) 2017-2022 Mottainai

Mottainai - Task/Job Build Service`

	srvExamples = `$> mottainai-server web -c mottainai-server.yaml

$> mottainai-server daemon -c mottainai-server.yaml

$> mottainai-server daemon -r -e http://127.0.0.1:4001 -c mottainai1/mottainai-server.yaml
`
)

func initConfig(config *s.Config) {
	// Set env variable
	config.Viper.SetEnvPrefix(s.MOTTAINAI_ENV_PREFIX)
	config.Viper.BindEnv("config")
	config.Viper.SetDefault("config", "")
	config.Viper.SetDefault("etcd-config", false)

	config.Viper.AutomaticEnv()

	// Create EnvKey Replacer for handle complex structure
	replacer := strings.NewReplacer(".", "__")
	config.Viper.SetEnvKeyReplacer(replacer)

	// Set config file name (without extension)
	config.Viper.SetConfigName(s.MOTTAINAI_CONFIGNAME)

	config.Viper.SetTypeByDefaultValue(true)
}

func initCommand(rootCmd *cobra.Command, config *s.Config) {
	var pflags = rootCmd.PersistentFlags()

	pflags.StringP("config", "c", "/etc/mottainai/mottainai-server.yaml",
		"Mottainai Server configuration file or Etcd path")
	pflags.BoolP("remote-config", "r", false,
		"Enable etcd remote config provider")
	pflags.StringP("etcd-endpoint", "e", "http://127.0.0.1:4001",
		"Etcd Server Address")
	pflags.String("etcd-keyring", "",
		"Etcd Keyring (Ex: /etc/secrets/mykeyring.gpg)")

	config.Viper.BindPFlag("config", pflags.Lookup("config"))
	config.Viper.BindPFlag("etcd-config", pflags.Lookup("remote-config"))
	config.Viper.BindPFlag("etcd-endpoint", pflags.Lookup("etcd-endpoint"))
	config.Viper.BindPFlag("etcd-keyring", pflags.Lookup("etcd-keyring"))

	rootCmd.AddCommand(
		newDaemonCommand(config),
		newPrintCommand(config),
		newWebCommand(config),
		newWebHookCommand(config),
	)
}

func Execute() {
	// Create Main Instance Config object
	var config *s.Config = s.NewConfig(nil)

	initConfig(config)

	var rootCmd = &cobra.Command{
		Short:        srvName,
		Version:      fmt.Sprintf("%s-g%s %s %s", s.MOTTAINAI_VERSION, s.BuildCommit, s.BuildTime, s.BuildGoVersion),
		Example:      srvExamples,
		Args:         cobra.OnlyValidArgs,
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				os.Exit(0)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var err error
			var pwd string
			var v *viper.Viper = config.Viper

			if v.GetBool("etcd-config") {
				if v.Get("etcd-keyring") != "" {
					v.AddSecureRemoteProvider("etcd", v.GetString("etcd-endpoint"),
						v.GetString("config"), v.GetString("etcd-keyring"))
				} else {
					v.AddRemoteProvider("etcd", v.GetString("etcd-endpoint"),
						v.GetString("config"))
				}
				v.SetConfigType("yml")
			} else {
				if v.Get("config") == "" {
					// Set config path list
					pwd, err = os.Getwd()
					utils.CheckError(err)
					v.AddConfigPath(pwd)
					v.AddConfigPath(s.MOTTAINAI_CONFIGPATH)
				} else {
					v.SetConfigFile(v.Get("config").(string))
				}
			}

			// Parse configuration file
			err = config.Unmarshal()
			utils.CheckError(err)
		},
	}

	initCommand(rootCmd, config)

	// Start command execution
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
