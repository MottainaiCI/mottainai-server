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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	anagent "github.com/mudler/anagent"

	"github.com/MottainaiCI/mottainai-server/pkg/utils"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	mottainai "github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	tasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	user "github.com/MottainaiCI/mottainai-server/pkg/user"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	ggithub "github.com/google/go-github/github"
	webhooks "gopkg.in/go-playground/webhooks.v3"

	"gopkg.in/go-playground/webhooks.v3/github"
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
	appName       = "MottainaiCI"
	task_file     = ".mottainai"
	pipeline_file = ".mottainai-pipeline"
)

// HandlePullRequest handles GitHub pull_request events
func HandlePullRequest(payload interface{}, header webhooks.Header, m *mottainai.Mottainai) {

	fmt.Println("Handling Pull Request")
	pl := payload.(github.PullRequestPayload)
	if pl.Action == "closed" {
		return
	}
	m.Invoke(func(client *ggithub.Client, db *database.Database) {

		if err := SendTask("pull_request", client, db, m, payload); err != nil {
			fmt.Println("Failed sending task", err)
		}
		if err := SendPipeline("pull_request", client, db, m, payload); err != nil {
			fmt.Println("Failed sending task", err)
		}
	})
}

// HandlePush handles GitHub push events
func HandlePush(payload interface{}, header webhooks.Header, m *mottainai.Mottainai) {

	fmt.Println("Handling Push")

	m.Invoke(func(client *ggithub.Client, db *database.Database) {

		if err := SendTask("push", client, db, m, payload); err != nil {
			fmt.Println("Failed sending task", err)
		}
		if err := SendPipeline("push", client, db, m, payload); err != nil {
			fmt.Println("Failed sending task", err)
		}
	})
}

func SendTask(kind string, client *ggithub.Client, db *database.Database, m *mottainai.Mottainai, payload interface{}) error {
	gitc, err := prepareTemp(kind, client, db, m, payload)
	if err != nil {
		return err
	}
	defer os.RemoveAll(gitc.Dir)
	gitdir := gitc.Dir
	pruid, commit, owner, user_repo, repo, ref := gitc.Uid, gitc.Commit, gitc.Owner, gitc.UserRepo, gitc.Repo, gitc.Ref

	var t *tasks.Task
	exists, _ := utils.Exists(path.Join(gitdir, task_file+".json"))
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

	if exists {
		t.Owner = gitc.StoredUser.ID
		t.Namespace = "" // do not allow automatic tag from PR
		t.TagNamespace = ""
		t.Source = user_repo
		t.Commit = commit
		t.Queue = setting.Configuration.WebHookDefaultQueue

		docID, err := db.Driver.CreateTask(t.ToMap())
		if err != nil {
			return err
		}

		url := m.Url() + "/tasks/display/" + docID
		m.SendTask(docID)

		// Create the 'pending' status and send it
		status1 := &ggithub.RepoStatus{State: &pending, TargetURL: &url, Description: &pendingDesc, Context: &appName}

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
	}

	return nil

}

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

func prepareTemp(kind string, client *ggithub.Client, db *database.Database, m *mottainai.Mottainai, payload interface{}) (*GitContext, error) {
	var pruid, commit, owner, user_repo, checkout, repo, ref, gh_user, clone_url string
	if kind == "pull_request" {
		pl := payload.(github.PullRequestPayload)
		clone_url = pl.PullRequest.Base.Repo.CloneURL
		gh_user = strconv.FormatInt(pl.PullRequest.User.ID, 10)
		user_repo = pl.PullRequest.Head.Repo.CloneURL

		commit = pl.PullRequest.Head.Sha

		number := pl.PullRequest.Number
		pruid = pl.PullRequest.Head.Sha + strconv.FormatInt(number, 10) + repo

		owner = pl.PullRequest.Base.User.Login
		repo = pl.PullRequest.Base.Repo.Name
		ref = pl.PullRequest.Head.Sha
		checkout = strconv.FormatInt(number, 10)
	} else {
		push := payload.(github.PushPayload)

		owner = push.Repository.Owner.Name
		gh_user = strconv.FormatInt(push.Sender.ID, 10)
		user_repo = push.Repository.CloneURL
		clone_url = user_repo
		repo = push.Repository.Name
		ref = push.HeadCommit.ID
		pruid = ref + repo
		checkout = ref
		commit = ref
	}

	ctx := &GitContext{Uid: pruid, Commit: commit, Owner: owner, UserRepo: user_repo, Checkout: checkout, Repo: repo, Ref: ref, User: gh_user}

	// Check setting if we have to process this.
	uuu, err := db.Driver.GetSettingByKey(setting.SYSTEM_WEBHOOK_ENABLED)
	if err == nil {
		if uuu.IsDisabled() {
			fmt.Println("Webhooks disabled from system settings")
			return ctx, errors.New("Webhooks disabled")
		}
	}

	// TODO: Check in users the enabled repository hooks
	// Later, with organizations and projects will be easier to link them.
	u, err := db.Driver.GetUserByIdentity("github", gh_user)
	if err != nil {
		status2 := &ggithub.RepoStatus{State: &failure, Description: &noPermDesc, Context: &appName}
		client.Repositories.CreateStatus(stdctx.Background(), owner, repo, ref, status2)

		return ctx, err
	}

	ctx.StoredUser = &u

	gitdir, err := ioutil.TempDir(setting.Configuration.BuildPath, "webhook_fetch"+repo)
	if err != nil {
		return ctx, errors.New("Failed creating tempdir: " + err.Error())
	}
	ctx.Dir = gitdir

	_, err = utils.Git([]string{"clone", clone_url, gitdir}, ".")
	if err != nil {
		return ctx, errors.New("Failed cloning repo: " + clone_url + " " + gitdir + " " + err.Error())
	}
	if kind == "pull_request" {
		_, err = utils.Git([]string{"fetch", "origin", "pull/" + checkout + "/head:CI_test"}, gitdir)
		if err != nil {
			os.RemoveAll(gitdir)
			return ctx, errors.New("Failed fetching repo: " + err.Error())
		}

		_, err = utils.Git([]string{"checkout", "CI_test"}, gitdir)
		if err != nil {
			os.RemoveAll(gitdir)
			return ctx, errors.New("Failed checkout repo: " + err.Error())
		}
	} else {
		_, err = utils.Git([]string{"checkout", checkout}, gitdir)
		if err != nil {
			os.RemoveAll(gitdir)
			return ctx, errors.New("Failed checkout repo: " + err.Error())
		}
	}

	return ctx, nil
}

