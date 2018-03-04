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
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	"github.com/MottainaiCI/mottainai-server/pkg/db"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"
)

func NamespaceTag(ctx *context.Context, db *database.Database) (string, error) {
	name := ctx.Params(":name")
	taskid := ctx.ParamsInt(":taskid")
	name, _ = utils.Strip(name)

	if len(taskid) == 0 || len(name) == 0 {
		return ":(", err
	}

	task, err := db.GetTask(taskid)
	if err != nil {
		return "", err
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
	source := filepath.Join(setting.Configuration.ArtefactPath, strconv.Itoa(task.ID))

	err = filepath.Walk(source, func(path string, f os.FileInfo, err error) error {
		_, file := filepath.Split(path)
		rel := strings.Replace(path, source, "", 1)
		rel = strings.Replace(rel, file, "", 1)

		fi, err := os.Stat(path)
		if err != nil {
			return err
		}
		switch mode := fi.Mode(); {
		case mode.IsDir():
			// do directory stuff
			return err
		case mode.IsRegular():
			os.MkdirAll(filepath.Join(setting.Configuration.NamespacePath, name, rel), os.ModePerm)
			utils.CopyFile(
				path,
				filepath.Join(setting.Configuration.NamespacePath, name, rel, file),
			)
		}
		return nil
	})
	if err != nil {
		return ":(", err
	}
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
