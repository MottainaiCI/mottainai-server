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

package context

import (
	"runtime"
	"strings"

	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	utils "github.com/MottainaiCI/mottainai-server/pkg/utils"

	event "github.com/MottainaiCI/mottainai-server/pkg/event"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	storage "github.com/MottainaiCI/mottainai-server/pkg/storage"
	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	task "github.com/MottainaiCI/mottainai-server/pkg/tasks"
)

const NameSpacesPrefix = "::"

var noperm = map[string]string{
	"message": "It seems you don't have enough permissions to perform this operation, i'm sorry.",
}

func (c *Context) CheckPlanPermissions(plan *task.Plan) bool {
	return c.CheckTaskPermissions(plan.Task)
}

func (c *Context) CheckPipelinePermissions(pip *task.Pipeline) bool {
	if c.User.IsManagerOrAdmin() {
		return true
	}

	// Return true if Admin or Owner of it
	if c.User.ID == pip.Owner {
		return true
	}

	c.NoPermission()
	return false
}

func (c *Context) CheckUser() bool {
	if c.User != nil {
		return true
	}
	c.NoPermission()
	return false
}

func (c *Context) CheckUserOrManager() bool {
	if c.User == nil {
		return false
	}

	if c.User.IsManagerOrAdmin() {
		return true
	}
	return false
}

func (c *Context) CheckTaskPermissions(task *task.Task) bool {
	if !c.CheckUser() {
		return false
	}

	if c.User.IsManagerOrAdmin() {
		return true
	}

	// Return true if Admin or Owner of it
	if c.User.ID == task.Owner {
		return true
	}

	c.NoPermission()
	return false
}

func (c *Context) CheckStoragePermissions(storage *storage.Storage) bool {
	if c.User.IsManagerOrAdmin() {
		return true
	}

	// Return true if Admin or Owner of it
	if c.User.ID == storage.Owner {
		return true
	}

	return false
}

// namepath checks
func (c *Context) CheckStorageBelongs(storage string) bool {
	if len(storage) > 0 &&
		!c.User.IsManagerOrAdmin() &&
		!strings.HasPrefix(storage, c.User.Name+NameSpacesPrefix) {

		c.NoPermission()
		return false
	}

	return true
}

func (c *Context) CheckNamespaceBelongs(namespace string) bool {
	if len(namespace) > 0 &&
		!c.User.IsManagerOrAdmin() &&
		!strings.HasPrefix(namespace, c.User.Name+NameSpacesPrefix) {

		c.NoPermission()
		return false
	}

	return true
}

// STATIC ROUTES AUTH CHECK
func CheckArtefactPermission(ctx *Context) bool {
	var err error
	var task agenttasks.Task
	file := ctx.Req.URL.Path
	if !ctx.IsLogged {
		ctx.NoPermission()
		return false
	}

	db := database.Instance().Driver
	ctx.Invoke(func(config *setting.Config) {
		file, _ = config.GetWeb().NormalizePath(file)
	})
	segments := strings.Split(file, "/")

	r := utils.NoEmptySlice(segments)
	if len(r) < 2 {
		return false
	}
	id := r[1]

	ctx.Invoke(func(config *setting.Config) {
		task, err = db.GetTask(config, id)
	})
	if err != nil {
		ctx.ServerError(err.Error(), err)
		return false
	}

	cModeSet, _ := db.GetSettingByKey(setting.SYSTEM_COMMUNITY_ENABLED)
	cMode := cModeSet.Value == "true"
	if !cMode && !task.IsOwner(ctx.User.ID) && !ctx.User.IsManagerOrAdmin() {
		ctx.NoPermission()
		return false
	}

	return true
}

func CheckNamespacePermission(ctx *Context) bool {
	if ctx.IsLogged {
		return true
	}
	return false
}

func CheckStoragePermission(ctx *Context) bool {
	file := ctx.Req.URL.Path
	if !ctx.IsLogged {
		ctx.NoPermission()
		return false
	}
	db := database.Instance().Driver
	ctx.Invoke(func(config *setting.Config) {
		file, _ = config.GetWeb().NormalizePath(file)
	})
	segments := strings.Split(file, "/")

	r := utils.NoEmptySlice(segments)
	if len(r) < 2 {
		return false
	}
	name := r[1]

	storage, err := db.SearchStorage(name)
	if err != nil {
		ctx.NotFound()
		return false
	}
	if !storage.IsOwner(ctx.User.ID) && !ctx.User.IsManagerOrAdmin() {
		ctx.NoPermission()
		return false
	}

	return true
}

// API Helpers
func (c *Context) APIActionFailed(id, objType, err, data string, statusCode int) {
	c.APIEventReport(event.APIResponse{
		ObjType:   objType,
		Processed: "true",
		Status:    "ko",
		Data:      data,
		ID:        id,
		Error:     err,
	}, statusCode)
}

func (c *Context) APIActionSuccess() {
	c.APIEventReport(event.APIResponse{ObjType: "action", Processed: "true", Status: "ok"}, 200)
}

func (c *Context) APICreationSuccess(id, objType string) {
	c.APIEventReport(event.APIResponse{ID: id, ObjType: objType, Processed: "true", Status: "ok"}, 200)
}

func (c *Context) APIPayload(id, objType, data string) {
	c.APIEventReport(event.APIResponse{Data: data, ID: id, ObjType: objType, Processed: "true", Status: "ok"}, 200)
}

func (c *Context) APIEventData(data string) {
	c.APIEventReport(event.APIResponse{Data: data, Processed: "true", Status: "ok"}, 200)
}

func (c *Context) APIEventReport(e event.APIResponse, statusCode int) {
	pc, _, _, ok := runtime.Caller(2)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		// TODO: do this only with debug enable
		e.Event = details.Name()
	}
	c.JSON(statusCode, e)
}
