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

package tiedot

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	dbcommon "github.com/MottainaiCI/mottainai-server/pkg/db/common"
	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"

	"github.com/MottainaiCI/mottainai-server/pkg/artefact"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
)

var TaskColl = "Tasks"

func (d *Database) IndexTask() {
	d.AddIndex(TaskColl, []string{"status"})
	d.AddIndex(TaskColl, []string{"queue"})
	d.AddIndex(TaskColl, []string{"result"})
	d.AddIndex(TaskColl, []string{"owner_id"})
	d.AddIndex(TaskColl, []string{"node_id"})
	d.AddIndex(TaskColl, []string{"result", "status"})
}

func (d *Database) InsertTask(t *agenttasks.Task) (string, error) {
	return d.CreateTask(t.ToMap())
}

func (d *Database) CreateTask(t map[string]interface{}) (string, error) {

	return d.InsertDoc(TaskColl, t)
}

func (d *Database) DeleteTask(config *setting.Config, docID string) error {

	t, err := d.GetTask(config, docID)
	if err != nil {
		return err
	}
	artefacts, err := d.GetTaskArtefacts(docID)
	if err != nil {
		return err
	}
	d.Invoke(func(config *setting.Config) {
		for _, artefact := range artefacts {
			artefact.CleanFromTask(config)
			d.DeleteArtefact(artefact.ID)
		}
		t.Clear(config.GetStorage().ArtefactPath, config.GetWeb().LockPath)
	})
	return d.DeleteDoc(TaskColl, docID)
}

func (d *Database) UpdateTask(docID string, t map[string]interface{}) error {
	return d.UpdateDoc(TaskColl, docID, t)
}

func (d *Database) GetTask(config *setting.Config, docID string) (agenttasks.Task, error) {
	doc, err := d.GetDoc(TaskColl, docID)
	if err != nil {
		return agenttasks.Task{}, err
	}
	t := agenttasks.NewTaskFromMap(doc)
	t.ID = docID
	return t, err
}

func (d *Database) GetTaskByStatus(config *setting.Config, status string) ([]agenttasks.Task, error) {
	var res []agenttasks.Task

	var query string
	switch status {
	case "running":
		query = `[{"eq": "running", "in": ["status"]}]`
	case "waiting":
		query = `[{"eq": "waiting", "in": ["status"]}]`
	case "stop":
		query = `[{"eq": "stop", "in": ["status"]}]`
	case "stopped":
		query = `[{"eq": "stopped", "in": ["status"]}]`
	case "error":
		query = `[{"eq": "error", "in": ["result"]}]`
	case "failed":
		query = `[{"eq": "failed", "in": ["result"]}]`
	case "success":
		query = `[{"eq": "success", "in": ["result"]}]`
	default:
		return res, errors.New("No valid status supplied")
	}

	queryResult, e := d.FindDoc(TaskColl, query)
	if e != nil {
		return res, e
	}

	// Query result are document IDs
	for docid := range queryResult {
		// Read document
		t, err := d.GetTask(config, docid)
		if err != nil {
			return []agenttasks.Task{}, err
		}
		res = append(res, t)

	}
	return res, nil
}

func (d *Database) GetTaskArtefacts(id string) ([]artefact.Artefact, error) {
	queryResult, err := d.FindDoc(ArtefactColl, `[{"eq": `+id+`, "in": ["task"]}]`)
	var res []artefact.Artefact
	if err != nil {
		return []artefact.Artefact{}, err
	}

	// Query result are document IDs
	for docid := range queryResult {
		// Read document
		art, err := d.GetArtefact(docid)
		if err != nil {
			return []artefact.Artefact{}, err
		}

		res = append(res, art)
	}
	return res, nil
}

func (d *Database) ListTasks() []dbcommon.DocItem {
	return d.ListDocs(TaskColl)
}

func (d *Database) AllTasks(config *setting.Config) []agenttasks.Task {
	tasks := d.DB().Use(TaskColl)
	tasks_id := make([]agenttasks.Task, 0)

	tasks.ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		t := agenttasks.NewTaskFromJson(docContent)
		t.ID = strconv.Itoa(id)
		tasks_id = append(tasks_id, t)
		return true
	})
	return tasks_id
}

