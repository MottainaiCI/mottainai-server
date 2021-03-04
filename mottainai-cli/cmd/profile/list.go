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

package profile

import (
	"fmt"
	"os"

	common "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	tablewriter "github.com/olekukonko/tablewriter"
	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
)

func newProfileListCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list [OPTIONS]",
		Short: "List profiles",
		Args:  cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			var v *viper.Viper = config.Viper

			if v.Get("profiles") == nil {
				fmt.Println("No profiles available.")
				return
			}

			var conf common.ProfileConf
			err = v.Unmarshal(&conf)
			tools.CheckError(err)

			if conf.Profiles == nil {
				fmt.Println("No profiles available.")
				return
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetBorders(tablewriter.Border{
				Left:   true,
				Top:    true,
				Right:  true,
				Bottom: true})
			table.SetCenterSeparator("|")
			table.SetHeader([]string{"Name", "Master URL", "ApiKey"})

			for k, val := range conf.Profiles {
				table.Append([]string{k, val.GetMaster(), val.GetApiKey()})
			}

			table.Render()
		},
	}

	return cmd
}
