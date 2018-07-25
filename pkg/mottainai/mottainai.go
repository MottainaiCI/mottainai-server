/*

Copyright (C) 2018  Ettore Di Giacinto <mudler@gentoo.org>
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
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"

	log "log"

	template "github.com/MottainaiCI/mottainai-server/pkg/template"

	context "github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	static "github.com/MottainaiCI/mottainai-server/pkg/static"

	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	"github.com/go-macaron/cache"
	"github.com/go-macaron/csrf"
	"github.com/go-macaron/session"

	machinery "github.com/RichardKnop/machinery/v1"
	config "github.com/RichardKnop/machinery/v1/config"
	cron "github.com/robfig/cron"

	"github.com/go-macaron/captcha"
	rabbithole "github.com/michaelklishin/rabbit-hole"
	macaron "gopkg.in/macaron.v1"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
)

type Mottainai struct {
	*macaron.Macaron
}

func New() *Mottainai {
	return &Mottainai{Macaron: macaron.New()}
}

func Classic() *Mottainai {
	cl := macaron.New()
	m := &Mottainai{Macaron: cl}

	m.Use(macaron.Logger())
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

	m.Use(session.Sessioner(session.Options{
		// Name of provider. Default is "memory".
		Provider: "memory",
		// Provider configuration, it's corresponding to provider.
		ProviderConfig: "",
		// Cookie name to save session ID. Default is "MacaronSession".
		CookieName: "MottainaiSession",
		// Cookie path to store. Default is "/".
		CookiePath: "/",
		// GC interval time in seconds. Default is 3600.
		Gclifetime: 3600,
		// Max life time in seconds. Default is whatever GC interval time is.
		Maxlifetime: 3600,
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
	}))
	m.Use(csrf.Csrfer(csrf.Options{ // HTTP header used to set and get token. Default is "X-CSRFToken".
		Header: "X-CSRFToken",
		// Form value used to set and get token. Default is "_csrf".
		Form: "_csrf",
		// Cookie value used to set and get token. Default is "_csrf".
		Cookie: "_csrf",
		// Cookie path. Default is "/".
		CookiePath: "/",
		// Key used for getting the unique ID per user. Default is "uid".
		SessionKey: "uid",
		// If true, send token via header. Default is false.
		SetHeader: false,
		// If true, send token via cookie. Default is false.
		SetCookie: false,
		// Set the Secure flag to true on the cookie. Default is false.
		Secure: false,
		// Disallow Origin appear in request header. Default is false.
		Origin: false,
		// The function called when Validate fails. Default is a simple error print.
		ErrorFunc: func(w http.ResponseWriter) {
			http.Error(w, "Invalid csrf token.", http.StatusBadRequest)
		},
	}))

	// XXX: Workaround
	// Set TMPDIR to /var/tmp by default
	// to prevent large files to be stored in ram instead of disk
	if os.Getenv("TMPDIR") == "" {
		os.Setenv("TMPDIR", "/var/tmp")
	}
	template.Setup(m.Macaron)

	m.Invoke(func(l *log.Logger) {
		l.SetPrefix("[ Mottainai ] ")
	})
	m.Use(captcha.Captchaer(captcha.Options{
		SubURL: setting.Configuration.AppSubURL,
	}))

	m.Use(context.Contexter())
	m.SetStatic()

	return m
}

func (m *Mottainai) SetStatic() {
	m.Use(static.AuthStatic(context.CheckArtefactPermission,
		path.Join(setting.Configuration.ArtefactPath),
		macaron.StaticOptions{
			Prefix: "artefact",
		},
	))

	m.Use(static.AuthStatic(context.CheckNamespacePermission,
		path.Join(setting.Configuration.NamespacePath),
		macaron.StaticOptions{
			Prefix: "namespace",
		},
	))
	m.Use(static.AuthStatic(context.CheckStoragePermission,
		path.Join(setting.Configuration.StoragePath),
		macaron.StaticOptions{
			Prefix: "storage",
		},
	))

	m.Use(static.Static(
		path.Join(setting.Configuration.StaticRootPath, "public"),
		macaron.StaticOptions{},
	))
}

func (m *Mottainai) Start() error {

	m.SetAutoHead(true)

	server := NewServer()

	server.Add(setting.Configuration.BrokerDefaultQueue)
	if setting.Configuration.BrokerType == "amqp" {
		rmqc, r_error := rabbithole.NewClient(setting.Configuration.BrokerURI, setting.Configuration.BrokerUser, setting.Configuration.BrokerPass)
		if r_error != nil {
			panic(r_error)
		}
		m.Map(rmqc)
	}

	th := agenttasks.DefaultTaskHandler()
	fmt.Println("DB  with " + setting.Configuration.DBPath)

	database.NewDatabase("tiedot")

	c := cron.New()

	m.Map(database.DBInstance)
	m.Map(server)
	m.Map(th)
	m.Map(c)
	m.Map(m)
	c.Start()

	m.LoadPlans()

	var listenAddr = fmt.Sprintf("%s:%s", setting.Configuration.HTTPAddr, setting.Configuration.HTTPPort)
	log.Printf("Listen: %v://%s", setting.Configuration.Protocol, listenAddr)

	//m.Run()
	var err error
	if len(setting.Configuration.TLSCert) > 0 && len(setting.Configuration.TLSKey) > 0 {
		err = http.ListenAndServeTLS(listenAddr, setting.Configuration.TLSCert, setting.Configuration.TLSKey, m)
	} else {
		err = http.ListenAndServe(listenAddr, m)
	}

	if err != nil {
		log.Fatal(4, "Fail to start server: %v", err)
	}
	c.Stop()
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

func (m *Mottainai) SendTask(docID int) (bool, error) {
	result := true
	var err error
	m.Invoke(func(d *database.Database, server *MottainaiServer, th *agenttasks.TaskHandler) {

		task, err := d.GetTask(docID)
		if err != nil {
			result = false
			return
		}
		task.ClearBuildLog()
		var broker *Broker
		if len(task.Queue) > 0 {
			broker = server.Get(task.Queue)
			log.Println("Sending task to queue ", task.Queue)

		} else {
			broker = server.Get(setting.Configuration.BrokerDefaultQueue)
			log.Println("Sending task to queue ", setting.Configuration.BrokerDefaultQueue)

		}

		d.UpdateTask(docID, map[string]interface{}{"status": "waiting", "result": "none"})

		fmt.Printf("Task Source: %v, Script: %v, Directory: %v, TaskName: %v", task.Source, task.Script, task.Directory, task.TaskName)

		if !th.Exists(task.TaskName) {
			fmt.Printf("Could not send task: Invalid task name")
			result = false
			return
		}

		_, err = broker.SendTask(&BrokerSendOptions{Delayed: task.Delayed, TaskName: task.TaskName, TaskID: docID})
		if err != nil {
			fmt.Printf("Could not send task: %s", err.Error())
			d.UpdateTask(docID, map[string]interface{}{
				"result": "error",
				"status": "done",
				"output": "Backend error, could not send task to broker: " + err.Error(),
			})

			result = false
			return
		}
	})
	return result, err
}

func (m *Mottainai) LoadPlans() {
	m.Invoke(func(c *cron.Cron, d *database.Database) {

		for _, plan := range d.AllPlans() {
			fmt.Println("Loading plan: ", plan.Task, plan)
			id := plan.ID
			c.AddFunc(plan.Planned, func() {
				uid, _ := strconv.Atoi(id)
				plan, _ := d.GetPlan(uid)
				plan.Task.Reset()
				docID, _ := d.CreateTask(plan.Task.ToMap())
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

func NewMachineryServer(queue string) (*machinery.Server, error) {

	var amqpConfig *config.AMQPConfig
	if setting.Configuration.BrokerType == "amqp" {
		amqpConfig = &config.AMQPConfig{
			Exchange:     setting.Configuration.BrokerExchange,
			ExchangeType: setting.Configuration.BrokerExchangeType,
			BindingKey:   queue + "_key",
			//BindingKey:   setting.Configuration.BrokerBindingKey,
		}

	}
	var cnf = &config.Config{
		Broker:          setting.Configuration.Broker,
		DefaultQueue:    queue,
		ResultBackend:   setting.Configuration.BrokerResultBackend,
		ResultsExpireIn: setting.Configuration.ResultsExpireIn,
		AMQP:            amqpConfig,
	}

	server, err := machinery.NewServer(cnf)
	return server, err
}
