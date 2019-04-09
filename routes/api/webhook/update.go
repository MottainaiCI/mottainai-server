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

package apiwebhook

import (
	"bytes"
	"encoding/gob"
	"errors"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
)

func UpdatePipelineWebHook(ctx *context.Context, db *database.Database, pipeform agenttasks.PipelineForm) error {
	id := ctx.Params(":id")

	webhook, err := db.Driver.GetWebHook(id)
	if err != nil {
		ctx.NotFound()
		return err
	}

	e := errors.New("Insufficient permission to remove webhook pipeline")

	if ctx.IsLogged {
		if webhook.OwnerId != ctx.User.ID && !ctx.User.IsAdmin() {
			ctx.ServerError("Failed updating webhook pipeline", e)
			return e
		}
	} else {
		ctx.ServerError("Failed updating webhook pipeline", e)
		return e
	}

	var tasks map[string]agenttasks.Task
	d := gob.NewDecoder(bytes.NewBuffer([]byte(pipeform.Tasks)))
	if err := d.Decode(&tasks); err != nil {
		panic(err)
	}
	opts := pipeform.Pipeline
	opts.Tasks = tasks
	opts.Reset()
	webhook.SetPipeline(opts)

	err = db.Driver.UpdateWebHook(id, webhook.ToMap())
	if err != nil {
		ctx.ServerError("Failed updating webhook pipeline", err)
		return err
	}
	return nil
}

func UpdateTaskWebHook(ctx *context.Context, db *database.Database, task agenttasks.Task) error {
	id := ctx.Params(":id")

	webhook, err := db.Driver.GetWebHook(id)
	if err != nil {
		ctx.NotFound()
		return err
	}

	e := errors.New("Insufficient permission to remove webhook pipeline")

	if ctx.IsLogged {
		if webhook.OwnerId != ctx.User.ID && !ctx.User.IsAdmin() {
			ctx.ServerError("Failed updating webhook pipeline", e)
			return e
		}
	} else {
		ctx.ServerError("Failed updating webhook pipeline", e)
		return e
	}
	webhook.SetTask(&task)
	err = db.Driver.UpdateWebHook(id, webhook.ToMap())
	if err != nil {
		ctx.ServerError("Failed updating webhook", err)
		return err
	}
	return nil
}

type WebhookUpdate struct {
	Id    string `form:"id" binding:"Required"`
	Value string `form:"value"`
	Key   string ` form:"key"`
}

func UpdateWebHook(upd WebhookUpdate, ctx *context.Context, db *database.Database) error {
	id := upd.Id

	webhook, err := db.Driver.GetWebHook(id)
	if err != nil {
		ctx.NotFound()
		return err
	}

	e := errors.New("Insufficient permission to update webhook")

	if ctx.IsLogged {
		if webhook.OwnerId != ctx.User.ID && !ctx.User.IsAdmin() {
			ctx.ServerError("Failed updating webhook pipeline", e)
			return e
		}
	} else {
		ctx.ServerError("Failed updating webhook pipeline", e)
		return e
	}

	values := webhook.ToMap()
	values[upd.Key] = upd.Value

	err = db.Driver.UpdateWebHook(id, values)
	if err != nil {
		ctx.ServerError("Failed updating webhook", err)
		return err
	}
	return nil
}

func UpdateTask(ctx *context.Context, db *database.Database, task agenttasks.Task) string {
	err := UpdateTaskWebHook(ctx, db, task)
	if err != nil {
		return ":("
	}

	return "OK"
}

func SetWebHookField(ctx *context.Context, db *database.Database, upd WebhookUpdate) string {
	err := UpdateWebHook(upd, ctx, db)
	if err != nil {
		return ":("
	}

	return "OK"
}

func UpdatePipeline(ctx *context.Context, db *database.Database, pipeform agenttasks.PipelineForm) string {
	err := UpdatePipelineWebHook(ctx, db, pipeform)
	if err != nil {
		return ":("
	}

	return "OK"
}
