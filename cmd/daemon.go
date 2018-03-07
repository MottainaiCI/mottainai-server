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

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	"github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	"github.com/MottainaiCI/mottainai-server/pkg/template"
	"github.com/MottainaiCI/mottainai-server/routes/api"
	"github.com/urfave/cli"
)

var Daemon = cli.Command{
	Name:        "daemon",
	Usage:       "Start api daemon",
	Description: `daemon - a lighter version, just api`,
	Action: func(c *cli.Context) {
		if c.IsSet("config") {
			newDaemon().Start(c.String("config"))
		} else {
			fmt.Println("No config file provided - running default")
			newDaemon().Start("")
		}
	},
	Flags: []cli.Flag{
		stringFlag("config, c", "custom/conf/app.yml", "Custom configuration file path"),
	},
}

func newDaemon() *mottainai.Mottainai {
	m := mottainai.Classic()
	context.Setup(m)
	api.Setup(m)
	template.Setup(m)
	return m
}
