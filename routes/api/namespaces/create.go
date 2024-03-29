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

package namespacesapi

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"
)

const NameSpacesPrefix = "::"

func NamespaceCreate(ctx *context.Context, db *database.Database) error {
	name := ctx.Params(":name")
	name, _ = utils.Strip(name)

	if !ctx.CheckNamespaceBelongs(name) {
		ctx.NoPermission()
		return nil
	}

	// docID, _ := db.CreateNamespace(map[string]interface{}{
	// 	"name": name,
	// 	"path": name,
	// })

	err := os.MkdirAll(filepath.Join(db.Config.GetStorage().NamespacePath, name), os.ModePerm)
	if err != nil {
		return err
	}

	ctx.APIActionSuccess()
	return nil
}

type NamespaceForm struct {
	Namespace  string                `form:"namespace" binding:"Required"`
	Name       string                `form:"name"`
	Path       string                `form:"path"`
	FileUpload *multipart.FileHeader `form:"file"`
}

func NamespaceUpload(uf NamespaceForm, ctx *context.Context, db *database.Database) error {

	file, err := uf.FileUpload.Open()
	defer file.Close()

	if err != nil {
		return err
	}

	if !ctx.CheckNamespaceBelongs(uf.Namespace) {
		ctx.NoPermission()
		return nil
	}

	os.MkdirAll(filepath.Join(db.Config.GetStorage().NamespacePath, uf.Namespace, uf.Path), os.ModePerm)
	f, err := os.OpenFile(filepath.Join(
		db.Config.GetStorage().NamespacePath, uf.Namespace, uf.Path, uf.Name),
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	io.Copy(f, file)

	ctx.APIActionSuccess()
	return nil
}
