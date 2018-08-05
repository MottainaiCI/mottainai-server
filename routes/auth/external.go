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

// TODO: Factor out in external/github.go

import (
  "errors"
  gothic "github.com/MottainaiCI/mottainai-server/pkg/providers"
  setting "github.com/MottainaiCI/mottainai-server/pkg/settings"

  "github.com/go-macaron/session"

  "github.com/MottainaiCI/mottainai-server/pkg/context"
  ciuser "github.com/MottainaiCI/mottainai-server/pkg/user"

 "net/http"


  database "github.com/MottainaiCI/mottainai-server/pkg/db"

)


func GithubLogout(c *context.Context, db *database.Database) error{
    c.Session.Delete("github")
    if c.IsLogged {
    gothic.GetProviderName = func (req *http.Request) (string, error) { return "github",nil}
      u, err := db.GetUser(c.User.ID)
      if err != nil {
        return err
      }
      u.RemoveIdentity("github")
      err=db.UpdateUser(c.User.ID, u.ToMap())
      if err != nil {
        return err
      }
      u.Password= ""
    c.Data["User"] = u
    c.Success(SHOW)
    return nil
  }
    return errors.New("User not logged")
}

func GithubLogin(c *context.Context, db *database.Database) error {
		// try to get the user without re-authenticating
    if c.IsLogged {
      //c.Session.Set("provider", interface{})
      gothic.GetProviderName = func (req *http.Request) (string, error) { return "github",nil}
		if gothUser, err := gothic.CompleteUserAuth(c); err == nil {
      u, err := db.GetUser(c.User.ID)

      if err != nil {
        return err
      }


      u.AddIdentity("github", &ciuser.Identity{ID: gothUser.UserID, Provider: "github"})
      err=db.UpdateUser(c.User.ID, u.ToMap())
      if err != nil {
        return err
      }
            u.Password = ""
      c.Data["User"] = u
      c.Success(SHOW)
      return nil
		} else {
       gothic.BeginAuthHandler(c)
      return nil
		}

  }
  return errors.New("User not logged")
}
// TODO: factor out in unique require check function from DB settings.
// in ctx/auth.go
// this is duplicated in webhook/github now
func RequiresIntegrationSetting(c *context.Context, db *database.Database)  error {
  // Check setting if we have to process this.
  err := errors.New("Third party integration disabled")
  uuu, err := db.GetSettingByKey(setting.SYSTEM_THIRDPARTY_INTEGRATION_ENABLED)
  if err == nil {
    if uuu.IsDisabled() {
      c.ServerError("Third party integration disabled", err)
      return err
    }
  }
  return nil
}
func GithubAuthCallback(s session.Store,c *context.Context, db *database.Database) error {

  if c.IsLogged {

    gothic.GetProviderName = func (req *http.Request) (string, error) { return "github",nil}
    user, err := gothic.CompleteUserAuth(c)
    if err != nil {
  return err
    }
    u, err := db.GetUser(c.User.ID)

    if err != nil {
      return err
    }

    u.AddIdentity("github", &ciuser.Identity{ID: user.UserID, Provider: "github"})
    err=db.UpdateUser(c.User.ID, u.ToMap())
    if err != nil {
      return err
    }
          u.Password = ""
    c.Data["User"] = u
    c.Success(SHOW)
    return nil
}

    return errors.New("User not logged")

}
