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

package agenttasks

import (
	"errors"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"

	"github.com/MottainaiCI/mottainai-server/pkg/client"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"
)

//TODO:
type ExecutorContext struct {
	ArtefactDir, StorageDir, NamespaceDir         string
	BuildDir, SourceDir, RootTaskDir, RealRootDir string
	DocID                                         string
	StandardOutput                                bool
}

const ABORT_EXECUTION_ERROR = "Aborting execution"
const ABORT_DUPLICATE_ERROR = "Task picked up twice"

func NewExecutorContext() *ExecutorContext {
	return &ExecutorContext{StandardOutput: true}
}

type TaskExecutor struct {
	MottainaiClient client.HttpClient
	Context         *ExecutorContext
	Config          *setting.Config
}

func (d *TaskExecutor) DownloadArtefacts(artefactdir, storagedir string) error {
	fetcher := d.MottainaiClient
	task_info := DefaultTaskHandler(d.Config).FetchTask(fetcher)
	var err error
	if len(task_info.RootTask) > 0 {
		for _, f := range strings.Split(task_info.RootTask, ",") {
			err = fetcher.DownloadArtefactsFromTask(f, artefactdir)
		}
	}

	if len(task_info.Namespace) > 0 {
		for _, f := range strings.Split(task_info.Namespace, ",") {
			err = fetcher.DownloadArtefactsFromNamespace(f, artefactdir)
		}
	}

	if len(task_info.Storage) > 0 {
		for _, f := range strings.Split(task_info.Storage, ",") {
			err = fetcher.DownloadArtefactsFromStorage(f, storagedir)
		}
	}
	return err
}

func (d *TaskExecutor) UploadArtefacts(folder string) error {
	err := filepath.Walk(folder, func(path string, f os.FileInfo, err error) error {
		return d.MottainaiClient.UploadFile(path, folder)
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

	task_info := DefaultTaskHandler(d.Config).FetchTask(d.MottainaiClient)
	if task_info.Status != setting.TASK_STATE_ASK_STOP {
		d.MottainaiClient.FinishTask()
	} else {
		d.MottainaiClient.AbortTask()
	}
	d.MottainaiClient.ErrorTask()

}

func (d *TaskExecutor) Success(status int) {
	task_info := DefaultTaskHandler(d.Config).FetchTask(d.MottainaiClient)
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
		log.Println(v)
	}
}

func (d *TaskExecutor) Setup(docID string) error {
	d.Context.DocID = docID
	fetcher := d.MottainaiClient
	fetcher.Doc(docID)
	ID := utils.GenID()
	hostname := utils.Hostname()

	th := DefaultTaskHandler(d.Config)
	task_info := th.FetchTask(fetcher)
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
	tmp_buildpath := path.Join(d.Context.RootTaskDir, "temp")
	dir := path.Join(tmp_buildpath, "root")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	artdir := path.Join(tmp_buildpath, "artefact")
	if err := os.MkdirAll(artdir, os.ModePerm); err != nil {
		return err
	}

	storagetmp := path.Join(tmp_buildpath, "storage")
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
		// TODO: This should go in a go routine and wait for ending
		r, err := git.PlainClone(d.Context.SourceDir, false, &git.CloneOptions{
			URL:      task_info.Source,
			Progress: os.Stdout,
		})
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

	return nil
}
