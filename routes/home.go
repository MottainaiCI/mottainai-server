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
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

func writeImage(w http.ResponseWriter, img string) error {

	infile, err := os.Open(img)
	if err != nil {
		// replace this with real error handling
		return err
	}
	defer infile.Close()
	ext := filepath.Ext(strings.TrimSpace(img))
	// Decode will figure out what type of image is in the file on its own.
	// We just have to be sure all the image packages we want are imported.
	src, _, err := image.Decode(infile)
	if err != nil {
		return err
	}

	buffer := new(bytes.Buffer)

	if ext == ".png" {
		if err := png.Encode(buffer, src); err != nil {
			return err
		}

		w.Header().Set("Content-Type", "image/png")
	} else if ext == ".jpeg" {
		if err := jpeg.Encode(buffer, src, nil); err != nil {
			return err
		}

		w.Header().Set("Content-Type", "image/jpeg")
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		return err
	}
	return nil
}

func Setup(m *macaron.Macaron) {

	m.NotFound(NotFound)
	m.InternalServerError(ServerError)

	// setup templates
	// m.Use(macaron.Renderer())

	m.Invoke(func(config *setting.Config) {
		m.Group(config.GetWeb().GroupAppPath(), func() {
			m.Get("/images/logo", func(ctx *context.Context, db *database.Database) error {
				if config.GetWeb().AppBrandingLogo != "" {
					return writeImage(ctx.Resp, config.GetWeb().AppBrandingLogo)

				}
				ctx.Redirect("/images/mottainai_logo.png")
				return nil
				//return writeImage(ctx.Resp, path.Join(config.GetWeb().StaticRootPath, "public", "images", "mottainai_logo.png"))
			})
			m.Get("/images/logo_small", func(ctx *context.Context, db *database.Database) error {
				if config.GetWeb().AppBrandingLogoSmall != "" {
					return writeImage(ctx.Resp, config.GetWeb().AppBrandingLogoSmall)

				}
				ctx.Redirect("/images/mottainai_logo_small.png")
				return nil
				//return writeImage(ctx.Resp, path.Join(config.GetWeb().StaticRootPath, "public", "images", "mottainai_logo_small.png"))
			})
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
		})
	})

	tasks.Setup(m)
	nodesroute.Setup(m)
	namespaceroute.Setup(m)
	tokenroute.Setup(m)
	api.Setup(m)
}
