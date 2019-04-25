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
	"bytes"
	"encoding/gob"
	"errors"
	"sort"

	"github.com/ghodss/yaml"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	task "github.com/MottainaiCI/mottainai-server/pkg/tasks"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	"github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	"github.com/robfig/cron"
)

func AllPipelines(ctx *context.Context, db *database.Database) ([]task.Pipeline, []task.Pipeline) {

	var all []task.Pipeline
	var mine []task.Pipeline

	if ctx.IsLogged {
		if ctx.User.IsAdmin() {
			all = db.Driver.AllPipelines(db.Config)
		}
		mine, _ = db.Driver.AllUserPipelines(db.Config, ctx.User.ID)

	}

	sort.Slice(all[:], func(i, j int) bool {
		return all[i].CreatedTime > all[j].CreatedTime
	})

	sort.Slice(mine[:], func(i, j int) bool {
		return mine[i].CreatedTime > mine[j].CreatedTime
	})
	return all, mine
}

func ShowAllPipelines(ctx *context.Context, db *database.Database) {

	all, mine := AllPipelines(ctx, db)

	if ctx.IsLogged {
		if ctx.User.IsAdmin() {
			ctx.JSON(200, all)
		} else {
			ctx.JSON(200, mine)
		}
	}

}

func APIPipelineShow(ctx *context.Context, db *database.Database) error {
	pip, err := PipelineShow(ctx, db)
	if err != nil {
		return err
	}

	ctx.JSON(200, pip)
	return nil
}

func PipelineShow(ctx *context.Context, db *database.Database) (*task.Pipeline, error) {
	id := ctx.Params(":id")
	pip, err := db.Driver.GetPipeline(db.Config, id)
	if err != nil {
		return &task.Pipeline{}, err
	}

	if !ctx.CheckPipelinePermissions(&pip) {
		return &task.Pipeline{}, errors.New("Moar permissions are required for this user")
	}

	for k, t := range pip.Tasks {
		ta, err := db.Driver.GetTask(db.Config, t.ID)
		if err != nil {
			return &task.Pipeline{}, err
		}
		pip.Tasks[k] = ta
	}

	return &pip, nil
}

func PipelineYaml(ctx *context.Context, db *database.Database) string {
	id := ctx.Params(":id")
	task, err := db.Driver.GetPipeline(db.Config, id)
	if err != nil {
		ctx.NotFound()
		return ""
	}
	if !ctx.CheckPipelinePermissions(&task) {
		return ""
	}

	y, err := yaml.Marshal(task)
	if err != nil {
		ctx.ServerError(err.Error(), err)
		return ""
	}

	return string(y)
}

func Pipeline(m *mottainai.Mottainai, c *cron.Cron, th *task.TaskHandler, ctx *context.Context, db *database.Database, o task.PipelineForm) error {
	var tasks map[string]task.Task

	d := gob.NewDecoder(bytes.NewBuffer([]byte(o.Tasks)))
	if err := d.Decode(&tasks); err != nil {
		return err
	}

	opts := o.Pipeline
	opts.Tasks = tasks
	opts.Reset()
	// XX: aggiornare i task!
	for i, t := range opts.Tasks {
		f := opts.Tasks[i]

		if ctx.IsLogged {
			f.Owner = ctx.User.ID
		}
		if !ctx.CheckNamespaceBelongs(t.TagNamespace) {
			ctx.NoPermission()
			return nil
		}
		f.Status = setting.TASK_STATE_WAIT

		id, err := db.Driver.CreateTask(f.ToMap())
		if err != nil {
			return err
		}
		f.ID = id
		opts.Tasks[i] = f
	}
	if ctx.IsLogged {
		opts.Owner = ctx.User.ID
	}

	fields := opts.ToMap(false)

	docID, err := db.Driver.CreatePipeline(fields)
	if err != nil {
		return err
	}
	m.ProcessPipeline(docID)

	ctx.APICreationSuccess(docID, "pipeline")
	return nil
}

func PipelineDelete(m *mottainai.Mottainai, ctx *context.Context, db *database.Database, c *cron.Cron) error {
	id := ctx.Params(":id")
	pips, err := db.Driver.GetPipeline(db.Config, id)
	if err != nil {
		ctx.NotFound()
	}

	if !ctx.CheckPipelinePermissions(&pips) {
		ctx.NoPermission()
		return nil
	}

	err = db.Driver.DeletePipeline(id)
	if err != nil {
		return err
	}

	return nil
}
