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
	"log"

	client "github.com/MottainaiCI/mottainai-server/pkg/client"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
)

func newTaskDownloadCommand(config *setting.Config) *cobra.Command {
	var filters []string

	var cmd = &cobra.Command{
		Use:   "download <taskid> <target> [OPTIONS]",
		Short: "Download task artefacts",
		Args:  cobra.RangeArgs(2, 2),
		Run: func(cmd *cobra.Command, args []string) {
			var v *viper.Viper = config.Viper

			id := args[0]
			target := args[1]
			if len(id) == 0 || len(target) == 0 {
				log.Fatalln("You need to define a task id and a target")
			}
			fetcher := client.NewTokenClient(v.GetString("master"), v.GetString("apikey"), config)
			fetcher.SetActiveReport(true)
			if err := fetcher.DownloadArtefactsFromTask(id, target, filters); err != nil {
				log.Fatalln(err)
			}
		},
	}

	cmd.Flags().StringArrayVarP(&filters, "filter", "f", []string{},
		"Define regex rule for filter artefacts to download.")
	return cmd
}
