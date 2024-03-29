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
	//"github.com/MottainaiCI/mottainai-server/pkg/template"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	"github.com/MottainaiCI/mottainai-server/routes/api"
	auth "github.com/MottainaiCI/mottainai-server/routes/auth"

	"github.com/MottainaiCI/mottainai-server/routes/webhook"
	macaron "gopkg.in/macaron.v1"
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
		})
	})

	api.Setup(m)
}
