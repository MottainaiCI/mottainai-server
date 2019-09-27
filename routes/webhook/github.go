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
	"fmt"
	"net/http"
	"strconv"

	user "github.com/MottainaiCI/mottainai-server/pkg/user"
	mhook "github.com/MottainaiCI/mottainai-server/pkg/webhook"

	logrus "github.com/sirupsen/logrus"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	logging "github.com/MottainaiCI/mottainai-server/pkg/logging"
	mottainai "github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	tasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	ggithub "github.com/google/go-github/github"
	webhooks "gopkg.in/go-playground/webhooks.v3"
	"gopkg.in/go-playground/webhooks.v3/github"
)

// HandlePullRequest handles GitHub pull_request events
func HandlePullRequest(m *mottainai.Mottainai, sessionHook *GitHubWebHook) {
	pl := sessionHook.Payload.(github.PullRequestPayload)

	m.Invoke(func(l *logging.Logger, client *ggithub.Client, db *database.Database) {
		sessionHook.Client = client
		l.WithFields(sessionHook.GetLogFields("")).Debug("Pull request received")

		if pl.Action == "closed" {
			return
		}

		sessionHook.HandleEvent(m, l, db)
	})
}

// HandlePush handles GitHub push events
func HandlePush(m *mottainai.Mottainai, sessionHook *GitHubWebHook) {
	m.Invoke(func(client *ggithub.Client, l *logging.Logger, db *database.Database) {
		sessionHook.Client = client
		l.WithFields(sessionHook.GetLogFields("")).Debug("Push received")
		sessionHook.HandleEvent(m, l, db)
	})
}

func NewGitContextGitHub(kind string, payload interface{}) *GitContext {
	var ans *GitContext
	if kind == "pull_request" {
		pl := payload.(github.PullRequestPayload)
		repo := pl.PullRequest.Base.Repo.Name
		ans = &GitContext{
			// TODO:
			CloneSSHUrl:  pl.PullRequest.Base.Repo.CloneURL,
			CloneHTTPUrl: pl.PullRequest.Base.Repo.CloneURL,
			User:         strconv.FormatInt(pl.PullRequest.User.ID, 10),
			UserRepo:     pl.PullRequest.Head.Repo.CloneURL,
			Commit:       pl.PullRequest.Head.Sha,
			Uid:          pl.PullRequest.Head.Sha + strconv.FormatInt(pl.PullRequest.Number, 10) + repo,
			Owner:        pl.PullRequest.Base.User.Login,
			Ref:          pl.PullRequest.Head.Sha,
			// allow to create complex filters: pull_request-develop, pull_request-master
			FilterRef: kind + "-" + pl.PullRequest.Base.Ref,
			Repo:      repo,
			Checkout:  strconv.FormatInt(pl.PullRequest.Number, 10),
		}

	} else {
		push := payload.(github.PushPayload)
		ans = &GitContext{
			Owner:        push.Repository.Owner.Name,
			CloneSSHUrl:  push.Repository.CloneURL,
			CloneHTTPUrl: push.Repository.CloneURL,
			UserRepo:     push.Repository.CloneURL,
			User:         strconv.FormatInt(push.Sender.ID, 10),
			Uid:          push.HeadCommit.ID + push.Repository.Name,
			Commit:       push.HeadCommit.ID,
			Checkout:     push.HeadCommit.ID,
			Repo:         push.Repository.Name,
			Ref:          push.HeadCommit.ID,
			FilterRef:    push.Ref,
		}
	}

	ans.KindEvent = kind

	return ans
}

//TODO: handle this with separate objs
type GitHubWebHook struct {
	*GitWebHook
	Client *ggithub.Client
}

func (h *GitHubWebHook) SetFailureStatus(errdescr string) {
	h.SetStatus(&failure, &errdescr, nil)
}

func (h *GitHubWebHook) SetStatus(state, errdescr, targetUrl *string) {
	status := ggithub.RepoStatus{
		State:       state,
		Description: errdescr,
		Context:     &h.AppName,
	}
	if targetUrl != nil {
		status.TargetURL = targetUrl
	}

	h.Client.Repositories.CreateStatus(
		stdctx.Background(),
		h.Context.Owner,
		h.Context.Repo,
		h.Context.Ref,
		&status,
	)
}

