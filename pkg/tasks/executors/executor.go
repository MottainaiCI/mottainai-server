/*
Copyright (C) 2017-2020  Ettore Di Giacinto <mudler@gentoo.org>
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

package agenttasks

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/MottainaiCI/mottainai-server/pkg/client"
	"github.com/MottainaiCI/mottainai-server/pkg/secret"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	tasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"
	"github.com/MottainaiCI/mottainai-server/routes/schema"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"

	"golang.org/x/crypto/ssh"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	ssh2 "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

const ABORT_EXECUTION_ERROR = "Aborting execution"
const ABORT_DUPLICATE_ERROR = "Task picked up twice"

type TaskExecutor struct {
	MottainaiClient client.HttpClient
	Context         *ExecutorContext
	Config          *setting.Config
}

type StateRequest struct {
	ContainerID   string
	CacheImage    string
	ImagesToClean []string
	Prune         bool
}

// StateExecution it's used for trace execution of
// a task. Possible states:
// [prepare] --> [running] --> [done]
//                  |--------> [error]
type StateExecution struct {
	Request *StateRequest
	Status  string
	Error   error
	Result  int
}

func (e *StateExecution) UpdateState(status string, err error, res int) {
	e.Status = status
	if err != nil {
		e.Error = err
	}
	e.Result = res
}

func (d *TaskExecutor) HandleTaskStop(timedout bool) (int, error) {
	if timedout {
		d.Report("!! Task timeout!")
	}
	d.Report(ABORT_EXECUTION_ERROR)
	d.MottainaiClient.AbortTask()
	if timedout { // Only timeouts are considered real errors
		return 1, errors.New(ABORT_EXECUTION_ERROR)
	}
	return 0, errors.New(ABORT_EXECUTION_ERROR)
}

func (d *TaskExecutor) DownloadArtefacts(artefactdir, storagedir string) error {
	var err error
	fetcher := d.MottainaiClient
	task_info, err := tasks.FetchTask(fetcher)
	if err != nil {
		d.Report("Couldn't get task info ", err.Error())
		return err
	}

	if len(task_info.RootTask) > 0 {
		for _, f := range strings.Split(task_info.RootTask, ",") {
			if len(f) > 0 {
				if task_info.IsNamespaceMerged() {
					err = fetcher.DownloadArtefactsFromTask(f, artefactdir, task_info.NamespaceFilters)
				} else {
					err = fetcher.DownloadArtefactsFromTask(f,
						path.Join(artefactdir, f),
						task_info.NamespaceFilters,
					)
				}
				if err != nil {
					d.Report("Error on download artefacts from task " + f)
					return err
				}
			}
		}
	}

	if len(task_info.Namespace) > 0 {
		for _, f := range strings.Split(task_info.Namespace, ",") {
			if len(f) > 0 {
				if task_info.IsNamespaceMerged() {
					err = fetcher.DownloadArtefactsFromNamespace(f, artefactdir, task_info.NamespaceFilters)
				} else {
					err = fetcher.DownloadArtefactsFromNamespace(f,
						path.Join(artefactdir, f),
						task_info.NamespaceFilters,
					)
				}
				if err != nil {
					d.Report("Error on download namespace " + f)
					return err
				}
			}
		}
	}

	if len(task_info.Storage) > 0 {
		for _, f := range strings.Split(task_info.Storage, ",") {
			if len(f) > 0 {
				err = fetcher.DownloadArtefactsFromStorage(f, storagedir)
				if err != nil {
					d.Report("Error on download data from storage " + f)
					return err
				}
			}
		}
	}

	return nil
}

func (d *TaskExecutor) UploadArtefacts(folder string) error {
	var filterRegexp []*regexp.Regexp = make([]*regexp.Regexp, 0)
	fetcher := d.MottainaiClient
	task_info, err := tasks.FetchTask(fetcher)
	if err != nil {
		d.Report("Couldn't get task info ", err.Error())
		return err
	}

	for _, filter := range task_info.ArtefactPushFilters {
		r, e := regexp.Compile(filter)
		if e != nil {
			d.Report("Failed compiling regex (" + filter + "):" + e.Error())
			return err
		}
		filterRegexp = append(filterRegexp, r)
	}

	err = filepath.Walk(folder, func(path string, f os.FileInfo, err error) error {

		skipped := false
		if len(task_info.ArtefactPushFilters) > 0 {
			skipped = true
			// Check if artefacts match with filter
			for _, filter := range filterRegexp {
				if filter.MatchString(path) {
					skipped = false
					break
				}
			}
		}

		if skipped {
			d.Report("[Upload] File " + path + " filtered.")
			return nil
		}

		e := d.MottainaiClient.UploadFile(path, folder)
		if e != nil {
			d.Report(fmt.Sprintf("Error on upload file %s: ", path) + e.Error())
		}

		return e
	})

	if err != nil {
		d.MottainaiClient.FailTask(err.Error())
	}
	return err
}

func (d *TaskExecutor) Clean() error {
	if len(d.Context.SourceDir) > 0 {
		if err := os.RemoveAll(d.Context.SourceDir); err != nil {
			d.Report("Warn: (cleanup failed on " + d.Context.SourceDir + ") : " + err.Error())
		}
	}
	if len(d.Context.ArtefactDir) > 0 {
		if err := os.RemoveAll(d.Context.ArtefactDir); err != nil {
			d.Report("Warn: (cleanup failed on " + d.Context.ArtefactDir + ") : " + err.Error())
		}
	}
	if len(d.Context.StorageDir) > 0 {
		if err := os.RemoveAll(d.Context.StorageDir); err != nil {
			d.Report("Warn: (cleanup failed on " + d.Context.StorageDir + ") : " + err.Error())
		}
	}
	if len(d.Context.BuildDir) > 0 {
		if err := os.RemoveAll(d.Context.BuildDir); err != nil {
			d.Report("Warn: (cleanup failed on " + d.Context.BuildDir + ") : " + err.Error())
		}
	}
	if len(d.Context.RootTaskDir) > 0 {
		if err := os.RemoveAll(d.Context.RootTaskDir); err != nil {
			d.Report("Warn: (cleanup failed on " + d.Context.RootTaskDir + ") : " + err.Error())
		}
	}
	return nil
}

func (d *TaskExecutor) Fail(errstring string) {

	task_info, _ := tasks.FetchTask(d.MottainaiClient)
	if task_info.Status != setting.TASK_STATE_ASK_STOP {
		d.MottainaiClient.FinishTask()
	} else {
		d.MottainaiClient.AbortTask()
	}
	d.MottainaiClient.ErrorTask()
}

func (d *TaskExecutor) Success(status int) {
	task_info, _ := tasks.FetchTask(d.MottainaiClient)
	if task_info.Status == setting.TASK_STATE_ASK_STOP {
		d.MottainaiClient.AbortTask()
		return
	}

	if status != 0 {
		d.MottainaiClient.FailTask("Exited with " + strconv.Itoa(status))
	} else {
		d.MottainaiClient.SuccessTask()
	}

}

func (d *TaskExecutor) ExitStatus(i int) {
	d.MottainaiClient.SetTaskField("exit_status", strconv.Itoa(i))
}

func (d *TaskExecutor) Report(v ...interface{}) {
	for _, val := range v {
		if out, ok := val.(string); ok {
			d.MottainaiClient.AppendTaskOutput(out)
		}
	}
	if d.Context.StandardOutput {
		log.Println(v...)
	}
}

func (d *TaskExecutor) Setup(docID string) error {
	var dir, artdir, storagetmp string

	// Handle Pre-execution task commands
	for _, k := range d.Config.GetAgent().PreTaskHookExec {
		d.Report("> Pre-Executing: " + k)
		args := strings.Split(k, " ")
		cmdName := args[0]
		out, stderr, err := utils.Cmd(cmdName, args[1:])
		if err != nil {
			d.Report("!! Error: " + err.Error() + ": " + stderr)
		}
		d.Report(out)
	}

	d.Context.DocID = docID
	fetcher := d.MottainaiClient
	d.MottainaiClient.SetUploadChunkSize(d.Config.GetAgent().UploadChunkSize)
	fetcher.Doc(docID)
	ID := utils.GenID()
	hostname := utils.Hostname()

	task_info, err := tasks.FetchTask(d.MottainaiClient)
	if err != nil {
		return errors.Wrap(err, "Failed to fetch task")
	}

	if task_info.Working() {
		d.Report(">>> WARNING! <<<", ABORT_DUPLICATE_ERROR, ">> NODE <<", ID+" ( "+hostname+" ) ")
		return errors.New(ABORT_DUPLICATE_ERROR)
	}

	fetcher.SetupTask()
	d.Report("Node: " + ID + " ( " + hostname + " ) ")
	fetcher.SetTaskField("nodeid", ID)

	fetcher.RunTask()
	fetcher.SetTaskField("start_time", time.Now().Format("20060102150405"))
	d.Report("> Build started!\n")

	d.Context.RootTaskDir = path.Join(d.Config.GetAgent().BuildPath, task_info.ID)
	d.Context.TaskRelativeDir = task_info.Directory

	tmp_buildpath := path.Join(d.Context.RootTaskDir, "temp")
	dir = path.Join(tmp_buildpath, "root")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	artdir = path.Join(tmp_buildpath, "artefact")
	if err := os.MkdirAll(artdir, os.ModePerm); err != nil {
		return err
	}

	storagetmp = path.Join(tmp_buildpath, "storage")
	if err := os.MkdirAll(storagetmp, os.ModePerm); err != nil {
		return err
	}

	d.Context.BuildDir = dir
	d.Context.ArtefactDir = artdir
	d.Context.StorageDir = storagetmp

	// Fetch git repo (for now only one supported) and checkout commit
	if len(task_info.Source) > 0 {
		d.Report("> Cloning git repo: " + task_info.Source + " in : " + tmp_buildpath)

		d.Context.SourceDir = path.Join(tmp_buildpath, "target_repo")
		if err := os.Mkdir(d.Context.SourceDir, os.ModePerm); err != nil {
			return err
		}
		opts := &git.CloneOptions{
			URL:      task_info.Source,
			Progress: d,
		}

		if task_info.PrivKey != "" {

			auth := task_info.PrivKey
			var s secret.Secret
			req := schema.Request{
				Route:   v1.Schema.GetSecretRoute("show"),
				Target:  &s,
				Options: map[string]interface{}{"id": task_info.PrivKey},
			}
			err := fetcher.Handle(req)

			if err == nil && s.Secret != "" {
				d.Report("Found secret by id.")
				auth = s.Secret
			} else {
				req := schema.Request{
					Route:   v1.Schema.GetSecretRoute("show_by_name"),
					Target:  &s,
					Options: map[string]interface{}{"name": task_info.PrivKey},
				}
				err := fetcher.Handle(req)
				if err == nil && s.Secret != "" {
					d.Report("Found secret by name.")
					auth = s.Secret
				}
			}

			if strings.HasPrefix(auth, "auth:") {
				d.Report("Found secret with auth prefix.")
				a := strings.TrimPrefix(auth, "auth:")
				data := strings.Split(a, ":")
				if len(data) != 2 {
					return errors.New("Invalid credentials")
				}
				opts.Auth = &http.BasicAuth{Username: data[0], Password: data[1]}

			} else {
				d.Report("Found private key for repository.")
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
				opts.Auth = sshAuth
			}
		}
		// TODO: This should go in a go routine and wait for ending
		r, err := git.PlainClone(d.Context.SourceDir, false, opts)
		if err != nil {
			return err
		}

		if len(task_info.Commit) > 0 {

			w, err := r.Worktree()
			if err != nil {
				return err
			}

			err = w.Checkout(&git.CheckoutOptions{
				Hash: plumbing.NewHash(task_info.Commit),
			})
			if err != nil {
				return err
			}

		}
	}

	//cwd, _ := os.Getwd()
	os.Chdir(d.Context.SourceDir)

	/*
		Setup completed create these paths:

		* d.Config.RootTaskDir = <AGENT_BUILD_PATH> + '/' + <TASK_ID>
		* tmp_buildpath = d.Config.RootTaskDir + '/temp'
		* dir = tmp_buildpath + '/root'
		* d.Config.ArtefactDir = tmp_buildpath + '/artefact'
		* d.Config.StorageDir = tmp_buildpath + '/storage'
		* d.Config.BuildDir = dir
		* d.Config.SourceDir = tmp_buildpath + '/target_repo' (only if Source is defined)

		Examples:

		RootTaskDir = /mottainai/build/12345678
		tmp_buildpath = /mottainai/build/12345678/temp
		dir = /mottainai/build/12345678/temp/root
		ArtefactDir = /mottainai/build/12345678/temp/artefact
		StorageDir = /mottainai/build/12345678/temp/storage
		BuildDir = /mottainai/build/12345678/temp/root
		SourceDir = /mottainai/build/12345678/temp/target_repo

	*/

	return nil
}

// Implement Write method as io.Writer
func (t *TaskExecutor) Write(p []byte) (int, error) {
	t.Report(string(p[0 : len(p)-1]))
	return len(p), nil
}

func (t *TaskExecutor) Close() error {
	t.Report(">>>>> Execution completed")
	return nil
}

func (t *TaskExecutor) CreateSharedImageName(task *tasks.Task) (string, error) {
	var OriginalSharedName string
	image := task.Image

	u, err := url.Parse(task.Source)
	if err != nil {
		OriginalSharedName = image + task.Directory
	} else {
		OriginalSharedName = image + u.Path + task.Directory
	}

	return utils.StrictStrip(OriginalSharedName)
}
