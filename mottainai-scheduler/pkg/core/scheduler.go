/*
Copyright (C) 2021 Daniele Rondina <geaaru@sabayonlinux.org>

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

package core

import (
	"errors"
	"fmt"
	"time"

	setting "github.com/MottainaiCI/mottainai-server/mottainai-scheduler/pkg/config"
	"github.com/MottainaiCI/mottainai-server/mottainai-scheduler/pkg/scheduler"
	specs "github.com/MottainaiCI/mottainai-server/mottainai-scheduler/pkg/specs"
	client "github.com/MottainaiCI/mottainai-server/pkg/client"
	logging "github.com/MottainaiCI/mottainai-server/pkg/logging"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"

	"github.com/mudler/anagent"
	logrus "github.com/sirupsen/logrus"
)

type MottainaiScheduler struct {
	*anagent.Anagent
	Client client.HttpClient

	ID       string
	Hostname string
}

func NewScheduler() *MottainaiScheduler {
	return &MottainaiScheduler{Anagent: anagent.New()}
}

func (m *MottainaiScheduler) InitializeTimer(config *setting.Config, sched specs.TaskScheduler) error {
	var tid anagent.TimerID = "keepalive"

	err := sched.Setup()
	if err != nil {
		msg := "Something goes wrong with scheduler setup: " + err.Error()
		return errors.New(msg)
	}

	duration, err := time.ParseDuration(fmt.Sprintf("%ds", config.GetScheduler().ScheduleTimerSec))
	if err != nil {
		return errors.New("Error on parse scheduler timer seconds: " + err.Error())
	}

	dSyncDflQueue, err := time.ParseDuration(
		fmt.Sprintf("%ds", config.GetScheduler().SyncDefaultQueueTimerSec),
	)
	if err != nil {
		return errors.New("Error on parse sync_default_queue_sec value: " + err.Error())
	}

	dSyncNodes, err := time.ParseDuration(
		fmt.Sprintf("%ds", config.GetScheduler().SyncNodesTimerSec),
	)
	if err != nil {
		return errors.New("Error on parse sync_nodes_sec value: " + err.Error())
	}

	m.Timer(tid, time.Now(), duration,
		true,
		func(a *anagent.Anagent) {

			err = sched.Schedule()
			if err != nil {
				fmt.Println("Error on schedule tasks " + err.Error())
			}
		})

	m.Timer("align_nodes", time.Now(), dSyncNodes,
		true,
		func(a *anagent.Anagent) {
			err = sched.RetrieveNodes()
			if err != nil {
				fmt.Println("Error on retrieve nodes " + err.Error())
			}
		})

	m.Timer("sync_default_queue", time.Now().Add(dSyncDflQueue), dSyncDflQueue,
		true,
		func(a *anagent.Anagent) {
			err = sched.RetrieveDefaultQueue()
			if err != nil {
				fmt.Println("Error on retrieve default queue " + err.Error())
			}
		})

	return nil
}

func (m *MottainaiScheduler) Run() error {
	var err error

	m.Invoke(func(config *setting.Config) {
		mcfg := config.ToMottainaiConfig()
		logger := logging.New()
		logger.SetupWithConfig(true, mcfg)
		logger.WithFields(logrus.Fields{
			"component": "scheduler",
		}).Info("Starting")
		m.Map(logger)

		// Create Scheduler
		s := scheduler.NewDefaultTaskScheduler(config, m.Anagent)
		m.Map(s)

		m.ID = utils.GenID()
		m.Hostname = utils.Hostname()

		err = m.InitializeTimer(config, s)
	})

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	m.Start()
	return errors.New("Scheduler stopped")
}
