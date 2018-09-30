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
	"errors"
	"os"
	"path/filepath"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"
)

func NamespaceDelete(ctx *context.Context, db *database.Database) (string, error) {
	name := ctx.Params(":name")
	name, _ = utils.Strip(name)

	if !ctx.CheckNamespaceBelongs(name) {
		return ":(", errors.New("Moar permissions are required for this user")
	}

	//err := db.DeleteNamespace(id)
	err := os.RemoveAll(filepath.Join(db.Config.GetStorage().NamespacePath, name))
	if err != nil {
		return ":(", err
	}
	return "OK", nil
}

func NamespaceRemovePath(ctx *context.Context, db *database.Database) (string, error) {
	name := ctx.Params(":name")
	name, _ = utils.Strip(name)
	path := ctx.Params(":path")

	if !ctx.CheckNamespaceBelongs(name) {
		return ":(", errors.New("Moar permissions are required for this user")
	}

	err := os.RemoveAll(filepath.Join(db.Config.GetStorage().NamespacePath, name, path))
	if err != nil {
		return ":(", err
	}
	return "OK", nil
}
