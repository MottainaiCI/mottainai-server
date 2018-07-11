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
	"strconv"
	"time"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	tasksapi "github.com/MottainaiCI/mottainai-server/routes/api/tasks"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	"github.com/MottainaiCI/mottainai-server/pkg/template"
)

func DisplayTask(ctx *context.Context, db *database.Database) {
	id := ctx.ParamsInt(":id")
	task, err := db.GetTask(id)
	if err != nil {
		ctx.NotFound()
		return
	}
	if !ctx.CheckTaskPermissions(&task) {
		return
	}
	ctx.Data["Task"] = task

	if ctx.IsLogged && (ctx.User.IsAdmin() || ctx.User.IsManager()) {
		uid, err := strconv.Atoi(task.Owner)
		if err == nil {
			u, err := db.GetUser(uid)
			if err == nil {
				u.Password = ""
				ctx.Data["TaskOwner"] = u
			}
		}
		nid, err := strconv.Atoi(task.Node)
		if err == nil {
			n, err := db.GetNode(nid)
			if err == nil {
				ctx.Data["TaskNode"] = n
			}
		}
	}

	if len(task.CreatedTime) > 0 && len(task.StartTime) > 0 {
		created, err := strconv.Atoi(task.CreatedTime)
		if err == nil {
			started, err := strconv.Atoi(task.StartTime)
			if err == nil {
				ctx.Data["WaitingTime"] = started - created
			}
		}
	}

	if len(task.StartTime) > 0 {
		now, err := strconv.Atoi(time.Now().Format("20060102150405"))
		if err == nil {
			started, err := strconv.Atoi(task.StartTime)
			if err == nil {
				ctx.Data["RunningTime"] = now - started
			}
		}
	}
	ctx.Data["Artefacts"] = task.Artefacts()
	template.TemplatePreview(ctx, "tasks/display")
}

func ShowAll(ctx *context.Context, db *database.Database) {
	all, mine := tasksapi.All(ctx, db)

	ctx.Data["Tasks"] = all
	ctx.Data["UserTasks"] = mine

	template.TemplatePreview(ctx, "tasks")
}

func ShowArtefacts(ctx *context.Context, db *database.Database) {
	id := ctx.ParamsInt(":id")

	tasks_info, err := db.GetTask(id)

	if err != nil {
		panic(err)
	}

	ctx.Data["Artefacts"] = tasks_info.Artefacts()
	ctx.Data["Task"] = id
	ctx.Data["TaskDetail"] = tasks_info

	template.TemplatePreview(ctx, "tasks/artefacts")
}
