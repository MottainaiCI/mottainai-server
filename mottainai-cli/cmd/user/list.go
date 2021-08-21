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
	"os"

	schema "github.com/MottainaiCI/mottainai-server/routes/schema"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"

	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	client "github.com/MottainaiCI/mottainai-server/pkg/client"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	user "github.com/MottainaiCI/mottainai-server/pkg/user"
	tablewriter "github.com/olekukonko/tablewriter"
	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
)

func newUserListCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list [OPTIONS]",
		Short: "List users",
		Args:  cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			var tlist []user.User
			var task_table [][]string
			var quiet bool
			var v *viper.Viper = config.Viper

			fetcher := client.NewTokenClient(v.GetString("master"), v.GetString("apikey"), config)

			req := &schema.Request{
				Route:  v1.Schema.GetUserRoute("show_all"),
				Target: &tlist,
			}
			err = fetcher.Handle(req)
			tools.CheckError(err)

			quiet, err = cmd.Flags().GetBool("quiet")
			tools.CheckError(err)

			if quiet {
				for _, i := range tlist {
					fmt.Println(i.ID)
				}
				return
			}

			for _, i := range tlist {
				task_table = append(task_table, []string{i.ID, i.Name, i.Email, i.Admin, i.Manager})
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
			table.SetCenterSeparator("|")
			table.SetHeader([]string{"ID", "Name", "Email", "Admin", "Manager"})

			for _, v := range task_table {
				table.Append(v)
			}
			table.Render()

		},
	}

	var flags = cmd.Flags()
	flags.BoolP("quiet", "q", false, "Quiet Output")

	return cmd
}
