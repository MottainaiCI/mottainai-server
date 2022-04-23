/*

Copyright (C) 2017-2022  Ettore Di Giacinto <mudler@gentoo.org>
                         Daniele Rondina <geaaru@funtoo.org>

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

package node

import (
	"encoding/json"
	"fmt"
	"os"

	utils "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/utils"
	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	cobra "github.com/spf13/cobra"
)

func newNodeCreateCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create [OPTIONS]",
		Short: "Create a new node",
		Args:  cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			jsonOutput, _ := cmd.Flags().GetBool("json")

			fetcher, err := utils.CreateClient(config)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			resp, err := fetcher.CreateNode()
			tools.CheckError(err)

			if jsonOutput {
				data, _ := json.Marshal(resp)
				fmt.Println(string(data))
			} else {
				tools.PrintResponse(resp)
			}
		},
	}

	var flags = cmd.Flags()
	flags.Bool("json", false, "JSON output")
	return cmd
}
