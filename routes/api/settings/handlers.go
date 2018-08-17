/*

Copyright (C) 2018  Ettore Di Giacinto <mudler@gentoo.org>
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

package settingsapi

import (
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
)

func ShowAll(ctx *context.Context, db *database.Database) {
	ctx.JSON(200, db.Driver.AllSettings())
}

func APICreate(ctx *context.Context, db *database.Database, s Setting) string {

	id, err := db.Driver.InsertSetting(&setting.Setting{Key: s.Key, Value: s.Value})
	if err != nil {
		ctx.ServerError("Failed insert", err)
		return ""
	}
	return id
}

func APIRemove(db *database.Database, ctx *context.Context) {
	key := ctx.Params(":key")
	uuu, err := db.Driver.GetSettingByKey(key)
	if err != nil {
		ctx.ServerError("Failed remove", err)
	}

	err = db.Driver.DeleteSetting(uuu.ID)
	if err != nil {
		ctx.ServerError("Failed remove", err)
	}
}

func APIUpdate(ctx *context.Context, db *database.Database, s Setting) {
	uuu, err := db.Driver.GetSettingByKey(s.Key)
	if err != nil {
		ctx.ServerError("Failed update", err)
	}
	uuu.Key = s.Key
	uuu.Value = s.Value
	db.Driver.UpdateSetting(uuu.ID, uuu.ToMap())
}
