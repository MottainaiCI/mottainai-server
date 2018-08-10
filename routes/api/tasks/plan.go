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
	"errors"
	"strconv"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	"github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	"github.com/robfig/cron"
)

func PlannedTasks(ctx *context.Context, db *database.Database) {
	plans := db.Driver.AllPlans()

	ctx.JSON(200, plans)
}

func PlannedTask(ctx *context.Context, db *database.Database) error {
	id := ctx.ParamsInt(":id")
	plan, err := db.Driver.GetPlan(id)
	if err != nil {
		return err
	}
	if !ctx.CheckPlanPermissions(&plan) {
		return errors.New("Moar permissions are required for this user")
	}

	ctx.JSON(200, plan)
	return nil
}

func Plan(m *mottainai.Mottainai, c *cron.Cron, th *agenttasks.TaskHandler, ctx *context.Context, db *database.Database, opts agenttasks.Plan) (string, error) {
	opts.Reset()
	fields := opts.ToMap()

	if !ctx.CheckNamespaceBelongs(opts.TagNamespace) || !ctx.CheckPlanPermissions(&opts) {
		return ":(", errors.New("Moar permissions are required for this user")
	}

	docID, err := db.Driver.CreatePlan(fields)
	if err != nil {
		return "", err
	}

	m.ReloadCron()
	return strconv.Itoa(docID), nil
}

func PlanDelete(m *mottainai.Mottainai, ctx *context.Context, db *database.Database, c *cron.Cron) error {
	id := ctx.ParamsInt(":id")
	plan, err := db.Driver.GetPlan(id)
	if err != nil {
		ctx.NotFound()
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
