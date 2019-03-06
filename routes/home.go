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
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

	"github.com/MottainaiCI/mottainai-server/routes/plans"
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

	m.Invoke(func(config *setting.Config) {
		if config.GetWeb().EmbedWebHookServer {
			AddWebHook(m)
		}
	})
	return m
}

func writeImage(w http.ResponseWriter, req *http.Request, img, url string) error {

	info, err := os.Stat(img)
	if err != nil || info == nil || info.IsDir() {
		return err
	}

	file, err := ioutil.ReadFile(img)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(file)
	ext := filepath.Ext(strings.TrimSpace(img))

	if ext == ".png" {
		w.Header().Set("Content-Type", "image/png")
	} else if ext == ".jpeg" {
		w.Header().Set("Content-Type", "image/jpeg")
	} else if ext == ".ico" {
		w.Header().Set("Content-Type", "image/x-icon")
	} else {
		return errors.New("File format specified for " + img + " is not supported")
	}
	http.ServeContent(w, req, url, info.ModTime(), reader)
	return nil
}

func Setup(m *macaron.Macaron) {

	m.NotFound(NotFound)
	m.InternalServerError(ServerError)

	// setup templates
	// m.Use(macaron.Renderer())

	m.Invoke(func(config *setting.Config) {
		m.Group(config.GetWeb().GroupAppPath(), func() {
			m.Get("/favicon", func(ctx *context.Context, db *database.Database) error {
				if config.GetWeb().AppBrandingFavicon != "" {
					return writeImage(ctx.Resp, ctx.Req.Request, config.GetWeb().AppBrandingFavicon, "/images/favicon")
				}
				ctx.Redirect(config.GetWeb().BuildURI("/favicon.ico"))
				return nil
			})
			m.Get("/images/logo", func(ctx *context.Context, db *database.Database) error {
				if config.GetWeb().AppBrandingLogo != "" {
					return writeImage(ctx.Resp, ctx.Req.Request, config.GetWeb().AppBrandingLogo, "/images/logo")
				}
				ctx.Redirect(config.GetWeb().BuildURI("/images/mottainai_logo.png"))
				return nil
			})
			m.Get("/images/logo_small", func(ctx *context.Context, db *database.Database) error {
				if config.GetWeb().AppBrandingLogoSmall != "" {
					return writeImage(ctx.Resp, ctx.Req.Request, config.GetWeb().AppBrandingLogoSmall, "/images/logo_small")
				}
				ctx.Redirect(config.GetWeb().BuildURI("/images/mottainai_logo_small.png"))
				return nil
			})
			m.Get("/", func(ctx *context.Context, db *database.Database) error {
				rtasks, e := db.Driver.GetTaskByStatus(db.Config, "running")
				if e != nil {
					return e
				}
				running_tasks := len(rtasks)
				wtasks, e := db.Driver.GetTaskByStatus(db.Config, "waiting")
				if e != nil {
					return e
				}
				waiting_tasks := len(wtasks)
				etasks, e := db.Driver.GetTaskByStatus(db.Config, "error")
				if e != nil {
					return e
				}
				error_tasks := len(etasks)
				ftasks, e := db.Driver.GetTaskByStatus(db.Config, "failed")
				if e != nil {
					return e
				}
				failed_tasks := len(ftasks)
				stasks, e := db.Driver.GetTaskByStatus(db.Config, "success")
				if e != nil {
					return e
				}
				succeeded_tasks := len(stasks)
				stoppedtasks, e := db.Driver.GetTaskByStatus(db.Config, "stopped")
				if e != nil {
					return e
				}
				instoptasks, e := db.Driver.GetTaskByStatus(db.Config, "stop")
				if e != nil {
					return e
				}

				stopped_tasks := len(stoppedtasks)
				instop_tasks := len(instoptasks)

				ctx.Data["TotalTasks"] = len(db.Driver.ListDocs("Tasks"))
				ctx.Data["RunningTasks"] = running_tasks
				ctx.Data["WaitingTasks"] = waiting_tasks
				ctx.Data["ErroredTasks"] = error_tasks
				ctx.Data["SucceededTasks"] = succeeded_tasks
				ctx.Data["FailedTasks"] = failed_tasks
				ctx.Data["StoppedTasks"] = stopped_tasks
				ctx.Data["InStopTasks"] = instop_tasks

				template.TemplatePreview(ctx, "index", db.Config)
				return nil
			})
		})
	})

	tasks.Setup(m)
	plans.Setup(m)
	nodesroute.Setup(m)
	namespaceroute.Setup(m)
	tokenroute.Setup(m)
	api.Setup(m)
}
