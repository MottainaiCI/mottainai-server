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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"

	user "github.com/MottainaiCI/mottainai-server/pkg/user"
	mhook "github.com/MottainaiCI/mottainai-server/pkg/webhook"

	anagent "github.com/mudler/anagent"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	mottainai "github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	utils "github.com/MottainaiCI/mottainai-server/pkg/utils"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	ggithub "github.com/google/go-github/github"
	webhooks "gopkg.in/go-playground/webhooks.v3"

	"gopkg.in/go-playground/webhooks.v3/github"
)

// HandlePullRequest handles GitHub pull_request events
func HandlePullRequest(u *user.User, payload interface{}, header webhooks.Header, m *mottainai.Mottainai) {

	fmt.Println("Handling Pull Request")
	pl := payload.(github.PullRequestPayload)
	if pl.Action == "closed" {
		return
	}
	m.Invoke(func(mo *mottainai.Mottainai, client *ggithub.Client, db *database.Database) {
		if err := SendTask(u, "pull_request", client, db, mo, payload); err != nil {
			fmt.Println("Failed sending task", err)
		}
		if err := SendPipeline(u, "pull_request", client, db, mo, payload); err != nil {
			fmt.Println("Failed sending task", err)
		}
	})
}

// HandlePush handles GitHub push events
func HandlePush(u *user.User, payload interface{}, header webhooks.Header, m *mottainai.Mottainai) {

	fmt.Println("Handling Push")

	m.Invoke(func(client *ggithub.Client, db *database.Database) {

		if err := SendTask(u, "push", client, db, m, payload); err != nil {
			fmt.Println("Failed sending task", err)
		}
		if err := SendPipeline(u, "push", client, db, m, payload); err != nil {
			fmt.Println("Failed sending task", err)
		}
	})
}

func prepareTemp(u *user.User, kind string, client *ggithub.Client, db *database.Database, m *mottainai.Mottainai, payload interface{}) (*GitContext, error) {
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
	ctx.StoredUser = u

	gitdir, err := ioutil.TempDir(setting.Configuration.BuildPath, path.Join("webhook_fetch", repo))
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

//TODO: handle this with separate objs
type GitHubWebHook struct{}

func GenGitHubHook(m *mottainai.Mottainai, w *mhook.WebHook, u *user.User) *github.Webhook {
	secret := w.Key
	//owner := u
	hook := github.New(&github.Config{Secret: secret})
	hook.RegisterEvents(func(payload interface{}, header webhooks.Header) {
		fmt.Println("Received webhook for PR")
		HandlePullRequest(u, payload, header, m)
	}, github.PullRequestEvent)

	hook.RegisterEvents(func(payload interface{}, header webhooks.Header) {
		fmt.Println("Received webhook for push")
		HandlePush(u, payload, header, m)
	}, github.PushEvent)
	return hook
}

func SetupGitHub(m *mottainai.Mottainai) {

	m.Invoke(func(client *ggithub.Client, a *anagent.Anagent, db *database.Database) {
		GlobalWatcher(client, a, db, m.Url())
	})

	// TODO: Generate tokens for  each user.
	// Let user add repo in specific collection, and check against that
	m.Post("/webhook/:uid/github", RequiresWebHookSetting, func(ctx *context.Context, db *database.Database, resp http.ResponseWriter, req *http.Request) {
		uid := ctx.Params(":uid")
		fmt.Println("Received payload for ", uid)
		w, err := db.Driver.GetWebHook(uid)
		if err != nil {
			fmt.Println("No webhook found for ", uid)

			return
		}
		u, err := db.Driver.GetUser(w.OwnerId)
		if err != nil {
			fmt.Println("No user found for ", uid)

			return
		}
		hook := GenGitHubHook(m, &w, &u)
		hook.ParsePayload(resp, req)
	})

}
