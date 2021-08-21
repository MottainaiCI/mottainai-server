/*
Copyright (C) 2021  Daniele Rondina <geaaru@sabayonlinux.org>

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

func newRunCommand(config *s.Config) *cobra.Command {

	var cmd = &cobra.Command{
		Use:     "run [OPTIONS]",
		Aliases: []string{"r"},
		Short:   "Importer Mottainai Database",
		Args:    cobra.OnlyValidArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			targetDir, _ := cmd.Flags().GetString("target-dir")
			if targetDir == "" {
				fmt.Println("Invalid target-dir parameter.")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			targetDir, _ := cmd.Flags().GetString("target-dir")

			m := mottainai.NewImporter(config)
			err := m.Import(targetDir)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println("All done.")
		},
	}

	flags := cmd.Flags()
	flags.String("target-dir", "",
		"Specify path of the directory where retrieve the backup data.")

	return cmd
}
