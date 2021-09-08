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

package namespace

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	schema "github.com/MottainaiCI/mottainai-server/routes/schema"

	utils "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/utils"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"
	cobra "github.com/spf13/cobra"
)

func newNamespaceShowCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "show <namespace> [OPTIONS]",
		Short: "Show artefacts belonging to namespace",
		Args:  cobra.RangeArgs(1, 1),
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				log.Fatalln("You need to define a namespace name")
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			var tlist []string

			jsonOutput, _ := cmd.Flags().GetBool("json")

			fetcher, err := utils.CreateClient(config)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			ns := args[0]
			req := &schema.Request{
				Route:  v1.Schema.GetNamespaceRoute("show_artefacts"),
				Target: &tlist,
				Options: map[string]interface{}{
					":name": ns,
				},
			}
			err = fetcher.Handle(req)
			if err != nil {
				if req.Response != nil {
					fmt.Println("ERROR: ", req.Response.StatusCode)
					fmt.Println(string(req.ResponseRaw))
					os.Exit(1)
				}

				log.Fatalln("error:", err)
			}

			if jsonOutput {
				data, _ := json.Marshal(tlist)
				fmt.Println(string(data))
			} else {

				for _, i := range tlist {
					fmt.Println("- " + i)
				}
			}
		},
	}

	var flags = cmd.Flags()
	flags.Bool("json", false, "JSON output")

	return cmd
}
