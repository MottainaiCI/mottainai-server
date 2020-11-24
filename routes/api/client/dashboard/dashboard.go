/*

Copyright (C) 2017-2018  Ettore Di Giacinto <mudler@gentoo.org>
Copyright (C) 2020       Adib Saad <adib.saad@gmail.com>
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

package dashboard

import (
  database "github.com/MottainaiCI/mottainai-server/pkg/db"
  v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"
  "gopkg.in/macaron.v1"
)

type DashboardData struct{
  Total int `json:"total"`
  Running int `json:"running"`
  Waiting int `json:"waiting"`
  Error int `json:"error"`
  Failed int `json:"failed"`
  Success int `json:"success"`
  Stopped int `json:"stopped"`
  Stop int `json:"stop"`
}

func Stats(ctx *macaron.Context, db *database.Database) error {
  rtasks, e := db.Driver.GetTaskByStatus(db.Config, "running")
  if e != nil {
    return e
  }
  running_tasks := len(rtasks)
  wtasks, e := db.Driver.GetTaskByStatus(db.Config, "waiting")
  if e != nil {
    return e
  }
  waiting_tasks := len(wtasks)
  etasks, e := db.Driver.GetTaskByStatus(db.Config, "error")
  if e != nil {
    return e
  }
  error_tasks := len(etasks)
  ftasks, e := db.Driver.GetTaskByStatus(db.Config, "failed")
  if e != nil {
    return e
  }
  failed_tasks := len(ftasks)
  stasks, e := db.Driver.GetTaskByStatus(db.Config, "success")
  if e != nil {
    return e
  }
  succeeded_tasks := len(stasks)
  stoppedtasks, e := db.Driver.GetTaskByStatus(db.Config, "stopped")
  if e != nil {
    return e
  }
  stopped_tasks := len(stoppedtasks)

  instoptasks, e := db.Driver.GetTaskByStatus(db.Config, "stop")
  if e != nil {
    return e
  }
  instop_tasks := len(instoptasks)

  p := DashboardData{
    len(db.Driver.ListDocs("Tasks")),
    running_tasks,
    waiting_tasks,
    error_tasks,
    succeeded_tasks,
    failed_tasks,
    stopped_tasks,
    instop_tasks,
  }
  ctx.JSON(200, &p)
  return nil
}

func Setup(m *macaron.Macaron) {
  v1.Schema.GetClientRoute("dashboard_stats").ToMacaron(m, Stats)
}
