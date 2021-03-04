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
	"encoding/json"
	"fmt"
	"log"

	user "github.com/MottainaiCI/mottainai-server/pkg/user"
	schema "github.com/MottainaiCI/mottainai-server/routes/schema"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"

	client "github.com/MottainaiCI/mottainai-server/pkg/client"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
)

func newUserShowCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "show <id>",
		Short: "Display user information",
		Args:  cobra.OnlyValidArgs,
		// TODO: PreRun check of minimal args if --json is not present
		Run: func(cmd *cobra.Command, args []string) {
			var v *viper.Viper = config.Viper

			if len(args) == 0 {
				log.Fatalln("You need to define a user id")
			}
			id := args[0]
			if len(id) == 0 {
				log.Fatalln("You need to define a user id")
			}

			fetcher := client.NewTokenClient(v.GetString("master"), v.GetString("apikey"), config)

			var t *user.User
			req := schema.Request{
				Route: v1.Schema.GetUserRoute("show"),
				Options: map[string]interface{}{
					":id": id,
				},
				Target: &t,
			}

			err := fetcher.Handle(req)
			if err != nil {
				log.Fatalln("error:", err)
			}
			b, err := json.MarshalIndent(t, "", "  ")
			if err != nil {
				log.Fatalln("error:", err)
			}
			fmt.Println(string(b))
		},
	}
	var flags = cmd.Flags()
	flags.String("type", "t", "Set the user id permission to the type ( e.g 'user' or 'admin')")

	return cmd
}
