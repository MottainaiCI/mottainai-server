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
)

func APIStop(ctx *context.Context, db *database.Database) {
	err := Stop(ctx, db)
	if err != nil {
		ctx.ServerError("Failed stopping task", err)
	}
	ctx.APIActionSuccess()
}

func Stop(ctx *context.Context, db *database.Database) error {
	// XXX: Nothing to see here for now
	// ok, id := ValidateNodeKey(&f, db)
	//
	// if ok == false {
	// 	ctx.NotFound()
	// 	return ":( "
	// }
	id := ctx.Params(":id")
	mytask, err := db.Driver.GetTask(db.Config, id)
	if err != nil {
		return errors.New("Task not found")
	}
	if !ctx.CheckTaskPermissions(&mytask) {
		return errors.New("More permissions required")
	}
	err = db.Driver.UpdateTask(id, map[string]interface{}{
		"status": "stop",
	})
	if err != nil {
		return errors.New("Failed updating database")
	}
	//ctx.Redirect("/tasks")
	//ctx.Redirect("/tasks/display/" + strconv.Itoa(id))

	return nil
	//ShowAll(ctx, db)
}
