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

package tasks

import (
	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	"github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	tasksapi "github.com/MottainaiCI/mottainai-server/routes/api/tasks"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	"github.com/MottainaiCI/mottainai-server/pkg/template"
)

// TODO: Add dup.

func Create(m *mottainai.Mottainai, th *agenttasks.TaskHandler, ctx *context.Context, db *database.Database, opts agenttasks.Task) {
	docID, err := tasksapi.Create(m, th, ctx, db, opts)
	if err != nil {
		ctx.NotFound()
	} else {
		ctx.Redirect("/tasks/display/" + docID)
	}
}

func Clone(m *mottainai.Mottainai, th *agenttasks.TaskHandler, ctx *context.Context, db *database.Database) {

	docID, err := tasksapi.CloneTask(m, th, ctx, db)
	if err != nil {
		ctx.NotFound()
	} else {
		ctx.Redirect("/tasks/display/" + docID)
	}
}

func Add(ctx *context.Context, config *setting.Config) {

	available_tasks := make([]string, 0)
	th := agenttasks.DefaultTaskHandler()
	for i, _ := range th.Tasks {
		if i != "error" && i != "success" {
			available_tasks = append(available_tasks, i)
		}
	}

	ctx.Data["AvailableTasks"] = available_tasks
	template.TemplatePreview(ctx, "tasks/add", config)
}
