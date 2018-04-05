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

func (m *MottainaiAgent) Run(config string) error {

	setting.GenDefault()
	//agent := mottainai.NewAgent()
	if len(config) > 0 {
		setting.LoadFromFileEnvironment(config)
	}

	server, m_error := New().NewMachineryServer()
	if m_error != nil {
		panic(m_error)
	}

	th := agenttasks.DefaultTaskHandler()
	th.RegisterTasks(server)
	m.Map(th)
	ID := utils.GenID()
	hostname := utils.Hostname()
	log.INFO.Println("Worker ID: " + ID)
	log.INFO.Println("Worker Hostname: " + hostname)

	worker := server.NewWorker(ID, setting.Configuration.AgentConcurrency)
	fetcher := client.NewClient()
	fetcher.RegisterNode(ID, hostname)
	m.Map(fetcher)
	// agent.TimerSeconds(int64(200), true, func(l *corelog.Logger) {
	// 		Register(ID)
	// 	})

	go func(w *machinery.Worker, a *MottainaiAgent) {
		a.Map(w)
		a.Start()
	}(worker, m)

	return worker.Launch()
}
