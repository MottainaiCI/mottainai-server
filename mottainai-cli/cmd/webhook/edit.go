/*

Copyright (C) 2017-2021  Ettore Di Giacinto <mudler@gentoo.org>
                         Daniele Rondina <geaaru@sabayonlinux.org>

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

package webhook

import (
	"fmt"
	"log"
	"os"

	utils "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/utils"
	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"

	cobra "github.com/spf13/cobra"
)

func newWebHookEditCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "edit <id> <key> <value>",
		Short: "Edit a webhook",
		//Args:  cobra.OnlyValidArgs,

		Args: cobra.RangeArgs(3, 3),
		// TODO: PreRun check of minimal args if --json is not present
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			id := args[0]
			key := args[1]
			value := args[2]

			dat := make(map[string]interface{})

			if len(args) != 3 {
				log.Fatalln("You need to define a webhook id and a key and a value to update")
			}
			dat["key"] = key
			dat["value"] = value
			dat["id"] = id

			fetcher, err := utils.CreateClient(config)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			res, err := fetcher.WebHookEdit(dat)
			tools.CheckError(err)
			tools.PrintResponse(res)
		},
	}

	return cmd
}
