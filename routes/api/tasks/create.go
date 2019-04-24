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
	"time"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	"github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
)

// TODO: Add dup.

func APICreate(m *mottainai.Mottainai, th *agenttasks.TaskHandler, ctx *context.Context, db *database.Database, opts agenttasks.Task) string {
	docID, err := Create(m, th, ctx, db, opts)
	if err != nil {
		ctx.NotFound()
		return ""
	}
	return docID
}

func Create(m *mottainai.Mottainai, th *agenttasks.TaskHandler, ctx *context.Context, db *database.Database, opts agenttasks.Task) (string, error) {
	opts.Reset()

	opts.Output = ""
	opts.Result = "none"
	opts.ExitStatus = ""
	opts.CreatedTime = time.Now().Format("20060102150405")

	if ctx.IsLogged {
		opts.Owner = ctx.User.ID
	}

	if !ctx.CheckNamespaceBelongs(opts.TagNamespace) {
		return ":(", errors.New("Moar permissions are required for this user")
	}

	docID, err := db.Driver.InsertTask(&opts)
	if err != nil {
		return "", err
	}

	if _, err := m.SendTask(docID); err != nil {
		return "Error sending task: " + err.Error(), err
	}
	return docID, nil
}

func CloneTask(m *mottainai.Mottainai, th *agenttasks.TaskHandler, ctx *context.Context, db *database.Database) (string, error) {
	id := ctx.Params(":id")

	task, err := db.Driver.GetTask(db.Config, id)
	if err != nil {
		return "", err
	}

	if !ctx.CheckNamespaceBelongs(task.TagNamespace) {
		return ":(", errors.New("Moar permissions are required for this user")
	}

	docID, err := db.Driver.CloneTask(db.Config, id)
	if err != nil {
		return "", err
	}

	if ctx.IsLogged {
		db.Driver.UpdateTask(docID, map[string]interface{}{
			"owner_id": ctx.User.ID,
		})
	}

	m.SendTask(docID)

	return docID, nil
}
