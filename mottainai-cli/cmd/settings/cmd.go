/*

Copyright (C) 2018  Ettore Di Giacinto <mudler@gentoo.org>
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

package settingcmd

import (
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/spf13/cobra"
)

func NewSettingCommand(config *setting.Config) *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "setting [command] [OPTIONS]",
		Short: "Manage Infrastructure settings",
	}

	cmd.AddCommand(
		newSettingCreateCommand(config),
		newSettingListCommand(config),
		newSettingRemoveCommand(config),
		newSettingUpdateCommand(config),
	)

	return cmd
}
