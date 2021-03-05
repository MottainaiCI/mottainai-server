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

package arangodb

import (
	"errors"
	"fmt"
	"github.com/MottainaiCI/mottainai-server/pkg/artefact"
	dbcommon "github.com/MottainaiCI/mottainai-server/pkg/db/common"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
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

func (d *Database) CloneTask(config *setting.Config, t string) (string, error) {
	tdata, err := d.GetTask(config, t)
	if err != nil {
		return "", err
	}
	tdata.Reset()
	tdata.ID = ""
	return d.InsertTask(&tdata)
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

// todo: deprecate
func (d *Database) GetTaskByStatus(config *setting.Config, status string) ([]agenttasks.Task, error) {
	var res []agenttasks.Task

	var query string
	switch status {
	case "running":
		query = `FOR c IN ` + TaskColl + `
			FILTER c.status == "running"
			RETURN c`
	case "waiting":
		query = `FOR c IN ` + TaskColl + `
			FILTER c.status == "waiting"
			RETURN c`
	case "stop":
		query = `FOR c IN ` + TaskColl + `
			FILTER c.status == "stop"
			RETURN c`
	case "stopped":
		query = `FOR c IN ` + TaskColl + `
			FILTER c.status == "stopped"
			RETURN c`
	case "error":
		query = `FOR c IN ` + TaskColl + `
			FILTER c.result == "error"
			RETURN c`
	case "failed":
		query = `FOR c IN ` + TaskColl + `
			FILTER c.result == "failed"
			RETURN c`
	case "success":
		query = `FOR c IN ` + TaskColl + `
			FILTER c.result == "success"
			RETURN c`
	default:
		return res, errors.New("No valid status supplied")
	}

	queryResult, err := d.FindDoc("", query)
	if err != nil {
		return res, err
	}

	// Query result are document IDs
	for id, v := range queryResult {
		t := agenttasks.NewTaskFromMap(v.(map[string]interface{}))
		t.ID = id
		res = append(res, t)
	}
	return res, nil

}

func (d *Database) GetTaskArtefacts(id string) ([]artefact.Artefact, error) {

	queryResult, err := d.FindDoc("", `FOR c IN `+ArtefactColl+`
		FILTER c.task == "`+id+`"
		RETURN c`)

	var res []artefact.Artefact
	if err != nil {
		return res, err
	}

	// Query result are document IDs
	for id, v := range queryResult {
		t := artefact.NewFromMap(v.(map[string]interface{}))
		t.ID = id
		res = append(res, t)
	}
	return res, nil
}

func (d *Database) ListTasks() []dbcommon.DocItem {
	return d.ListDocs(TaskColl)
}

func (d *Database) AllTasks(config *setting.Config) []agenttasks.Task {
	tasks_id := make([]agenttasks.Task, 0)
	docs, err := d.FindDoc("", "FOR c IN "+TaskColl+" return c")
	if err != nil {
		return tasks_id
	}

	for k, v := range docs {
		t := agenttasks.NewTaskFromMap(v.(map[string]interface{}))
		t.ID = k
		tasks_id = append(tasks_id, t)
	}

	return tasks_id
}

func (d *Database) AllTasksFiltered(config *setting.Config, f dbcommon.TaskFilter) (res dbcommon.TaskResult, err error) {
	sortClause := fmt.Sprintf("SORT c.%s %s", f.Sort, f.SortOrder)

	query :=
		fmt.Sprintf(
			"FOR c IN %s %s LIMIT %d, %d RETURN c",
			TaskColl,
			sortClause,
			f.PageIndex*f.PageSize, f.PageSize)

	docs, err := d.FindDocSorted(query)
	if err != nil {
		return res, err
	}

	countRes, err := d.FindDocSorted(fmt.Sprintf("RETURN LENGTH(%s)", TaskColl))

	var tasks []agenttasks.Task
	for _, v := range docs {
		doc := v.(map[string]interface{})
		t := agenttasks.NewTaskFromMap(doc)
		t.ID = doc["_key"].(string)
		tasks = append(tasks, t)
	}

	res.Tasks = tasks
	if i, ok := countRes[0].(float64); ok {
		res.Total = int(i)
	}

	return res, nil
}

func (d *Database) AllNodeTask(config *setting.Config, id string) ([]agenttasks.Task, error) {
	queryResult, err := d.FindDoc("", `FOR c IN `+TaskColl+`
		FILTER c.node_id == "`+id+`"
		RETURN c`)

	var res []agenttasks.Task
	if err != nil {
		return res, err
	}

	// Query result are document IDs
	for k, v := range queryResult {
		t := agenttasks.NewTaskFromMap(v.(map[string]interface{}))
		t.ID = k
		res = append(res, t)
	}
	return res, nil

}

func (d *Database) AllUserTask(config *setting.Config, id string) ([]agenttasks.Task, error) {

	queryResult, err := d.FindDoc("", `FOR c IN `+TaskColl+`
		FILTER c.owner_id == "`+id+`"
		RETURN c`)

	var res []agenttasks.Task
	if err != nil {
		return res, err
	}

	// Query result are document IDs
	for id, v := range queryResult {
		t := agenttasks.NewTaskFromMap(v.(map[string]interface{}))
		t.ID = id
		res = append(res, t)
	}
	return res, nil
}

func (d *Database) AllUserFiltered(config *setting.Config, id string, f dbcommon.TaskFilter) (res dbcommon.TaskResult, err error) {
	sortClause := fmt.Sprintf("SORT c.%s %s", f.Sort, f.SortOrder)

	query :=
		fmt.Sprintf(
			"FOR c IN %s %s LIMIT %d, %d RETURN c",
			TaskColl,
			sortClause,
			f.PageIndex*f.PageSize, f.PageSize)

	queryResult, err := d.FindDoc("",
		fmt.Sprintf("FOR c IN %s FILTER c.owner_id == %s %s LIMIT %d, %d RETURN c",
			TaskColl, id,
			query,
			f.PageIndex*f.PageSize, f.PageSize),
	)

	countRes, err := d.FindDoc("", fmt.Sprintf("RETURN LENGTH(%s)", TaskColl))
	fmt.Printf("%v\n", countRes)

	if err != nil {
		return res, err
	}

	// Query result are document IDs
	var tasks []agenttasks.Task
	for id, v := range queryResult {
		t := agenttasks.NewTaskFromMap(v.(map[string]interface{}))
		t.ID = id
		tasks = append(tasks, t)
	}

	res.Tasks = tasks
	res.Total = 0

	return res, nil
}

func (d *Database) GetTaskMetrics() (map[string]interface{}, error) {
	statusRes, err := d.CollectCount(TaskColl, "status")
	if err != nil {
		return nil, err
	}
	resultRes, err := d.CollectCount(TaskColl, "result")
	if err != nil {
		return nil, err
	}
	total, err := d.Count(TaskColl)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status": statusRes,
		"result": resultRes,
		"total":  total,
	}, nil
}
