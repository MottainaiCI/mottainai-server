/*

Copyright (C) 2017-2018  Ettore Di Giacinto <mudler@gentoo.org>
Copyright (C) 2020       Adib Saad <adib.saad@gmail.com>
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
  "github.com/MottainaiCI/mottainai-server/pkg/context"
  database "github.com/MottainaiCI/mottainai-server/pkg/db"
  setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
  user "github.com/MottainaiCI/mottainai-server/pkg/user"
  "github.com/go-macaron/binding"
  "gopkg.in/macaron.v1"

  v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"
)

type SignIn struct {
  UserName    string `binding:"Required;MaxSize(254)"`
  Password    string `binding:"Required;MaxSize(255)"`
  LoginSource int64
  Remember    bool
}

type UserResp struct {
  ID   string `json:"id"`
  Name string `json:"name"`
  Email    string `json:"email"`
  Admin    string `json:"is_admin"`
  Manager  string `json:"is_manager"`
}

type ErrorResp struct {
  Error string `json:"error"`
}

func Login(c *context.Context, f SignIn, db *database.Database) {
  var err error
  var u user.User

  onlyuser_val, err := db.Driver.GetSettingByKey(
    setting.SYSTEM_SIGNIN_ONLY_USERVALIDATION)
  if err == nil {
    if onlyuser_val.IsEnabled() {
      u, err = db.Driver.GetUserByName(f.UserName)
    } else {
      u, err = db.Driver.SignIn(f.UserName, f.Password)
    }
  } else {
    u, err = db.Driver.SignIn(f.UserName, f.Password)
  }

  if err != nil {
    c.JSON(400, ErrorResp{err.Error()})
    return
  }

  c.Invoke(func(config *setting.Config) {
    if f.Remember {
      days := 86400 * 30
      c.SetCookie("u_name", u.Name, days, config.GetWeb().AppSubURL, "", true, true)
      c.SetSuperSecureCookie(u.Password, "r_name", u.Name, days, config.GetWeb().AppSubURL, "", true, true)
    }

    c.Session.Set("uid", u.ID)
    c.Session.Set("uname", u.Name)

    // Clear whatever CSRF has right now, force to generate a new one
    c.SetCookie("_csrf", "", -1, config.GetWeb().AppSubURL)
  })

  c.JSONSuccess(&u)
}

func Logout(c *context.Context) {
  c.Invoke(func(config *setting.Config) {
    c.Session.Delete("uid")
    c.Session.Delete("uname")
    c.SetCookie("u_name", "", -1, config.GetWeb().AppSubURL)
    c.SetCookie("r_name", "", -1, config.GetWeb().AppSubURL)
    c.SetCookie("_csrf", "", -1, config.GetWeb().AppSubURL)
    c.SubURLRedirect("/")
  })
}

func User(c *context.Context) {
  c.JSONSuccess(&UserResp{
    c.User.ID,
    c.User.Name,
    c.User.Email,
    c.User.Admin,
    c.User.Manager,
  })
}

func Setup(m *macaron.Macaron) {
  m.Invoke(func(config *setting.Config) {
    reqSignOut := context.Toggle(&context.ToggleOptions{
     SignOutRequired: true,
     Config:          config,
     BaseURL:         config.GetWeb().AppSubURL})
    bindIgnErr := binding.BindIgnErr
    reqSignIn := context.Toggle(&context.ToggleOptions{
      SignInRequired: true,
      Config:         config,
      BaseURL:        config.GetWeb().AppSubURL})

    v1.Schema.GetClientRoute("auth_login").ToMacaron(m, reqSignOut, bindIgnErr(SignIn{}), Login)
    v1.Schema.GetClientRoute("auth_user").ToMacaron(m, reqSignIn, User)
    v1.Schema.GetClientRoute("auth_logout").ToMacaron(m, reqSignIn, Logout)
  })
}
