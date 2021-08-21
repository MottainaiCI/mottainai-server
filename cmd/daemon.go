/*

Copyright (C) 2017-2018  Ettore Di Giacinto <mudler@gentoo.org>
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

	"github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	"github.com/MottainaiCI/mottainai-server/routes"

	s "github.com/MottainaiCI/mottainai-server/pkg/settings"
	cobra "github.com/spf13/cobra"
)

func newDaemonCommand(config *s.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "daemon",
		Short: "Start api daemon",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			setupAdmin, _ := cmd.Flags().GetBool("setup-admin-user")

			m := mottainai.Classic(config)
			routes.SetupDaemon(m)

			if setupAdmin {
				err := m.SetupAdminUser()
				if err != nil {
					fmt.Println("Error on setup admin user: " + err.Error())
					os.Exit(1)
				}
			}
			m.Start()
		},
	}
	flags := cmd.Flags()
	flags.Bool("setup-admin-user", false, `Automatically create the admin user if no users are present.

Use variables:
MOTTAINAI_ADMIN_USER
MOTTAINAI_ADMIN_PASS
MOTTAINAI_ADMIN_EMAIL
`)

	return cmd
}
