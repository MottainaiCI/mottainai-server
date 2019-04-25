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
	"errors"

	"github.com/MottainaiCI/mottainai-server/pkg/context"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
)

func RemoveWebHook(ctx *context.Context, db *database.Database) error {
	id := ctx.Params(":id")

	webhook, err := db.Driver.GetWebHook(id)
	if err != nil {
		ctx.NotFound()
		return err
	}

	e := errors.New("Insufficient permission to remove webhook")

	if ctx.IsLogged {
		if webhook.OwnerId != ctx.User.ID && !ctx.User.IsAdmin() {
			ctx.ServerError("Failed removing webhook", e)
			return e
		}
	} else {
		ctx.ServerError("Failed removing webhook", e)
		return e
	}

	err = db.Driver.DeleteWebHook(id)
	if err != nil {
		ctx.ServerError("Failed removing webhook", err)
		return err
	}
	return nil
}

func Remove(ctx *context.Context, db *database.Database) error {
	err := RemoveWebHook(ctx, db)
	if err != nil {
		return err
	}

	ctx.APIActionSuccess()
	return nil
}
