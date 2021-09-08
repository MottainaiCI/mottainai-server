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

	utils "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/utils"
	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	schema "github.com/MottainaiCI/mottainai-server/routes/schema"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"

	cobra "github.com/spf13/cobra"
)

func newQueueDelPipelineInWaitingCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "del-pipeline [OPTIONS]",
		Short: "Del pipeline in waiting to a queue (dev only)",
		Args:  cobra.OnlyValidArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			qid, _ := cmd.Flags().GetString("qid")
			pid, _ := cmd.Flags().GetString("pipelineid")

			if qid == "" {
				fmt.Println("Missing queue id field")
				os.Exit(1)
			}
			if pid == "" {
				fmt.Println("Missing pipeline id field")
				os.Exit(1)
			}

		},
		Run: func(cmd *cobra.Command, args []string) {
			fetcher, err := utils.CreateClient(config)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			qid, _ := cmd.Flags().GetString("qid")
			pid, _ := cmd.Flags().GetString("pipelineid")

			req := &schema.Request{
				Route: v1.Schema.GetQueueRoute("del_pipeline"),
				Options: map[string]interface{}{
					":qid": qid,
					":pid": pid,
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
	flags.StringP("pipelineid", "t", "", "Pipeline ID")

	return cmd
}
