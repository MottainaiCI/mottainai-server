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

package routes

import (
	context "github.com/MottainaiCI/mottainai-server/pkg/context"
	"github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/MottainaiCI/mottainai-server/pkg/template"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	"github.com/MottainaiCI/mottainai-server/routes/api"
	auth "github.com/MottainaiCI/mottainai-server/routes/auth"
	namespaceroute "github.com/MottainaiCI/mottainai-server/routes/namespaces"
	nodesroute "github.com/MottainaiCI/mottainai-server/routes/nodes"
	tokenroute "github.com/MottainaiCI/mottainai-server/routes/token"

	"github.com/MottainaiCI/mottainai-server/routes/webhook"
	macaron "gopkg.in/macaron.v1"

	"github.com/MottainaiCI/mottainai-server/routes/tasks"
)

func NotFound(c *context.Context) {
	c.NotFound()
}

func ServerError(c *context.Context, e error) {
	c.ServerError("Internal server error", e)
}

func SetupDaemon(m *mottainai.Mottainai) *mottainai.Mottainai {
	api.Setup(m.Macaron)
	return m
}

func SetupWebHookServer(m *mottainai.WebHookServer) *mottainai.WebHookServer {
	webhook.Setup(m.Mottainai)
	m.Invoke(webhook.GlobalWatcher)
	return m
}

func AddWebHook(m *mottainai.Mottainai) {
	webhook.Setup(m)
	m.Invoke(webhook.GlobalWatcher)
}

func SetupWebUI(m *mottainai.Mottainai) *mottainai.Mottainai {
	Setup(m.Macaron)
	auth.Setup(m.Macaron)

	if setting.Configuration.EmbedWebHookServer {
		AddWebHook(m)
	}
	return m
}

func Setup(m *macaron.Macaron) {

	m.NotFound(NotFound)
	m.InternalServerError(ServerError)

	// setup templates
	// m.Use(macaron.Renderer())

	m.Get("/", func(ctx *context.Context, db *database.Database) error {
		rtasks, e := db.Driver.FindDoc("Tasks", `[{"eq": "running", "in": ["status"]}]`)
		if e != nil {
			return e
		}
		running_tasks := len(rtasks)
		wtasks, e := db.Driver.FindDoc("Tasks", `[{"eq": "waiting", "in": ["status"]}]`)
		if e != nil {
			return e
		}
		waiting_tasks := len(wtasks)
		etasks, e := db.Driver.FindDoc("Tasks", `[{"eq": "error", "in": ["result"]}]`)
		if e != nil {
			return e
		}
		error_tasks := len(etasks)
		ftasks, e := db.Driver.FindDoc("Tasks", `[{"eq": "failed", "in": ["result"]}]`)
		if e != nil {
			return e
		}
		failed_tasks := len(ftasks)
		stasks, e := db.Driver.FindDoc("Tasks", `[{"eq": "success", "in": ["result"]}]`)
		if e != nil {
			return e
		}
		succeeded_tasks := len(stasks)

		ctx.Data["TotalTasks"] = len(db.Driver.ListDocs("Tasks"))

		ctx.Data["RunningTasks"] = running_tasks
		ctx.Data["WaitingTasks"] = waiting_tasks
		ctx.Data["ErroredTasks"] = error_tasks
		ctx.Data["SucceededTasks"] = succeeded_tasks
		ctx.Data["FailedTasks"] = failed_tasks

		template.TemplatePreview(ctx, "index", db.Config)
		return nil
	})

	tasks.Setup(m)
	nodesroute.Setup(m)
	namespaceroute.Setup(m)
	tokenroute.Setup(m)
	api.Setup(m)
}
