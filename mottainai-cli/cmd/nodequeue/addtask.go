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
	"fmt"
	"os"

	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	client "github.com/MottainaiCI/mottainai-server/pkg/client"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"

	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
)

func newNodeQueueAddTaskCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "add-task [OPTIONS]",
		Short: "Add a task to a node queue (dev only)",
		Args:  cobra.OnlyValidArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			akey, _ := cmd.Flags().GetString("agent-key")
			nid, _ := cmd.Flags().GetString("node-id")
			tid, _ := cmd.Flags().GetString("tid")
			queue, _ := cmd.Flags().GetString("queue")

			if akey == "" {
				fmt.Println("Missing agent-key field")
				os.Exit(1)
			}

			if nid == "" {
				fmt.Println("Missing node-id field")
				os.Exit(1)
			}

			if queue == "" {
				fmt.Println("Missing queue field")
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
				v.GetString("master"),
				v.GetString("apikey"),
				config,
			)

			akey, _ := cmd.Flags().GetString("agent-key")
			nid, _ := cmd.Flags().GetString("node-id")
			tid, _ := cmd.Flags().GetString("tid")
			queue, _ := cmd.Flags().GetString("queue")

			resp, err := fetcher.NodeQueueAddTask(akey, nid, queue, tid)

			tools.CheckError(err)
			tools.PrintResponse(resp)
		},
	}

	var flags = cmd.Flags()
	flags.String("agent-key", "", "Agent Key of the node")
	flags.String("node-id", "", "NodeID of the node")
	flags.StringP("queue", "q", "", "Queue name")
	flags.StringP("tid", "t", "", "Task ID")

	return cmd
}
