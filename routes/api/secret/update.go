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

package apisecret

import (
	"errors"

	"github.com/MottainaiCI/mottainai-server/pkg/context"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
)

type SecretUpdate struct {
	Id    string `form:"id" binding:"Required"`
	Value string `form:"value"`
	Key   string ` form:"key"`
}

func UpdateSecret(upd SecretUpdate, ctx *context.Context, db *database.Database) error {
	id := upd.Id

	secret, err := db.Driver.GetSecret(id)
	if err != nil {
		ctx.NotFound()
		return err
	}

	e := errors.New("Insufficient permission to update secret")

	if ctx.IsLogged {
		if secret.OwnerId != ctx.User.ID && !ctx.User.IsAdmin() {
			ctx.ServerError("Failed updating secret pipeline", e)
			return e
		}
	} else {
		ctx.ServerError("Failed updating secret pipeline", e)
		return e
	}

	values := secret.ToMap()
	values[upd.Key] = upd.Value

	err = db.Driver.UpdateSecret(id, values)
	if err != nil {
		ctx.ServerError("Failed updating secret", err)
		return err
	}
	return nil
}

func SetSecretField(ctx *context.Context, db *database.Database, upd SecretUpdate) error {
	err := UpdateSecret(upd, ctx, db)
	if err != nil {
		return err
	}

	ctx.APIActionSuccess()
	return nil
}
