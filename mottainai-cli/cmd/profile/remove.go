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

	common "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
)

func newProfileRemoveCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "remove <profile-name> [OPTIONS]",
		Short: "Remove a profile",
		Args:  cobra.RangeArgs(1, 1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			var name, f string
			var conf common.ProfileConf
			var p *common.Profile
			var v *viper.Viper = config.Viper

			name = args[0]

			if v.Get("profiles") == nil {
				// POST: No configuration file found

				fmt.Printf("Profile %s is not present.\n", name)
				return

			} else {
				// POST: A configuration file is already present.

				err = v.Unmarshal(&conf)
				tools.CheckError(err)

				p, _ = conf.GetProfile(name)
				if p == nil {
					fmt.Printf("Profile %s is not present.\n", name)
					return
				}

				p = conf.RemoveProfile(name)
			}

			// Create new viper configuration to avoid
			// write of command line arguments/settings
			viper := viper.New()
			viper.SetConfigType("yaml")
			viper.Set("profiles", conf.Profiles)

			f = v.ConfigFileUsed()

			err = viper.WriteConfigAs(f)
			tools.CheckError(err)

			fmt.Printf("Profile %s with master %s removed correctly.\n",
				name, p.GetMaster())
		},
	}

	return cmd
}
