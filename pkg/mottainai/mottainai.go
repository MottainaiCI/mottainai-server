/*

Copyright (C) 2018-2021  Ettore Di Giacinto <mudler@gentoo.org>
                         Daniele Rondina <geaaru@sabayonlinux.org>

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

package mottainai

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	context "github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	logging "github.com/MottainaiCI/mottainai-server/pkg/logging"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	static "github.com/MottainaiCI/mottainai-server/pkg/static"
	agenttask "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	taskmanager "github.com/MottainaiCI/mottainai-server/pkg/tasks/manager"
	template "github.com/MottainaiCI/mottainai-server/pkg/template"
	logrus "github.com/sirupsen/logrus"

	"github.com/go-macaron/cache"
	"github.com/go-macaron/csrf"
	"github.com/go-macaron/session"
	_ "github.com/go-macaron/session/redis"
	cron "github.com/robfig/cron"

	"github.com/go-macaron/captcha"
	"github.com/gofrs/uuid"
	macaron "gopkg.in/macaron.v1"
)

type Mottainai struct {
	*macaron.Macaron
}

func New() *Mottainai {
	return &Mottainai{Macaron: macaron.New()}
}

func Classic(config *setting.Config) *Mottainai {
	cl := macaron.New()
	m := &Mottainai{Macaron: cl}
	logger := logging.New()
	logger.SetupWithConfig(true, config)
	logger.WithFields(logrus.Fields{
		"component": "core",
	}).Info("Starting")

	m.Map(logger)
	m.Map(config)
	cl.Map(config)
	database.NewDatabase(config)

	m.Map(database.DBInstance)
	m.Use(logging.MacaronLogger())
	m.Use(macaron.Recovery())

	// TODO: This down deserve config section. Note for _csrf is duplicated in auth
	m.Use(cache.Cacher(cache.Options{ // Name of adapter. Default is "memory".
		Adapter: "memory",
		// Adapter configuration, it's corresponding to adapter.
		AdapterConfig: "",
		// GC interval time in seconds. Default is 60.
		Interval: 60,
		// Configuration section name. Default is "cache".
		Section: "cache",
	}))

	sessionStore := "memory"
	sessionConfig := ""
	switch s := config.GetWeb().SessionProvider; s {
	case "redis":
		sessionStore = s
		sessionConfig = config.GetWeb().SessionProviderConfig
	}
	sesopts := session.Options{
		// Name of provider. Default is "memory".
		Provider: sessionStore,
		// Provider configuration, it's corresponding to provider.
		ProviderConfig: sessionConfig,
		// Cookie name to save session ID. Default is "MacaronSession".
		CookieName: "MottainaiSession",
		// Cookie path to store. Default is "/".
		CookiePath: "/",
		// GC interval time in seconds. Default is 3600.
		Gclifetime: 3600,
		// Max life time in seconds. Default is whatever GC interval time is.
		Maxlifetime: 60 * 60 * 24 * 14, // two weeks
		// Use HTTPS only. Default is false.
		Secure: false,
		// Cookie life time. Default is 0.
		CookieLifeTime: 0,
		// Cookie domain name. Default is empty.
		Domain: "",
		// Session ID length. Default is 16.
		IDLength: 16,
		// Configuration section name. Default is "session".
		Section: "session",
	}

	// send csrf through header so client can save it
	csrfopts := csrf.Options{
		Header:    "X-CSRFToken",
		SetHeader: true,
	}

	if config.GetWeb().GetProtocol() == "https" {
		m.Invoke(func(s session.Store) {
			sesopts.Secure = true
			csrfopts.Secure = true
		})
	}

	m.Use(session.Sessioner(sesopts))
	m.Use(csrf.Csrfer(csrfopts))

	// Honoring first Posix TMPDIR and if
	// TMPDIR is not set I use web.upload_tmpdir
	// to prevent large files to be stored in ram instead of disk
	if os.Getenv("TMPDIR") == "" {
		os.Setenv("TMPDIR", config.GetWeb().UploadTmpDir)
	}
	template.Setup(m.Macaron)

	m.Use(captcha.Captchaer(captcha.Options{
		URLPrefix: config.GetWeb().BuildURI("/captcha/"),
	}))

	m.Use(context.Contexter())
	m.SetStatic()

	if config.GetWeb().EmbedWebHookServer {
		SetupWebHook(m)
	}

	return m
}

func (m *Mottainai) SetStatic() {
	m.Invoke(func(c *setting.Config) {
		m.Use(static.AuthStatic(context.CheckArtefactPermission,
			path.Join(c.GetStorage().ArtefactPath),
			c.GetWeb().AccessControlAllowOrigin, c,
			macaron.StaticOptions{
				Prefix: "artefact",
			},
		))

		m.Use(static.AuthStatic(context.CheckNamespacePermission,
			path.Join(c.GetStorage().NamespacePath),
			c.GetWeb().AccessControlAllowOrigin, c,
			macaron.StaticOptions{
				Prefix: "namespace",
			},
		))
		m.Use(static.AuthStatic(context.CheckStoragePermission,
			path.Join(c.GetStorage().StoragePath),
			c.GetWeb().AccessControlAllowOrigin, c,
			macaron.StaticOptions{
				Prefix: "storage",
			},
		))

		m.Use(static.Static(
			path.Join(c.GetWeb().StaticRootPath, "public"),
			c.GetWeb().AccessControlAllowOrigin, c,
			macaron.StaticOptions{},
		))
	})
}

func (m *Mottainai) listenAddr() string {
	var ans string
	m.Invoke(func(config *setting.Config) {
		ans = fmt.Sprintf("%s:%s", config.GetWeb().HTTPAddr, config.GetWeb().HTTPPort)
	})

	return ans
}

func (m *Mottainai) url() string {
	var ans string

	m.Invoke(func(config *setting.Config) {
		ans = fmt.Sprintf("%s://%s", config.GetWeb().Protocol,
			m.listenAddr())
	})
	return ans
}

func (m *Mottainai) Start() error {

	m.SetAutoHead(true)

	m.Invoke(func(config *setting.Config, l *logging.Logger) {
		th := taskmanager.DefaultTaskHandler(config)
		l.WithFields(logrus.Fields{
			"component": "core",
			"path":      config.GetDatabase().DBPath,
		}).Info("Database Configuration")

		c := cron.New()

		m.Map(th)
		m.Map(c)
		m.Map(m)
		m.Map(m.Macaron)
		c.Start()
		m.LoadPlans()
		// For now
		if config.GetWeb().EmbedWebHookServer {
			SetupWebHookAgent(m)
		}

		l.WithFields(logrus.Fields{
			"component": "core",
			"url":       m.url(),
		}).Info("WebUI listening")
		m.HealthCheckRun(config.GetWeb().HealthCheckInterval) // Start server HealthCheck daemon

		//m.Run()
		var err error
		if len(config.GetGeneral().TLSCert) > 0 && len(config.GetGeneral().TLSKey) > 0 {
			err = http.ListenAndServeTLS(m.listenAddr(),
				config.GetGeneral().TLSCert, config.GetGeneral().TLSKey, m)
		} else {
			err = http.ListenAndServe(m.listenAddr(), m)
		}

		if err != nil {
			l.WithFields(logrus.Fields{
				"component": "web",
				"error":     err,
			}).Fatal("Failed to start server")
		}
		c.Stop()
	})

	return nil
}

func (m *Mottainai) WrapF(f http.HandlerFunc) macaron.Handler {
	return func(c *context.Context) {
		f(c.Resp, c.Req.Request)
	}
}

func (m *Mottainai) WrapH(h http.Handler) macaron.Handler {
	return func(c *context.Context) {
		h.ServeHTTP(c.Resp, c.Req.Request)
	}
}
func (m *Mottainai) ProcessPipeline(docID string) (bool, error) {
	result := true
	var rerr error
	m.Invoke(func(d *database.Database, th *taskmanager.TaskHandler,
		config *setting.Config, l *logging.Logger) {
		pip, err := d.Driver.GetPipeline(config, docID)
		if err != nil {
			rerr = err
			result = false
			return
		}

		errOnSend := false
		if len(pip.Chain) > 0 {
			// POST: Chain pipeline
			//       In this this i send only the first task
			fmt.Println("Processing chain pipeline " + pip.ID)

			_, err := m.SendTask(pip.Tasks[pip.Chain[0]].ID)
			if err != nil {
				m.FailTask(pip.Tasks[pip.Chain[0]].ID, err.Error())
				errOnSend = true
				rerr = err
			}

			if errOnSend {
				for _, t := range pip.Chain[1:] {
					m.FailTask(pip.Tasks[t].ID,
						"Task in error a cause to errors with other tasks of the pipeline: "+err.Error())
				}
			}

		} else {
			// POST: Group pipeline or chord pipeline
			//       In this case i can send all tasks to the queue.
			for _, t := range pip.Group {
				if !errOnSend {
					_, err := m.SendTask(pip.Tasks[t].ID)
					if err != nil {
						m.FailTask(pip.Tasks[t].ID, err.Error())
						errOnSend = true
						rerr = err
					}
				} else {
					m.FailTask(pip.Tasks[t].ID,
						"Task in error a cause to errors with other tasks of the pipeline: "+err.Error())
				}
			}

			if errOnSend && len(pip.Chord) > 0 {
				for _, t := range pip.Chord {
					m.FailTask(pip.Tasks[t].ID,
						"Task in error a cause to errors with other tasks of the pipeline: "+err.Error())
				}
			}

		}

		if errOnSend {
			result = false
		}
	})

	return result, rerr
}

func (m *Mottainai) FailTask(task, reason string) {
	m.Invoke(func(d *database.Database, l *logging.Logger) {
		l.WithFields(logrus.Fields{
			"component": "core",
			"task_id":   task,
			"error":     reason,
		}).Error(reason)
		d.Driver.UpdateTask(task, map[string]interface{}{
			"result": "error",
			"status": "done",
			"output": reason,
		})
	})
}

func (m *Mottainai) processableTask(docID string) error {
	// TODO: Add proper locks to db schema, this is racy
	var resultError error
	m.Invoke(func(d *database.Database) {

		// Check setting if we have to process this.
		if protectOverwrite, _ := d.Driver.GetSettingByKey(setting.SYSTEM_PROTECT_NAMESPACE_OVERWRITE); protectOverwrite.IsDisabled() {
			return
		}

		_, err := d.Driver.GetTask(d.Config, docID)
		if err != nil {
			resultError = errors.New("Failed to get task information while checking if it is processable or not")
			return
		}
	})
	return resultError
}

func (m *Mottainai) CreateTask(task *agenttask.Task) error {
	var ans error
	m.Invoke(func(db *database.Database) {
		if task.Queue == "" {
			// Retrieve default queue.
			defaultQueue, _ := db.Driver.GetSettingByKey(
				setting.SYSTEM_TASKS_DEFAULT_QUEUE,
			)

			if defaultQueue.Value == "" {
				task.Queue = "general"
			} else {
				task.Queue = defaultQueue.Value
			}
		}

		docID, err := db.Driver.InsertTask(task)
		if err != nil {
			ans = err
			return
		}

		task.ID = docID
	})

	return ans
}

func (m *Mottainai) PrepareTaskQueue(task agenttask.Task) (string, error) {
	var ansErr error
	var ansQid string

	m.Invoke(func(d *database.Database) {
		// Check if exists the queue
		q, err := d.Driver.GetQueueByKey(task.Queue)
		if err != nil {
			ansErr = errors.New("Failed on retrieve queue data")
			return

		} else if q.Qid == "" {
			// POST: create the queue
			ct := time.Now().UTC().Format("20060102150405")
			qid, _ := uuid.NewV4()

			_, err = d.Driver.CreateQueue(map[string]interface{}{
				"qid":              qid.String(),
				"name":             task.Queue,
				"tasks_waiting":    []string{},
				"tasks_inprogress": []string{},
				"creation_date":    ct,
				"update_date":      ct,
			})

			if err != nil {
				ansErr = errors.New("Error on create the task queue: " + err.Error())
				return
			}
			ansQid = qid.String()
		} else {
			ansQid = q.Qid
		}
	})

	return ansQid, ansErr
}

func (m *Mottainai) SendTask(docID string) (bool, error) {
	result := false
	var err error
	m.Invoke(func(d *database.Database, l *logging.Logger, th *taskmanager.TaskHandler, config *setting.Config) {

		if err := m.processableTask(docID); err != nil {
			m.FailTask(docID, err.Error())
			return
		}

		task, err := d.Driver.GetTask(config, docID)
		if err != nil {
			err = errors.New("Failed to get task information while checking if it is processable or not")
			return
		}

		if !th.Exists(task.Type) {
			err = errors.New("Could not send task: Invalid task type")
			m.FailTask(docID, err.Error())
			return
		}

		qid, err := m.PrepareTaskQueue(task)
		if err != nil {
			return
		}

		// POST: add task to the queue
		err = d.Driver.AddTaskInWaiting2Queue(qid, docID)
		if err != nil {
			err = errors.New("Error on add task in queue: " + err.Error())
			return
		}
		fmt.Println(fmt.Sprintf("Added waiting task %s in queue %s.", docID, qid))

		task.ClearBuildLog(config.GetStorage().ArtefactPath)

		l.WithFields(logrus.Fields{
			"component": "core",
			"task_id":   docID,
			"queue":     task.Queue,
		}).Info("Sending task")

		d.Driver.UpdateTask(docID, map[string]interface{}{
			"status": "waiting",
			"result": "none",
		})

		l.WithFields(logrus.Fields{
			"component": "core",
			"task_id":   docID,
			"type":      task.Type,
		}).Debug("Task")

		//

	})

	return result, err
}

func (m *Mottainai) LoadPlans() {
	m.Invoke(func(c *cron.Cron, d *database.Database, l *logging.Logger, config *setting.Config) {

		for _, plan := range d.Driver.AllPlans(config) {
			l.WithFields(logrus.Fields{
				"component": "core",
				"plan_id":   plan.ID,
			}).Debug("Loading plan")
			id := plan.ID
			c.AddFunc(plan.Planned, func() {
				plan, _ := d.Driver.GetPlan(config, id)
				plan.Task.Reset()
				docID, _ := d.Driver.CreateTask(plan.Task.ToMap())
				m.SendTask(docID)
			})
		}

	})
}

func (m *Mottainai) ReloadCron() {

	m.Invoke(func(c *cron.Cron, d *database.Database) {
		c.Stop()
		c = cron.New()
		m.Map(c)
		m.LoadPlans()
		c.Start()
	})

}
