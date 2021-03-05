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
	"io/ioutil"

	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	client "github.com/MottainaiCI/mottainai-server/pkg/client"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	task "github.com/MottainaiCI/mottainai-server/pkg/tasks"

	"github.com/ghodss/yaml"
	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
)

func newTaskCreateCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create [OPTIONS]",
		Short: "Create a new task",
		Args:  cobra.OnlyValidArgs,
		// TODO: PreRun check of minimal args if --json is not present
		Run: func(cmd *cobra.Command, args []string) {

			var v *viper.Viper = config.Viper

			fetcher := client.NewTokenClient(v.GetString("master"), v.GetString("apikey"), config)
			to, _ := cmd.Flags().GetString("to")
			dat := make(map[string]interface{})
			t := &task.Task{}

			jsonfile, err := cmd.Flags().GetString("json")
			tools.CheckError(err)
			yamlfile, err := cmd.Flags().GetString("yaml")
			tools.CheckError(err)

			if jsonfile != "" {
				content, err := ioutil.ReadFile(jsonfile)
				if err != nil {
					panic(err)
				}
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

			var value string
			var flagsName []string = []string{
				"name", "script", "storage", "source", "directory", "type", "image",
				"namespace", "storage_path", "artefact_path", "tag_namespace",
				"prune", "queue", "cache_image",
			}

			for _, n := range flagsName {
				if cmd.Flag(n).Changed {
					value, err = cmd.Flags().GetString(n)
					tools.CheckError(err)
					dat[n] = value
				}
			}
			var created = make(map[string]bool)
			if len(to) > 0 {
				created = GenerateTasks(fetcher, dat, to)
			} else {
				res, err := fetcher.CreateTask(dat)
				tools.CheckError(err)

				tid := res.ID
				if tid == "" {
					tools.PrintResponse(res)
					panic("Failed creating task")
				}
				created[tid] = false

				fmt.Println("-------------------------")
				fmt.Println("Task " + tid + " has been created")
				fmt.Println("-------------------------")
				fmt.Println("Live log: ", tools.BuildCmdArgs(cmd, "task attach "+tid))
				fmt.Println("Information: ", tools.BuildCmdArgs(cmd, "task show "+tid))
				fmt.Println("URL:", " "+fetcher.GetBaseURL()+"/tasks/display/"+tid)
				fmt.Println("Build Log:", " "+fetcher.GetBaseURL()+"/artefact/"+tid+"/build_"+tid+".log")
				fmt.Println("-------------------------")
			}
			if monitor, err := cmd.Flags().GetBool("monitor"); err == nil && monitor {
				fmt.Println("Monitoring task state")
				MonitorTasks(fetcher, created)
			}

		},
	}

	var flags = cmd.Flags()
	flags.String("json", "", "Decode parameters from a JSON file ( e.g. /path/to/file.json )")
	flags.String("yaml", "", "Decode parameters from a YAML file ( e.g. /path/to/file.yaml )")
	flags.String("script", "", "Entrypoint script")
	flags.String("storage", "", "Storage ID")
	flags.StringP("source", "s", "", "Repository url ( e.g. https://github.com/foo/bar.git )")
	flags.StringP("directory", "d", "", "Directory inside repository url ( e.g. /test )")
	flags.StringP("type", "t", "docker_execute", "Task type ( default: docker_execute )")
	flags.StringP("name", "", "my_task", "Task Name ( default: empty )")
	flags.StringP("image", "i", "", "Image used from the task ( e.g. my/docker-image:latest")
	flags.StringP("namespace", "n", "", "Specify a namespace the task will be started on")
	flags.StringP("storage_path", "S", "storage", "Specify the storage path in the task")
	flags.StringP("artefact_path", "A", "artefacts", "Specify the artefacts path in the task")
	flags.StringP("tag_namespace", "T", "", "Automatically to the specified namespace on success")
	flags.StringP("prune", "P", "yes", "Perform pruning actions after execution")
	flags.StringP("queue", "q", "", "Queue where to send the task to")
	flags.String("to", "", "Regex match pattern for nodes, it will create a task for each one")
	flags.Bool("monitor", false, "Monitor task after creation (returns same exit status as task)")

	flags.StringP("cache_image", "C", "yes",
		"Cache image after execution inside the host for later reuse.")

	return cmd
}
