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

	webhook "github.com/MottainaiCI/mottainai-server/pkg/webhook"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
)

func CreateWebHook(ctx *context.Context, db *database.Database) (*webhook.WebHook, error) {
	webtype := ctx.Params(":type")
	var t *webhook.WebHook
	var err error
	if ctx.IsLogged {
		t, err = webhook.GenerateUserWebHook(ctx.User.ID)
		if err != nil {
			ctx.ServerError("Failed creating webhook", err)
			return t, err
		}
		t.Type = webtype
	} else {
		ctx.ServerError("Failed creating webhook", errors.New("Insufficient permission for creating a webhook"))
		return t, err
	}
	return t, nil
}

func Create(ctx *context.Context, db *database.Database) string {
	t, err := CreateWebHook(ctx, db)
	if err != nil {
		return ":("
	}
	_, err = db.Driver.InsertWebHook(t)
	if err != nil {
		ctx.ServerError("Failed creating webhook", err)
		return ":("
	}

	return t.Key
}
