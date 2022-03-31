/*

Copyright (C) 2021-2022  Daniele Rondina <geaaru@sabayonlinux.org>

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
	"log"
	"os"

	utils "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/utils"
	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"

	cobra "github.com/spf13/cobra"
)

func newNodeQueueDeleteByIdCommand(config *setting.Config) *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "delete_byid <id> [OPTIONS]",
		Short: "delete a node queue by id (dev only)",
		Args:  cobra.RangeArgs(1, 1),
		Run: func(cmd *cobra.Command, args []string) {
			fetcher, err := utils.CreateClient(config)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			id := args[0]
			if len(id) == 0 {
				log.Fatalln("You need to defina a node id")
			}
			resp, err := fetcher.NodeQueueDelById(id)

			if err != nil {
				if resp.Request != nil && resp.Request.Response != nil {
					fmt.Println("ERROR: ", resp.Request.Response.StatusCode)
					fmt.Println(string(resp.Request.ResponseRaw))
					os.Exit(1)
				} else {
					tools.CheckError(err)
				}
			}

			tools.PrintResponse(resp)
		},
	}

	return cmd
}
