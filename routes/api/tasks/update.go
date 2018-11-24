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
	"time"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"

	rabbithole "github.com/michaelklishin/rabbit-hole"
)

type UpdateTaskForm struct {
	Id         string `form:"id" binding:"Required"`
	Status     string `form:"status"`
	Result     string `form:"result"`
	Output     string `form:"output"`
	ExitStatus string `form:"exit_status"`
	Field      string `form:"field"`
	Value      string `form:"value"`
	Key        string ` form:"key"`
}

func UpdateTaskField(f UpdateTaskForm, rmqc *rabbithole.Client, ctx *context.Context, db *database.Database) error {
	mytask, err := db.Driver.GetTask(db.Config, f.Id)
	if err != nil {
		return err
	}
	if !ctx.CheckTaskPermissions(&mytask) {
		return errors.New("Moar permissions are required for this user")
	}
	if len(f.Field) > 0 && len(f.Value) > 0 {
		db.Driver.UpdateTask(f.Id, map[string]interface{}{
			f.Field: f.Value,
		})
		// Set state once we have task's exit status
		if f.Field == "exit_status" {

			// TODO: To change to properly handling of fields as well, but for now
			// we can cope with it.

			db.Driver.UpdateTask(f.Id, map[string]interface{}{
				"exit_status": f.Value,
			})

			if !mytask.IsStopped() {
				db.Driver.UpdateTask(f.Id, map[string]interface{}{
					"result": mytask.DecodeStatus(f.Value),
				})

				t, err := db.Driver.GetTask(db.Config, f.Id)
				if err != nil {
					return err
				}
				t.HandleStatus(db.Config.GetStorage().NamespacePath, db.Config.GetStorage().ArtefactPath)
			}
		}
	}

	return nil
}

func SetNode(f UpdateTaskForm, ctx *context.Context, db *database.Database) error {
	mytask, err := db.Driver.GetTask(db.Config, f.Id)
	if err != nil {
		return err
	}
	if !ctx.CheckTaskPermissions(&mytask) {
		return errors.New("Moar permissions are required for this user")
	}
	node, err := db.Driver.GetNodeByKey(f.Key)
	if err != nil {
		return err
	}
	db.Driver.UpdateTask(f.Id, map[string]interface{}{
		"node_id": node.ID,
	})

	return nil
}

func AppendToTask(f UpdateTaskForm, rmqc *rabbithole.Client, ctx *context.Context, db *database.Database) string {

	if len(f.Output) > 0 {

		mytask, err := db.Driver.GetTask(db.Config, f.Id)
		if err != nil {
			return ":("
		}
		err = mytask.AppendBuildLog(f.Output, db.Config.GetStorage().ArtefactPath, db.Config.GetWeb().LockPath)
		if err != nil {
			fmt.Println("Can't write to buildlog: ", err.Error())
			return "Error: " + err.Error()
		}
	}
	return "OK"
}

func UpdateTask(f UpdateTaskForm, rmqc *rabbithole.Client, ctx *context.Context, db *database.Database) string {

	if len(f.Status) > 0 {
		db.Driver.UpdateTask(f.Id, map[string]interface{}{
			"status": f.Status,
		})
	}

	if len(f.Output) > 0 {
		db.Driver.UpdateTask(f.Id, map[string]interface{}{
			"output": f.Output,
		})
	}

	if len(f.Result) > 0 {
		db.Driver.UpdateTask(f.Id, map[string]interface{}{
			"result":   f.Result,
			"end_time": time.Now().Format("20060102150405"),
		})
	}

	t, err := db.Driver.GetTask(db.Config, f.Id)
	if err != nil {
		return ":( "
	}
	t.HandleStatus(db.Config.GetStorage().NamespacePath, db.Config.GetStorage().ArtefactPath)

	return "OK"
}
