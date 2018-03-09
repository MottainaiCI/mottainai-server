/*

Copyright (C) 2018  Ettore Di Giacinto <mudler@gentoo.org>
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

package mottainai

import (
	"fmt"
	"net/http"
	"path"

	log "log"

	"github.com/MottainaiCI/mottainai-server/pkg/db"
	machinery "github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"

	"github.com/MottainaiCI/mottainai-server/pkg/tasks"
	"github.com/michaelklishin/rabbit-hole"
	macaron "gopkg.in/macaron.v1"

	"github.com/MottainaiCI/mottainai-server/pkg/settings"
)

type Mottainai struct {
	*macaron.Macaron
}

func New() *Mottainai {
	return &Mottainai{Macaron: macaron.New()}
}

func Classic() *Mottainai {
	return &Mottainai{Macaron: macaron.Classic()}
}

func (m *Mottainai) SetStatic() {
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
}

func (m *Mottainai) Start(fileconfig string) error {
	setting.GenDefault()

	if len(fileconfig) > 0 {
		setting.LoadFromFileEnvironment(fileconfig)
	}

	m.SetStatic()

	rabbit, m_error := m.NewMachineryServer()
	if m_error != nil {
		panic(m_error)
	}

	rmqc, r_error := rabbithole.NewClient(setting.Configuration.AMQPURI, setting.Configuration.AMQPUser, setting.Configuration.AMQPPass)
	if r_error != nil {
		panic(r_error)
	}

	th := agenttasks.DefaultTaskHandler()
	th.RegisterTasks(rabbit)
	fmt.Println("DB  with " + setting.Configuration.DBPath)

	database.NewDatabase("tiedot")

	m.Map(database.DBInstance)
	m.Map(rmqc)
	m.Map(rabbit)
	m.Map(th)

	var listenAddr = fmt.Sprintf("%s:%s", setting.Configuration.HTTPAddr, setting.Configuration.HTTPPort)
	log.Printf("Listen: %v://%s", setting.Configuration.Protocol, listenAddr)

	//m.Run()
	err := http.ListenAndServe(listenAddr, m)

	if err != nil {
		log.Fatal(4, "Fail to start server: %v", err)
	}
	return nil
}

func (m *Mottainai) NewMachineryServer() (*machinery.Server, error) {
	var cnf = &config.Config{
		Broker:          setting.Configuration.AMQPBroker,
		DefaultQueue:    setting.Configuration.AMQPDefaultQueue,
		ResultBackend:   setting.Configuration.AMQPResultBackend,
		ResultsExpireIn: setting.Configuration.ResultsExpireIn,
		AMQP: &config.AMQPConfig{
			Exchange:     setting.Configuration.AMQPExchange,
			ExchangeType: setting.Configuration.AMQPExchangeType,
			BindingKey:   setting.Configuration.AMQPBindingKey,
		},
	}

	server, err := machinery.NewServer(cnf)
	return server, err
}
