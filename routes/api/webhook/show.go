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
	webhook "github.com/MottainaiCI/mottainai-server/pkg/webhook"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
)

func GetWebHooks(ctx *context.Context, db *database.Database) ([]webhook.WebHook, []webhook.WebHook, error) {
	var all []webhook.WebHook
	var mine []webhook.WebHook

	var err error
	if ctx.IsLogged {
		if ctx.User.IsAdmin() {
			all = db.Driver.AllWebHooks()
			return all, mine, nil
		}
		mine, err = db.Driver.GetWebHooksByUserID(ctx.User.ID)
		if err != nil {
			ctx.ServerError("Failed finding webhook", err)
			return all, mine, err
		}
	}
	return all, mine, nil
}

func ShowSingle(ctx *context.Context, db *database.Database) error {
	id := ctx.Params(":id")

	w, err := db.Driver.GetWebHook(id)
	if err != nil {
		ctx.NotFound()
		return err
	}
	s := &webhook.WebHookSingle{}
	if w.HasTask() {
		t, err := w.ReadTask()
		if err != nil {
			ctx.ServerError("Failed rendering webhook", err)
			return err
		}
		s.Task = t
	}
	if w.HasPipeline() {
		p, err := w.ReadPipeline()
		if err != nil {
			ctx.ServerError("Failed rendering webhook", err)
			return err
		}
		s.Pipeline = p
	}

	s.WebHook = &w

	ctx.JSON(200, s)
	return nil
}

func ShowAll(ctx *context.Context, db *database.Database) {
	all, mine, err := GetWebHooks(ctx, db)
	if err != nil {
		ctx.ServerError("Failed finding webhook", err)
		return
	}

	all = append(all, mine...)

	ctx.JSON(200, all)
}
