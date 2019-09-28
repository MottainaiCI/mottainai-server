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

package webhook

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"regexp"
	"strings"

	anagent "github.com/mudler/anagent"

	logrus "github.com/sirupsen/logrus"

	logging "github.com/MottainaiCI/mottainai-server/pkg/logging"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	utils "github.com/MottainaiCI/mottainai-server/pkg/utils"
	mhook "github.com/MottainaiCI/mottainai-server/pkg/webhook"

	"github.com/MottainaiCI/mottainai-server/pkg/context"
	database "github.com/MottainaiCI/mottainai-server/pkg/db"
	mottainai "github.com/MottainaiCI/mottainai-server/pkg/mottainai"
	tasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	user "github.com/MottainaiCI/mottainai-server/pkg/user"

	"golang.org/x/crypto/ssh"
	webhooks "gopkg.in/go-playground/webhooks.v3"
	git "gopkg.in/src-d/go-git.v4"
	gith "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	ssh2 "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

var (
	pending       = "pending"
	success       = "success"
	failure       = "error"
	pendingDesc   = "Build in progress, please wait."
	noPermDesc    = "Insufficient permissions"
	successDesc   = "Build successful."
	failureDesc   = "Build failed."
	notfoundDesc  = "No mottainai file found on repo"
	task_file     = ".mottainai"
	pipeline_file = ".mottainai-pipeline"
)

type GitContext struct {
	Dir      string
	Uid      string
	Commit   string
	Owner    string
	UserRepo string
	Checkout string
	Repo     string
	Ref      string
	User     string

	CloneSSHUrl  string
	CloneHTTPUrl string
	FilterRef    string
	KindEvent    string

	Envs []string

	StoredUser *user.User
}

type WebHookCallbacks interface {
	SetFailureStatus(string)
	SetStatus(*string, *string, *string)
	SetPendingStatus()
	LoadEventEnvs2Task(*tasks.Task)
	GetLogFields(string) logrus.Fields
}

type GitWebHook struct {
	Context   *GitContext
	Payload   interface{}
	User      *user.User
	Header    webhooks.Header
	Hook      *mhook.WebHook
	AppName   string
	BuildPath string
	CBHandler WebHookCallbacks // workaround for call children struct from father.
}

func newGitWebHook(payload interface{}, w *mhook.WebHook, u *user.User, header webhooks.Header) *GitWebHook {
	return &GitWebHook{
		Payload: payload,
		User:    u,
		Header:  header,
		Hook:    w,
	}
}

func NewGitContext(hookType, kindEvent string, payload interface{}) *GitContext {
	if hookType == "github" {
		return NewGitContextGitHub(kindEvent, payload)
	} else {
		return NewGitContextGitLab(kindEvent, payload)
	}
}

func (ctx *GitContext) IsEventFiltered(webHookFilter string) (bool, error) {
	// Filter by ref: If hook contains a filter defined, it has to match the ref of the push, or we discard it
	if len(webHookFilter) > 0 {
		includeRegex, err := regexp.Compile(webHookFilter)
		if err != nil {
			return true, errors.New("Webhook filter invalid")
		}
		if !includeRegex.Match([]byte(ctx.FilterRef)) {
			return true, nil
		}
	}

	return false, nil
}

func (h *GitWebHook) HandleEvent(m *mottainai.Mottainai, l *logging.Logger, db *database.Database) (string, string) {
	var err error
	var idTask, idPipeline string

	err = h.CheckIfHookIsAdmit(db)
	if err != nil {
		l.WithFields(h.CBHandler.GetLogFields(err.Error())).Error(
			fmt.Sprintf("Error on processing hook filter regex for user %s.",
				h.User.ID))
		return "", ""
	}

	if idTask, err = h.SendTask(db, m, l); err != nil {
		l.WithFields(h.CBHandler.GetLogFields(err.Error())).Error("Failed sending task")
	}

	if idPipeline, err = h.SendPipeline(db, m, l); err != nil {
		l.WithFields(h.CBHandler.GetLogFields(err.Error())).Error("Failed sending pipeline")
	}

	return idTask, idPipeline
}

