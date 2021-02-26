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

package cmd

import (
	"fmt"
	"os"

	"github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	s "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/spf13/cobra"
)

func newAgentCommand(config *s.Config) *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "agent [OPTIONS]",
		Short: "Start agent",
		Args:  cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {

			var err error
			m := mottainai.NewAgent()
			m.Map(config)
			err = m.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	return cmd
}
