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
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/MottainaiCI/mottainai-server/pkg/utils"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	mottainai "github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	tasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	ggithub "github.com/google/go-github/github"
	webhooks "gopkg.in/go-playground/webhooks.v3"
	macaron "gopkg.in/macaron.v1"

	"gopkg.in/go-playground/webhooks.v3/github"
)

var (
	pending      = "pending"
	success      = "success"
	failure      = "error"
	targetUrl    = "https://xyz.ngrok.com/status"
	pendingDesc  = "Build/testing in progress, please wait."
	successDesc  = "Build/testing successful."
	failureDesc  = "Build or Unit Test failed."
	notfoundDesc = "No mottainai.json found on repo"
	appName      = "MottainaiCI"
)

// HandlePullRequest handles GitHub pull_request events
func HandlePullRequest(payload interface{}, header webhooks.Header, m *mottainai.Mottainai) {

	fmt.Println("Handling Pull Request")

	pl := payload.(github.PullRequestPayload)

	fmt.Printf("%+v", pl)

	m.Invoke(func(client *ggithub.Client, db *database.Database) {

		// Create the 'pending' status and send it
		status1 := &ggithub.RepoStatus{State: &pending, TargetURL: &targetUrl, Description: &pendingDesc, Context: &appName}

		client.Repositories.CreateStatus(stdctx.Background(), pl.PullRequest.Base.User.Login, pl.PullRequest.Base.Repo.Name, pl.PullRequest.Head.Sha, status1)

		repo := pl.PullRequest.Base.Repo.CloneURL
		//commit := pl.PullRequest.Head.Sha
		number := pl.PullRequest.Number
		gitdir, err := ioutil.TempDir(setting.Configuration.TempWorkDir, "git"+pl.PullRequest.Head.Sha)
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(gitdir)

		out, err := utils.Git([]string{"clone", repo, gitdir}, ".")
		fmt.Println(out)
		if err != nil {
			panic(err)
		}

		out, err = utils.Git([]string{"fetch", "origin", "pull/" + strconv.FormatInt(number, 10) + "/head:CI_test"}, gitdir)
		fmt.Println(out)
		if err != nil {
			panic(err)
		}

		out, err = utils.Git([]string{"checkout", "CI_test"}, gitdir)
		fmt.Println(out)
		if err != nil {
			panic(err)
		}

		if exists, _ := utils.Exists(path.Join(gitdir, "mottainai.json")); exists == true {

			t := tasks.FromFile(path.Join(gitdir, "mottainai.json"))
			t.Namespace = "" // do not allow automatic tag from PR
			log.Println(t)

			docID, err := db.CreateTask(t.ToMap())
			if err != nil {
				return "", err
			}
			m.SendTask(docID)

			// We should update the status when we know the result of the task.
			log.Println("Returning Success")
			status2 := &ggithub.RepoStatus{State: &success, TargetURL: &targetUrl, Description: &successDesc, Context: &appName}
			client.Repositories.CreateStatus(stdctx.Background(), pl.PullRequest.Base.User.Login, pl.PullRequest.Base.Repo.Name, pl.PullRequest.Head.Sha, status2)

		} else {
			// Create the 'success' status and send it
			log.Println("mottainai.json not present")
			status2 := &ggithub.RepoStatus{State: &failure, TargetURL: &targetUrl, Description: &notfoundDesc, Context: &appName}
			client.Repositories.CreateStatus(stdctx.Background(), pl.PullRequest.Base.User.Login, pl.PullRequest.Base.Repo.Name, pl.PullRequest.Head.Sha, status2)

		}

	})
}

func SetupGitHub(m *macaron.Macaron) {

	hook := github.New(&github.Config{Secret: setting.Configuration.WebHookGitHubSecret})
	hook.RegisterEvents(func(payload interface{}, header webhooks.Header) {
		fmt.Println("Received webhook for PR")
		m.Invoke(func(mo *mottainai.Mottainai) { HandlePullRequest(payload, header, mo) })
	}, github.PullRequestEvent)

	m.Post("/webhook/github", func(ctx *context.Context, db *database.Database, resp http.ResponseWriter, req *http.Request) {
		hook.ParsePayload(resp, req)
	})

}
