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

package tasksapi

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
)

type ArtefactForm struct {
	TaskID     int                   `form:"taskid" binding:"Required"`
	Name       string                `form:"name"`
	Path       string                `form:"path"`
	FileUpload *multipart.FileHeader `form:"file"`
}

func AllArtefactList(ctx *context.Context, db *database.Database) {
	artefacts := db.Driver.AllArtefacts()
	ctx.JSON(200, artefacts)
}

func ArtefactList(ctx *context.Context, db *database.Database) {
	id := ctx.ParamsInt(":id")
	// artefacts, err := db.Driver.GetTaskArtefacts(id)
	// if err != nil {
	// 	panic(err)
	// }

	// ns, err := db.SearchNamespace(name)
	// if err != nil {
	// 	ctx.JSON(200, ns)
	// }
	t, err := db.Driver.GetTask(id)
	if !ctx.CheckTaskPermissions(&t) {
		return
	}
	if err != nil {
		panic(err)
	}
	artefacts := t.Artefacts()

	ctx.JSON(200, artefacts)
}

func ArtefactUpload(uf ArtefactForm, ctx *context.Context, db *database.Database) error {

	file, err := uf.FileUpload.Open()
	defer file.Close()

	task, err := db.Driver.GetTask(uf.TaskID)
	if err != nil {
		return err
	}

	if !ctx.CheckTaskPermissions(&task) {
		return errors.New("Insufficient permissions")
	}

	os.MkdirAll(filepath.Join(setting.Configuration.ArtefactPath, task.ID, uf.Path), os.ModePerm)
	f, err := os.OpenFile(filepath.Join(setting.Configuration.ArtefactPath, task.ID, uf.Path, uf.Name), os.O_WRONLY|os.O_CREATE, os.ModePerm)

	defer f.Close()
	io.Copy(f, file)

	db.Driver.CreateArtefact(map[string]interface{}{
		"name": uf.Name,
		"path": uf.Path,
		"task": task.ID,
		//"namespace": task.Namespace,
	})
	return nil
}
