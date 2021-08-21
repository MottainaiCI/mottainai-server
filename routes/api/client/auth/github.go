/*

Copyright (C) 2017-2021  Ettore Di Giacinto <mudler@gentoo.org>
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

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	"net/http"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
)

func GithubIntegrationUrl(c *context.Context, db *database.Database) error {
	if c.IsLogged {
		gothic.GetProviderName = func(req *http.Request) (string, error) { return "github", nil }
		gothic.GetGithubUrl(c, db)
		return nil
	}
	return errors.New("user not logged in")
}

func GithubLogout(c *context.Context, db *database.Database) error {
	if c.IsLogged {
		gothic.GetProviderName = func(req *http.Request) (string, error) { return "github", nil }
		u, err := db.Driver.GetUser(c.User.ID)
		if err != nil {
			return err
		}
		u.RemoveIdentity("github")
		err = db.Driver.UpdateUser(c.User.ID, u.ToMap())
		if err != nil {
			return err
		}
		c.JSON(200, "")
		return nil
	}
	return errors.New("user not logged in")
}

// TODO: factor out in unique require check function from DB settings.
// in ctx/auth.go
// this is duplicated in webhook/github now
func RequiresIntegrationSetting(c *context.Context, db *database.Database) error {
	// Check setting if we have to process this.
	err := errors.New("Third party integration disabled")
	uuu, err := db.Driver.GetSettingByKey(setting.SYSTEM_THIRDPARTY_INTEGRATION_ENABLED)
	if err == nil {
		if uuu.IsDisabled() {
			c.ServerError("Third party integration disabled", err)
			return err
		}
	}
	return nil
}
