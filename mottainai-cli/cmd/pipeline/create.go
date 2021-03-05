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

package pipeline

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

func newPipelineCreateCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create [OPTIONS]",
		Short: "Create a new pipeline",
		Args:  cobra.OnlyValidArgs,
		// TODO: PreRun check of minimal args if --json is not present
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			var jsonfile string
			var v *viper.Viper = config.Viper

			fetcher := client.NewTokenClient(v.GetString("master"), v.GetString("apikey"), config)
			var p = &task.Pipeline{}
			dat := make(map[string]interface{})

			jsonfile, err = cmd.Flags().GetString("json")
			tools.CheckError(err)
			yamlfile, err := cmd.Flags().GetString("yaml")
			tools.CheckError(err)

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
				dat = p.ToMap(false)
			}

			res, err := fetcher.PipelineCreate(dat)
			tools.CheckError(err)

			tid := res.ID
			if tid == "" {
				tools.PrintResponse(res)
				panic("Failed creating task")
			}

			fmt.Println("-------------------------")
			fmt.Println("Pipeline " + tid + " has been created")
			fmt.Println("-------------------------")
			fmt.Println("Information: ", tools.BuildCmdArgs(cmd, "pipeline show "+tid))
			fmt.Println("-------------------------")
		},
	}

	var flags = cmd.Flags()
	flags.String("json", "", "Decode parameters from a JSON file ( e.g. /path/to/file.json )")
	flags.String("yaml", "", "Decode parameters from a YAML file ( e.g. /path/to/file.yaml )")

	return cmd
}
