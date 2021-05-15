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

package queue

import (
	"fmt"
	"log"
	"os"

	tools "github.com/MottainaiCI/mottainai-server/mottainai-cli/common"
	client "github.com/MottainaiCI/mottainai-server/pkg/client"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"

	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
)

func newQueueCreateCommand(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create [OPTIONS]",
		Short: "Create queue (dev only)",
		Args:  cobra.OnlyValidArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			queue, _ := cmd.Flags().GetString("queue")

			if queue == "" {
				fmt.Println("Missing queue field")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			var v *viper.Viper = config.Viper

			fetcher := client.NewTokenClient(
				v.GetString("master"), v.GetString("apikey"), config,
			)

			queue, _ := cmd.Flags().GetString("queue")

			resp, err := fetcher.QueueCreate(queue)
			if err != nil {
				log.Fatalln("error:", err)
			}

			tools.PrintResponse(resp)
		},
	}

	var flags = cmd.Flags()
	flags.String("queue", "", "Name of the queue to create")

	return cmd
}
