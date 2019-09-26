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
	"net/http"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	logging "github.com/MottainaiCI/mottainai-server/pkg/logging"
	mottainai "github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	user "github.com/MottainaiCI/mottainai-server/pkg/user"
	mhook "github.com/MottainaiCI/mottainai-server/pkg/webhook"

	logrus "github.com/sirupsen/logrus"
	webhooks "gopkg.in/go-playground/webhooks.v3"
	gitlab "gopkg.in/go-playground/webhooks.v3/gitlab"
)

func HandleGitLabMergeRequest(u *user.User, payload interface{}, header webhooks.Header,
	m *mottainai.Mottainai, w *mhook.WebHook) {
}

func HandleGitLabPush(u *user.User, payload interface{}, header webhooks.Header,
	m *mottainai.Mottainai, w *mhook.WebHook) {
}

func HandleGitLabTag(u *user.User, payload interface{}, header webhooks.Header,
	m *mottainai.Mottainai, w *mhook.WebHook) {
}

func HandleGitLabIssue(u *user.User, payload interface{}, header webhooks.Header,
	m *mottainai.Mottainai, w *mhook.WebHook) {
}

func GenGitLabHook(db *database.Database, m *mottainai.Mottainai, w *mhook.WebHook,
	u *user.User) *gitlab.Webhook {
	secret := w.Key
	hook := gitlab.New(&gitlab.Config{Secret: secret})

	uuu, err := db.Driver.GetSettingByKey(setting.SYSTEM_WEBHOOK_PR_ENABLED)
	if err == nil && !uuu.IsDisabled() {
		hook.RegisterEvents(func(payload interface{}, header webhooks.Header) {
			HandleGitLabMergeRequest(u, payload, header, m, w)
		}, gitlab.MergeRequestEvents)
	}

	// Register handler for PushEvents
	hook.RegisterEvents(func(payload interface{}, header webhooks.Header) {
		HandleGitLabPush(u, payload, header, m, w)
	}, gitlab.PushEvents)

	// Register handler for TagEvents
	hook.RegisterEvents(func(payload interface{}, header webhooks.Header) {
		HandleGitLabTag(u, payload, header, m, w)
	}, gitlab.TagEvents)

	// Register handler for IssuesEvents
	hook.RegisterEvents(func(payload interface{}, header webhooks.Header) {
		HandleGitLabIssue(u, payload, header, m, w)
	}, gitlab.IssuesEvents)

	return hook
}

func SetupGitLabHub(m *mottainai.Mottainai) {
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
