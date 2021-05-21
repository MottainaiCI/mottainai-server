/*

Copyright (C) 2018-2021  Ettore Di Giacinto <mudler@gentoo.org>
                         Daniele Rondina <geaaru@sabayonlinux.org>

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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	client "github.com/MottainaiCI/mottainai-server/pkg/client"
	logging "github.com/MottainaiCI/mottainai-server/pkg/logging"
	taskmanager "github.com/MottainaiCI/mottainai-server/pkg/tasks/manager"
	logrus "github.com/sirupsen/logrus"

	nodes "github.com/MottainaiCI/mottainai-server/pkg/nodes"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/mudler/anagent"

	"github.com/MottainaiCI/mottainai-server/pkg/utils"
)

type MottainaiAgent struct {
	*anagent.Anagent
	Client client.HttpClient

	ID           string
	Hostname     string
	PrivateQueue string
}

func NewAgent() *MottainaiAgent {
	return &MottainaiAgent{Anagent: anagent.New()}
}

const MAXTIMER = 720
const MINTIMER = 10
const R = 3.81199961
const STEPS = 215

func (m *MottainaiAgent) SetKeepAlive(ID, hostname string, config *setting.Config) {
	queues := config.GetAgent().Queues
	queues[m.PrivateQueue] = config.GetAgent().PrivateQueue
	m.Client.RegisterNode(
		ID, hostname,
		config.GetAgent().StandAlone,
		config.GetAgent().Queues,
		config.GetAgent().SupportedExecutors,
		config.GetAgent().AgentConcurrency,
	)

	var tid anagent.TimerID = "keepalive"
	var registerResponse nodes.NodeRegisterResponse

	m.Timer(tid, time.Now(), time.Duration(MINTIMER*time.Second), true,
		func(a *anagent.Anagent, c *client.Fetcher, tm *taskmanager.TaskManager) {
			queues := config.GetAgent().Queues
			queues[m.PrivateQueue] = config.GetAgent().PrivateQueue

			res, err := c.RegisterNode(
				ID, hostname,
				config.GetAgent().StandAlone,
				queues,
				config.GetAgent().SupportedExecutors,
				config.GetAgent().AgentConcurrency,
			)

			if err == nil && res.Request.Response.StatusCode == 200 && res.Status == "ok" {
				d := time.Duration(MINTIMER * time.Second)

				// Parse response
				err = json.Unmarshal([]byte(res.Data), &registerResponse)
				if err != nil {
					fmt.Println("Error on parse server response " + err.Error())
					return
				}
				// Readjust keepalive timer based on how many nodes are in the cluster.
				pop := utils.FeatureScaling(
					float64(registerResponse.Position),
					float64(registerResponse.NumNodes), 0, 1,
				)
				scale_factor := float64(registerResponse.NumNodes)
				timer := utils.FeatureScaling(
					utils.LogisticMapSteps(STEPS, R, pop)*scale_factor,
					float64(registerResponse.NumNodes),
					MINTIMER, MAXTIMER,
				)
				//fmt.Println("Timer set to", timer)
				if timer < MAXTIMER && timer > MINTIMER {
					d = time.Duration(timer) * time.Second
				}
				m.GetTimer(tid).After(d)

				if registerResponse.TaskInQueue {
					tm.NodeUniqueId = registerResponse.NodeUniqueId
					tm.NodeId = ID
					err := tm.GetTasks()
					if err != nil {
						fmt.Println("Unexpected error on process tasks: " + err.Error())
					}
				} else {
					// Check for expired tasks
					emptyMap := make(map[string][]string, 0)
					err := tm.AnalyzeQueues(emptyMap)
					if err != nil {
						fmt.Println("Unexpected error on process tasks: " + err.Error())
					}
				}

			} else {
				if err != nil {
					if res.Request != nil && res.Request.Response != nil {
						fmt.Println(fmt.Sprintf("%s: Error on registrer node: %s",
							res.Request.Response.Status, err.Error()))
					} else {
						fmt.Println(fmt.Sprintf("Error on registrer node: %s",
							err.Error()))
					}
				} else {
					fmt.Println(fmt.Sprintf("%s: Error on registrer node: %s",
						res.Request.Response.Status, res.Error))
				}
				//log.ERROR.Println("Error on register node ", err.Error())
			}

		})
}

func (m *MottainaiAgent) Run() error {

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-signalChannel
		switch sig {
		case os.Interrupt:
			fmt.Println("Received SIGINT event. Shutdown.")
		case syscall.SIGTERM:
			fmt.Println("Received SIGTERM event. Shutdown.")
		}

		m.Stop()
	}()

	m.Invoke(func(config *setting.Config) {
		logger := logging.New()
		logger.SetupWithConfig(true, config)
		logger.WithFields(logrus.Fields{
			"component": "agent",
		}).Info("Starting")
		//log.Set(logger)
		m.Map(logger)
		tm := taskmanager.NewTaskManager(config)
		m.Map(tm)
		fetcher := client.NewTokenClient(
			config.GetWeb().AppURL,
			config.GetAgent().ApiKey, config)
		m.Client = fetcher
		m.Map(fetcher)

		ID := utils.GenID()
		if config.GetAgent().ForceAgentId != "" {
			ID = config.GetAgent().ForceAgentId
		}

		m.ID = ID
		m.Hostname = utils.Hostname()
		//log.INFO.Println("Worker ID: " + ID)
		//log.INFO.Println("Worker Hostname: " + hostname)

		if config.GetAgent().PrivateQueue != 0 {
			m.PrivateQueue = m.Hostname + ID
			//log.INFO.Println("Listening on private queue: " + m.PrivateQueue)
		}

		m.SetKeepAlive(ID, m.Hostname, config)
	})

	m.Start()
	return errors.New("Agent stopped")
}
