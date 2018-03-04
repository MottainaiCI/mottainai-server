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

package storagesapi

import (
	"path/filepath"

	"github.com/MottainaiCI/mottainai-server/pkg/utils"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	"github.com/MottainaiCI/mottainai-server/pkg/db"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
)

func StorageList(ctx *context.Context, db *database.Database) {
	ns := db.AllStorages()

	//source := filepath.Join(setting.Configuration.StoragePath)
	//ns, _ := utils.ListDirs(source)

	ctx.JSON(200, ns)
}

func StorageShow(ctx *context.Context, db *database.Database) {
	id := ctx.ParamsInt(":id")
	ns, err := db.GetStorage(id)
	if err != nil {
		ctx.NotFound()
		return
	}
	//source := filepath.Join(setting.Configuration.StoragePath)
	//ns, _ := utils.ListDirs(source)

	ctx.JSON(200, ns)
}

func StorageListArtefacts(ctx *context.Context, db *database.Database) {
	id := ctx.ParamsInt(":id")

	st, err := db.GetStorage(id)
	if err != nil {
		ctx.NotFound()
		return
	}
	// ns, err := db.SearchStorage(name)
	// if err != nil {
	// 	ctx.JSON(200, ns)
	// }
	source := filepath.Join(setting.Configuration.StoragePath, st.Path)

	artefacts := utils.TreeList(source)

	// artefacts, err := db.GetStorageArtefacts(ns.ID)
	// if err != nil {
	// 	ctx.JSON(200, artefacts)
	// }
	ctx.JSON(200, artefacts)
}
