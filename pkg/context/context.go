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

package context

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"

	logging "github.com/MottainaiCI/mottainai-server/pkg/logging"
	logrus "github.com/sirupsen/logrus"

	auth "github.com/MottainaiCI/mottainai-server/pkg/auth"
	event "github.com/MottainaiCI/mottainai-server/pkg/event"
	user "github.com/MottainaiCI/mottainai-server/pkg/user"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/go-macaron/cache"
	"github.com/go-macaron/csrf"
	"github.com/go-macaron/session"

	macaron "gopkg.in/macaron.v1"
)

// Context represents context of a request.
type Context struct {
	*macaron.Context
	Cache       cache.Cache
	csrf        csrf.CSRF
	Flash       *session.Flash
	Session     session.Store
	Link        string // Current request URL
	IsLogged    bool
	User        *user.User
	IsBasicAuth bool
}

// HTML responses template with given status.
func (c *Context) HTML(status int, name string) {
	c.Context.HTML(status, name)
}

// Title sets "Title" field in template data.
func (c *Context) Title(locale string) {
	c.Data["Title"] = locale
}

// RenderWithErr used for page has form validation but need to prompt error to users.
func (c *Context) RenderWithErr(msg, tpl string) {

	c.Flash.ErrorMsg = msg
	c.Data["Flash"] = c.Flash
	c.HTML(http.StatusOK, tpl)
}

// RenderWithInfo used for page has form validation but need to prompt error to users.
func (c *Context) RenderWithInfo(msg, tpl string) {

	c.Flash.InfoMsg = msg
	c.Data["Flash"] = c.Flash
	c.HTML(http.StatusOK, tpl)
}

// ServerError renders the 500 page.
func (c *Context) ServerError(title string, err error) {
	var webconfig *setting.WebConfig
	c.Invoke(func(config *setting.Config) {
		webconfig = config.GetWeb()
	})

	// Restrict API calls with error message.
	if auth.IsAPIPath(c.Req.URL.Path, webconfig) {
		c.JSON(500,
			event.APIResponse{
				Error:     err.Error(),
				Status:    "error",
				Processed: "true",
			})
		return
	}

	c.Data["ErrorMsg"] = err
	c.Handle(http.StatusInternalServerError, title, err)
}

func (c *Context) NotFound() {
	var webconfig *setting.WebConfig
	err := "Page not found"

	c.Invoke(func(config *setting.Config) {
		webconfig = config.GetWeb()
	})
	// Restrict API calls with error message.
	if auth.IsAPIPath(c.Req.URL.Path, webconfig) {
		c.JSON(404, event.APIResponse{
			Error:     err,
			Status:    "error",
			Processed: "true",
		})
		return
	}

	c.Data["Title"] = err
	c.Handle(http.StatusNotFound, err, errors.New(err))
}

// Handle handles and logs error by given status.
func (c *Context) Handle(status int, title string, err error) {
	switch status {
	case http.StatusNotFound:
		c.Data["Title"] = "Page Not Found"
	case http.StatusInternalServerError:
		c.Data["Title"] = "Internal Server Error"
		c.Invoke(func(logger *logging.Logger) {
			logger.WithFields(logrus.Fields{
				"component": "core",
				"error":     err,
			}).Error("Error on ", title)
		})
		if c.IsLogged && c.User.IsAdmin() {
			c.Data["ErrorMsg"] = err
		}
	}
	c.HTML(status, fmt.Sprintf("status/%d", status))
}

// FormErr sets "Err_xxx" field in template data.
func (c *Context) FormErr(names ...string) {
	for i := range names {
		c.Data["Err_"+names[i]] = true
	}
}

// HasError returns true if error occurs in form validation.
func (c *Context) HasError() bool {
	hasErr, ok := c.Data["HasError"]
	if !ok {
		return false
	}
	c.Flash.ErrorMsg = c.Data["ErrorMsg"].(string)
	c.Data["Flash"] = c.Flash
	return hasErr.(bool)
}

