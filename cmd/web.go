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
	"net/http"
	"path"

	log "log"

	"github.com/MottainaiCI/mottainai-server/pkg/agentconn"
	"github.com/MottainaiCI/mottainai-server/pkg/context"
	"github.com/MottainaiCI/mottainai-server/pkg/tasks"
	"github.com/michaelklishin/rabbit-hole"
	macaron "gopkg.in/macaron.v1"

	"github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/MottainaiCI/mottainai-server/pkg/template"
	"github.com/MottainaiCI/mottainai-server/routes"
	"github.com/urfave/cli"
)

var Web = cli.Command{
	Name:  "web",
	Usage: "Start web server",
	Description: `Mottainai web server is the only thing you need to run,
and it takes care of all the other things for you`,
	Action: runWeb,
	Flags: []cli.Flag{
		stringFlag("config, c", "custom/conf/app.yml", "Custom configuration file path"),
	},
}

// newMacaron initializes Macaron instance.
func newMacaron() *macaron.Macaron {

	m := macaron.Classic()

	m.Use(macaron.Static(
		path.Join(setting.Configuration.ArtefactPath),
		macaron.StaticOptions{
			Prefix: "artefact",
		},
	))

	m.Use(macaron.Static(
		path.Join(setting.Configuration.NamespacePath),
		macaron.StaticOptions{
			Prefix: "namespace",
		},
	))
	m.Use(macaron.Static(
		path.Join(setting.Configuration.StoragePath),
		macaron.StaticOptions{
			Prefix: "storage",
		},
	))
	//	m.Use(toolbox.Toolboxer(m))
	m.Use(macaron.Static(
		path.Join(setting.Configuration.StaticRootPath, "public"),
		macaron.StaticOptions{},
	))
	template.Setup(m)

	context.Setup(m)
	routes.Setup(m)

	return m
}

func runWeb(c *cli.Context) error {
	setting.GenDefault()
	if c.IsSet("config") {
		setting.LoadFromFileEnvironment(c.String("config"))
	}

	m := newMacaron()
	rabbit, m_error := agentconn.NewMachineryServer()
	if m_error != nil {
		panic(m_error)
	}

	rmqc, r_error := rabbithole.NewClient(setting.Configuration.AMQPURI, setting.Configuration.AMQPUser, setting.Configuration.AMQPPass)
	if r_error != nil {
		panic(r_error)
	}

	m.Map(rmqc)
	agenttasks.RegisterTasks(rabbit)
	m.Map(rabbit)

	var listenAddr = fmt.Sprintf("%s:%s", setting.Configuration.HTTPAddr, setting.Configuration.HTTPPort)
	log.Printf("Listen: %v://%s%s", setting.Configuration.Protocol, listenAddr, setting.Configuration.AppSubURL)

	//m.Run()
	err := http.ListenAndServe(listenAddr, m)

	if err != nil {
		log.Fatal(4, "Fail to start server: %v", err)
	}
	return nil
}
