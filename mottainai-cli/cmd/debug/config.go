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

package debug

import (
	"fmt"

	common "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	cobra "github.com/spf13/cobra"
)

func title(t string) string {
	return fmt.Sprintf("%s\n%s\n%s",
		"=========================================",
		t,
		"=========================================",
	)
}

func newDebugConfigCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "config",
		Short: "Show configuration params",
		Args:  cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(common.MCLI_HEADER)
			fmt.Println("")
			fmt.Println(fmt.Sprintf("%s\n%s\n",
				title("General Options:"), config.General.String()))

			fmt.Println(title("Mottainai Server Options:"))
			fmt.Println(fmt.Sprintf("profile: %s", config.Viper.GetString("profile")))
			fmt.Println(fmt.Sprintf("master:  %s", config.Viper.GetString("master")))
			if config.Viper.GetBool("all") {
				fmt.Println(fmt.Sprintf("\n%s\n%s\n",
					title("Mottainai Agent Options:"), config.Agent.String()))
			}
		},
	}

	var flags = cmd.Flags()

	flags.BoolP("all", "a", false, "Additional options for agent mode.")
	config.Viper.BindPFlag("all", flags.Lookup("all"))

	return cmd
}
