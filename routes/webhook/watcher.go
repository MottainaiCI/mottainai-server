/*

Copyright (C) 2017-2018  Ettore Di Giacinto <mudler@gentoo.org>
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

package webhook

import (
	"time"

	logrus "github.com/sirupsen/logrus"

	logging "github.com/MottainaiCI/mottainai-server/pkg/logging"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	anagent "github.com/mudler/anagent"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	ggithub "github.com/google/go-github/github"
)

type WatcherEvent struct {
	EventType string
	EventId   string
	Handler   WebHookCallbacks
}

func NewWatcherEvent(eType, eId string, handler WebHookCallbacks) *WatcherEvent {
	return &WatcherEvent{
		EventType: eType,
		EventId:   eId,
		Handler:   handler,
	}
}

func GetDefaultLogFields(pipelineId, taskId, status, err string, handler WebHookCallbacks) logrus.Fields {
	ans := handler.GetLogFields(err)
	ans["component"] = "webhook_global_watcher"
	if pipelineId != "" {
		ans["pipeline"] = pipelineId
	}
	if taskId != "" {
		ans["task"] = taskId
	}
	if status != "" {
		ans["status"] = status
	}
	return ans
}

func GlobalWatcher(client *ggithub.Client, a *anagent.Anagent, db *database.Database, config *setting.Config, logger *logging.Logger) {
	logger.WithFields(logrus.Fields{
		"component": "webhook_global_watcher",
	}).Info("Starting")
	var tid anagent.TimerID = anagent.TimerID("global_watcher")
	watch := make(map[string]*WatcherEvent)

	a.Map(watch)

	a.Timer(tid, time.Now(), time.Duration(30*time.Second), true, func(w map[string]*WatcherEvent) {
		logger.WithFields(logrus.Fields{
			"component": "webhook_global_watcher",
		}).Debug("Check for pending tasks")

		//	defer a.Unlock()
		// Checking for PR that needs update
		for k, v := range w {

			if v.EventType == "pipeline" {

				url := config.GetWeb().BuildAbsURL("/pipeline/" + v.EventId)
				pip, err := db.Driver.GetPipeline(db.Config, v.EventId)
				if err != nil { // XXX:
					delete(w, k)
					return
				}

				done := 0
				fail := false
				for _, t := range pip.Tasks {

					ta, err := db.Driver.GetTask(db.Config, t.ID)
					if err != nil {
						delete(w, k)
						return
					}

					if ta.IsDone() || ta.IsStopped() {
						done++
						if !ta.IsSuccess() {
							fail = true
						}

					}
				}

				if done == len(pip.Tasks) {
					delete(w, k)
					if fail == false {
						fields := GetDefaultLogFields(v.EventId, "", "success", "", v.Handler)
						logger.WithFields(fields).Info("Pipeline successfully executed")

						v.Handler.SetStatus(&success, &successDesc, &url)

					} else {
						fields := GetDefaultLogFields(v.EventId, "", "failure", "", v.Handler)
						logger.WithFields(fields).Info("Pipeline failed ")

						v.Handler.SetStatus(&failure, &failureDesc, &url)
					}
				}

				// Handle task events
			} else if v.EventType == "task" {

				url := config.GetWeb().BuildAbsURL("/tasks/display/" + v.EventId)
				task, err := db.Driver.GetTask(db.Config, v.EventId)
				if err == nil {
					if task.IsDone() || task.IsStopped() {
						if task.IsSuccess() {
							fields := GetDefaultLogFields("", v.EventId, "success", "", v.Handler)
							logger.WithFields(fields).Info("Task succeeded")
							v.Handler.SetStatus(&success, &successDesc, &url)
						} else {
							fields := GetDefaultLogFields("", v.EventId, "failure", "", v.Handler)
							logger.WithFields(fields).Info("Task failed ")
							v.Handler.SetStatus(&failure, &failureDesc, &url)
						}
						delete(w, k)
					}

					return
				} else {
					delete(w, k)
				}

			} else {
				logger.Error("Unknown event %s", v)
				delete(w, k)
			}
		}

	})
}
