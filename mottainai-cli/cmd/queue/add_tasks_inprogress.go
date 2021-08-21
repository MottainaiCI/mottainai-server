/*

Copyright (C) 2021  Daniele Rondina <geaaru@sabayonlinux.org>

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

package queue

import (
	"fmt"
	"log"
	"os"

	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	client "github.com/MottainaiCI/mottainai-server/pkg/client"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	schema "github.com/MottainaiCI/mottainai-server/routes/schema"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"

	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
)

func newQueueAddTaskInProgressCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "add-task-in-progress [OPTIONS]",
		Short: "Add task in progress to a queue (dev only)",
		Args:  cobra.OnlyValidArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			qid, _ := cmd.Flags().GetString("qid")
			tid, _ := cmd.Flags().GetString("taskid")

			if qid == "" {
				fmt.Println("Missing queue id field")
				os.Exit(1)
			}
			if tid == "" {
				fmt.Println("Missing task id field")
				os.Exit(1)
			}

		},
		Run: func(cmd *cobra.Command, args []string) {
			var v *viper.Viper = config.Viper

			fetcher := client.NewTokenClient(
				v.GetString("master"), v.GetString("apikey"), config,
			)

			qid, _ := cmd.Flags().GetString("qid")
			tid, _ := cmd.Flags().GetString("taskid")

			req := &schema.Request{
				Route: v1.Schema.GetQueueRoute("add_task_in_progress"),
				Options: map[string]interface{}{
					":qid": qid,
					":tid": tid,
				},
			}

			resp, err := fetcher.HandleAPIResponse(req)
			if err != nil {

				if req.Response != nil {
					fmt.Println("ERROR: ", req.Response.StatusCode)
					fmt.Println(string(req.ResponseRaw))
					os.Exit(1)
				}

				log.Fatalln("error:", err)
			}
			tools.PrintResponse(resp)
		},
	}

	var flags = cmd.Flags()
	flags.String("qid", "", "Queue ID")
	flags.StringP("taskid", "t", "", "Task ID")

	return cmd
}
