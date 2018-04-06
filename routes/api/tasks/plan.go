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
	"strconv"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	"github.com/MottainaiCI/mottainai-server/pkg/db"
	"github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	"github.com/MottainaiCI/mottainai-server/pkg/tasks"
	"github.com/robfig/cron"

	machinery "github.com/RichardKnop/machinery/v1"
)

func PlannedTasks(ctx *context.Context, db *database.Database) {
	plans := db.AllPlans()

	ctx.JSON(200, plans)
}

func PlannedTask(ctx *context.Context, db *database.Database) {
	id := ctx.ParamsInt(":id")
	plan, err := db.GetPlan(id)
	if err != nil {
		panic(err)
	}

	ctx.JSON(200, plan)
}

func Plan(m *mottainai.Mottainai, c *cron.Cron, th *agenttasks.TaskHandler, ctx *context.Context, rabbit *machinery.Server, db *database.Database, opts agenttasks.Plan) (string, error) {
	plan := opts.Planned
	opts.Reset()
	fields := opts.ToMap()

	docID, err := db.CreatePlan(fields)
	if err != nil {
		return "", err
	}

	m.ReloadCron()

	return strconv.Itoa(docID), nil
}

func PlanDelete(m *mottainai.Mottainai, ctx *context.Context, rabbit *machinery.Server, db *database.Database, c *cron.Cron) error {
	id := ctx.ParamsInt(":id")
	err := db.DeletePlan(id)
	if err != nil {
		return err
	}
	m.ReloadCron()
	return nil
}
