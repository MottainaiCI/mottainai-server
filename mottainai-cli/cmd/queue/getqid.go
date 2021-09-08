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

	schema "github.com/MottainaiCI/mottainai-server/routes/schema"

	utils "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/utils"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"
	cobra "github.com/spf13/cobra"
)

func newQueueGetQidCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "get-qid <queue-name> [OPTIONS]",
		Short: "Retrieve Queue Id by queue name (dev only)",
		Args:  cobra.OnlyValidArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("Missing queue name")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			var qid string

			fetcher, err := utils.CreateClient(config)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			req := &schema.Request{
				Route: v1.Schema.GetQueueRoute("get_qid"),
				Options: map[string]interface{}{
					":name": args[0],
				},
				Target: &qid,
			}
			err = fetcher.Handle(req)
			if err != nil {
				if req.Response != nil {
					fmt.Println("ERROR: ", req.Response.StatusCode)
					fmt.Println(string(req.ResponseRaw))
				} else {
					log.Fatalln("error:", err)
				}
				os.Exit(1)
			}

			fmt.Println(qid)
		},
	}

	return cmd
}
