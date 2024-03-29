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

package plan

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	schema "github.com/MottainaiCI/mottainai-server/routes/schema"

	utils "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/utils"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	citasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	"github.com/MottainaiCI/mottainai-server/routes/schema/v1"
	cobra "github.com/spf13/cobra"
)

func newPlanShowCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "show <paln-id> [OPTIONS]",
		Short: "Show a plan",
		Args:  cobra.RangeArgs(1, 1),
		Run: func(cmd *cobra.Command, args []string) {
			var t citasks.Plan

			id := args[0]
			if len(id) == 0 {
				log.Fatalln("You need to define a plan id")
			}

			fetcher, err := utils.CreateClient(config)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			req := &schema.Request{
				Route: v1.Schema.GetTaskRoute("plan_show"),
				Options: map[string]interface{}{
					":id": id,
				},
				Target: &t,
			}

			err = fetcher.Handle(req)
			if err != nil {
				log.Fatalln("error:", err)
			}
			b, err := json.MarshalIndent(t, "", "  ")
			if err != nil {
				log.Fatalln("error:", err)
			}
			fmt.Println(string(b))
		},
	}

	return cmd
}
