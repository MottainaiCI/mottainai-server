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

package plan

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	schema "github.com/MottainaiCI/mottainai-server/routes/schema"

	utils "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/utils"
	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	citasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"
	tablewriter "github.com/olekukonko/tablewriter"
	cobra "github.com/spf13/cobra"
)

func newPlanListCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list [OPTIONS]",
		Short: "List plans",
		Args:  cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			var tlist []citasks.Plan
			var task_table [][]string
			var quiet bool

			jsonOutput, _ := cmd.Flags().GetBool("json")
			quiet, _ = cmd.Flags().GetBool("quiet")

			fetcher, err := utils.CreateClient(config)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			req := &schema.Request{
				Route:  v1.Schema.GetTaskRoute("plan_list"),
				Target: &tlist,
			}
			err = fetcher.Handle(req)
			tools.CheckError(err)

			sort.Slice(tlist[:], func(i, j int) bool {
				return tlist[i].CreatedTime > tlist[j].CreatedTime
			})

			tools.CheckError(err)

			if jsonOutput {
				data, _ := json.Marshal(tlist)
				fmt.Println(string(data))
			} else {
				if quiet {
					for _, i := range tlist {
						fmt.Println(i.ID)
					}
					return
				}

				for _, i := range tlist {
					task_table = append(task_table, []string{i.ID, i.Planned, i.Namespace, i.TagNamespace, i.Source, i.Directory})
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.SetBorders(tablewriter.Border{
					Left: true, Top: false, Right: true, Bottom: false,
				})
				table.SetCenterSeparator("|")
				table.SetHeader([]string{
					"ID", "Planned", "From Namespace", "Tag to", "Source", "Dir",
				})

				for _, v := range task_table {
					table.Append(v)
				}
				table.Render()
			}

		},
	}

	var flags = cmd.Flags()
	flags.BoolP("quiet", "q", false, "Quiet Output")
	flags.Bool("json", false, "JSON output")

	return cmd
}