func SendPipeline(kind string, client *ggithub.Client, db *database.Database, m *mottainai.Mottainai, payload interface{}) error {

	gitc, err := prepareTemp(kind, client, db, m, payload)
	if err != nil {
		return err
	}
	defer os.RemoveAll(gitc.Dir)
	gitdir := gitc.Dir
	pruid, commit, owner, user_repo, repo, ref := gitc.Uid, gitc.Commit, gitc.Owner, gitc.UserRepo, gitc.Repo, gitc.Ref

	var t *tasks.Pipeline
	exists, _ := utils.Exists(path.Join(gitdir, pipeline_file+".json"))
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

	if exists {
		t.Owner = gitc.StoredUser.ID
		// XXX:
		// do not allow automatic tag from PR
		for i, p := range t.Tasks { // Duplicated in API.
			p.Namespace = ""
			p.TagNamespace = ""
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
		t.Queue = setting.Configuration.WebHookDefaultQueue

		docID, err := db.Driver.CreatePipeline(t.ToMap(false))
		if err != nil {
			return err
		}

		url := m.Url() + "/tasks/display/" + docID
		fmt.Println("Sending pipeline", docID)
		_, err = m.ProcessPipeline(docID)
		if err != nil {
			fmt.Println(err)
			return err
		}

		// Create the 'pending' status and send it
		status1 := &ggithub.RepoStatus{State: &pending, TargetURL: &url, Description: &pendingDesc, Context: &appName}

		client.Repositories.CreateStatus(stdctx.Background(), owner, repo, ref, status1)

		m.Invoke(func(a *anagent.Anagent) {
			data := strings.Join([]string{kind, owner, repo, ref, "pipeline", docID}, ",")
			a.Invoke(func(w map[string]string) {
				fmt.Println("Add event to global watcher")
				a.Lock()
				defer a.Unlock()
				w[pruid] = data
			})
		})

		//return
	}
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

func GlobalWatcher(client *ggithub.Client, a *anagent.Anagent, db *database.Database, url string) {
	var tid anagent.TimerID = anagent.TimerID("global_watcher")
	watch := make(map[string]string)

	a.Map(watch)

	a.Timer(tid, time.Now(), time.Duration(30*time.Second), true, func(w map[string]string) {

		//	defer a.Unlock()
		// Checking for PR that needs update
		for k, v := range w {
			data := strings.Split(v, ",")

			turl := url + "/" + data[4] + "/display/" + data[5]

			if data[4] == "pipeline" {

				pip, err := db.Driver.GetPipeline(data[5])
				if err != nil { // XXX:
					delete(w, k)
					return
				}

				done := 0
				fail := false
				for _, t := range pip.Tasks {

					ta, err := db.Driver.GetTask(t.ID)
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
						fmt.Println("Returning Success")
						status2 := &ggithub.RepoStatus{State: &success, TargetURL: &url, Description: &successDesc, Context: &appName}
						client.Repositories.CreateStatus(stdctx.Background(), data[1], data[2], data[3], status2)

					} else {
						fmt.Println("Returning Failure")
						status2 := &ggithub.RepoStatus{State: &failure, TargetURL: &url, Description: &failureDesc, Context: &appName}
						client.Repositories.CreateStatus(stdctx.Background(), data[1], data[2], data[3], status2)
					}
				}
			} else {

				task, err := db.Driver.GetTask(data[5])
				if err == nil {
					if task.IsDone() || task.IsStopped() {
						if task.IsSuccess() {
							fmt.Println("Returning Success")
							status2 := &ggithub.RepoStatus{State: &success, TargetURL: &turl, Description: &successDesc, Context: &appName}
							client.Repositories.CreateStatus(stdctx.Background(), data[1], data[2], data[3], status2)

						} else {
							fmt.Println("Returning Failure")
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
func SetupGitHub(m *mottainai.Mottainai) {

	m.Invoke(func(client *ggithub.Client, a *anagent.Anagent, db *database.Database) {
		GlobalWatcher(client, a, db, m.Url())
	})

	hook := github.New(&github.Config{Secret: setting.Configuration.WebHookToken})
	hook.RegisterEvents(func(payload interface{}, header webhooks.Header) {
		fmt.Println("Received webhook for PR")
		m.Invoke(func(mo *mottainai.Mottainai) { HandlePullRequest(payload, header, mo) })
	}, github.PullRequestEvent)

	hook.RegisterEvents(func(payload interface{}, header webhooks.Header) {
		fmt.Println("Received webhook for push")
		m.Invoke(func(mo *mottainai.Mottainai) { HandlePush(payload, header, mo) })
	}, github.PushEvent)
	// TODO: Generate tokens for  each user.
	// Let user add repo in specific collection, and check against that
	m.Post("/webhook/github", RequiresWebHookSetting, func(ctx *context.Context, db *database.Database, resp http.ResponseWriter, req *http.Request) {
		hook.ParsePayload(resp, req)
	})

}
