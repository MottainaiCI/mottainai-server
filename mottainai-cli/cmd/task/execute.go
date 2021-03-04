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
	"io"
	"log"

	schema "github.com/MottainaiCI/mottainai-server/routes/schema"

	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	client "github.com/MottainaiCI/mottainai-server/pkg/client"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	citasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	manager "github.com/MottainaiCI/mottainai-server/pkg/tasks/manager"

	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"
	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
)

func newTaskExecuteCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "execute <taskid> [OPTIONS]",
		Short: "execute task",
		Args:  cobra.RangeArgs(1, 1),
		Run: func(cmd *cobra.Command, args []string) {
			var v *viper.Viper = config.Viper

			fetcher := client.NewTokenClient(v.GetString("master"), v.GetString("apikey"), config)
			config.GetAgent().ApiKey = v.GetString("apikey")
			fetcher.SetActiveReport(true)
			id := args[0]
			if len(id) == 0 {
				log.Fatalln("You need to define a task id")
			}

			var t citasks.Task

			var err error
			req := schema.Request{
				Route: v1.Schema.GetTaskRoute("as_json"),
				Options: map[string]interface{}{
					":id": id,
				},
			}
			err = fetcher.HandleRaw(req, func(b io.ReadCloser) error {
				return json.NewDecoder(b).Decode(&t)
			})
			tools.CheckError(err)
			var fn func(string) (int, error)

			config.GetWeb().AppURL = v.GetString("master")
			fn = manager.DefaultTaskHandler(config).Handler(t.Type)
			fn(id)
		},
	}

	return cmd
}