func (c *Context) NoPermission() {
	var webconfig *setting.WebConfig
	c.Invoke(func(config *setting.Config) {
		webconfig = config.GetWeb()
	})
	if auth.IsAPIPath(c.Req.URL.Path, webconfig) {
		c.JSON(403, noperm)
	} else {
		c.Error(403)
	}
}

// PageIs sets "PageIsxxx" field in template data.
func (c *Context) PageIs(name string) {
	c.Data["PageIs"+name] = true
}

// Success responses template with status http.StatusOK.
func (c *Context) Success(name string) {
	c.HTML(http.StatusOK, name)
}

// JSONSuccess responses JSON with status http.StatusOK.
func (c *Context) JSONSuccess(data interface{}) {
	c.JSON(http.StatusOK, data)
}

// SubURLRedirect responses redirection wtih given location and status.
// It prepends setting.AppSubURL to the location string.
func (c *Context) SubURLRedirect(location string, status ...int) {
	c.Invoke(func(config *setting.Config) {
		c.Redirect(config.GetWeb().BuildURI(location))
	})
}

func (c *Context) ServeContent(name string, r io.ReadSeeker, params ...interface{}) {
	modtime := time.Now()
	for _, p := range params {
		switch v := p.(type) {
		case time.Time:
			modtime = v
		}
	}
	c.Resp.Header().Set("Content-Description", "File Transfer")
	c.Resp.Header().Set("Content-Type", "application/octet-stream")
	c.Resp.Header().Set("Content-Disposition", "attachment; filename="+name)
	c.Resp.Header().Set("Content-Transfer-Encoding", "binary")
	c.Resp.Header().Set("Expires", "0")
	c.Resp.Header().Set("Cache-Control", "must-revalidate")
	c.Resp.Header().Set("Pragma", "public")
	http.ServeContent(c.Resp, c.Req.Request, name, modtime, r)
}

func Setup(m *macaron.Macaron) {
	m.Use(Contexter())
}

func Contexter() macaron.Handler {
	return func(ctx *macaron.Context, config *setting.Config, sess session.Store, f *session.Flash, x csrf.CSRF, cache cache.Cache) {
		c := &Context{
			Context: ctx,
			Cache:   cache,
			csrf:    x,
			Flash:   f,
			Session: sess,
			Link:    ctx.Req.URL.Path,
		}
		c.Data["Link"] = c.Link
		c.Data["PageStartTime"] = time.Now()

		if len(config.GetWeb().AccessControlAllowOrigin) > 0 {
			// Set CORS headers for browser-based git clients
			ctx.Resp.Header().Set("Access-Control-Allow-Origin", config.GetWeb().AccessControlAllowOrigin)
			ctx.Resp.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			ctx.Header().Set("Access-Control-Allow-Origin", config.GetWeb().AccessControlAllowOrigin)
			c.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Header().Set("Access-Control-Max-Age", "3600")
			c.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
		}

		// Get user from session if logged in
		c.User, c.IsBasicAuth = auth.SignedInUser(c.Context, c.Session)
		if c.User != nil {
			c.IsLogged = true
			c.Data["IsLogged"] = c.IsLogged
			c.Data["LoggedUser"] = c.User
			c.Data["LoggedUserID"] = c.User.ID
			c.Data["LoggedUserName"] = c.User.Name
			c.Data["IsAdmin"] = c.User.Admin
			c.Data["IsManager"] = c.User.Manager
			if c.User.IsManager() || c.User.IsAdmin() {
				c.Data["IsManagerOrAdmin"] = true
			} else {
				c.Data["IsManagerOrAdmin"] = false
			}
		} else {
			c.Data["LoggedUserID"] = 0
			c.Data["LoggedUserName"] = ""
			c.Data["IsAdmin"] = "no"
		}

		c.Data["CSRFToken"] = x.GetToken()
		c.Data["CSRFTokenHTML"] = template.HTML(`<input type="hidden" name="_csrf" value="` + x.GetToken() + `">`)

		ctx.Map(c)
	}
}
