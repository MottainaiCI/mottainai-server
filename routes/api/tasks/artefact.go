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
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"

	"github.com/MottainaiCI/mottainai-server/pkg/db"

	"github.com/MottainaiCI/mottainai-server/pkg/settings"
)

type ArtefactForm struct {
	TaskID     int                   `form:"taskid" binding:"Required"`
	Name       string                `form:"name"`
	Path       string                `form:"path"`
	FileUpload *multipart.FileHeader `form:"file"`
}

func AllArtefactList(ctx *context.Context, db *database.Database) {
	artefacts := db.AllArtefacts()
	ctx.JSON(200, artefacts)
}

func ArtefactList(ctx *context.Context, db *database.Database) {
	id := ctx.ParamsInt(":id")
	// artefacts, err := db.GetTaskArtefacts(id)
	// if err != nil {
	// 	panic(err)
	// }

	// ns, err := db.SearchNamespace(name)
	// if err != nil {
	// 	ctx.JSON(200, ns)
	// }

	artefacts := utils.TreeList(filepath.Join(setting.Configuration.ArtefactPath, strconv.Itoa(id)))

	ctx.JSON(200, artefacts)
}

func ArtefactUpload(uf ArtefactForm, ctx *context.Context, db *database.Database) string {

	file, err := uf.FileUpload.Open()

	task, err := db.GetTask(uf.TaskID)
	defer file.Close()
	if err != nil {
		return err.Error()
	}

	os.MkdirAll(filepath.Join(setting.Configuration.ArtefactPath, strconv.Itoa(task.ID), uf.Path), os.ModePerm)
	f, err := os.OpenFile(filepath.Join(setting.Configuration.ArtefactPath, strconv.Itoa(task.ID), uf.Path, uf.Name), os.O_WRONLY|os.O_CREATE, os.ModePerm)

	defer f.Close()
	io.Copy(f, file)

	db.CreateArtefact(map[string]interface{}{
		"name": uf.Name,
		"path": uf.Path,
		"task": task.ID,
		//"namespace": task.Namespace,
	})
	return "OK"
}
