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

func APIDelete(ctx *context.Context, db *database.Database) string {
	err := Delete(ctx, db)
	if err != nil {
		ctx.NotFound()
		return ""
	}
	return "OK"
}

func Delete(ctx *context.Context, db *database.Database) error {
	id := ctx.Params(":id")

	task, err := db.Driver.GetTask(db.Config, id)
	if err != nil {
		return err
	}
	if !ctx.CheckTaskPermissions(&task) {
		return errors.New("Moar permissions are required for this user")
	}
	err = db.Driver.DeleteTask(db.Config, id)
	if err != nil {
		return err
	}
	return nil
}
