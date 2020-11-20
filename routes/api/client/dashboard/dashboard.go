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
