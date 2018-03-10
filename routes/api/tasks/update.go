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
	"time"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	"github.com/MottainaiCI/mottainai-server/pkg/db"

	machinery "github.com/RichardKnop/machinery/v1"
	rabbithole "github.com/michaelklishin/rabbit-hole"
)

type UpdateTaskForm struct {
	Id         int    `form:"id" binding:"Required"`
	Status     string `form:"status"`
	APIKey     string `form:"apikey" binding:"Required"`
	Result     string `form:"result"`
	Output     string `form:"output"`
	ExitStatus string `form:"exit_status"`
	Field      string `form:"field"`
	Value      string `form:"value"`
}

func UpdateTaskField(f UpdateTaskForm, rmqc *rabbithole.Client, ctx *context.Context, rabbit *machinery.Server, db *database.Database) string {

	ok, _ := ValidateNodeKey(&f, db)
	if ok == false {
		ctx.NotFound()
		return ":( "
	}
	if len(f.Field) > 0 && len(f.Value) > 0 {
		db.UpdateTask(f.Id, map[string]interface{}{
			f.Field: f.Value,
		})
	}

	return "OK"
}

func AppendToTask(f UpdateTaskForm, rmqc *rabbithole.Client, ctx *context.Context, rabbit *machinery.Server, db *database.Database) string {

	ok, _ := ValidateNodeKey(&f, db)
	if ok == false {
		ctx.NotFound()
		fmt.Println("Invalid KEY!!!!")
		return ":( "
	}

	if len(f.Output) > 0 {

		mytask, err := db.GetTask(f.Id)
		if err != nil {
			return ":("
		}
		mytask.AppendBuildLog(f.Output)
	}
	return "OK"
}

func UpdateTask(f UpdateTaskForm, rmqc *rabbithole.Client, ctx *context.Context, rabbit *machinery.Server, db *database.Database) string {
	ok, _ := ValidateNodeKey(&f, db)

	if ok == false {
		ctx.NotFound()
		return ":( "
	}

	if len(f.Status) > 0 {
		db.UpdateTask(f.Id, map[string]interface{}{
			"status": f.Status,
		})
	}

	if len(f.Output) > 0 {
		db.UpdateTask(f.Id, map[string]interface{}{
			"output": f.Output,
		})
	}

	if len(f.Result) > 0 {
		db.UpdateTask(f.Id, map[string]interface{}{
			"result":   f.Result,
			"end_time": time.Now().Format("20060102150405"),
		})
		t, err := db.GetTask(f.Id)
		if err != nil {
			return ":("
		}
		t.HandleStatus()
	}

	return "OK"
}
