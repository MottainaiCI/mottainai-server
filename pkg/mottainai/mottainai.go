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
	cron "github.com/robfig/cron"

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

	server, m_error := m.NewMachineryServer()
	if m_error != nil {
		panic(m_error)
	}

	if setting.Configuration.BrokerType == "amqp" {
		rmqc, r_error := rabbithole.NewClient(setting.Configuration.BrokerURI, setting.Configuration.BrokerUser, setting.Configuration.BrokerPass)
		if r_error != nil {
			panic(r_error)
		}
		m.Map(rmqc)
	}

	th := agenttasks.DefaultTaskHandler()
	th.RegisterTasks(server)
	fmt.Println("DB  with " + setting.Configuration.DBPath)

	database.NewDatabase("tiedot")

	c := cron.New()

	m.Map(database.DBInstance)
	m.Map(server)
	m.Map(th)
	m.Map(c)
	m.Map(m)
	c.Start()

	m.LoadPlans()

	var listenAddr = fmt.Sprintf("%s:%s", setting.Configuration.HTTPAddr, setting.Configuration.HTTPPort)
	log.Printf("Listen: %v://%s", setting.Configuration.Protocol, listenAddr)

	//m.Run()
	err := http.ListenAndServe(listenAddr, m)

	if err != nil {
		log.Fatal(4, "Fail to start server: %v", err)
	}
	c.Stop()
	return nil
}

func (m *Mottainai) SendTask(docID int, server *machinery.Server, d *database.Database) (bool, error) {

	task, err := d.GetTask(docID)
	if err != nil {
		return false, err
	}
	task.ClearBuildLog()

	d.UpdateTask(docID, map[string]interface{}{"status": "waiting", "result": "none"})

	fmt.Printf("Task Source: %v, Script: %v, Yaml: %v, Directory: %v, TaskName: %v", task.Source, task.Script, task.Yaml, task.Directory, task.TaskName)
	th := agenttasks.DefaultTaskHandler()

	_, err = th.SendTask(server, task.TaskName, docID)
	if err != nil {
		fmt.Printf("Could not send task: %s", err.Error())
		return false, err
	}
	return true, nil
}

func (m *Mottainai) LoadPlans() {
	m.Invoke(func(c *cron.Cron, d *database.Database, server *machinery.Server) {

		for _, plan := range d.AllPlans() {
			fmt.Println("Loading plan: ", plan.Task, plan)
			c.AddFunc(plan.Planned, func() {
				docID, _ := d.CreateTask(plan.Task.ToMap())
				m.SendTask(docID, server, d)
			})
		}

	})
}

func (m *Mottainai) ReloadCron() {

	m.Invoke(func(c *cron.Cron, d *database.Database, server *machinery.Server) {
		c.Stop()
		c = cron.New()
		m.Map(c)
		m.LoadPlans()
		c.Start()
	})

}

func (m *Mottainai) NewMachineryServer() (*machinery.Server, error) {

	var amqpConfig *config.AMQPConfig
	if setting.Configuration.BrokerType == "amqp" {
		amqpConfig = &config.AMQPConfig{
			Exchange:     setting.Configuration.BrokerExchange,
			ExchangeType: setting.Configuration.BrokerExchangeType,
			BindingKey:   setting.Configuration.BrokerBindingKey,
		}

	}
	var cnf = &config.Config{
		Broker:          setting.Configuration.Broker,
		DefaultQueue:    setting.Configuration.BrokerDefaultQueue,
		ResultBackend:   setting.Configuration.BrokerResultBackend,
		ResultsExpireIn: setting.Configuration.ResultsExpireIn,
		AMQP:            amqpConfig,
	}

	server, err := machinery.NewServer(cnf)
	return server, err
}
