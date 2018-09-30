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
	"path/filepath"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
)

func Namespaces(namespacePath string) []string {
	//ns := db.AllNamespaces()

	source := filepath.Join(namespacePath)
	ns, _ := utils.ListDirs(source)
	return ns
}

func NamespaceList(ctx *context.Context, db *database.Database) {

	var ns []string
	ctx.Invoke(func(config *setting.Config) {
		ns = Namespaces(config.GetStorage().NamespacePath)
	})
	ctx.JSON(200, ns)
}

func NamespaceArtefacts(namespace string, namespacePath string) []string {
	name, _ := utils.Strip(namespace)

	// ns, err := db.SearchNamespace(name)
	// if err != nil {
	// 	ctx.JSON(200, ns)
	// }
	source := filepath.Join(namespacePath, name)

	artefacts := utils.TreeList(source)
	return artefacts
}

func NamespaceListArtefacts(ctx *context.Context, db *database.Database) {
	name := ctx.Params(":name")

	var artefacts []string
	ctx.Invoke(func(config *setting.Config) {
		artefacts = NamespaceArtefacts(name, config.GetStorage().NamespacePath)
	})

	// ns, err := db.SearchNamespace(name)
	// if err != nil {
	// 	ctx.JSON(200, ns)
	// }

	// artefacts, err := db.GetNamespaceArtefacts(ns.ID)
	// if err != nil {
	// 	ctx.JSON(200, artefacts)
	// }
	ctx.JSON(200, artefacts)
}
