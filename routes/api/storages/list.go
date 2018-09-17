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

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
)

func StorageList(ctx *context.Context, db *database.Database) {
	ns := db.Driver.AllStorages()

	//source := filepath.Join(db.Config.StoragePath)
	//ns, _ := utils.ListDirs(source)

	ctx.JSON(200, ns)
}

func StorageShow(ctx *context.Context, db *database.Database) {
	id := ctx.Params(":id")
	ns, err := db.Driver.GetStorage(id)

	if err != nil {
		ctx.ServerError(err.Error(), err)
		//ctx.NotFound()
		return
	}
	if !ctx.CheckStoragePermissions(&ns) {
		ctx.NoPermission()
		return
	}

	//source := filepath.Join(db.Config.StoragePath)
	//ns, _ := utils.ListDirs(source)

	ctx.JSON(200, ns)
}

func StorageListArtefacts(ctx *context.Context, db *database.Database) {
	id := ctx.Params(":id")

	st, err := db.Driver.GetStorage(id)
	if err != nil {
		ctx.ServerError(err.Error(), err)
		return
	}
	if !ctx.CheckStoragePermissions(&st) {
		ctx.NoPermission()
		return
	}
	// ns, err := db.Driver.SearchStorage(name)
	// if err != nil {
	// 	ctx.JSON(200, ns)
	// }
	source := filepath.Join(db.Config.StoragePath, st.Path)
	artefacts := utils.TreeList(source)

	// artefacts, err := db.Driver.GetStorageArtefacts(ns.ID)
	// if err != nil {
	// 	ctx.JSON(200, artefacts)
	// }
	ctx.JSON(200, artefacts)
}
