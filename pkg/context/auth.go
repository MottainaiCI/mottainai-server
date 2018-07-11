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
	"net/url"

	"github.com/go-macaron/csrf"
	macaron "gopkg.in/macaron.v1"

	auth "github.com/MottainaiCI/mottainai-server/pkg/auth"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
)

type ToggleOptions struct {
	SignInRequired  bool
	SignOutRequired bool
	AdminRequired   bool
	ManagerRequired bool
	DisableCSRF     bool
}

func Toggle(options *ToggleOptions) macaron.Handler {
	return func(c *Context) {

		// Redirect to dashboard if user tries to visit any non-login page.
		if options.SignOutRequired && c.IsLogged && c.Req.RequestURI != "/" {
			c.Redirect(setting.Configuration.AppSubURL + "/")
			return
		}

		if !options.SignOutRequired && !options.DisableCSRF && c.Req.Method == "POST" && !auth.IsAPIPath(c.Req.URL.Path) {
			csrf.Validate(c.Context, c.csrf)
			if c.Written() {
				return
			}
		}
		if options.SignInRequired {
			if !c.IsLogged {
				// Restrict API calls with error message.
				if auth.IsAPIPath(c.Req.URL.Path) {
					c.JSON(403, map[string]string{
						"message": "Only signed in user is allowed to call APIs.",
					})
					return
				}

				c.SetCookie("redirect_to", url.QueryEscape(setting.Configuration.AppSubURL+c.Req.RequestURI), 0, setting.Configuration.AppSubURL)
				c.Redirect(setting.Configuration.AppSubURL + "/user/login")
				return
			}
		}

		// Redirect to log in page if auto-signin info is provided and has not signed in.
		if !options.SignOutRequired && !c.IsLogged && !auth.IsAPIPath(c.Req.URL.Path) &&
			len(c.GetCookie("u_name")) > 0 {
			c.SetCookie("redirect_to", url.QueryEscape(setting.Configuration.AppSubURL+c.Req.RequestURI), 0, setting.Configuration.AppSubURL)
			c.Redirect(setting.Configuration.AppSubURL + "/user/login")
			return
		}

		if options.ManagerRequired {
			if !c.User.IsManager() && !c.User.IsAdmin() {

				c.NoPermission()
				return
			}
			c.Data["PageIsManager"] = true
		}

		if options.AdminRequired {
			if !c.User.IsAdmin() {

				c.NoPermission()
				return
			}
			c.Data["PageIsAdmin"] = true
		}
	}
}
