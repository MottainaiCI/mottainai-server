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
	"errors"
	"os"
	"path"
	"strings"

	anagent "github.com/mudler/anagent"

	logrus "github.com/sirupsen/logrus"

	logging "github.com/MottainaiCI/mottainai-server/pkg/logging"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	utils "github.com/MottainaiCI/mottainai-server/pkg/utils"
	mhook "github.com/MottainaiCI/mottainai-server/pkg/webhook"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	mottainai "github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	tasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	user "github.com/MottainaiCI/mottainai-server/pkg/user"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	ggithub "github.com/google/go-github/github"
)

var (
	pending       = "pending"
	success       = "success"
	failure       = "error"
	pendingDesc   = "Build in progress, please wait."
	noPermDesc    = "Insufficient permissions"
	successDesc   = "Build successful."
	failureDesc   = "Build failed."
	notfoundDesc  = "No mottainai file found on repo"
	task_file     = ".mottainai"
	pipeline_file = ".mottainai-pipeline"
)

type GitContext struct {
	Dir      string
	Uid      string
	Commit   string
	Owner    string
	UserRepo string
	Checkout string
	Repo     string
	Ref      string
	User     string

	StoredUser *user.User
}

func SendTask(u *user.User, kind string, client *ggithub.Client, db *database.Database, m *mottainai.Mottainai, payload interface{}, w *mhook.WebHook) error {

	var appName string
	var logger *logging.Logger
	m.Invoke(func(config *setting.Config, l *logging.Logger) {
		appName = config.GetWeb().AppName
		logger = l
	})

	gitc, err := prepareTemp(u, kind, client, db, m, payload, w)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"component": "webhook_global_watcher",
			"error":     err.Error(),
		}).Error("Error while preparing temp directory")
		strerr := err.Error()
		if gitc != nil {
			client.Repositories.CreateStatus(stdctx.Background(), gitc.Owner, gitc.Repo, gitc.Ref, &ggithub.RepoStatus{State: &failure, Description: &strerr, Context: &appName})

		}
		return err
	}
	defer os.RemoveAll(gitc.Dir)

	gitdir := gitc.Dir
	pruid, commit, owner, user_repo, repo, ref := gitc.Uid, gitc.Commit, gitc.Owner, gitc.UserRepo, gitc.Repo, gitc.Ref

	// Create the 'pending' status and send it
	status1 := &ggithub.RepoStatus{State: &pending, Description: &pendingDesc, Context: &appName}

	client.Repositories.CreateStatus(stdctx.Background(), owner, repo, ref, status1)

	var t *tasks.Task
	exists := false
	if w.HasTask() {
		t, err = w.ReadTask()
		if err != nil {
			return err
		}
		exists = true
	} else {
		exists, _ = utils.Exists(path.Join(gitdir, task_file+".json"))
		if exists == true {
			t, err = tasks.FromFile(path.Join(gitdir, task_file+".json"))
			if err != nil {
				return err
			}
		} else {

			exists, _ = utils.Exists(path.Join(gitdir, task_file+".yaml"))
			if exists == true {
				t, err = tasks.FromYamlFile(path.Join(gitdir, task_file+".yaml"))
				if err != nil {
					return err
				}
			}
		}
	}
	if !exists {
		return errors.New("Task not found")
	}

	t.Owner = gitc.StoredUser.ID
	t.Source = user_repo
	t.Commit = commit
	t.Queue = QueueSetting(db)

	docID, err := db.Driver.CreateTask(t.ToMap())
	if err != nil {
		return err
	}
	var url string
	m.Invoke(func(config *setting.Config) {
		url = config.GetWeb().BuildAbsURL("/tasks/display/" + docID)
	})

	m.SendTask(docID)
	// Create the 'pending' status and send it
	status1 = &ggithub.RepoStatus{State: &pending, TargetURL: &url, Description: &pendingDesc, Context: &appName}

	client.Repositories.CreateStatus(stdctx.Background(), owner, repo, ref, status1)

	m.Invoke(func(a *anagent.Anagent) {
		data := strings.Join([]string{kind, owner, repo, ref, "tasks", docID}, ",")
		a.Invoke(func(w map[string]string) {
			a.Lock()
			defer a.Unlock()
			w[pruid] = data
		})

	})

	//return nil

	return nil

}

