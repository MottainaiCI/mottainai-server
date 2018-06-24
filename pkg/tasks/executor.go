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
	"strings"
	"time"

	"github.com/MottainaiCI/mottainai-server/pkg/client"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"
)

//TODO:
type ExecutorContext struct {
	ArtefactDir, StorageDir, NamespaceDir, BuildDir, SourceDir string
}

type TaskExecutor struct {
	MottainaiClient *client.Fetcher
}

func (d *TaskExecutor) DownloadArtefacts(artefactdir, storagedir string) error {
	fetcher := d.MottainaiClient
	task_info := DefaultTaskHandler().FetchTask(fetcher)
	if len(task_info.RootTask) > 0 {
		for _, f := range strings.Split(task_info.RootTask, ",") {
			fetcher.DownloadArtefactsFromTask(f, artefactdir)
		}
	}

	if len(task_info.Namespace) > 0 {
		for _, f := range strings.Split(task_info.Namespace, ",") {
			fetcher.DownloadArtefactsFromNamespace(f, artefactdir)
		}
	}

	if len(task_info.Storage) > 0 {
		for _, f := range strings.Split(task_info.Storage, ",") {
			fetcher.DownloadArtefactsFromStorage(f, storagedir)
		}
	}
	return nil
}

func (d *TaskExecutor) Setup(docID string) error {
	fetcher := client.NewFetcher(docID)
	fetcher.SetTaskStatus("setup")
	ID := utils.GenID()
	hostname := utils.Hostname()
	fetcher.AppendTaskOutput("Node: " + ID + " ( " + hostname + " ) ")
	fetcher.SetTaskField("nodeid", ID)

	d.MottainaiClient = fetcher
	th := DefaultTaskHandler()
	task_info := th.FetchTask(fetcher)
	if task_info.Status == "running" {
		fetcher.SetTaskStatus("failure")
		msg := "Task picked twice"
		fetcher.AppendTaskOutput(msg)
		return errors.New(msg)
	}

	fetcher.SetTaskStatus("running")
	fetcher.SetTaskField("start_time", time.Now().Format("20060102150405"))
	fetcher.AppendTaskOutput("Build started!\n")
	return nil
}
