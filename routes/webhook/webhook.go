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
	"log"
	"net/http"
	"time"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	mottainai "github.com/MottainaiCI/mottainai-server/pkg/mottainai"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	ggithub "github.com/google/go-github/github"
	webhooks "gopkg.in/go-playground/webhooks.v3"
	macaron "gopkg.in/macaron.v1"

	"gopkg.in/go-playground/webhooks.v3/github"
)

func Setup(m *macaron.Macaron) {

	hook := github.New(&github.Config{Secret: setting.Configuration.WebHookGitHubSecret})
	hook.RegisterEvents(func(payload interface{}, header webhooks.Header) {
		fmt.Println("Received webhook for PR")
		m.Invoke(func(mo *mottainai.Mottainai) { HandlePullRequest(payload, header, mo) })
	}, github.PullRequestEvent)

	m.Post("/webhook/github", func(ctx *context.Context, db *database.Database, resp http.ResponseWriter, req *http.Request) {
		hook.ParsePayload(resp, req)
	})

}

var (
	pending     = "pending"
	success     = "success"
	failure     = "error"
	targetUrl   = "https://xyz.ngrok.com/status"
	pendingDesc = "Build/testing in progress, please wait."
	successDesc = "Build/testing successful."
	failureDesc = "Build or Unit Test failed."
	appName     = "MottainaiCI"
)

// HandlePullRequest handles GitHub pull_request events
func HandlePullRequest(payload interface{}, header webhooks.Header, m *mottainai.Mottainai) {

	fmt.Println("Handling Pull Request")

	pl := payload.(github.PullRequestPayload)

	// Do whatever you want from here...
	fmt.Printf("%+v", pl)

	m.Invoke(func(client *ggithub.Client) {

		// Create the 'pending' status and send it
		status1 := &ggithub.RepoStatus{State: &pending, TargetURL: &targetUrl, Description: &pendingDesc, Context: &appName}

		client.Repositories.CreateStatus(stdctx.Background(), pl.PullRequest.Base.User.Login, pl.PullRequest.Base.Repo.Name, pl.PullRequest.Head.Sha, status1)

		// sleep for 30 seconds
		time.Sleep(30 * time.Second)

		// Create the 'success' status and send it
		log.Println("Returning Success")
		status2 := &ggithub.RepoStatus{State: &success, TargetURL: &targetUrl, Description: &successDesc, Context: &appName}
		client.Repositories.CreateStatus(stdctx.Background(), pl.PullRequest.Base.User.Login, pl.PullRequest.Base.Repo.Name, pl.PullRequest.Head.Sha, status2)

	})
}