func (h *GitWebHook) CheckIfHookIsAdmit(db *database.Database) error {

	filtered, err := h.Context.IsEventFiltered(h.Hook.Filter)
	if err != nil {
		return err
	} else if filtered {
		return errors.New("Webhook filtered")
	}

	// Check setting if we have to process this.
	uuu, err := db.Driver.GetSettingByKey(setting.SYSTEM_WEBHOOK_ENABLED)
	if err != nil || (err == nil && uuu.IsDisabled()) {
		var strerr string
		if err == nil {
			strerr = "Webhooks disabled"
			err = errors.New(strerr)
		} else {
			strerr = "Unexpected error on handle webhook"
		}
		h.CBHandler.SetFailureStatus(strerr)
		return err
	}

	uuu, err = db.Driver.GetSettingByKey(setting.SYSTEM_WEBHOOK_INTERNAL_ONLY)
	if err == nil && uuu.IsEnabled() &&
		reflect.TypeOf(h.CBHandler) == reflect.TypeOf((*GitHubWebHook)(nil)) {
		u, err := db.Driver.GetUserByIdentity("github", h.Context.User)
		if err != nil {
			h.CBHandler.SetFailureStatus(noPermDesc)
			return err
		}
		h.Context.StoredUser = &u
	} else if err != nil {
		strerr := "Unexpected error on handle webhook"
		h.CBHandler.SetFailureStatus(strerr)
		return err
	} else {
		// TODO: Check in users the enabled repository hooks
		// Later, with organizations and projects will be easier to link them.
		h.Context.StoredUser = h.User
	}

	return nil
}

func (h *GitWebHook) PrepareGitDir(db *database.Database) error {
	err := os.MkdirAll(path.Join(h.BuildPath, "webhook_fetch", h.Context.Repo), os.ModePerm)
	if err != nil {
		err = errors.New("Failed creating webhook_fetch temp dir (Set your buildpath): " + err.Error())
		return err
	}

	h.Context.Dir, err = ioutil.TempDir(h.BuildPath, path.Join("webhook_fetch", h.Context.Repo))
	if err != nil {
		err = errors.New("Failed creating tempdir: " + err.Error())
		return err
	}

	opts := &git.CloneOptions{
		// By default for now I use HTTP Git URL.
		URL: h.Context.CloneHTTPUrl,
	}

	if h.Hook.Auth != "" {
		auth := h.Hook.Auth
		secret, err := db.Driver.GetSecret(h.Hook.Auth)
		if err == nil {
			auth = secret.Secret
		} else {
			secret, err := db.Driver.GetSecretByName(h.Hook.Auth)
			if err == nil {
				auth = secret.Secret
			}
		}

		if strings.HasPrefix(auth, "auth:") {
			a := strings.TrimPrefix(auth, "auth:")
			data := strings.Split(a, ":")
			if len(data) != 2 {
				err = errors.New("Invalid credentials")
				return err
			}
			opts.Auth = &gith.BasicAuth{Username: data[0], Password: data[1]}

		} else {
			signer, err := ssh.ParsePrivateKey([]byte(auth))
			if err != nil {
				return err
			}
			sshAuth := &ssh2.PublicKeys{
				User:   "git",
				Signer: signer,
				// TODO: This could be avoid if we use a directory that
				// contains valid certificates. See if there is a way to
				// accept only valid certificate and/or configure this through
				// agent configuration option.
				HostKeyCallbackHelper: ssh2.HostKeyCallbackHelper{
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				},
			}

			// With credential I use SSH Repo url
			opts.URL = h.Context.CloneSSHUrl
			opts.Auth = sshAuth
		}
	}

	r, err := git.PlainClone(h.Context.Dir, false, opts)
	if err != nil {
		os.RemoveAll(h.Context.Dir)
		err = errors.New("Failed cloning repo: " + opts.URL + " " + h.Context.Dir + " " + err.Error())
		h.CBHandler.SetFailureStatus(err.Error())
		return err
	}

	if h.Context.KindEvent == "pull_request" {
		err = utils.GitCheckoutPullRequest(r, "origin", h.Context.Checkout)
		if err != nil {
			os.RemoveAll(h.Context.Dir)
			err = errors.New("Failed checkout repo: " + err.Error())
			h.CBHandler.SetFailureStatus(err.Error())
			return err
		}
	} else if h.Context.KindEvent == "merge_request" {
		err = utils.GitCheckoutMergeRequest(r, "origin", h.Context.Checkout)
		if err != nil {
			os.RemoveAll(h.Context.Dir)
			err = errors.New("Failed checkout repo: " + err.Error())
			h.CBHandler.SetFailureStatus(err.Error())
			return err
		}
	} else {
		err = utils.GitCheckoutCommit(r, h.Context.Checkout)
		if err != nil {
			os.RemoveAll(h.Context.Dir)
			err = errors.New("Failed checkout repo: " + err.Error())
			h.CBHandler.SetFailureStatus(err.Error())
			return err
		}
	}

	return nil
}

