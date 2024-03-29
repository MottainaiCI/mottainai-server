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

package api

import (
	"github.com/MottainaiCI/mottainai-server/routes/api/callbacks"
	client "github.com/MottainaiCI/mottainai-server/routes/api/client"
	namespacesapi "github.com/MottainaiCI/mottainai-server/routes/api/namespaces"
	nodesapi "github.com/MottainaiCI/mottainai-server/routes/api/nodes"
	queuesapi "github.com/MottainaiCI/mottainai-server/routes/api/queues"
	apisecret "github.com/MottainaiCI/mottainai-server/routes/api/secret"
	settingsroute "github.com/MottainaiCI/mottainai-server/routes/api/settings"
	stats "github.com/MottainaiCI/mottainai-server/routes/api/stats"
	storagesapi "github.com/MottainaiCI/mottainai-server/routes/api/storages"
	tasksapi "github.com/MottainaiCI/mottainai-server/routes/api/tasks"
	apitoken "github.com/MottainaiCI/mottainai-server/routes/api/token"
	apiwebhook "github.com/MottainaiCI/mottainai-server/routes/api/webhook"

	userapi "github.com/MottainaiCI/mottainai-server/routes/api/user"
	macaron "gopkg.in/macaron.v1"
)

func Setup(m *macaron.Macaron) {
	client.Setup(m)
	userapi.Setup(m)
	nodesapi.Setup(m)
	tasksapi.Setup(m)
	namespacesapi.Setup(m)
	apitoken.Setup(m)
	storagesapi.Setup(m)
	stats.Setup(m)
	settingsroute.Setup(m)
	apiwebhook.Setup(m)
	apisecret.Setup(m)
	queuesapi.Setup(m)
	callbacks.Setup(m)
}
