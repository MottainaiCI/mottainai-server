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
	"errors"
	"strconv"
	"strings"
	"time"

	client "github.com/MottainaiCI/mottainai-server/pkg/client"
	logging "github.com/MottainaiCI/mottainai-server/pkg/logging"
	taskmanager "github.com/MottainaiCI/mottainai-server/pkg/tasks/manager"
	logrus "github.com/sirupsen/logrus"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/mudler/anagent"

	"github.com/MottainaiCI/mottainai-server/pkg/utils"
	machinery "github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/log"
)

type MottainaiAgent struct {
	*anagent.Anagent
	Client client.HttpClient
}

func NewAgent() *MottainaiAgent {
	return &MottainaiAgent{Anagent: anagent.New()}
}

const MAXTIMER = 720
const MINTIMER = 50
const R = 3.81199961
const STEPS = 215

func (m *MottainaiAgent) SetKeepAlive(ID, hostname string) {
	m.Client.RegisterNode(ID, hostname)

	var tid anagent.TimerID = "keepalive"

	m.Timer(tid, time.Now(), time.Duration(MINTIMER*time.Second), true, func(a *anagent.Anagent, c *client.Fetcher) {
		if res, err := c.RegisterNode(ID, hostname); err == nil {
			d := time.Duration(MINTIMER * time.Second)
			population := strings.Split(res.Data, ",")
			if len(population) == 2 {
				nodes, e := strconv.Atoi(population[0])
				if e != nil {
					return
				}
				i, e := strconv.Atoi(population[1])
				if e != nil {
					return
				}
				// Readjust keepalive timer based on how many nodes are in the cluster.
				pop := utils.FeatureScaling(float64(i), float64(nodes), 0, 1)
				scale_factor := float64(nodes)
				timer := utils.FeatureScaling(utils.LogisticMapSteps(STEPS, R, pop)*scale_factor, float64(nodes), MINTIMER, MAXTIMER)
				//fmt.Println("Timer set to", timer)
				if timer < MAXTIMER && timer > MINTIMER {
					d = time.Duration(timer) * time.Second
				}
				m.GetTimer(tid).After(d)
			}

		}

	})
}

func (m *MottainaiAgent) Run() error {

	var defaultWorker *machinery.Worker
	var is_standalone bool = false
	server := NewServer()

	m.Invoke(func(config *setting.Config) {

		logger := logging.New()
		logger.SetupWithConfig(true, config)
		logger.WithFields(logrus.Fields{
			"component": "agent",
		}).Info("Starting")
		log.Set(logger)
		m.Map(logger)

		broker := server.Add(config.GetBroker().BrokerDefaultQueue, config)
		th := taskmanager.DefaultTaskHandler(config)
		fetcher := client.NewTokenClient(
			config.GetWeb().AppURL,
			config.GetAgent().ApiKey, config)
		m.Client = fetcher
		m.Map(server)
		m.Map(th)
		m.Map(fetcher)

		ID := utils.GenID()
		hostname := utils.Hostname()
		log.INFO.Println("Worker ID: " + ID)
		log.INFO.Println("Worker Hostname: " + hostname)

		if config.GetAgent().PrivateQueue != 0 {
			privqueue := hostname + ID
			b := server.Add(privqueue, config)
			w := b.NewWorker(privqueue, config.GetAgent().PrivateQueue)
			log.INFO.Println("Listening on private queue: " + privqueue)
			go w.Launch()
		}

		defaultWorker = broker.NewWorker(ID, config.GetAgent().AgentConcurrency)
		m.SetKeepAlive(ID, hostname)

		for q, concurrent := range config.GetAgent().Queues {
			log.INFO.Println("Listening on queue ", q, " with concurrency ", concurrent)
			b := server.Add(q, config)
			w := b.NewWorker(ID, concurrent)
			go w.Launch()
		}

		is_standalone = config.GetAgent().StandAlone
	})

	if is_standalone {
		m.Start()
		return errors.New("Agent stopped")
	}

	go func(w *machinery.Worker, a *MottainaiAgent) {
		a.Map(w)
		a.Start()
	}(defaultWorker, m)

	return defaultWorker.Launch()
}
