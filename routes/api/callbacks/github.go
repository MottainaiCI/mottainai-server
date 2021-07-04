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

package callbacks

import (
  "github.com/MottainaiCI/mottainai-server/pkg/context"
  database "github.com/MottainaiCI/mottainai-server/pkg/db"
  gothic "github.com/MottainaiCI/mottainai-server/pkg/providers"
  ciuser "github.com/MottainaiCI/mottainai-server/pkg/user"
  "net/http"
)

func GithubIntegration (ctx *context.Context, db *database.Database) {
  gothic.GetProviderName = func(req *http.Request) (string, error) { return "github", nil }
  u, gothUser, err := gothic.CompleteUserAuth(ctx, db)
  if err != nil {
    ctx.Data["ErrorMsg"] = err.Error()
    ctx.HTML(http.StatusInternalServerError, "status/500")
    return
  }

  u.AddIdentity("github", &ciuser.Identity{ID: gothUser.UserID, Provider: "github"})
  err = db.Driver.UpdateUser(ctx.User.ID, u.ToMap())
  if err != nil {
   ctx.RenderWithErr("Could not complete integration", "status/500")
   return
  }

  ctx.HTML(200, "callbacks/integrations/github")
}
