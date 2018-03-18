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
	"github.com/MottainaiCI/mottainai-server/pkg/context"
	"github.com/MottainaiCI/mottainai-server/pkg/db"
	"github.com/MottainaiCI/mottainai-server/pkg/namespace"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"
)

func NamespaceTag(ctx *context.Context, db *database.Database) (string, error) {
	name := ctx.Params(":name")
	taskid := ctx.ParamsInt(":taskid")
	name, _ = utils.Strip(name)

	if len(name) == 0 {
		return ":( No namespace name given", nil
	}

	task, err := db.GetTask(taskid)
	if err != nil {
		return "", err
	}
	ns := namespace.NewFromMap(map[string]interface{}{"name": name, "path": name})
	err = ns.Tag(task.ID)
	if err != nil {
		return ":(", err
	}
	// artefacts, err := db.GetTaskArtefacts(taskid)
	// if err != nil {
	// 	return "", err
	// }
	// ns, err := db.SearchNamespace(name)
	// if err != nil {
	// 	return "", err
	// }
	// os.RemoveAll(filepath.Join(setting.Configuration.NamespacePath, ns.Name))
	// os.MkdirAll(filepath.Join(setting.Configuration.NamespacePath, ns.Name), os.ModePerm)

	// for _, artefact := range artefacts {
	// 	fmt.Println("Moving artefact: " + artefact.Name)
	// 	db.UpdateArtefact(artefact.ID, map[string]interface{}{
	// 		"namespace": ns.ID,
	// 	})
	//
	// 	os.MkdirAll(filepath.Join(setting.Configuration.NamespacePath, ns.Name, artefact.Path), os.ModePerm)
	// 	utils.CopyFile(
	// 		filepath.Join(setting.Configuration.ArtefactPath, strconv.Itoa(task.ID), artefact.Path, artefact.Name),
	// 		filepath.Join(setting.Configuration.NamespacePath, ns.Name, artefact.Path, artefact.Name),
	// 	)
	// 	fmt.Println("Copy: ",
	// 		filepath.Join(setting.Configuration.ArtefactPath, strconv.Itoa(task.ID), artefact.Path, artefact.Name),
	// 		filepath.Join(setting.Configuration.NamespacePath, ns.Name, artefact.Path, artefact.Name),
	// 	)
	//
	// }

	return "OK", nil
}

func NamespaceClone(ctx *context.Context, db *database.Database) (string, error) {
	name := ctx.Params(":name")
	from := ctx.Params(":from")
	name, _ = utils.Strip(name)
	from, _ = utils.Strip(from)

	if len(name) == 0 {
		return ":( No namespace name given", nil
	}

	ns := namespace.NewFromMap(map[string]interface{}{"name": name, "path": name})
	err := ns.Clone(namespace.NewFromMap(map[string]interface{}{"name": from, "path": from}))
	if err != nil {
		return ":(", err
	}

	return "OK", nil
}
