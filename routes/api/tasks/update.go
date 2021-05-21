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

	logging "github.com/MottainaiCI/mottainai-server/pkg/logging"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	logrus "github.com/sirupsen/logrus"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
)

type UpdateTaskForm struct {
	Id        string `form:"id" binding:"Required"`
	Status    string `form:"status"`
	Result    string `form:"result"`
	Output    string `form:"output"`
	xitStatus string `form:"exit_status"`
	Field     string `form:"field"`
	Value     string `form:"value"`
	Key       string ` form:"key"`
}

func SyncTaskLastUpdate(id string, db *database.Database) {
	db.Driver.UpdateTask(id, map[string]interface{}{
		"last_update_time": time.Now().UTC().Format("20060102150405"),
	})
}

func UpdateTaskField(f UpdateTaskForm, ctx *context.Context, db *database.Database) {
	mytask, err := db.Driver.GetTask(db.Config, f.Id)
	if err != nil {
		ctx.ServerError("Failed getting task", err)
		return
	}
	if !ctx.CheckTaskPermissions(&mytask) {
		ctx.NoPermission()
		return
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
					ctx.ServerError("Failed getting task", err)
					return
				}
				t.HandleStatus(db.Config.GetStorage().NamespacePath, db.Config.GetStorage().ArtefactPath)
			}
		}

		SyncTaskLastUpdate(f.Id, db)
	}

	ctx.APIActionSuccess()
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
		"node_id":          node.ID,
		"last_update_time": time.Now().UTC().Format("20060102150405"),
	})

	ctx.APIActionSuccess()

	return nil
}

func AppendToTask(logger *logging.Logger, f UpdateTaskForm, ctx *context.Context, db *database.Database) error {
	mytask, err := db.Driver.GetTask(db.Config, f.Id)
	if err != nil {
		return err
	}
	if !ctx.CheckTaskPermissions(&mytask) {
		ctx.NoPermission()
		return nil
	}
	if len(f.Output) > 0 {
		mytask, err := db.Driver.GetTask(db.Config, f.Id)
		if err != nil {
			return errors.New("Task not found")
		}
		err = mytask.AppendBuildLog(f.Output, db.Config.GetStorage().ArtefactPath, db.Config.GetWeb().LockPath)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"component": "api",
				"error":     err.Error(),
			}).Error("Can't write to buildlog")
			return err
		}
	}
	SyncTaskLastUpdate(f.Id, db)

	ctx.APIActionSuccess()
	return nil
}

func UpdateTask(f UpdateTaskForm, ctx *context.Context, db *database.Database) error {
	t, err := db.Driver.GetTask(db.Config, f.Id)
	if err != nil {
		return errors.New("Task not found")
	}

	upd := map[string]interface{}{}
	updtime := time.Now().UTC().Format("20060102150405")

	if len(f.Status) > 0 {
		upd["status"] = f.Status
		if f.Status == setting.TASK_STATE_RUNNING {
			upd["start_time"] = updtime
		}
	}

	if len(f.Output) > 0 {
		upd["output"] = f.Output
	}

	if len(f.Result) > 0 {
		upd["result"] = f.Result
		upd["end_time"] = updtime
	}

	if len(upd) > 0 {
		upd["last_update_time"] = updtime

		fmt.Println(fmt.Sprintf("Task %s update: %s", f.Id, upd))
		db.Driver.UpdateTask(f.Id, upd)

		t.HandleStatus(
			db.Config.GetStorage().NamespacePath,
			db.Config.GetStorage().ArtefactPath,
		)
	} else {
		fmt.Println(fmt.Sprintf("For Task %s no updates received", f.Id))
	}

	if f.Status == setting.TASK_STATE_RUNNING && t.PipelineID != "" {
		// Retrieve pipeline
		pipeline, err := db.Driver.GetPipeline(db.Config, t.PipelineID)
		if err == nil && pipeline.ID != "" {
			// Check if the pipeline is already in queue as in progress
			queue, err := db.Driver.GetQueueByKey(pipeline.Queue)
			if err != nil {
				fmt.Println("Error on retrieve queue data for queue " + pipeline.Queue +
					": " + err.Error())
			} else {
				if queue.Qid == "" {
					fmt.Println("Error on retrieve queue data for queue " + pipeline.Queue + ".")
				} else if !queue.HasPipelineRunning(t.PipelineID) {
					err = db.Driver.AddPipelineInProgress2Queue(queue.Qid, t.PipelineID)
					if err != nil {
						fmt.Println("Error on add pipeline " + t.PipelineID +
							" in the waiting queue for queue " + queue.Qid)
					}

					// Update start time of the pipeline
					err = db.Driver.UpdatePipeline(t.PipelineID, map[string]interface{}{
						"start_time": updtime,
					})
					if err != nil {
						fmt.Println("Error on update pipeline start time " + err.Error())
					}
				}
			}

		} else if err != nil {
			// else ignoring invalid pipeline
			fmt.Println("Error on retrieve pipeline data for task " +
				t.ID + ": " + err.Error())
		}

	}

	ctx.APIActionSuccess()
	return nil
}
