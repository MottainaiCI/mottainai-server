/*

Copyright (C) 2017-2019  Daniele Rondina <geaaru@sabayonlinux.org>
                         Ettore Di Giacinto <mudler@gentoo.org>
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
	"fmt"
	"net/http"
	"strconv"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	logging "github.com/MottainaiCI/mottainai-server/pkg/logging"
	mottainai "github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	tasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	user "github.com/MottainaiCI/mottainai-server/pkg/user"
	mhook "github.com/MottainaiCI/mottainai-server/pkg/webhook"

	logrus "github.com/sirupsen/logrus"
	webhooks "gopkg.in/go-playground/webhooks.v3"
	gitlab "gopkg.in/go-playground/webhooks.v3/gitlab"
)

type GitLabWebhookDataResp struct {
	TaskId     string `json:"task_id,omitempty"`
	PipelineId string `json:"pipeline_id,omitempty"`
	Note       string `json:"note,omitempty"`
	Processed  bool   `json:"processed"`
	Status     bool   `json:"status"`
}

type GitLabWebHook struct {
	*GitWebHook
}

func NewGitLabWebHook(payload interface{}, w *mhook.WebHook, u *user.User, header webhooks.Header, kindEvent string) *GitLabWebHook {
	ans := &GitLabWebHook{
		GitWebHook: newGitWebHook(payload, w, u, header),
	}
	ans.Context = NewGitContext("gitlab", kindEvent, payload)
	ans.GitWebHook.CBHandler = ans
	return ans
}

func (h *GitLabWebHook) LoadEventEnvs2Task(task *tasks.Task) {
	var envs []string

	if h.Context.KindEvent == "merge_request" {
		mr := h.Payload.(gitlab.MergeRequestEventPayload)
		envs = []string{
			"GITLAB_EVENT_MRID=" + strconv.FormatInt(mr.ObjectAttributes.IID, 10),
			"GITLAB_EVENT_TYPE=" + h.Context.KindEvent,
			"GITLAB_EVENT_USERNAME=" + mr.User.UserName,
			"GITLAB_EVENT_PROJECT_NAME=" + mr.Project.Name,
			"GITLAB_EVENT_REPO_HOMEPAGE=" + mr.Repository.Homepage,
			"GITLAB_EVENT_REPO_URL=" + mr.Repository.URL,
			"GITLAB_EVENT_SOURCE_GITSSH_URL=" + mr.ObjectAttributes.Source.GitSSHURL,
			"GITLAB_EVENT_SOURCE_GITHTTP_URL=" + mr.ObjectAttributes.Source.GitHTTPURL,
			"GITLAB_EVENT_SOURCE_BRANCH=" + mr.ObjectAttributes.SourceBranch,
			"GITLAB_EVENT_TARGET_GITSSH_URL=" + mr.ObjectAttributes.Target.GitSSHURL,
			"GITLAB_EVENT_TARGET_GITHTTP_URL=" + mr.ObjectAttributes.Target.GitHTTPURL,
			"GITLAB_EVENT_TARGET_BRANCH=" + mr.ObjectAttributes.TargetBranch,
			"GITLAB_EVENT_MR_TITLE=" + mr.ObjectAttributes.Title,
			"GITLAB_EVENT_MR_STATE=" + mr.ObjectAttributes.State,
			"GITLAB_EVENT_MR_ACTION=" + mr.ObjectAttributes.Action,
			"GITLAB_EVENT_LAST_COMMIT=" + mr.ObjectAttributes.LastCommit.ID,
			"GITLAB_EVENT_AUTHORID=" + strconv.FormatInt(mr.ObjectAttributes.AuthorID, 10),
		}

		// Addend info to task
		if task.Name == "" {
			task.Name = fmt.Sprintf("%s %s #%d",
				mr.Project.Name, h.Context.KindEvent, mr.ObjectAttributes.IID)
		} else {
			task.Name = fmt.Sprintf("%s - %s %s #%d",
				task.Name, mr.Project.Name, h.Context.KindEvent, mr.ObjectAttributes.IID,
			)
		}

	} else if h.Context.KindEvent == "push" {
		push := h.Payload.(gitlab.PushEventPayload)

		envs = []string{
			"GITLAB_EVENT_TYPE=" + h.Context.KindEvent,
			"GITLAB_EVENT_USERNAME=" + push.UserName,
			"GITLAB_EVENT_PROJECT_NAME=" + push.Project.Name,
			"GITLAB_EVENT_REPO_HOMEPAGE=" + push.Repository.Homepage,
			"GITLAB_EVENT_REPO_URL=" + push.Repository.URL,
			"GITLAB_EVENT_REPO_GITSSH_URL=" + push.Project.GitSSSHURL,
			"GITLAB_EVENT_REPO_GITHTTP_URL=" + push.Project.GitHTTPURL,
			"GITLAB_EVENT_USERID=" + strconv.FormatInt(push.UserID, 10),
			"GITLAB_EVENT_REF=" + push.Ref,
			"GITLAB_EVENT_COMMIT_BEFORE=" + push.Before,
			"GITLAB_EVENT_CHECKOUT_SHA=" + push.CheckoutSHA,
			"GITLAB_EVENT_TOT_COMMIT_COUNT=" + strconv.FormatInt(push.TotalCommitsCount, 10),
		}

		// Addend info to task
		if task.Name == "" {
			task.Name = fmt.Sprintf("%s %s %s",
				push.Project.Name, h.Context.KindEvent, push.CheckoutSHA)
		} else {
			task.Name = fmt.Sprintf("%s - %s %s %s",
				task.Name, push.Project.Name, h.Context.KindEvent, push.CheckoutSHA,
			)
		}

	} else if h.Context.KindEvent == "tag" {
		tag := h.Payload.(gitlab.TagEventPayload)

		envs = []string{
			"GITLAB_EVENT_TYPE=" + h.Context.KindEvent,
			"GITLAB_EVENT_USERNAME=" + tag.UserName,
			"GITLAB_EVENT_PROJECT_NAME=" + tag.Project.Name,
			"GITLAB_EVENT_REPO_HOMEPAGE=" + tag.Repository.Homepage,
			"GITLAB_EVENT_REPO_URL=" + tag.Repository.URL,
			"GITLAB_EVENT_REPO_GITSSH_URL=" + tag.Project.GitSSSHURL,
			"GITLAB_EVENT_REPO_GITHTTP_URL=" + tag.Project.GitHTTPURL,
			"GITLAB_EVENT_USERID=" + strconv.FormatInt(tag.UserID, 10),
			"GITLAB_EVENT_REF=" + tag.Ref,
			"GITLAB_EVENT_CHECKOUT_SHA=" + tag.CheckoutSHA,
		}

		// Addend info to task
		if task.Name == "" {
			task.Name = fmt.Sprintf("%s %s %s",
				tag.Project.Name, h.Context.KindEvent, tag.Ref)
		} else {
			task.Name = fmt.Sprintf("%s - %s %s %s",
				task.Name, tag.Project.Name, h.Context.KindEvent, tag.Ref,
			)
		}

	} else if h.Context.KindEvent == "issue" {

		issue := h.Payload.(gitlab.IssueEventPayload)

		envs = []string{
			"GITLAB_EVENT_TYPE=" + h.Context.KindEvent,
			"GITLAB_EVENT_USERNAME=" + issue.User.UserName,
			"GITLAB_EVENT_PROJECT_NAME=" + issue.Project.Name,
			"GITLAB_EVENT_PROJECT_GITSSH_URL=" + issue.Project.GitSSSHURL,
			"GITLAB_EVENT_PROJECT_GITHTTP_URL=" + issue.Project.GitHTTPURL,
			"GITLAB_EVENT_REPO_HOMEPAGE=" + issue.Repository.Homepage,
			"GITLAB_EVENT_REPO_URL=" + issue.Repository.URL,
			"GITLAB_EVENT_ISSUE_ID=" + strconv.FormatInt(issue.ObjectAttributes.IID, 10),
			"GITLAB_EVENT_ISSUE_TITLE=" + issue.ObjectAttributes.Title,
			"GITLAB_EVENT_ISSUE_CREATION_DATE=" + fmt.Sprintf("%s", issue.ObjectAttributes.CreatedAt),
			"GITLAB_EVENT_ISSUE_UPDATE_DATE=" + fmt.Sprintf("%s", issue.ObjectAttributes.UpdatedAt),
			"GITLAB_EVENT_ISSUE_AUTHORID=" + strconv.FormatInt(issue.ObjectAttributes.AuthorID, 10),
			"GITLAB_EVENT_ISSUE_DESCR=" + issue.ObjectAttributes.Description,
			"GITLAB_EVENT_ISSUE_STATE=" + issue.ObjectAttributes.State,
			"GITLAB_EVENT_ISSUE_ACTION=" + issue.ObjectAttributes.Action,
		}

		// Addend info to task
		if task.Name == "" {
			task.Name = fmt.Sprintf("%s %s #%d",
				issue.Project.Name, h.Context.KindEvent, issue.ObjectAttributes.IID)
		} else {
			task.Name = fmt.Sprintf("%s - %s %s #%d",
				task.Name, issue.Project.Name, h.Context.KindEvent, issue.ObjectAttributes.IID,
			)
		}
	}

	task.Environment = append(task.Environment, envs...)
}

func (h *GitLabWebHook) SetPendingStatus() {
	// This will be used with integration to Gitlab Job/Pipeline
}

func (h *GitLabWebHook) SetStatus(state, errdescr, targetUrl *string) {
	// This will be used with integration to Gitlab Job/Pipeline
}

func (h *GitLabWebHook) SetFailureStatus(errdescr string) {
	// This will be used with integration to Gitlab Job/Pipeline
}

func (h *GitLabWebHook) GetLogFields(err string) logrus.Fields {
	ans := logrus.Fields{
		"component": "webhook",
		"event":     fmt.Sprintf("gitlab_%s", h.Context.KindEvent),
		"wid":       h.Hook.ID,
	}
	if err != "" {
		ans["error"] = err
	}
	return ans
}

func NewGitContextGitLab(kindEvent string, payload interface{}) *GitContext {
	var ans *GitContext

	if kindEvent == "merge_request" {
		mr := payload.(gitlab.MergeRequestEventPayload)

		repo := mr.Project.Name
		ans = &GitContext{
			Owner:        mr.User.UserName,
			CloneSSHUrl:  mr.ObjectAttributes.Source.GitSSHURL,
			CloneHTTPUrl: mr.ObjectAttributes.Source.GitHTTPURL,
			// TODO: Show what url to use
			UserRepo:  mr.ObjectAttributes.Source.GitHTTPURL,
			User:      strconv.FormatInt(mr.ObjectAttributes.AuthorID, 10),
			Uid:       mr.ObjectAttributes.LastCommit.ID + repo,
			Commit:    mr.ObjectAttributes.LastCommit.ID,
			Checkout:  strconv.FormatInt(mr.ObjectAttributes.IID, 10),
			Repo:      repo,
			Ref:       mr.ObjectAttributes.SourceBranch,
			FilterRef: fmt.Sprintf("%s-%s", kindEvent, mr.ObjectAttributes.SourceBranch),
		}

	} else if kindEvent == "push" {
		push := payload.(gitlab.PushEventPayload)

		repo := push.Project.Name
		ans = &GitContext{
			// with webhooks.v5 use user_username instead of user_name
			Owner: push.UserName,
			// TODO: show if store both git and https urls.
			CloneSSHUrl:  push.Project.GitSSSHURL,
			CloneHTTPUrl: push.Project.GitHTTPURL,
			// TODO: Show what url to use
			UserRepo:  push.Project.GitHTTPURL,
			User:      strconv.FormatInt(push.UserID, 10),
			Uid:       push.CheckoutSHA + repo,
			Commit:    push.CheckoutSHA,
			Checkout:  push.CheckoutSHA,
			Repo:      repo,
			Ref:       push.Ref,
			FilterRef: fmt.Sprintf("%s-%s", kindEvent, push.Ref),
		}

	} else if kindEvent == "tag" {
		tag := payload.(gitlab.TagEventPayload)

		repo := tag.Project.Name
		ans = &GitContext{
			// with webhooks.v5 use user_username instead of user_name
			Owner: tag.UserName,
			// TODO: show if store both git and https urls.
			CloneSSHUrl:  tag.Project.GitSSSHURL,
			CloneHTTPUrl: tag.Project.GitHTTPURL,
			// TODO: Show what url to use
			UserRepo:  tag.Project.GitHTTPURL,
			User:      strconv.FormatInt(tag.UserID, 10),
			Uid:       tag.CheckoutSHA + repo,
			Commit:    tag.CheckoutSHA,
			Checkout:  tag.CheckoutSHA,
			Repo:      repo,
			Ref:       tag.Ref,
			FilterRef: fmt.Sprintf("%s-%s", kindEvent, tag.Ref),
		}

	} else if kindEvent == "issue" {

		issue := payload.(gitlab.IssueEventPayload)

		repo := issue.Project.Name
		ans = &GitContext{
			Owner: issue.Assignee.Username,
			// TODO: show if store both git and https urls.
			CloneSSHUrl:  issue.Project.GitSSSHURL,
			CloneHTTPUrl: issue.Project.GitHTTPURL,
			// TODO: Show what url to use
			UserRepo:  issue.Project.GitHTTPURL,
			User:      strconv.FormatInt(issue.ObjectAttributes.AuthorID, 10),
			Uid:       strconv.FormatInt(issue.ObjectAttributes.IID, 10) + "-issue-" + repo,
			Commit:    issue.Project.DefaultBranch,
			Checkout:  issue.Project.DefaultBranch,
			Repo:      repo,
			Ref:       issue.Project.DefaultBranch,
			FilterRef: fmt.Sprintf("%s-%s", kindEvent, issue.Project.DefaultBranch),
		}
	}

	ans.KindEvent = kindEvent

	return ans
}

func HandleGitLabMergeRequest(m *mottainai.Mottainai, sessionHook *GitLabWebHook) {
	mr := sessionHook.Payload.(gitlab.MergeRequestEventPayload)

	m.Invoke(func(mo *mottainai.Mottainai, l *logging.Logger, db *database.Database) {
		fields := sessionHook.GetLogFields("")
		fields["payload"] = mr
		l.WithFields(fields).Debug(fmt.Sprintf(
			"Merge request for hook id %s received.", sessionHook.Hook.ID,
		))

		if mr.ObjectAttributes.State == "closed" {
			return
		}

		sessionHook.HandleEvent(m, l, db)
	})
}

func HandleGitLabPush(m *mottainai.Mottainai, sessionHook *GitLabWebHook) {
	push := sessionHook.Payload.(gitlab.PushEventPayload)

	m.Invoke(func(mo *mottainai.Mottainai, l *logging.Logger, db *database.Database) {
		fields := sessionHook.GetLogFields("")
		fields["payload"] = push
		l.WithFields(fields).Debug(fmt.Sprintf(
			"Push event request for user %s received.", sessionHook.Hook.ID,
		))

		sessionHook.HandleEvent(m, l, db)
	})
}

func HandleGitLabTag(m *mottainai.Mottainai, sessionHook *GitLabWebHook) {
	tag := sessionHook.Payload.(gitlab.TagEventPayload)

	m.Invoke(func(mo *mottainai.Mottainai, l *logging.Logger, db *database.Database) {
		fields := sessionHook.GetLogFields("")
		fields["payload"] = tag
		l.WithFields(fields).Debug(fmt.Sprintf(
			"Tag event for user %s received.", sessionHook.Hook.ID,
		))

		sessionHook.HandleEvent(m, l, db)
	})
}

func HandleGitLabIssue(m *mottainai.Mottainai, sessionHook *GitLabWebHook) {
	pl := sessionHook.Payload.(gitlab.IssueEventPayload)

	m.Invoke(func(mo *mottainai.Mottainai, l *logging.Logger, db *database.Database) {
		fields := sessionHook.GetLogFields("")
		fields["payload"] = pl
		l.WithFields(fields).Debug(fmt.Sprintf(
			"Issue event for user %s received.", sessionHook.Hook.ID,
		))

		sessionHook.HandleEvent(m, l, db)
	})
}

func GenGitLabHook(db *database.Database, m *mottainai.Mottainai, w *mhook.WebHook,
	u *user.User) *gitlab.Webhook {
	secret := w.Key
	hook := gitlab.New(&gitlab.Config{Secret: secret})

	var appName, buildPath string
	m.Invoke(func(config *setting.Config) {
		appName = config.GetWeb().AppName
		buildPath = config.GetAgent().BuildPath
	})

	uuu, err := db.Driver.GetSettingByKey(setting.SYSTEM_WEBHOOK_PR_ENABLED)
	if err == nil && !uuu.IsDisabled() {
		hook.RegisterEvents(func(payload interface{}, header webhooks.Header) {
			sessionHook := NewGitLabWebHook(payload, w, u, header, "merge_request")
			sessionHook.AppName = appName
			sessionHook.BuildPath = buildPath
			HandleGitLabMergeRequest(m, sessionHook)
		}, gitlab.MergeRequestEvents)
	}

	// Register handler for PushEvents
	hook.RegisterEvents(func(payload interface{}, header webhooks.Header) {
		sessionHook := NewGitLabWebHook(payload, w, u, header, "push")
		sessionHook.AppName = appName
		sessionHook.BuildPath = buildPath
		HandleGitLabPush(m, sessionHook)
	}, gitlab.PushEvents)

	// Register handler for TagEvents
	hook.RegisterEvents(func(payload interface{}, header webhooks.Header) {
		sessionHook := NewGitLabWebHook(payload, w, u, header, "tag")
		sessionHook.AppName = appName
		sessionHook.BuildPath = buildPath
		HandleGitLabTag(m, sessionHook)
	}, gitlab.TagEvents)

	// Register handler for IssuesEvents
	hook.RegisterEvents(func(payload interface{}, header webhooks.Header) {
		sessionHook := NewGitLabWebHook(payload, w, u, header, "issue")
		sessionHook.AppName = appName
		sessionHook.BuildPath = buildPath
		HandleGitLabIssue(m, sessionHook)
	}, gitlab.IssuesEvents)

	return hook
}

func SetupGitLab(m *mottainai.Mottainai) {
	webHookHandler := func(l *logging.Logger, ctx *context.Context,
		db *database.Database, resp http.ResponseWriter, req *http.Request) {
		uid := ctx.Params(":uid")
		l.WithFields(logrus.Fields{
			"component": "webhook",
			"event":     "gitlab_post",
			"uid":       uid,
		}).Debug("Received payload")

		w, err := db.Driver.GetWebHook(uid)
		if err != nil {
			l.WithFields(logrus.Fields{
				"component": "webhook",
				"event":     "gitlab_post",
				"uid":       uid,
			}).Error("No webhook found")
			return
		}
		u, err := db.Driver.GetUser(w.OwnerId)
		if err != nil {
			l.WithFields(logrus.Fields{
				"component": "webhook",
				"event":     "gitlab_post",
				"uid":       uid,
			}).Error("No user found")
			return
		}
		hook := GenGitLabHook(db, m, &w, &u)
		hook.ParsePayload(resp, req)
	}

	m.Invoke(func(config *setting.Config) {
		m.Group(config.GetWeb().GroupAppPath(), func() {
			m.Post("/webhook/:uid/gitlab", RequiresWebHookSetting, webHookHandler)
		})
	})
}
