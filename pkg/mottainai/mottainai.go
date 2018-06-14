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

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	static "github.com/MottainaiCI/mottainai-server/pkg/static"
	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"

	machinery "github.com/RichardKnop/machinery/v1"
	config "github.com/RichardKnop/machinery/v1/config"
	cron "github.com/robfig/cron"

	rabbithole "github.com/michaelklishin/rabbit-hole"
	macaron "gopkg.in/macaron.v1"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
)

type Mottainai struct {
	*macaron.Macaron
}

func New() *Mottainai {
	return &Mottainai{Macaron: macaron.New()}
}

func Classic() *Mottainai {
	cl := macaron.New()
	cl.Use(macaron.Logger())
	cl.Use(macaron.Recovery())
	cl.Invoke(func(l *log.Logger) {
		l.SetPrefix("[ Mottainai ] ")
	})
	return &Mottainai{Macaron: cl}
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

	m.Use(static.Static(
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
	m.SetAutoHead(true)

	server := NewServer()

	server.Add(setting.Configuration.BrokerDefaultQueue)
	if setting.Configuration.BrokerType == "amqp" {
		rmqc, r_error := rabbithole.NewClient(setting.Configuration.BrokerURI, setting.Configuration.BrokerUser, setting.Configuration.BrokerPass)
		if r_error != nil {
			panic(r_error)
		}
		m.Map(rmqc)
	}

	th := agenttasks.DefaultTaskHandler()
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
	var err error
	if len(setting.Configuration.TLSCert) > 0 && len(setting.Configuration.TLSKey) > 0 {
		err = http.ListenAndServeTLS(listenAddr, setting.Configuration.TLSCert, setting.Configuration.TLSKey, m)
	} else {
		err = http.ListenAndServe(listenAddr, m)
	}

	if err != nil {
		log.Fatal(4, "Fail to start server: %v", err)
	}
	c.Stop()
	return nil
}

func (m *Mottainai) SendTask(docID int) (bool, error) {
	result := true
	var err error
	m.Invoke(func(d *database.Database, server *MottainaiServer, th *agenttasks.TaskHandler) {

		task, err := d.GetTask(docID)
		if err != nil {
			result = false
			return
		}
		task.ClearBuildLog()
		var broker *Broker
		if len(task.Queue) > 0 {
			broker = server.Get(task.Queue)
			log.Println("Sending task to queue ", task.Queue)

		} else {
			broker = server.Get(setting.Configuration.BrokerDefaultQueue)
			log.Println("Sending task to queue ", setting.Configuration.BrokerDefaultQueue)

		}

		d.UpdateTask(docID, map[string]interface{}{"status": "waiting", "result": "none"})

		fmt.Printf("Task Source: %v, Script: %v, Directory: %v, TaskName: %v", task.Source, task.Script, task.Directory, task.TaskName)

		if !th.Exists(task.TaskName) {
			fmt.Printf("Could not send task: Invalid task name")
			result = false
			return
		}

		_, err = broker.SendTask(task.TaskName, docID)
		if err != nil {
			fmt.Printf("Could not send task: %s", err.Error())
			d.UpdateTask(docID, map[string]interface{}{
				"result": "error",
				"status": "done",
				"output": "Backend error, could not send task to broker: " + err.Error(),
			})

			result = false
			return
		}
	})
	return result, err
}

func (m *Mottainai) LoadPlans() {
	m.Invoke(func(c *cron.Cron, d *database.Database) {

		for _, plan := range d.AllPlans() {
			fmt.Println("Loading plan: ", plan.Task, plan)
			id := plan.ID
			c.AddFunc(plan.Planned, func() {
				plan, _ := d.GetPlan(id)
				plan.Task.Reset()
				docID, _ := d.CreateTask(plan.Task.ToMap())
				m.SendTask(docID)
			})
		}

	})
}

func (m *Mottainai) ReloadCron() {

	m.Invoke(func(c *cron.Cron, d *database.Database) {
		c.Stop()
		c = cron.New()
		m.Map(c)
		m.LoadPlans()
		c.Start()
	})

}

func NewMachineryServer(queue string) (*machinery.Server, error) {

	var amqpConfig *config.AMQPConfig
	if setting.Configuration.BrokerType == "amqp" {
		amqpConfig = &config.AMQPConfig{
			Exchange:     setting.Configuration.BrokerExchange,
			ExchangeType: setting.Configuration.BrokerExchangeType,
			BindingKey:   queue + "_key",
			//BindingKey:   setting.Configuration.BrokerBindingKey,
		}

	}
	var cnf = &config.Config{
		Broker:          setting.Configuration.Broker,
		DefaultQueue:    queue,
		ResultBackend:   setting.Configuration.BrokerResultBackend,
		ResultsExpireIn: setting.Configuration.ResultsExpireIn,
		AMQP:            amqpConfig,
	}

	server, err := machinery.NewServer(cnf)
	return server, err
}
