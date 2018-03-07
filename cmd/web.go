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

	"github.com/MottainaiCI/mottainai-server/routes"
	"github.com/urfave/cli"
)

var Web = cli.Command{
	Name:        "web",
	Usage:       "Start web server",
	Description: `Full-blown webui`,
	Action: func(c *cli.Context) {
		if c.IsSet("config") {
			newWebUI().Start(c.String("config"))
		} else {
			fmt.Println("No config file provided - running default")
			newWebUI().Start("")
		}
	},
	Flags: []cli.Flag{
		stringFlag("config, c", "custom/conf/app.yml", "Custom configuration file path"),
	},
}

func newWebUI() *mottainai.Mottainai {

	m := mottainai.Classic()
	template.Setup(m)
	context.Setup(m)
	routes.Setup(m)

	return m
}
