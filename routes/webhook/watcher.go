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
	stdctx "context"
	"strings"
	"time"

	logrus "github.com/sirupsen/logrus"

	logging "github.com/MottainaiCI/mottainai-server/pkg/logging"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	anagent "github.com/mudler/anagent"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	ggithub "github.com/google/go-github/github"
)

func GlobalWatcher(client *ggithub.Client, a *anagent.Anagent, db *database.Database, config *setting.Config, logger *logging.Logger) {
	url := config.GetWeb().AppURL
	appName := config.GetWeb().AppName
	logger.WithFields(logrus.Fields{
		"component": "webhook_global_watcher",
	}).Info("Starting")
	var tid anagent.TimerID = anagent.TimerID("global_watcher")
	watch := make(map[string]string)

	a.Map(watch)

	a.Timer(tid, time.Now(), time.Duration(30*time.Second), true, func(w map[string]string) {
		logger.WithFields(logrus.Fields{
			"component": "webhook_global_watcher",
		}).Debug("Check for pending tasks")

		//	defer a.Unlock()
		// Checking for PR that needs update
		for k, v := range w {
			data := strings.Split(v, ",")

			turl := url + "/" + data[4] + "/display/" + data[5]

			if data[4] == "pipeline" {

				pip, err := db.Driver.GetPipeline(db.Config, data[5])
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
						logger.WithFields(logrus.Fields{
							"component":   "webhook_global_watcher",
							"pipeline_id": data[5],
							"status":      "success",
						}).Info("Pipeline successfully executed")

						status2 := &ggithub.RepoStatus{State: &success, TargetURL: &url, Description: &successDesc, Context: &appName}
						client.Repositories.CreateStatus(stdctx.Background(), data[1], data[2], data[3], status2)

					} else {
						logger.WithFields(logrus.Fields{
							"component":   "webhook_global_watcher",
							"pipeline_id": data[5],
							"status":      "failure",
						}).Info("Pipeline failed ")
						status2 := &ggithub.RepoStatus{State: &failure, TargetURL: &url, Description: &failureDesc, Context: &appName}
						client.Repositories.CreateStatus(stdctx.Background(), data[1], data[2], data[3], status2)
					}
				}
			} else {

				task, err := db.Driver.GetTask(db.Config, data[5])
				if err == nil {
					if task.IsDone() || task.IsStopped() {
						if task.IsSuccess() {
							logger.WithFields(logrus.Fields{
								"component": "webhook_global_watcher",
								"task_id":   data[5],
								"status":    "success",
							}).Info("Task succeeded")
							status2 := &ggithub.RepoStatus{State: &success, TargetURL: &turl, Description: &successDesc, Context: &appName}
							client.Repositories.CreateStatus(stdctx.Background(), data[1], data[2], data[3], status2)
						} else {
							logger.WithFields(logrus.Fields{
								"component": "webhook_global_watcher",
								"task_id":   data[5],
								"status":    "failure",
							}).Info("Task failed ")
							status2 := &ggithub.RepoStatus{State: &failure, TargetURL: &turl, Description: &failureDesc, Context: &appName}
							client.Repositories.CreateStatus(stdctx.Background(), data[1], data[2], data[3], status2)
						}
						delete(w, k)
					}

					return
				} else {
					delete(w, k)
				}

			}
		}

	})
}
