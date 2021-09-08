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

	utils "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/utils"
	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"

	cobra "github.com/spf13/cobra"
)

func newNodeQueueCreateCommand(config *setting.Config) *cobra.Command {
	var queues []string

	var cmd = &cobra.Command{
		Use:   "create [OPTIONS]",
		Short: "Create a new node queue (dev only)",
		Args:  cobra.OnlyValidArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			akey, _ := cmd.Flags().GetString("agent-key")
			nid, _ := cmd.Flags().GetString("node-id")

			if akey == "" {
				fmt.Println("Missing agent-key field")
				os.Exit(1)
			}

			if nid == "" {
				fmt.Println("Missing node-id field")
				os.Exit(1)
			}

			if len(queues) == 0 {
				fmt.Println("At least one queue is required")
				os.Exit(1)
			}

		},
		Run: func(cmd *cobra.Command, args []string) {
			fetcher, err := utils.CreateClient(config)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			akey, _ := cmd.Flags().GetString("agent-key")
			nid, _ := cmd.Flags().GetString("node-id")

			// Create empty queues map
			mq := make(map[string][]string, 0)
			for _, q := range queues {
				mq[q] = []string{}
			}

			resp, err := fetcher.NodeQueueCreate(akey, nid, mq)

			tools.CheckError(err)
			tools.PrintResponse(resp)
		},
	}

	var flags = cmd.Flags()
	flags.StringSliceVarP(&queues, "queue", "q", []string{},
		"Define the queue of the node.")
	flags.String("agent-key", "", "Agent Key of the node")
	flags.String("node-id", "", "NodeID of the node")

	return cmd
}
