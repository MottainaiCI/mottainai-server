/*

Copyright (C) 2018-2021  Ettore Di Giacinto <mudler@gentoo.org>
                         Daniele Rondina <geaaru@sabayonlinux.org>

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

package mottainai

import (
	"fmt"
	"os"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	user "github.com/MottainaiCI/mottainai-server/pkg/user"
)

func (m *Mottainai) SetupAdminUser() error {
	var ans error

	adminUser := os.Getenv(fmt.Sprintf("%s_ADMIN_USER", setting.MOTTAINAI_ENV_PREFIX))
	adminPass := os.Getenv(fmt.Sprintf("%s_ADMIN_PASS", setting.MOTTAINAI_ENV_PREFIX))
	adminEmail := os.Getenv(fmt.Sprintf("%s_ADMIN_EMAIL", setting.MOTTAINAI_ENV_PREFIX))

	if adminUser != "" && adminPass != "" && adminEmail != "" {
		m.Invoke(func(d *database.Database) {
			nUsers := d.Driver.CountUsers()
			if nUsers > 0 {
				fmt.Println("Users already present. Skipping creation of the admin user.")
			} else {

				user := &user.User{
					Name:     adminUser,
					Email:    adminEmail,
					Password: adminPass,
					// For now create only normal user. Upgrade to admin/manager
					// is done through specific api call.
					Admin:   "yes",
					Manager: "yes",
				}

				_, err := d.Driver.InsertAndSaltUser(user)
				if err != nil {
					ans = err
					return
				}

				fmt.Println(fmt.Sprintf("User admin %s correctly created.", adminUser))
			}
		})
	}

	return ans
}
