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
	dbcommon "github.com/MottainaiCI/mottainai-server/pkg/db/common"
	"sort"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	task "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	"github.com/ghodss/yaml"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
)

func GetTaskYaml(ctx *context.Context, db *database.Database) string {
	id := ctx.Params(":id")
	task, err := db.Driver.GetTask(db.Config, id)
	if err != nil {
		ctx.NotFound()
		return ""
	}
	if !ctx.CheckTaskPermissions(&task) {
		ctx.NoPermission()
		return ""
	}

	y, err := yaml.Marshal(task)
	if err != nil {
		ctx.ServerError(err.Error(), err)
		return ""
	}

	return string(y)
}

func GetTaskJson(ctx *context.Context, db *database.Database) {
	id := ctx.Params(":id")
	task, err := db.Driver.GetTask(db.Config, id)
	if err != nil {
		ctx.NotFound()
		return
	}
	if !ctx.CheckTaskPermissions(&task) {
		ctx.NoPermission()
		return
	}
	ctx.JSON(200, task)
}

func APIShowTaskByStatus(ctx *context.Context, db *database.Database) {
	ctx.JSON(200, ShowTaskByStatus(ctx, db))
}

// TODO: We shouldn't have queries here but in the db interface
func ShowTaskByStatus(ctx *context.Context, db *database.Database) []task.Task {

	status := ctx.Params(":status")
	tasks, e := db.Driver.GetTaskByStatus(db.Config, status)
	if e != nil {
		return []task.Task{}
	}
	var res []task.Task

	// Query result are document IDs
	for _, task := range tasks {
		// Read document

		if ctx.CheckUserOrManager() || ctx.CheckTaskPermissions(&task) {
			res = append(res, task)
		}
	}
	return res
}

func StreamOutputTask(ctx *context.Context, db *database.Database) string {
	id := ctx.Params(":id")
	pos := ctx.ParamsInt(":pos")

	task, err := db.Driver.GetTask(db.Config, id)
	if err != nil {
		ctx.NotFound()
		return ""
	}
	if !ctx.CheckTaskPermissions(&task) {
		ctx.NoPermission()

		return ""
	}

	return task.GetLogPart(pos, db.Config.GetStorage().ArtefactPath, db.Config.GetWeb().LockPath)
}

func TailTask(ctx *context.Context, db *database.Database) string {
	id := ctx.Params(":id")
	pos := ctx.ParamsInt(":pos")

	task, err := db.Driver.GetTask(db.Config, id)
	if err != nil {
		ctx.NotFound()
		return ""
	}
	if !ctx.CheckTaskPermissions(&task) {
		ctx.NoPermission()
		return ""
	}

	return task.TailLog(pos, db.Config.GetStorage().ArtefactPath, db.Config.GetWeb().LockPath)
}

func All(ctx *context.Context, db *database.Database) []task.Task {
	var all []task.Task

	if ctx.IsLogged {
		if ctx.User.IsAdmin() {
			all = db.Driver.AllTasks(db.Config)
		} else {
			all, _ = db.Driver.AllUserTask(db.Config, ctx.User.ID)
		}

	}

	sort.Slice(all[:], func(i, j int) bool {
		return all[i].CreatedTime > all[j].CreatedTime
	})

	return all
}

func AllFiltered(ctx *context.Context, db *database.Database) (result dbcommon.TaskResult) {
	f := dbcommon.CreateTaskFilter(
		ctx.QueryInt("pageIndex"),
		ctx.QueryInt("pageSize"),
		ctx.Query("sort"),
		ctx.Query("sortOrder"),
	)

	if ctx.IsLogged {
		if ctx.User.IsAdmin() {
			result, _ = db.Driver.AllTasksFiltered(db.Config, f)
		} else {
			result, _ = db.Driver.AllUserFiltered(db.Config, ctx.User.ID, f)
		}
	}

	return result
}

func ShowAll(ctx *context.Context, db *database.Database) {
	if ctx.IsLogged {
		all := All(ctx, db)
		ctx.JSON(200, all)
	}
}

func ShowAllFiltered(ctx *context.Context, db *database.Database) {
	if ctx.IsLogged {
		result := AllFiltered(ctx, db)
		ctx.JSON(200, result)
	}
}
