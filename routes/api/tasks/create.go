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
	"fmt"
	"strings"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	"github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
)

func APICreate(m *mottainai.Mottainai, ctx *context.Context, db *database.Database, opts agenttasks.Task) error {
	docID, err := Create(m, ctx, db, opts)
	if err != nil {
		return err
	}

	ctx.APICreationSuccess(docID, "task")
	return nil
}

func Create(m *mottainai.Mottainai, ctx *context.Context,
	db *database.Database, opts agenttasks.Task) (string, error) {

	opts.Reset()
	opts.Result = "none"

	if ctx.IsLogged {
		opts.Owner = ctx.User.ID
	}

	// Check storage permissions
	if opts.Storage != "" {
		storages := strings.Split(opts.Storage, ",")
		for _, s := range storages {

			storage, err := db.Driver.SearchStorage(strings.TrimSpace(s))
			if err != nil {
				return "", errors.New("Invalid storage " + s)
			}

			if !ctx.CheckStoragePermissions(&storage) {
				return "", errors.New("More permissions requires to use storage " + s)
			}
		}
	}

	if !ctx.CheckNamespaceBelongs(opts.TagNamespace) {
		return "", errors.New("More permissions required")
	}

	err := m.CreateTask(&opts)
	if err != nil {
		return "", err
	}

	fmt.Println(fmt.Sprintf(
		"Created task %s for queue %s.", opts.ID, opts.Queue,
	))

	if _, err := m.SendTask(opts.ID); err != nil {
		return "", err
	}
	return opts.ID, nil
}

func CloneAndSend(id string, m *mottainai.Mottainai, ctx *context.Context, db *database.Database) (string, error) {

	task, err := db.Driver.GetTask(db.Config, id)
	if err != nil {
		return "", err
	}

	if task.ID == "" {
		return "", errors.New("Invalid task id")
	}

	// Check storage permissions
	if task.Storage != "" {
		storages := strings.Split(task.Storage, ",")
		for _, s := range storages {

			storage, err := db.Driver.SearchStorage(strings.TrimSpace(s))

			if err != nil {
				return "", errors.New("Invalid storage " + s)
			}

			if !ctx.CheckStoragePermissions(&storage) {
				return "", errors.New("More permissions requires to use storage " + s)
			}
		}
	}

	if !ctx.CheckNamespaceBelongs(task.TagNamespace) {
		ctx.NoPermission()
		return "", nil
	}

	task.Reset()

	if ctx.IsLogged {
		task.Owner = ctx.User.ID
	}
	task.ID = ""
	task.EndTime = ""
	task.StartTime = ""

	err = m.CreateTask(&task)
	if err != nil {
		return "", errors.New("Error on create task " + err.Error())
	}

	if _, err := m.SendTask(task.ID); err != nil {
		return "", err
	}
	return task.ID, nil
}

func CloneTask(m *mottainai.Mottainai, ctx *context.Context, db *database.Database) error {
	id := ctx.Params(":id")

	docID, err := CloneAndSend(id, m, ctx, db)
	if err != nil {
		return err
	}

	ctx.APICreationSuccess(docID, "task")
	return nil
}
