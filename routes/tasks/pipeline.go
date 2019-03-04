/*

Copyright (C) 2019  Ettore Di Giacinto <mudler@gentoo.org>
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

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	tasksapi "github.com/MottainaiCI/mottainai-server/routes/api/tasks"

	"github.com/MottainaiCI/mottainai-server/pkg/template"
)

func DisplayPipeline(ctx *context.Context, db *database.Database) {
	pip, err := tasksapi.PipelineShow(ctx, db)
	if err != nil {
		ctx.NotFound()
		return
	}

	ctx.Data["Pipeline"] = pip

	if ctx.IsLogged && (ctx.User.IsAdmin() || ctx.User.IsManager()) {

		u, err := db.Driver.GetUser(pip.Owner)
		if err == nil {
			u.Password = ""
			ctx.Data["Owner"] = u
		}
	}

	if len(pip.CreatedTime) > 0 && len(pip.StartTime) > 0 {
		created, err := strconv.Atoi(pip.CreatedTime)
		if err == nil {
			started, err := strconv.Atoi(pip.StartTime)
			if err == nil {
				ctx.Data["WaitingTime"] = started - created
			}
		}
	}

	if len(pip.StartTime) > 0 {
		now, err := strconv.Atoi(time.Now().Format("20060102150405"))
		if err == nil {
			started, err := strconv.Atoi(pip.StartTime)
			if err == nil {
				ctx.Data["RunningTime"] = now - started
			}
		}
	}
	template.TemplatePreview(ctx, "pipelines/display", db.Config)
}

func DisplayAllPipelines(ctx *context.Context, db *database.Database) {
	all, mine := tasksapi.AllPipelines(ctx, db)

	if ctx.IsLogged {
		if ctx.User.IsAdmin() {
			ctx.Data["Pipelines"] = all
		} else {
			ctx.Data["Pipelines"] = mine
		}
	}

	template.TemplatePreview(ctx, "pipelines", db.Config)
}