func (d *Database) AllTasksFiltered(config *setting.Config, f dbcommon.TaskFilter) (dbcommon.TaskResult, error) {
	tasks := d.DB().Use(TaskColl)
	tasks_map := make(map[string]agenttasks.Task, 0)
	tasks_keys := []string{}

	tasks.ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		t := agenttasks.NewTaskFromJson(docContent)
		t.ID = strconv.Itoa(id)
		key := t.ID

		// Check if there are filter to apply
		if f.Status != "" && t.Status != f.Status {
			goto ignore
		}

		if f.Result != "" && t.Result != f.Result {
			goto ignore
		}

		if f.Image != "" && !strings.Contains(t.Image, f.Image) {
			goto ignore
		}

		if f.ID != "" && !strings.Contains(t.ID, f.ID) {
			goto ignore
		}

		if f.Name != "" && !strings.Contains(t.Name, f.Name) {
			goto ignore
		}

		// NOTE: key that are unique use ID as postfix.
		switch f.Sort {
		case "_key":
			key = t.ID
		case "name":
			key = fmt.Sprintf("%s-%s", t.Name, t.ID)
		case "image":
			key = fmt.Sprintf("%s-%s", t.Image, t.ID)
		case "status":
			key = fmt.Sprintf("%s-%s", t.Status, t.ID)
		case "start_time":
			key = fmt.Sprintf("%s-%s", t.StartTime, t.ID)
		case "created_time":
			key = fmt.Sprintf("%s-%s", t.CreatedTime, t.ID)
		}

		tasks_map[key] = t
		tasks_keys = append(tasks_keys, key)

	ignore:

		return true
	})

	if f.SortOrder == "DESC" {
		sort.Sort(sort.Reverse(sort.StringSlice(tasks_keys)))
	} else {
		sort.Strings(tasks_keys)
	}

	startIndex := f.PageSize * f.PageIndex
	endIndex := f.PageSize * (f.PageIndex + 1)

	if endIndex > len(tasks_keys) {
		endIndex = len(tasks_keys)
	}

	stasks := []agenttasks.Task{}

	for _, k := range tasks_keys[startIndex:endIndex] {
		stasks = append(stasks, tasks_map[k])
	}

	// tiedot doesn't support limit
	ans := dbcommon.TaskResult{}

	ans.Total = len(tasks_keys)

	ans.Tasks = stasks

	return ans, nil
}

func (d *Database) AllNodeTask(config *setting.Config, id string) ([]agenttasks.Task, error) {
	queryResult, err := d.FindDoc(TaskColl, `[{"eq": "`+id+`", "in": ["node_id"]}]`)
	var res []agenttasks.Task
	if err != nil {
		return res, err
	}
	for docid := range queryResult {

		// Read document
		t, err := d.GetTask(config, docid)
		if err != nil {
			return res, err
		}

		res = append(res, t)
	}
	return res, nil
}

func (d *Database) AllUserTask(config *setting.Config, id string) ([]agenttasks.Task, error) {
	queryResult, err := d.FindDoc(TaskColl, `[{"eq": "`+id+`", "in": ["owner_id"]}]`)
	var res []agenttasks.Task
	if err != nil {
		return res, err
	}
	for docid := range queryResult {

		// Read document
		t, err := d.GetTask(config, docid)
		if err != nil {
			return res, err
		}

		res = append(res, t)
	}
	return res, nil
}

func (d *Database) AllUserFiltered(config *setting.Config, id string, f dbcommon.TaskFilter) (dbcommon.TaskResult, error) {
	// tiedot doesn't support limit
	ans := dbcommon.TaskResult{}

	tasks, err := d.AllUserTask(config, id)
	if err != nil {
		return ans, err
	}
	ans.Total = len(tasks)
	ans.Tasks = tasks

	return ans, nil
}

func (d *Database) GetTaskMetrics() (map[string]interface{}, error) {
	tasks := d.AllTasks(nil)

	statusRes := make(map[string]int, 0)
	resultRes := make(map[string]int, 0)

	for _, t := range tasks {
		if _, ok := statusRes[t.Status]; ok {
			statusRes[t.Status] = statusRes[t.Status] + 1
		} else {
			statusRes[t.Status] = 1
		}

		if _, ok := resultRes[t.Result]; ok {
			resultRes[t.Result] = resultRes[t.Result] + 1
		} else {
			resultRes[t.Result] = 1
		}
	}

	ans := map[string]interface{}{
		"status": statusRes,
		"result": resultRes,
		"total":  len(tasks),
	}

	return ans, nil
}
