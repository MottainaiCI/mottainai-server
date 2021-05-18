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

	client "github.com/MottainaiCI/mottainai-server/pkg/client"
	queues "github.com/MottainaiCI/mottainai-server/pkg/queues"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"
	tablewriter "github.com/olekukonko/tablewriter"
	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
)

func newQueueListCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list [OPTIONS]",
		Short: "List queues",
		Args:  cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			var n []queues.Queue
			var queues_table [][]string
			var v *viper.Viper = config.Viper

			fetcher := client.NewTokenClient(
				v.GetString("master"), v.GetString("apikey"), config,
			)

			req := &schema.Request{
				Route:  v1.Schema.GetQueueRoute("show_all"),
				Target: &n,
			}
			err := fetcher.Handle(req)
			if err != nil {
				log.Fatalln("error:", err)
			}

			for _, i := range n {
				queues_table = append(queues_table,
					[]string{
						i.Qid, i.Name,
						fmt.Sprintf("%d", len(i.Waiting)),
						fmt.Sprintf("%d", len(i.InProgress)),
						i.CreationDate, i.UpdateDate,
					},
				)
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
			table.SetCenterSeparator("|")
			table.SetHeader([]string{
				"Queue Id", "Queue Name", "# Waiting Tasks",
				"# In Progress Tasks", "Creation Date",
				"Update Date",
			})

			for _, v := range queues_table {
				table.Append(v)
			}
			table.Render()
		},
	}

	return cmd
}
