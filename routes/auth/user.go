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

package auth

import (
	"fmt"
	"net/url"

	user "github.com/MottainaiCI/mottainai-server/pkg/user"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"

	"github.com/go-macaron/captcha"
	log "gopkg.in/clog.v1"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
)

const (
	LOGIN  = "user/auth/login"
	SIGNUP = "user/auth/signup"
)

// AutoLogin reads cookie and try to auto-login.
func AutoLogin(c *context.Context, db *database.Database) (bool, error) {

	uname := c.GetCookie("u_name")
	if len(uname) == 0 {
		return false, nil
	}

	isSucceed := false
	defer func() {
		if !isSucceed {
			log.Trace("auto-login cookie cleared: %s", uname)
			c.SetCookie("u_name", "", -1, setting.Configuration.AppSubURL)
			c.SetCookie("r_name", "", -1, setting.Configuration.AppSubURL)
			c.SetCookie("s_name", "", -1, setting.Configuration.AppSubURL)
		}
	}()

	u, err := db.GetUserByName(uname)
	if err != nil {

		return false, fmt.Errorf("GetUserByName: %v", err)

	}

	if val, ok := c.GetSuperSecureCookie(u.Password, "r_name"); !ok || val != u.Name {
		return false, nil
	}

	isSucceed = true
	c.Session.Set("uid", u.ID)
	c.Session.Set("uname", u.Name)
	c.SetCookie("_csrf", "", -1, setting.Configuration.AppSubURL)
	//	if setting.EnableLoginStatusCookie {
	//		c.SetCookie(setting.LoginStatusCookieName, "true", 0, setting.Configuration.AppSubURL)
	//	}
	return true, nil
}

// isValidRedirect returns false if the URL does not redirect to same site.
// False: //url, http://url
// True: /url
func isValidRedirect(url string) bool {
	return len(url) >= 2 && url[0] == '/' && url[1] != '/'
}

func Login(c *context.Context, db *database.Database) {
	c.Title("sign_in")

	// Check auto-login
	isSucceed, err := AutoLogin(c, db)
	if err != nil {
		c.ServerError("AutoLogin", err)
		return
	}

	redirectTo := c.Query("redirect_to")
	if len(redirectTo) > 0 {
		c.SetCookie("redirect_to", redirectTo, 0, setting.Configuration.AppSubURL)
	} else {
		redirectTo, _ = url.QueryUnescape(c.GetCookie("redirect_to"))
	}

	if isSucceed {
		if isValidRedirect(redirectTo) {
			c.Redirect(redirectTo)
		} else {
			c.SubURLRedirect("/")
		}
		c.SetCookie("redirect_to", "", -1, setting.Configuration.AppSubURL)
		return
	}

	c.Success(LOGIN)
}

func afterLogin(c *context.Context, u user.User, remember bool) {
	if remember {
		days := 86400 * 30
		c.SetCookie("u_name", u.Name, days, setting.Configuration.AppSubURL, "", true, true)
		c.SetSuperSecureCookie(u.Password, "r_name", u.Name, days, setting.Configuration.AppSubURL, "", true, true)
	}

	c.Session.Set("uid", u.ID)
	c.Session.Set("uname", u.Name)

	// Clear whatever CSRF has right now, force to generate a new one
	c.SetCookie("_csrf", "", -1, setting.Configuration.AppSubURL)

	redirectTo, _ := url.QueryUnescape(c.GetCookie("redirect_to"))
	c.SetCookie("redirect_to", "", -1, setting.Configuration.AppSubURL)
	if isValidRedirect(redirectTo) {
		c.Redirect(redirectTo)
		return
	}

	c.SubURLRedirect("/")
}

func LoginPost(c *context.Context, f SignIn, db *database.Database) {
	c.Title("Sign in")

	if c.HasError() {
		c.Success(LOGIN)
		return
	}

	u, err := db.SignIn(f.UserName, f.Password)
	if err != nil {
		c.ServerError("UserLogin", err)
		return
	}

	afterLogin(c, u, f.Remember)
	return

}

func SignOut(c *context.Context) {
	c.Session.Delete("uid")
	c.Session.Delete("uname")
	c.SetCookie("u_name", "", -1, setting.Configuration.AppSubURL)
	c.SetCookie("r_name", "", -1, setting.Configuration.AppSubURL)
	c.SetCookie("_csrf", "", -1, setting.Configuration.AppSubURL)
	c.SubURLRedirect("/")
}

func SignUp(c *context.Context) {
	c.Title("Sign up")

	c.Data["EnableCaptcha"] = true

	c.Success(SIGNUP)
}

func SignUpPost(c *context.Context, cpt *captcha.Captcha, f Register, db *database.Database) {
	c.Title("sign_up")

	c.Data["EnableCaptcha"] = true

	if c.HasError() {
		c.Success(SIGNUP)
		return
	}
	//Captcha
	if !cpt.VerifyReq(c.Req) {
		c.FormErr("Captcha")
		return
	}

	if f.Password != f.Retype {
		c.FormErr("Password")
		return
	}

	u := &user.User{
		Name:     f.UserName,
		Email:    f.Email,
		Password: f.Password,
		//IsActive: !setting.Service.RegisterEmailConfirm,
	}
	if db.CountUsers() == 1 {
		u.MakeAdmin()
	}
	if _, err := db.InsertAndSaltUser(u); err != nil {
		c.ServerError("CreateUser", err)
		return
	}
	log.Trace("Account created: %s", u.Name)

	c.SubURLRedirect("/user/login")
}
