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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	utils "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/utils"
	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	event "github.com/MottainaiCI/mottainai-server/pkg/event"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	task "github.com/MottainaiCI/mottainai-server/pkg/tasks"

	cobra "github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newWebHookUpdateCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "update <webhook> (task|pipeline) [OPTIONS]",
		Short: "Update a webhook",
		//Args:  cobra.OnlyValidArgs,

		Args: cobra.RangeArgs(2, 2),
		// TODO: PreRun check of minimal args if --json is not present
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			var res event.APIResponse

			fetcher, err := utils.CreateClient(config)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			id := args[0]
			mytype := args[1]

			if len(id) == 0 {
				log.Fatalln("You need to define a webhook id")
			}
			if mytype != "task" && mytype != "pipeline" {
				log.Fatalln("You can delete a task or a pipeline associated to a webhook")
			}
			dat := make(map[string]interface{})

			jsonfile, err := cmd.Flags().GetString("json")
			tools.CheckError(err)
			yamlfile, err := cmd.Flags().GetString("yaml")
			tools.CheckError(err)
			switch mytype {
			case "pipeline":
				t := &task.Pipeline{}
				if jsonfile != "" {
					content, err := ioutil.ReadFile(jsonfile)
					tools.CheckError(err)

					if err := json.Unmarshal(content, &t); err != nil {
						panic(err)
					}
					dat = t.ToMap(false)
				} else if yamlfile != "" {
					content, err := ioutil.ReadFile(yamlfile)
					if err != nil {
						panic(err)
					}
					if err := yaml.Unmarshal(content, &t); err != nil {
						panic(err)
					}
					dat = t.ToMap(false)
				}

				res, err = fetcher.WebHookPipelineUpdate(id, dat)
			case "task":
				t := &task.Task{}

				if jsonfile != "" {
					content, err := ioutil.ReadFile(jsonfile)
					tools.CheckError(err)

					if err := json.Unmarshal(content, &t); err != nil {
						panic(err)
					}
					dat = t.ToMap()
				} else if yamlfile != "" {
					content, err := ioutil.ReadFile(yamlfile)
					if err != nil {
						panic(err)
					}
					if err := yaml.Unmarshal(content, &t); err != nil {
						panic(err)
					}
					dat = t.ToMap()
				}

				res, err = fetcher.WebHookTaskUpdate(id, dat)
			}

			tools.CheckError(err)

			tools.PrintResponse(res)
			if len(res.Error) > 0 {
				os.Exit(1)
			}
		},
	}

	var flags = cmd.Flags()
	flags.String("json", "", "Decode parameters from a JSON file ( e.g. /path/to/file.json )")
	flags.String("yaml", "", "Decode parameters from a YAML file ( e.g. /path/to/file.yaml )")

	return cmd
}
