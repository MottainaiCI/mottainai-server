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

	utils "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/utils"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	citasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	schema "github.com/MottainaiCI/mottainai-server/routes/schema"
	"github.com/MottainaiCI/mottainai-server/routes/schema/v1"

	tablewriter "github.com/olekukonko/tablewriter"
	cobra "github.com/spf13/cobra"
)

type TaskListFiltered struct {
	Total int            `json:"total"`
	Tasks []citasks.Task `json:"tasks"`
}

func newTaskListCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list [OPTIONS]",
		Short: "List tasks",
		Args:  cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			var tlist TaskListFiltered

			fetcher, err := utils.CreateClient(config)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			status, _ := cmd.Flags().GetString("status")
			name, _ := cmd.Flags().GetString("name")
			id, _ := cmd.Flags().GetString("id")
			image, _ := cmd.Flags().GetString("image")
			result, _ := cmd.Flags().GetString("result")
			pageSize, _ := cmd.Flags().GetInt32("page-size")

			options := make(map[string]interface{}, 0)

			options["pageSize"] = fmt.Sprintf("%d", pageSize)

			if status != "" {
				options["status"] = status
			}

			if result != "" {
				options["result"] = result
			}

			if name != "" {
				options["name"] = name
			}

			if id != "" {
				options["id"] = id
			}

			if image != "" {
				options["image"] = image
			}

			req := &schema.Request{
				Route:   v1.Schema.GetTaskRoute("show_all_filtered"),
				Target:  &tlist,
				Options: options,
			}
			err = fetcher.Handle(req)
			if err != nil {
				fmt.Println("Error: \n" + string(req.ResponseRaw))
				os.Exit(1)
			}

			sort.Slice(tlist.Tasks[:], func(i, j int) bool {
				return tlist.Tasks[i].CreatedTime > tlist.Tasks[j].CreatedTime
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
					for _, i := range tlist.Tasks {
						fmt.Println(i.ID)
					}
					return
				}

				var task_table [][]string

				for _, i := range tlist.Tasks {
					t, _ := time.Parse("20060102150405", i.CreatedTime)
					t2, _ := time.Parse("20060102150405", i.EndTime)
					task_table = append(task_table, []string{i.ID, i.Name, i.Type, i.Status, i.Result, t.String(), t2.String(), i.Image, i.Owner})
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.SetFooterAlignment(tablewriter.ALIGN_LEFT)
				table.SetHeader([]string{
					"ID", "Name", "Type", "Status", "Result", "Created", "End", "Image", "Owner",
				})
				table.SetFooter([]string{
					"Total Tasks", "", "", "", "", "", "", "", fmt.Sprintf("%d", tlist.Total),
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
	flags.BoolP("json", "j", false, "Json output")
	flags.String("status", "", "Filter tasks of the specificied status.")
	flags.String("name", "", "Filter tasks with name matching the value.")
	flags.String("id", "", "Filter tasks with id matching the value.")
	flags.String("image", "", "Filter tasks with image matching the value.")
	flags.Int32("page-size", 100, "Set page size. Max page size is based on server side config.")
	flags.String("result", "", "Filter tasks with the specifiied result.")

	return cmd
}
