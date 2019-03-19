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

package nodesroute

import (
	"sort"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"

	"github.com/MottainaiCI/mottainai-server/pkg/template"
)

func ShowAll(ctx *context.Context, db *database.Database) {
	//tasks := db.ListTasks()
	nodes := db.Driver.AllNodes()
	//ctx.Data["TasksIDs"] = tasks
	ctx.Data["Nodes"] = nodes
	template.TemplatePreview(ctx, "nodes", db.Config)
}

type BrokerConfig struct {
	Type, DefaultQueue, Broker, ResultBackend, Exchange string
}

func Show(ctx *context.Context, db *database.Database) {
	id := ctx.Params(":id")

	node, err := db.Driver.GetNode(id)
	if err != nil {
		ctx.NotFound()
		return
	}
	//p_queue := node.Hostname + node.NodeID

	ctx.Data["Node"] = node
	tasks, _ := db.Driver.AllNodeTask(db.Config, node.ID)
	//tasks, _ := db.Driver.FindDoc("Tasks", `[{"eq": "`+p_queue+`", "in": ["queue"]}]`)
	// var node_tasks = make([]agenttasks.Task, 0)
	// for i, _ := range tasks {
	// 	t, _ := db.Driver.GetTask(db.Config, strconv.Itoa(i))
	// 	node_tasks = append(node_tasks, t)
	// }
	sort.Slice(tasks[:], func(i, j int) bool {
		return tasks[i].CreatedTime > tasks[j].CreatedTime
	})

	ctx.Data["Tasks"] = tasks
	if ctx.CheckUserOrManager() {
		apikeys, err := db.Driver.GetTokensByUserID(ctx.User.ID)
		if err != nil {
			ctx.ServerError("Failed finding token", err)
			return
		}
		if len(apikeys) != 0 {
			ctx.Data["EphemeralApiKey"] = apikeys[0].Key
		}
		ctx.Invoke(func(config *setting.Config) {
			ctx.Data["EphemeralBrokerSettings"] = &BrokerConfig{DefaultQueue: config.GetBroker().BrokerDefaultQueue,
				ResultBackend: config.GetBroker().BrokerResultBackend,
				Broker:        config.GetBroker().Broker,
				Exchange:      config.GetBroker().BrokerExchange,
				Type:          config.GetBroker().Type,
			}
		})
	}

	template.TemplatePreview(ctx, "nodes/show", db.Config)
}