func (h *GitWebHook) GetHookTask(db *database.Database) (*tasks.Task, error) {
	var t *tasks.Task
	var err error

	exists := false
	if h.Hook.HasTask() {
		t, err = h.Hook.ReadTask()
		if err != nil {
			return nil, err
		}
		exists = true
	} else {
		exists, _ = utils.Exists(path.Join(h.Context.Dir, task_file+".json"))
		if exists == true {
			t, err = tasks.FromFile(path.Join(h.Context.Dir, task_file+".json"))
			if err != nil {
				return nil, err
			}
		} else {
			exists, _ = utils.Exists(path.Join(h.Context.Dir, task_file+".yaml"))
			if exists == true {
				t, err = tasks.FromYamlFile(path.Join(h.Context.Dir, task_file+".yaml"))
				if err != nil {
					return nil, err
				}
			}
		}
	}

	if !exists {
		return nil, errors.New("Task not found")
	}

	t.Owner = h.Context.StoredUser.ID
	t.Source = h.Context.UserRepo
	t.Commit = h.Context.Commit
	t.Queue = QueueSetting(db)

	return t, nil
}

func (h *GitWebHook) GetHookPipeline(db *database.Database) (*tasks.Pipeline, error) {
	var t *tasks.Pipeline
	var err error

	exists := false
	if h.Hook.HasPipeline() {
		t, err = h.Hook.ReadPipeline()
		if err != nil {
			return nil, err
		}
		exists = true
	} else {
		exists, _ = utils.Exists(path.Join(h.Context.Dir, pipeline_file+".json"))
		if exists == true {
			t, err = tasks.PipelineFromJsonFile(path.Join(h.Context.Dir, pipeline_file+".json"))
			if err != nil {
				return nil, err
			}
		} else {
			exists, _ = utils.Exists(path.Join(h.Context.Dir, pipeline_file+".yaml"))
			if exists == true {
				t, err = tasks.PipelineFromYamlFile(path.Join(h.Context.Dir, pipeline_file+".yaml"))
				if err != nil {
					return nil, err
				}
			}
		}
	}

	if !exists {
		// Pipeline not available on webhook. I ignore webhook.
		return nil, nil
	}

	t.Owner = h.Context.StoredUser.ID
	// XXX:
	t.Queue = QueueSetting(db)

	return t, nil
}

func (h *GitWebHook) CreateHookTask(m *mottainai.Mottainai, db *database.Database) (string, error) {
	t, err := h.GetHookTask(db)
	if err != nil {
		h.CBHandler.SetFailureStatus(err.Error())
		return "", err
	} else if t == nil {
		// POST: no pipeline available
		return "", nil
	}

	h.CBHandler.LoadEventEnvs2Task(t)

	docID, err := db.Driver.CreateTask(t.ToMap())
	if err != nil {
		h.CBHandler.SetFailureStatus(err.Error())
		return "", err
	}

	m.SendTask(docID)

	var url string
	m.Invoke(func(config *setting.Config) {
		url = config.GetWeb().BuildAbsURL("/tasks/display/" + docID)
	})

	// Create the 'pending' status and send it
	h.CBHandler.SetStatus(&pending, &pendingDesc, &url)

	m.Invoke(func(a *anagent.Anagent) {
		data := strings.Join([]string{
			h.Context.KindEvent,
			h.Context.Owner,
			h.Context.Repo,
			h.Context.Ref,
			"tasks", docID,
		}, ",")
		a.Invoke(func(w map[string]string) {
			a.Lock()
			defer a.Unlock()
			w[h.Context.Uid] = data
		})
	})

	return docID, nil
}

