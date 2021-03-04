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

package user

import (
	"log"

	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	client "github.com/MottainaiCI/mottainai-server/pkg/client"
	event "github.com/MottainaiCI/mottainai-server/pkg/event"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
)

func newUserSetCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "set [OPTIONS]",
		Short: "Set/unset admin flag to user",
		Args:  cobra.OnlyValidArgs,
		// TODO: PreRun check of minimal args if --json is not present
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			var res event.APIResponse
			var v *viper.Viper = config.Viper
			t, err := cmd.Flags().GetString("type")

			if len(args) == 0 {
				log.Fatalln("You need to define a user id")
			}

			id := args[0]
			if len(id) == 0 {
				log.Fatalln("You need to define a user id")
			}
			fetcher := client.NewTokenClient(v.GetString("master"), v.GetString("apikey"), config)

			if t == "admin" {
				res, err = fetcher.UserSet(id, "admin")
			} else if t == "user" {
				res, err = fetcher.UserUnset(id, "admin")
				res, err = fetcher.UserUnset(id, "manager")
			} else if t == "manager" {
				res, err = fetcher.UserSet(id, "manager")
			}

			tools.CheckError(err)
			tools.PrintResponse(res)
		},
	}
	var flags = cmd.Flags()
	flags.String("type", "t", "Set the user id permission to the type ( e.g 'user' or 'admin')")

	return cmd
}
