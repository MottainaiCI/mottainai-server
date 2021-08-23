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
	"fmt"
	"log"
	"os"

	utils "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/utils"
	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	user "github.com/MottainaiCI/mottainai-server/pkg/user"

	cobra "github.com/spf13/cobra"
)

func newUserEditCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "edit [OPTIONS]",
		Short: "Edit a user",
		Args:  cobra.OnlyValidArgs,
		// TODO: PreRun check of minimal args if --json is not present
		Run: func(cmd *cobra.Command, args []string) {
			dat := make(map[string]interface{})

			if len(args) == 0 {
				log.Fatalln("You need to define a user id")
			}

			id := args[0]
			if len(id) == 0 {
				log.Fatalln("You need to define a user id")
			}
			name, err := cmd.Flags().GetString("name")
			tools.CheckError(err)
			email, err := cmd.Flags().GetString("email")
			tools.CheckError(err)
			password, err := cmd.Flags().GetString("password")
			tools.CheckError(err)
			us := &user.User{}
			u := &user.UserForm{}
			if name != "" {
				u.Name = name
			}
			if email != "" {
				u.Email = email
			}
			if password != "" {
				us.Password = password
				us.SaltPassword()
				u.Password = us.Password
			}

			fetcher, err := utils.CreateClient(config)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			dat = u.ToMap()

			res, err := fetcher.UserUpdate(id, dat)
			tools.CheckError(err)
			tools.PrintResponse(res)
		},
	}
	var flags = cmd.Flags()
	flags.String("email", "", "Email of the user")
	flags.String("name", "", "Name of the user")
	flags.String("password", "", "Password of the user")

	return cmd
}