func (h *GitWebHook) CreateHookPipeline(m *mottainai.Mottainai, db *database.Database, l *logging.Logger) (string, error) {
	t, err := h.GetHookPipeline(db)
	if err != nil {
		h.CBHandler.SetFailureStatus(err.Error())
		return "", err
	} else if t == nil {
		return "", nil
	}

	// do not allow automatic tag from PR
	for i, p := range t.Tasks { // Duplicated in API.
		if h.Context.KindEvent == "pull_request" || h.Context.KindEvent == "merge_request" {
			// do not allow automatic tag from PR
			p.TagNamespace = ""
			p.Storage = ""
			p.Binds = []string{}
			p.RootTask = ""
		}
		p.Owner = h.Context.StoredUser.ID
		p.Source = h.Context.UserRepo
		p.Commit = h.Context.Commit
		p.Status = setting.TASK_STATE_WAIT

		h.CBHandler.LoadEventEnvs2Task(&p)

		id, err := db.Driver.CreateTask(p.ToMap())
		if err != nil {
			return "", err
		}
		p.ID = id
		t.Tasks[i] = p
	}

	docID, err := db.Driver.CreatePipeline(t.ToMap(false))
	if err != nil {
		return "", err
	}

	fields := h.CBHandler.GetLogFields("")
	fields["pipeline_id"] = docID
	l.WithFields(fields).Debug("Sending pipeline")

	_, err = m.ProcessPipeline(docID)
	if err != nil {
		fields["error"] = err.Error()
		l.WithFields(fields).Error("While sending")
		return "", err
	}

	var url string
	m.Invoke(func(config *setting.Config) {
		url = config.GetWeb().BuildURI("/pipeline/" + docID)
	})

	// Create the 'pending' status and send it
	h.CBHandler.SetStatus(&pending, &pendingDesc, &url)

	m.Invoke(func(a *anagent.Anagent) {
		data := strings.Join([]string{
			h.Context.KindEvent,
			h.Context.Owner,
			h.Context.Repo,
			h.Context.Ref,
			"pipeline", docID,
		}, ",")
		a.Invoke(func(w map[string]string) {
			l.WithFields(logrus.Fields{
				"component": "webhook_global_watcher",
				"event":     "add",
			}).Debug("Add event to global watcher")
			a.Lock()
			defer a.Unlock()
			w[h.Context.Uid] = data
		})
	})

	return docID, nil
}

func (h *GitWebHook) GetLogFields(err string) logrus.Fields {
	ans := logrus.Fields{
		"component": "webhook",
		"event":     fmt.Sprintf("general_%s", h.Context.KindEvent),
	}
	if h.Hook != nil {
		ans["wid"] = h.Hook.ID
	}
	if err != "" {
		ans["error"] = err
	}
	return ans
}

func (h *GitWebHook) SendTask(db *database.Database, m *mottainai.Mottainai, logger *logging.Logger) (string, error) {
	err := h.PrepareGitDir(db)
	var idTask string

	if err != nil {
		logger.WithFields(logrus.Fields{
			"component": "webhook_global_watcher",
			"error":     err.Error(),
		}).Error("Error while preparing temp directory")

		return "", err
	}
	defer os.RemoveAll(h.Context.Dir)

	// Create the 'pending' status and send it
	h.CBHandler.SetPendingStatus()

	idTask, err = h.CreateHookTask(m, db)
	if err != nil {
		return "", err
	}

	return idTask, nil
}

func (h *GitWebHook) SendPipeline(db *database.Database, m *mottainai.Mottainai, logger *logging.Logger) (string, error) {
	var pipelineId string

	err := h.PrepareGitDir(db)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"component": "webhook_global_watcher",
			"error":     err.Error(),
		}).Error("Error while preparing temp directory")

		return "", err
	}
	defer os.RemoveAll(h.Context.Dir)

	pipelineId, err = h.CreateHookPipeline(m, db, logger)
	if err != nil {
		return "", err
	}

	return pipelineId, nil
}

func RequiresWebHookSetting(c *context.Context, db *database.Database) error {
	// Check setting if we have to process this.
	err := errors.New("Webhook integration disabled")
	uuu, err := db.Driver.GetSettingByKey(setting.SYSTEM_WEBHOOK_ENABLED)
	if err == nil {
		if uuu.IsDisabled() {
			c.ServerError("Webhook integration disabled", err)
			return err
		}
	}
	return nil
}

func QueueSetting(db *database.Database) string {
	uuu, err := db.Driver.GetSettingByKey(setting.SYSTEM_WEBHOOK_DEFAULT_QUEUE)
	if err != nil {
		return "default_webhooks"
	}
	return uuu.Value
}

func Setup(m *mottainai.Mottainai) {
	SetupGitHub(m)
	SetupGitLab(m)
}
