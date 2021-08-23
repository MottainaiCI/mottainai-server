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

package storage

import (
	"fmt"
	"log"
	"os"

	utils "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/utils"
	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"

	cobra "github.com/spf13/cobra"
)

func newStorageRemoveCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "remove <storageid> <absolute_path> [OPTIONS]",
		Short: "Remove a given path from a storage",
		Args:  cobra.RangeArgs(2, 2),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			fetcher, err := utils.CreateClient(config)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			st := args[0]
			path := args[1]
			if len(st) == 0 || len(path) == 0 {
				log.Fatalln("You need to define a storage id and a path to delete")
			}

			res, err := fetcher.StorageRemovePath(st, path)
			tools.CheckError(err)
			tools.PrintResponse(res)
		},
	}

	return cmd
}
