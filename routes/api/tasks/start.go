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
	"fmt"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	"github.com/MottainaiCI/mottainai-server/pkg/db"
	"github.com/MottainaiCI/mottainai-server/pkg/tasks"

	machinery "github.com/RichardKnop/machinery/v1"
)

func APISendStartTask(ctx *context.Context, db *database.Database, rabbit *machinery.Server) string {
	_, err := SendStartTask(ctx, db, rabbit)
	if err != nil {
		ctx.NotFound()
		return ":("
	}
	return "OK"
}

func SendStartTask(ctx *context.Context, db *database.Database, rabbit *machinery.Server) (string, error) {
	id := ctx.ParamsInt(":id")
	fmt.Println("Starting task ", id)

	mytask, err := db.GetTask(id)
	if err != nil {
		return "", err
	}
	if mytask.IsWaiting() || mytask.IsRunning() {
		return "WAITING/RUNNING", nil
	}

	err = SendTask(db, rabbit, id)
	if err != nil {
		return ":( ", err
	} else {
		return "OK", nil
	}
}

func SendTask(db *database.Database, rabbit *machinery.Server, docID int) error {

	task, err := db.GetTask(docID)
	if err != nil {
		panic(err)
	}

	db.UpdateTask(docID, map[string]interface{}{"status": "waiting", "result": "none"})

	fmt.Printf("Task Source: %v, Script: %v, Yaml: %v, Directory: %v, TaskName: %v", task.Source, task.Script, task.Yaml, task.Directory, task.TaskName)

	_, err = agenttasks.SendTask(rabbit, task.TaskName, docID)
	if err != nil {
		fmt.Printf("Could not send task: %s", err.Error())
	}
	return err
}
