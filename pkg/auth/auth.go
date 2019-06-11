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
	"strings"

	log "gopkg.in/clog.v1"

	user "github.com/MottainaiCI/mottainai-server/pkg/user"

	utils "github.com/MottainaiCI/mottainai-server/pkg/utils"
	"github.com/go-macaron/session"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	macaron "gopkg.in/macaron.v1"
)

func IsAPIPath(url string, c *setting.WebConfig) bool {
	return strings.HasPrefix(url, c.BuildURI("/api/"))
}

// SignedInID returns the id of signed in user.
func SignedInID(c *macaron.Context, sess session.Store) string {
	db := database.Instance().Driver
	// Check access token.
	//if IsAPIPath(c.Req.URL.Path) {
	tokenSHA := c.Query("token")
	if len(tokenSHA) <= 0 {
		tokenSHA = c.Query("access_token")
	}
	if len(tokenSHA) == 0 {
		// Well, check with header again.
		auHead := c.Req.Header.Get("Authorization")
		if len(auHead) > 0 {

			auths := strings.Fields(auHead)
 			if len(auths) == 2 && strings.EqualFold(auths[0], "token") {
				tokenSHA = auths[1]
			}
		}
	}

	// Let's see if token is valid.
	if len(tokenSHA) > 0 {
		t, err := db.GetTokenByKey(tokenSHA)
		if err != nil {
			log.Error(2, "GetTokenByKey: %v", err)
			return ""
		}
		return t.UserId
	}
	//}

	uid := sess.Get("uid")
	if uid == nil {
		return ""
	}
	if id, ok := uid.(string); ok {
		if _, err := db.GetUser(id); err != nil {
			//	if !errors.New("User not found" + err) {
			log.Error(2, "GetUserByID: %v", err)
			//	}
			return ""
		}
		return id
	}
	return ""
}

// SignedInUser returns the user object of signed user.
// It returns a bool value to indicate whether user uses basic auth or not.
func SignedInUser(ctx *macaron.Context, sess session.Store) (*user.User, bool) {
	var u user.User
	var err error
	db := database.Instance().Driver

	uid := SignedInID(ctx, sess)

	if uid == "" {

		// Check with basic auth.
		baHead := ctx.Req.Header.Get("Authorization")
		if len(baHead) > 0 {
			auths := strings.Fields(baHead)
			if len(auths) == 2 && auths[0] == "Basic" {
				uname, passwd, _ := utils.BasicAuthDecode(auths[1])
				onlyuser_val, err := db.GetSettingByKey(
					setting.SYSTEM_SIGNIN_ONLY_USERVALIDATION)
				if err == nil {
					if onlyuser_val.IsEnabled() {
						u, err = db.GetUserByName(uname)
					} else {
						u, err = db.SignIn(uname, passwd)
					}
				} else {
					// If setting is not present erro is No settingname found
					// I consider so that user and password validation is enable
					u, err = db.SignIn(uname, passwd)
				}
				if err != nil {
					log.Error(4, "SignIn error : %v", err)
					return nil, false
				}

				return &u, true
			}
		}
		return nil, false
	}

	u, err = db.GetUser(uid)
	if err != nil {
		log.Error(4, "GetUser Error: %v", err)
		return nil, false
	}
	return &u, false
}
