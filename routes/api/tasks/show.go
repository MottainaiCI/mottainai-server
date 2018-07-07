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

package tasksapi

import (
	"sort"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	task "github.com/MottainaiCI/mottainai-server/pkg/tasks"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
)

func GetTaskJson(ctx *context.Context, db *database.Database) {
	id := ctx.ParamsInt(":id")
	task, err := db.GetTask(id)
	if err != nil {
		ctx.NotFound()
		return
	}
	ctx.JSON(200, task)
}

func StreamOutputTask(ctx *context.Context, db *database.Database) string {
	id := ctx.ParamsInt(":id")
	pos := ctx.ParamsInt(":pos")

	task, err := db.GetTask(id)
	if err != nil {
		ctx.NotFound()
		return ""
	}
	return task.GetLogPart(pos)
}

func TailTask(ctx *context.Context, db *database.Database) string {
	id := ctx.ParamsInt(":id")
	pos := ctx.ParamsInt(":pos")

	task, err := db.GetTask(id)
	if err != nil {
		ctx.NotFound()
		return ""
	}
	return task.TailLog(pos)
}

func All(ctx *context.Context, db *database.Database) ([]task.Task, []task.Task) {

	var all []task.Task
	var mine []task.Task

	if ctx.IsLogged {
		if ctx.User.IsAdmin() {
			all = db.AllTasks()
		}
		mine, _ = db.AllUserTask(ctx.User.ID)

	}

	sort.Slice(all[:], func(i, j int) bool {
		return all[i].CreatedTime > all[j].CreatedTime
	})

	sort.Slice(mine[:], func(i, j int) bool {
		return mine[i].CreatedTime > mine[j].CreatedTime
	})
	return all, mine
}

func ShowAll(ctx *context.Context, db *database.Database) {

	all, mine := All(ctx, db)

	if ctx.IsLogged {
		if ctx.User.IsAdmin() {
			ctx.JSON(200, all)
		} else {
			ctx.JSON(200, mine)
		}
	}

}
