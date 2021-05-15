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

package nodequeue

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	schema "github.com/MottainaiCI/mottainai-server/routes/schema"

	client "github.com/MottainaiCI/mottainai-server/pkg/client"
	queues "github.com/MottainaiCI/mottainai-server/pkg/queues"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"
	tablewriter "github.com/olekukonko/tablewriter"
	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
)

func newNodeQueueListCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list [OPTIONS]",
		Short: "List node queues",
		Args:  cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			var n []queues.NodeQueues
			var node_table [][]string
			var v *viper.Viper = config.Viper

			jsonOutput, _ := cmd.Flags().GetBool("json")

			fetcher := client.NewTokenClient(
				v.GetString("master"), v.GetString("apikey"), config,
			)

			req := &schema.Request{
				Route:  v1.Schema.GetNodeQueueRoute("show_all"),
				Target: &n,
			}

			err := fetcher.Handle(req)
			if err != nil {

				if req.Response != nil {
					fmt.Println("ERROR: ", req.Response.StatusCode)
					fmt.Println(string(req.ResponseRaw))
					os.Exit(1)
				}

				log.Fatalln("error:", err)
			}

			if jsonOutput {
				data, _ := json.Marshal(n)
				fmt.Println(string(data))
			} else {
				for _, i := range n {

					nq := []string{}
					for k, _ := range i.Queues {
						nq = append(nq, k)
					}

					node_table = append(node_table,
						[]string{
							i.ID,
							i.AgentKey,
							i.NodeId,
							fmt.Sprintf("%d", len(nq)),
							i.CreationDate,
						},
					)
				}

				table := tablewriter.NewWriter(os.Stdout)
				table.SetBorders(
					tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false},
				)
				table.SetCenterSeparator("|")
				table.SetHeader([]string{"ID", "Agent Key", "NodeId", "# Queues", "Creation Date"})

				for _, v := range node_table {
					table.Append(v)
				}
				table.Render()
			}
		},
	}

	var flags = cmd.Flags()
	flags.Bool("json", false, "JSON output")

	return cmd
}
