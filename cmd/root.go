/*

Copyright (C) 2017-2018  Ettore Di Giacinto <mudler@gentoo.org>
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

	"github.com/spf13/cobra"

	s "github.com/MottainaiCI/mottainai-server/pkg/settings"
	utils "github.com/MottainaiCI/mottainai-server/pkg/utils"
	viper "github.com/spf13/viper"
)

const (
	srvName = `Mottainai CLI
Copyright (c) 2017-2018 Mottainai

Mottainai - Task/Job Build Service`

	srvExamples = `$> mottainai-server web -c mottainai-server.yaml

$> mottainai-server daemon -c mottainai-server.yaml

$> mottainai-server daemon -r -e http://127.0.0.1:4001 -c mottainai1/mottainai-server.yaml
`
)

var rootCmd = &cobra.Command{
	Short:        srvName,
	Version:      s.MOTTAINAI_VERSION,
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
		var v *viper.Viper = s.Configuration.Viper

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
		err = s.Configuration.Unmarshal()
		utils.CheckError(err)
	},
}

func init() {
	var pflags = rootCmd.PersistentFlags()

	pflags.StringP("config", "c", "/etc/mottainai/mottainai-server.yaml",
		"Mottainai Server configuration file or Etcd path")
	pflags.BoolP("remote-config", "r", false,
		"Enable etcd remote config provider")
	pflags.StringP("etcd-endpoint", "e", "http://127.0.0.1:4001",
		"Etcd Server Address")
	pflags.String("etcd-keyring", "",
		"Etcd Keyring (Ex: /etc/secrets/mykeyring.gpg)")

	s.Configuration.Viper.BindPFlag("config", pflags.Lookup("config"))
	s.Configuration.Viper.BindPFlag("etcd-config", pflags.Lookup("remote-config"))
	s.Configuration.Viper.BindPFlag("etcd-endpoint", pflags.Lookup("etcd-endpoint"))
	s.Configuration.Viper.BindPFlag("etcd-keyring", pflags.Lookup("etcd-keyring"))

	rootCmd.AddCommand(
		newDaemonCommand(),
		newPrintCommand(),
		newWebCommand(),
		newWebHookCommand(),
	)
}

func Execute() {
	// Start command execution
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
