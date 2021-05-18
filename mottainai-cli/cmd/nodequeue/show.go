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
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	schema "github.com/MottainaiCI/mottainai-server/routes/schema"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"

	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	client "github.com/MottainaiCI/mottainai-server/pkg/client"
	queues "github.com/MottainaiCI/mottainai-server/pkg/queues"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
)

func newNodeQueueShowCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "show <nodequeue-id> [OPTIONS]",
		Short: "Show a node queue data",
		Args:  cobra.RangeArgs(1, 1),
		Run: func(cmd *cobra.Command, args []string) {
			var n queues.NodeQueues
			var v *viper.Viper = config.Viper

			fetcher := client.NewTokenClient(
				v.GetString("master"), v.GetString("apikey"), config,
			)

			id := args[0]
			if len(id) == 0 {
				log.Fatalln("You need to define a node queue id")
			}

			req := &schema.Request{
				Route: v1.Schema.GetNodeQueueRoute("show_byagent"),
				Options: map[string]interface{}{
					":nodeid": id,
				},
				Target: &n,
			}
			msg := map[string]interface{}{
				"akey":   "j1ZvSpC1405KjYfkCB6Uv1AKj2idwO",
				"nodeid": id,
			}
			b, _ := json.Marshal(msg)

			req.Body = bytes.NewBuffer(b)

			err := fetcher.Handle(req)
			fmt.Println("RES ", string(req.ResponseRaw))
			tools.CheckError(err)
			/*
				req := &schema.Request{
					Route: v1.Schema.GetNodeQueueRoute("show"),
					Options: map[string]interface{}{
						":id": id,
					},
					Target: &n,
				}
				err := fetcher.Handle(req)
				tools.CheckError(err)

				b, err := json.MarshalIndent(n, "", "  ")
				if err != nil {
					log.Fatalln("error:", err)
				}
				fmt.Println(string(b))
			*/
		},
	}

	return cmd
}