func SendPipeline(u *user.User, kind string, client *ggithub.Client, db *database.Database, m *mottainai.Mottainai, payload interface{}, w *mhook.WebHook) error {

	var logger *logging.Logger
	m.Invoke(func(l *logging.Logger) {
		logger = l
	})

	gitc, err := prepareTemp(u, kind, client, db, m, payload, w)
	if err != nil {
		return err
	}
	defer os.RemoveAll(gitc.Dir)
	gitdir := gitc.Dir
	pruid, commit, owner, user_repo, repo, ref := gitc.Uid, gitc.Commit, gitc.Owner, gitc.UserRepo, gitc.Repo, gitc.Ref

	var t *tasks.Pipeline
	exists := false
	if w.HasPipeline() {
		t, err = w.ReadPipeline()
		if err != nil {
			return err
		}
		exists = true
	} else {
		exists, _ = utils.Exists(path.Join(gitdir, pipeline_file+".json"))
		if exists == true {
			t, err = tasks.PipelineFromJsonFile(path.Join(gitdir, pipeline_file+".json"))
			if err != nil {
				return err
			}
		} else {
			exists, _ = utils.Exists(path.Join(gitdir, pipeline_file+".yaml"))
			if exists == true {
				t, err = tasks.PipelineFromYamlFile(path.Join(gitdir, pipeline_file+".yaml"))
				if err != nil {
					return err
				}
			}
		}
	}

	if !exists {
		return errors.New("Pipeline not found")
	}

	t.Owner = gitc.StoredUser.ID
	// XXX:
	t.Queue = QueueSetting(db)
	// do not allow automatic tag from PR
	for i, p := range t.Tasks { // Duplicated in API.
		//p.Namespace = ""
		//p.TagNamespace = ""
		if kind == "pull_request" {
			//	t.Namespace = "" // do not allow automatic tag from PR
			p.TagNamespace = ""
			p.Storage = ""
			p.Binds = []string{}
			p.RootTask = ""
		}
		p.Owner = gitc.StoredUser.ID
		p.Source = user_repo
		p.Commit = commit

		p.Status = setting.TASK_STATE_WAIT

		id, err := db.Driver.CreateTask(p.ToMap())
		if err != nil {
			return err
		}
		p.ID = id
		t.Tasks[i] = p
	}

	docID, err := db.Driver.CreatePipeline(t.ToMap(false))
	if err != nil {
		return err
	}

	var url string
	var appName string
	m.Invoke(func(config *setting.Config) {
		url = config.GetWeb().BuildURI("/tasks/display/" + docID)
		appName = config.GetWeb().AppName
	})
	logger.WithFields(logrus.Fields{
		"component":   "webhook_global_watcher",
		"pipeline_id": docID,
	}).Debug("Sending pipeline")

	_, err = m.ProcessPipeline(docID)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"component":   "webhook_global_watcher",
			"pipeline_id": docID,
			"error":       err.Error(),
		}).Error("While sending")
		return err
	}

	// Create the 'pending' status and send it
	status1 := &ggithub.RepoStatus{State: &pending, TargetURL: &url, Description: &pendingDesc, Context: &appName}

	client.Repositories.CreateStatus(stdctx.Background(), owner, repo, ref, status1)

	m.Invoke(func(a *anagent.Anagent) {
		data := strings.Join([]string{kind, owner, repo, ref, "pipeline", docID}, ",")
		a.Invoke(func(w map[string]string) {
			logger.WithFields(logrus.Fields{
				"component": "webhook_global_watcher",
				"event":     "add",
			}).Debug("Add event to global watcher")
			a.Lock()
			defer a.Unlock()
			w[pruid] = data
		})
	})

	return nil
}

func RequiresWebHookSetting(c *context.Context, db *database.Database) error {
	// Check setting if we have to process this.
	err := errors.New("Webhook integration disabled")
	uuu, err := db.Driver.GetSettingByKey(setting.SYSTEM_WEBHOOK_ENABLED)
	if err == nil {
		if uuu.IsDisabled() {
			c.ServerError("Webhook integration disabled", err)
			return err
		}
	}
	return nil
}

func QueueSetting(db *database.Database) string {

	uuu, err := db.Driver.GetSettingByKey(setting.SYSTEM_WEBHOOK_DEFAULT_QUEUE)
	if err != nil {
		return "default_webhooks"
	}
	return uuu.Value
}

func Setup(m *mottainai.Mottainai) {

	SetupGitHub(m)
}