func (h *GitHubWebHook) LoadEventEnvs2Task(task *tasks.Task) {
	var envs []string

	if h.Context.KindEvent == "pull_request" {

		pl := h.Payload.(github.PullRequestPayload)

		envs = []string{
			"GITHUB_EVENT_TYPE=" + h.Context.KindEvent,
			"GITHUB_EVENT_REPO_NAME=" + pl.PullRequest.Base.Repo.Name,
			"GITHUB_EVENT_REPO_GITSSH_URL=" + pl.PullRequest.Base.Repo.SSHURL,
			"GITHUB_EVENT_REPO_GITHTTP_URL=" + pl.PullRequest.Base.Repo.GitURL,
			"GITHUB_EVENT_REF=" + pl.PullRequest.Base.Ref,
			"GITHUB_EVENT_COMMIT=" + pl.PullRequest.Head.Sha,
		}

	} else {
		push := h.Payload.(github.PushPayload)

		envs = []string{
			"GITHUB_EVENT_TYPE=" + h.Context.KindEvent,
			"GITHUB_EVENT_REPO_NAME=" + push.Repository.Name,
			"GITHUB_EVENT_REF=" + push.HeadCommit.ID,
		}
	}

	(*task).Environment = append(task.Environment, envs...)
}

func (h *GitHubWebHook) SetPendingStatus() {
	h.SetStatus(&pending, &pendingDesc, nil)
}

func (h *GitHubWebHook) GetLogFields(err string) logrus.Fields {
	ans := logrus.Fields{
		"component": "webhook",
		"event":     fmt.Sprintf("github_%s", h.Context.KindEvent),
		"wid":       h.Hook.ID,
	}
	if err != "" {
		ans["error"] = err
	}
	return ans
}

func NewGitHubWebHook(payload interface{}, w *mhook.WebHook, u *user.User, header webhooks.Header, kindEvent string) *GitHubWebHook {
	ans := &GitHubWebHook{
		GitWebHook: newGitWebHook(payload, w, u, header),
	}
	ans.Context = NewGitContext("github", kindEvent, payload)
	ans.GitWebHook.CBHandler = ans
	return ans
}

func GenGitHubHook(db *database.Database, m *mottainai.Mottainai, w *mhook.WebHook, u *user.User) *github.Webhook {
	secret := w.Key
	hook := github.New(&github.Config{Secret: secret})

	var appName, buildPath string
	m.Invoke(func(config *setting.Config) {
		appName = config.GetWeb().AppName
		buildPath = config.GetAgent().BuildPath
	})

	uuu, err := db.Driver.GetSettingByKey(setting.SYSTEM_WEBHOOK_PR_ENABLED)
	if err == nil && !uuu.IsDisabled() {
		hook.RegisterEvents(func(payload interface{}, header webhooks.Header) {
			sessionHook := NewGitHubWebHook(payload, w, u, header, "pull_request")
			sessionHook.AppName = appName
			sessionHook.BuildPath = buildPath
			HandlePullRequest(m, sessionHook)
		}, github.PullRequestEvent)
	}

	hook.RegisterEvents(func(payload interface{}, header webhooks.Header) {
		sessionHook := NewGitHubWebHook(payload, w, u, header, "push")
		sessionHook.AppName = appName
		sessionHook.BuildPath = buildPath
		HandlePush(m, sessionHook)
	}, github.PushEvent)

	return hook
}

func SetupGitHub(m *mottainai.Mottainai) {
	// TODO: Generate tokens for  each user.
	// Let user add repo in specific collection, and check against that

	webHookHandler := func(l *logging.Logger, ctx *context.Context,
		db *database.Database, resp http.ResponseWriter, req *http.Request) {
		uid := ctx.Params(":uid")
		l.WithFields(logrus.Fields{
			"component": "webhook",
			"event":     "github_post",
			"uid":       uid,
		}).Debug("Received payload")

		w, err := db.Driver.GetWebHook(uid)
		if err != nil {
			l.WithFields(logrus.Fields{
				"component": "webhook",
				"event":     "github_post",
				"uid":       uid,
			}).Error("No webhook found")
			return
		}
		u, err := db.Driver.GetUser(w.OwnerId)
		if err != nil {
			l.WithFields(logrus.Fields{
				"component": "webhook",
				"event":     "github_post",
				"uid":       uid,
			}).Error("No user found")
			return
		}
		hook := GenGitHubHook(db, m, &w, &u)
		hook.ParsePayload(resp, req)
	}

	m.Invoke(func(config *setting.Config) {
		m.Group(config.GetWeb().GroupAppPath(), func() {
			m.Post("/webhook/:uid/github", RequiresWebHookSetting, webHookHandler)
		})
	})

}
