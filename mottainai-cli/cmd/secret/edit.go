/*

Copyright (C) 2017-2021  Ettore Di Giacinto <mudler@gentoo.org>
                         Daniele Rondina <geaaru@sabayonlinux.org>

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

package secret

import (
	"io/ioutil"
	"log"
	"os"

	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	client "github.com/MottainaiCI/mottainai-server/pkg/client"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
)

func newSecretEditCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "edit <id> <key> [<value>|-f file]",
		Short: "Edit a secret",
		Args:  cobra.RangeArgs(2, 3),

		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 2 && cmd.Flag("from-file").Value.String() == "" {
				log.Fatalln("Missing value or --from-file option")
			} else if len(args) < 2 {
				cmd.Help()
				os.Exit(0)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			var content []byte
			var v *viper.Viper = config.Viper
			var value string

			id := args[0]
			key := args[1]

			if len(args) > 2 {
				value = args[2]
			} else {
				// Read value from file
				content, err = ioutil.ReadFile(cmd.Flag("from-file").Value.String())
				if err != nil {
					log.Fatalln("Error on read file ", err)
				}
				value = string(content)
			}

			fetcher := client.NewTokenClient(v.GetString("master"), v.GetString("apikey"), config)
			dat := make(map[string]interface{})

			dat["key"] = key
			dat["value"] = value
			dat["id"] = id

			res, err := fetcher.SecretEdit(dat)
			tools.CheckError(err)
			tools.PrintResponse(res)
		},
	}

	var pflags = cmd.PersistentFlags()
	pflags.StringP("from-file", "f", "", "Read value from file.")

	return cmd
}
