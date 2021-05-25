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
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
)

type ArtefactForm struct {
	TaskID     string                `form:"taskid" binding:"Required"`
	Name       string                `form:"name"`
	Path       string                `form:"path"`
	FileUpload *multipart.FileHeader `form:"file"`
}

func AllArtefactList(ctx *context.Context, db *database.Database) {
	artefacts := db.Driver.AllArtefacts()
	ctx.JSON(200, artefacts)
}

func ArtefactList(ctx *context.Context, db *database.Database) error {
	id := ctx.Params(":id")
	// artefacts, err := db.Driver.GetTaskArtefacts(id)
	// if err != nil {
	// 	panic(err)
	// }

	// ns, err := db.SearchNamespace(name)
	// if err != nil {
	// 	ctx.JSON(200, ns)
	// }
	t, err := db.Driver.GetTask(db.Config, id)
	if !ctx.CheckTaskPermissions(&t) {
		ctx.NoPermission()
		return nil
	}
	if err != nil {
		return err
	}

	artefacts := t.Artefacts(db.Config.GetStorage().ArtefactPath)

	ctx.JSON(200, artefacts)
	return nil
}

func ArtefactUpload(uf ArtefactForm, ctx *context.Context, db *database.Database) error {

	if uf.TaskID == "" {
		return errors.New("Invalid artefact without task id")
	}

	fmt.Println(fmt.Sprintf("[%s] Receiving artefact %s for path %s.",
		uf.TaskID, uf.Name, uf.Path))

	file, err := uf.FileUpload.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	task, err := db.Driver.GetTask(db.Config, uf.TaskID)
	if err != nil {
		return err
	}

	if task.ID == "" {
		return errors.New("Invalid task id")
	}

	if !ctx.CheckTaskPermissions(&task) {
		ctx.NoPermission()
		return nil
	}

	var f *os.File
	ctx.Invoke(func(config *setting.Config) {
		os.MkdirAll(filepath.Join(config.GetStorage().ArtefactPath, task.ID, uf.Path), os.ModePerm)
		f, err = os.OpenFile(filepath.Join(config.GetStorage().ArtefactPath, task.ID, uf.Path, uf.Name), os.O_WRONLY|os.O_CREATE, os.ModePerm)
	})

	if err != nil {
		return err
	}

	defer f.Close()
	io.Copy(f, file)

	db.Driver.CreateArtefact(map[string]interface{}{
		"name": uf.Name,
		"path": uf.Path,
		"task": task.ID,
		//"namespace": task.Namespace,
	})

	ctx.APIActionSuccess()
	return nil
}
