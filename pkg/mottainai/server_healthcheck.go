/*

Copyright (C) 2019 Ettore Di Giacinto <mudler@gentoo.org>

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
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	logging "github.com/MottainaiCI/mottainai-server/pkg/logging"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/mudler/anagent"
	logrus "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

const timeformat = "20060102150405"

// Monitor state of tasks and nodes
// If a task hasn't been updated for a long time, and it's running we mark it as errored and
// we annotate in the task field the reason for the abort (agent unreachable).
// In the same way, we need to notice whenever an agent goes offline for long time, and abort all the running tasks assigned to it.

// Starts the watcher
func (m *Mottainai) HealthCheckRun(interval int) error {

	m.Invoke(func(l *logging.Logger) {
		l.WithFields(logrus.Fields{
			"component": "server_healthcheck",
			"interval":  interval,
		}).Info("Starting")
	})

	runner := anagent.New()
	runner.TimerSeconds(int64(interval), true, func() { m.Invoke(m.HealthCheck) })

	go runner.Start()

	return errors.New("Failed to start the healthcheck runner")
}

// Idempotent action
func (m *Mottainai) HealthCheck(d *database.Database, config *setting.Config) error {

	// In case of config reloads, service is still up, but we disable the actions if they are set to 0
	// We could also disable/enable the service as necessary, but that is low hanging fruit for later
	if config.GetWeb().TaskDeadline != 0 {
		// Deadline to 0 means to disable the check

		err := m.CheckTasksDeadline(d, config)
		if err != nil {
			return err
		}
	}

	if config.GetWeb().NodeDeadline != 0 {
		err := m.CheckNodesDeadline(d, config)
		if err != nil {
			return err
		}
	}
	return nil
}

func MarkTaskAborted(id, reason string, d *database.Database) error {
	return d.Driver.UpdateTask(id, map[string]interface{}{
		"status": setting.TASK_STATE_STOPPED,
		"result": setting.TASK_RESULT_ERROR,
		"output": "Task exceeded deadline: " + reason,
	})
}

func (m *Mottainai) CheckTasksDeadline(d *database.Database, config *setting.Config) error {

	tasks, e := d.Driver.GetTaskByStatus(config, setting.TASK_STATE_RUNNING)
	if e != nil {
		return e
	}
	for _, t := range tasks {
		now := time.Now()
		if len(t.UpdatedTime) > 0 {
			last_update, e := time.Parse(timeformat, t.UpdatedTime)
			if e != nil {
				return e
			}

			if int(now.Sub(last_update).Seconds()) > config.GetWeb().TaskDeadline {
				m.Invoke(func(l *logging.Logger) {
					l.WithFields(logrus.Fields{
						"component": "server_healthcheck",
						"action":    "abort",
						"task":      t.ID,
					}).Info("Exceeded task deadline (" + strconv.Itoa(config.GetWeb().TaskDeadline) + "s)")
				})
				e = MarkTaskAborted(t.ID, "No updates to the task since "+now.Sub(last_update).String(), d)
				if e != nil {
					return e
				}
			}
		}
	}

	return nil
}

func (m *Mottainai) CheckNodesDeadline(d *database.Database, config *setting.Config) error {
	// FIXME: O(n) for now, bad performance.
	//But usually this should be fine as we don't have as many hosts as tasks
	nodes := d.Driver.AllNodes()

	for _, n := range nodes {
		last_update, e := time.Parse(timeformat, n.LastReport)
		if e != nil {
			return e
		}
		now := time.Now()

		if int(now.Sub(last_update).Seconds()) > config.GetWeb().NodeDeadline {
			// If node is down, check among its tasks
			//  tasks, _ := d.Driver.AllNodeTask(db.Config, node.NodeID)
			// Probably there will be less running rather then hosts presents in the cluster
			tasks, e := d.Driver.GetTaskByStatus(config, setting.TASK_STATE_RUNNING)
			if e != nil {
				return e
			}
			for _, t := range tasks {
				if t.Node == n.ID {
					m.Invoke(func(l *logging.Logger) {
						l.WithFields(logrus.Fields{
							"component": "server_healthcheck",
							"action":    "abort",
							"task":      t.ID,
						}).Info("Exceeded node deadline (" + strconv.Itoa(config.GetWeb().TaskDeadline) + "s)")
					})
					e = MarkTaskAborted(t.ID, "Didn't heard from the node since "+now.Sub(last_update).String(), d)
					if e != nil {
						return e
					}
				}
			}
		}
	}
	return nil
}
