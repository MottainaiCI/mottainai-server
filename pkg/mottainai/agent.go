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
	"os"

	client "github.com/MottainaiCI/mottainai-server/pkg/client"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/mudler/anagent"

	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"
	machinery "github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/log"
)

type MottainaiAgent struct {
	*anagent.Anagent
}

func NewAgent() *MottainaiAgent {
	return &MottainaiAgent{Anagent: anagent.New()}
}

func (m *MottainaiAgent) HealthCheckRun() {
	m.HealthCheckSetup()

	m.Anagent.Start()
}

func (m *MottainaiAgent) Run() error {

	server := NewServer()
	broker := server.Add(setting.Configuration.BrokerDefaultQueue)
	th := agenttasks.DefaultTaskHandler()
	fetcher := client.NewClient(setting.Configuration.AppURL)

	m.Map(server)
	m.Map(th)
	m.Map(fetcher)

	ID := utils.GenID()
	hostname := utils.Hostname()
	log.INFO.Println("Worker ID: " + ID)
	log.INFO.Println("Worker Hostname: " + hostname)

	if setting.Configuration.PrivateQueue != 0 {
		privqueue := hostname + ID
		b := server.Add(privqueue)
		w := b.NewWorker(privqueue, setting.Configuration.PrivateQueue)
		log.INFO.Println("Listening on private queue: " + privqueue)
		go w.Launch()
	}

	os.MkdirAll(setting.Configuration.TempWorkDir, os.ModePerm)

	defaultWorker := broker.NewWorker(ID, setting.Configuration.AgentConcurrency)
	fetcher.RegisterNode(ID, hostname)

	m.TimerSeconds(int64(200), true, func(c *client.Fetcher) {
		c.RegisterNode(ID, hostname)
	})

	for q, concurrent := range setting.Configuration.Queues {
		log.INFO.Println("Listening on queue ", q, " with concurrency ", concurrent)
		b := server.Add(q)
		w := b.NewWorker(ID, concurrent)
		go w.Launch()
	}

	if setting.Configuration.StandAlone {
		m.Start()
		return nil
	}

	go func(w *machinery.Worker, a *MottainaiAgent) {
		a.Map(w)
		a.Start()
	}(defaultWorker, m)

	return defaultWorker.Launch()
}
