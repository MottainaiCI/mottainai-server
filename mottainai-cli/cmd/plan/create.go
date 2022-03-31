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
	"io/ioutil"
	"os"

	utils "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/utils"
	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	task "github.com/MottainaiCI/mottainai-server/pkg/tasks"

	cobra "github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newPlanCreateCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create [OPTIONS]",
		Short: "Create a new planning",
		Args:  cobra.OnlyValidArgs,
		// TODO: PreRun check of minimal args if --json is not present
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			var jsonfile string
			var value string
			var p = &task.Plan{}
			dat := make(map[string]interface{})
			var flagsName []string = []string{
				"script", "storage", "source", "directory", "task", "image",
				"namespace", "storage_path", "artefact_path", "tag_namespace",
				"prune", "queue", "cache_image", "planned",
			}

			fetcher, err := utils.CreateClient(config)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			jsonfile, _ = cmd.Flags().GetString("json")
			yamlfile, _ := cmd.Flags().GetString("yaml")

			if jsonfile != "" {
				content, err := ioutil.ReadFile(jsonfile)
				tools.CheckError(err)

				if err := json.Unmarshal(content, &dat); err != nil {
					panic(err)
				}
			} else if yamlfile != "" {
				content, err := ioutil.ReadFile(yamlfile)
				if err != nil {
					panic(err)
				}
				if err := yaml.Unmarshal(content, &p); err != nil {
					panic(err)
				}

				dat = p.ToMap()
			}

			for _, n := range flagsName {
				if cmd.Flag(n).Changed {
					value, err = cmd.Flags().GetString(n)
					tools.CheckError(err)
					dat[n] = value
				}
			}

			res, err := fetcher.PlanCreate(dat)
			tools.CheckError(err)

			tid := res.ID
			if tid == "" {
				tools.PrintResponse(res)
				panic("Failed creating task")
			}

			fmt.Println("-------------------------")
			fmt.Println("Plan " + tid + " has been created")
			fmt.Println("-------------------------")
			fmt.Println("Information: ", tools.BuildCmdArgs(cmd, "plan show "+tid))
			fmt.Println("-------------------------")

		},
	}

	var flags = cmd.Flags()
	flags.String("json", "", "Decode parameters from a JSON file ( e.g. /path/to/file.json )")
	flags.String("yaml", "", "Decode parameters from a YAML file ( e.g. /path/to/file.yaml )")
	flags.String("script", "", "Entrypoint script")
	flags.String("storage", "", "Storage ID")
	flags.StringP("source", "s", "", "Repository url ( e.g. https://github.com/foo/bar.git )")
	flags.StringP("directory", "d", "", "Directory inside repository url ( e.g. /test )")
	flags.StringP("task", "t", "docker_execute", "Task type ( default: docker_execute )")
	flags.StringP("image", "i", "", "Image used from the task ( e.g. my/docker-image:latest")
	flags.StringP("namespace", "n", "", "Specify a namespace the task will be started on")
	flags.StringP("storage_path", "S", "storage", "Specify the storage path in the task")
	flags.StringP("artefact_path", "A", "artefacts", "Specify the artefacts path in the task")
	flags.StringP("tag_namespace", "T", "", "Automatically to the specified namespace on success")
	flags.StringP("prune", "P", "yes", "Perform pruning actions after execution")
	flags.StringP("queue", "q", "", "Queue where to send the task to")
	flags.StringP("cache_image", "C", "yes",
		"Cache image after execution inside the host for later reuse.")
	// TODO: see how permit use of two char "pl"
	flags.String("planned", "", "Plan task creation with cron syntax ( e.g @every 1m )")

	return cmd
}
