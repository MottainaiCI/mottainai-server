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
	"time"

	setting "github.com/MottainaiCI/mottainai-server/mottainai-scheduler/pkg/config"
	"github.com/MottainaiCI/mottainai-server/mottainai-scheduler/pkg/scheduler"
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

func (m *MottainaiScheduler) InitializeTimer(config *setting.Config) {
	var tid anagent.TimerID = "keepalive"

	m.Timer(
		tid, time.Now(),
		time.Duration(config.GetScheduler().ScheduleTimerSec*time.Second),
		true,
		func(a *anagent.Anagent, c *client.Fetcher) {
		})
}

func (m *MottainaiScheduler) Run() error {

	m.Invoke(func(config *setting.Config) {
		mcfg := config.ToMottainaiConfig()
		logger := logging.New()
		logger.SetupWithConfig(true, mcfg)
		logger.WithFields(logrus.Fields{
			"component": "scheduler",
		}).Info("Starting")
		m.Map(logger)

		// Create Scheduler
		s := scheduler.NewDefaultTaskScheduler(config, m)
		m.Map(s)

		m.ID = utils.GenID()
		m.Hostname = utils.Hostname()

		m.InitializeTimer(config)
	})

	m.Start()
	return errors.New("Scheduler stopped")
}
