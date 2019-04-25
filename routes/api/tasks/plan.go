/*

Copyright (C) 2017-2019  Ettore Di Giacinto <mudler@gentoo.org>
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
	"errors"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	"github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	"github.com/robfig/cron"
)

func PlannedTasks(ctx *context.Context, db *database.Database) {
	plans := db.Driver.AllPlans(db.Config)

	ctx.JSON(200, plans)
}

func PlannedTask(ctx *context.Context, db *database.Database) error {
	id := ctx.Params(":id")
	plan, err := db.Driver.GetPlan(db.Config, id)
	if err != nil {
		return err
	}
	if !ctx.CheckPlanPermissions(&plan) {
		ctx.NoPermission()
		return nil
	}

	ctx.JSON(200, plan)
	return nil
}

func Plan(m *mottainai.Mottainai, c *cron.Cron, th *agenttasks.TaskHandler, ctx *context.Context, db *database.Database, opts agenttasks.Plan) error {
	opts.Reset()
	fields := opts.ToMap()

	if !ctx.CheckNamespaceBelongs(opts.TagNamespace) || !ctx.CheckPlanPermissions(&opts) {
		ctx.NoPermission()
		return nil
	}

	docID, err := db.Driver.CreatePlan(fields)
	if err != nil {
		return err
	}

	m.ReloadCron()

	ctx.APICreationSuccess(docID, "plan")
	return nil
}

func PlanDeleteById(id string, db *database.Database, m *mottainai.Mottainai, ctx *context.Context) error {
	plan, err := db.Driver.GetPlan(db.Config, id)
	if err != nil {
		ctx.NotFound()
		return err
	}

	if !ctx.CheckNamespaceBelongs(plan.TagNamespace) || !ctx.CheckPlanPermissions(&plan) {
		return errors.New("Moar permissions are required for this user")
	}

	err = db.Driver.DeletePlan(id)
	if err != nil {
		return err
	}

	m.ReloadCron()
	return nil
}

func PlanDelete(m *mottainai.Mottainai, ctx *context.Context, db *database.Database) error {
	return PlanDeleteById(ctx.Params(":id"), db, m, ctx)
}
