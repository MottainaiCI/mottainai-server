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

package task

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	utils "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/utils"
	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	citasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	schema "github.com/MottainaiCI/mottainai-server/routes/schema"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"

	cobra "github.com/spf13/cobra"
)

func newTaskAttachCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "attach <taskid> [OPTIONS]",
		Short: "attach to a task output",
		Args:  cobra.RangeArgs(1, 1),
		Run: func(cmd *cobra.Command, args []string) {
			fetcher, err := utils.CreateClient(config)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			id := args[0]
			if len(id) == 0 {
				log.Fatalln("You need to define a task id")
			}
			var pos = 0

			for {
				time.Sleep(time.Second + 2)
				var t citasks.Task

				var err error
				req := &schema.Request{
					Route: v1.Schema.GetTaskRoute("as_json"),
					Options: map[string]interface{}{
						":id": id,
					},
				}
				err = fetcher.HandleRaw(req, func(b io.ReadCloser) error {
					return json.NewDecoder(b).Decode(&t)
				})
				tools.CheckError(err)

				if t.Status != "running" {
					if t.Status == "done" && pos == 0 {
						buff, err := fetcher.TaskLog(id)
						tools.CheckError(err)
						tools.PrintBuff(buff)
					} else {
						if t.Status == "none" || t.Status == "waiting" {
							// Sleeping...
							fmt.Println("Task not running... sleeping.")
							continue
						} else {
							fmt.Println("Build status: " + t.Status + " Can't attach to any live stream.")
						}
					}
					return
				}

				buff, err := fetcher.TaskStream(id, strconv.Itoa(pos))
				tools.CheckError(err)
				pos += len(buff)
				tools.PrintBuff(buff)
			}

		},
	}

	return cmd
}
