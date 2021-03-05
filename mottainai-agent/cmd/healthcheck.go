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
	"github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	s "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/spf13/cobra"
)

func newHealthcheckCommand(config *s.Config) *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "health [OPTIONS]",
		Short: "Start HealthCheck service",
		Args:  cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {

			var err error
			var oneshot, force bool

			oneshot, err = cmd.Flags().GetBool("oneshot")
			if err != nil {
				panic(err)
			}
			force, err = cmd.Flags().GetBool("force")
			if err != nil {
				panic(err)
			}

			m := mottainai.NewAgent()
			m.Map(config)
			if oneshot {
				m.HealthCheckSetup(force)
				if force || !m.AgentIsBusy() {
					m.HealthClean(force)
				}
			} else {
				m.HealthCheckRun(force)
			}
		},
	}

	var flags = cmd.Flags()
	flags.BoolP("oneshot", "o", false, "Execute once")
	flags.BoolP("force", "f", false, "Do not check for agent state (oneshot)")

	return cmd
}
