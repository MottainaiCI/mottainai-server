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

package task

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	client "github.com/MottainaiCI/mottainai-server/pkg/client"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	citasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	schema "github.com/MottainaiCI/mottainai-server/routes/schema"
	"github.com/MottainaiCI/mottainai-server/routes/schema/v1"

	tablewriter "github.com/olekukonko/tablewriter"
	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
)

func newTaskListCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list [OPTIONS]",
		Short: "List tasks",
		Args:  cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			var v *viper.Viper = config.Viper

			fetcher := client.NewTokenClient(v.GetString("master"), v.GetString("apikey"), config)

			var tlist []citasks.Task
			req := &schema.Request{
				Route:  v1.Schema.GetTaskRoute("show_all"),
				Target: &tlist,
			}
			err = fetcher.Handle(req)
			tools.CheckError(err)

			sort.Slice(tlist[:], func(i, j int) bool {
				return tlist[i].CreatedTime > tlist[j].CreatedTime
			})
			quiet, _ := cmd.Flags().GetBool("quiet")
			jsonOutput, _ := cmd.Flags().GetBool("json")

			if jsonOutput {

				data, err := json.Marshal(tlist)
				if err != nil {
					fmt.Println(fmt.Errorf("Error on convert data to json: %s", err.Error()))
					os.Exit(1)
				}
				fmt.Println(string(data))

			} else {
				if quiet {
					for _, i := range tlist {
						fmt.Println(i.ID)
					}
					return
				}

				var task_table [][]string

				for _, i := range tlist {
					t, _ := time.Parse("20060102150405", i.CreatedTime)
					t2, _ := time.Parse("20060102150405", i.EndTime)
					task_table = append(task_table, []string{i.ID, i.Name, i.Type, i.Status, i.Result, t.String(), t2.String(), i.Source, i.Directory})
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
				table.SetCenterSeparator("|")
				table.SetHeader([]string{"ID", "Name", "Type", "Status", "Result", "Created", "End", "Source", "Dir"})

				for _, v := range task_table {
					table.Append(v)
				}
				table.Render()
			}
		},
	}

	var flags = cmd.Flags()
	flags.BoolP("quiet", "q", false, "Quiet Output")
	flags.BoolP("json", "j", false, "Json output")

	return cmd
}
