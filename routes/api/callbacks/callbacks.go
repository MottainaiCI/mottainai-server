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
  setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
  v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"
  "gopkg.in/macaron.v1"
)

func Setup(m *macaron.Macaron) {
  m.Invoke(func(config *setting.Config) {
    m.Group(config.GetWeb().GroupAppPath(), func() {
      v1.Schema.GetCallbackRoute("cb_int_gh").ToMacaron(m, GithubIntegration)
    })
  })
}
