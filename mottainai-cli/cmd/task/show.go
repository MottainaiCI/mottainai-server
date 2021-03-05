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
	"log"

	schema "github.com/MottainaiCI/mottainai-server/routes/schema"

	client "github.com/MottainaiCI/mottainai-server/pkg/client"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	citasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"
	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
)

func newTaskShowCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "show <taskid> [OPTIONS]",
		Short: "Show a task",
		Args:  cobra.RangeArgs(1, 1),
		Run: func(cmd *cobra.Command, args []string) {
			var v *viper.Viper = config.Viper

			fetcher := client.NewTokenClient(v.GetString("master"), v.GetString("apikey"), config)

			id := args[0]
			if len(id) == 0 {
				log.Fatalln("You need to define a task id")
			}
			var t citasks.Task

			req := schema.Request{
				Route: v1.Schema.GetTaskRoute("as_json"),
				Options: map[string]interface{}{
					":id": id,
				},
				Target: &t,
			}

			err := fetcher.Handle(req)
			if err != nil {
				panic(err)
			}
			b, err := json.MarshalIndent(t, "", "  ")
			if err != nil {
				fmt.Println("error:", err)
			}
			fmt.Println(string(b))
			//for _, i := range tlist {
			//	fmt.Println(strconv.Itoa(i.ID) + " " + i.Status)
			//}
		},
	}

	return cmd
}
