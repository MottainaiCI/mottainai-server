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

	secret "github.com/MottainaiCI/mottainai-server/pkg/secret"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
)

func CreateSecret(ctx *context.Context, db *database.Database) (*secret.Secret, error) {
	name := ctx.Params(":name")
	var t *secret.Secret
	var err error
	if ctx.IsLogged {
		t = &secret.Secret{Name: name, OwnerId: ctx.User.ID}
	} else {
		ctx.ServerError("Failed creating secret", errors.New("Insufficient permission for creating a secret"))
		return t, err
	}
	return t, nil
}

func Create(ctx *context.Context, db *database.Database) error {
	t, err := CreateSecret(ctx, db)
	if err != nil {
		return err
	}
	id, err := db.Driver.InsertSecret(t)
	if err != nil {
		ctx.ServerError("Failed creating secret", err)
		return err
	}

	ctx.APIPayload(id, "secret", t.Secret)
	return nil
}
