/*
Copyright (C) 2021 Daniele Rondina <geaaru@sabayonlinux.org>

Credits goes also to Gogs authors, some code portions and re-implemented design
are also coming from the Gogs project, which is using the go-macaron framework
and was really source of ispiration. Kudos to them!

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

	s "github.com/MottainaiCI/mottainai-server/mottainai-scheduler/pkg/config"
	core "github.com/MottainaiCI/mottainai-server/mottainai-scheduler/pkg/core"

	cobra "github.com/spf13/cobra"
)

func newDaemonCommand(config *s.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "daemon",
		Short: "Start scheduler daemon.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			m := core.NewScheduler()
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
