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

func newUserCreateCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create --name [user] --email [email] --password [password] [OPTIONS]",
		Short: "Create a user",
		Args:  cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fetcher, err := utils.CreateClient(config)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			dat := make(map[string]interface{})
			u := &user.UserForm{}

			name, err := cmd.Flags().GetString("name")
			tools.CheckError(err)
			email, err := cmd.Flags().GetString("email")
			tools.CheckError(err)
			password, err := cmd.Flags().GetString("password")
			tools.CheckError(err)

			if name == "" {
				log.Fatalln("Missing mandatory parameter name")
			}
			if email == "" {
				log.Fatalln("Missing mandatory parameter email")
			}
			if password == "" {
				log.Fatalln("Missing mandatory parameter password")
			}

			u.Name = name
			u.Email = email
			u.Password = password

			dat = u.ToMap()

			res, err := fetcher.UserCreate(dat)
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
