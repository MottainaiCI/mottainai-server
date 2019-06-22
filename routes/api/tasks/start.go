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

	database "github.com/MottainaiCI/mottainai-server/pkg/db"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	"github.com/MottainaiCI/mottainai-server/pkg/mottainai"
)

func APISendStartTask(m *mottainai.Mottainai, ctx *context.Context, db *database.Database) {
	err := SendStartTask(m, ctx, db)
	if err != nil {
		ctx.ServerError("Failed starting task", err)
	}
	ctx.APIActionSuccess()
}

func SendStartTask(m *mottainai.Mottainai, ctx *context.Context, db *database.Database) error {
	id := ctx.Params(":id")
	mytask, err := db.Driver.GetTask(db.Config, id)
	if err != nil {
		return errors.New("Task not found")
	}
	if !ctx.CheckTaskPermissions(&mytask) {
		return errors.New("More permission required")
	}

	if mytask.IsWaiting() || mytask.IsRunning() {
		return errors.New("Task is already waiting or running, refusing to start")
	}

	_, err = m.SendTask(id)
	if err != nil {
		return err
	}

	return nil
}
