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
	"strconv"

	dbcommon "github.com/MottainaiCI/mottainai-server/pkg/db/common"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
)

var PipelinesColl = "Pipelines"

func (d *Database) IndexPipeline() {
	d.AddIndex(PipelinesColl, []string{"status"})
	d.AddIndex(PipelinesColl, []string{"result"})
	d.AddIndex(PipelinesColl, []string{"result", "status"})
}

func (d *Database) InsertPipeline(t *agenttasks.Pipeline) (string, error) {
	return d.CreatePipeline(t.ToMap(false))
}

func (d *Database) CreatePipeline(t map[string]interface{}) (string, error) {
	return d.InsertDoc(PipelinesColl, t)
}

func (d *Database) ClonePipeline(config *setting.Config, t string) (string, error) {
	task, err := d.GetPipeline(config, t)
	if err != nil {
		return "", err
	}

	return d.InsertPipeline(&task)
}

func (d *Database) DeletePipeline(docID string) error {
	return d.DeleteDoc(PipelinesColl, docID)
}

func (d *Database) AllUserPipelines(config *setting.Config, id string) ([]agenttasks.Pipeline, error) {
	queryResult, err := d.FindDoc(TaskColl, `[{"eq": "`+id+`", "in": ["owner_id"]}]`)
	var res []agenttasks.Pipeline
	if err != nil {
		return res, err
	}
	for docid := range queryResult {

		// Read document
		t, err := d.GetPipeline(config, docid)
		t.ID = docid
		if err != nil {
			return res, err
		}

		res = append(res, t)
	}
	return res, nil
}

func (d *Database) UpdatePipeline(docID string, t map[string]interface{}) error {
	return d.UpdateDoc(PipelinesColl, docID, t)
}

func (d *Database) UpdateLivePipelineData(config *setting.Config, pip *agenttasks.Pipeline) error {
	for k, t := range pip.Tasks {
		ta, err := d.GetTask(config, t.ID)
		if err != nil {
			return err
		}
		pip.Tasks[k] = ta
	}
	return nil
}

func (d *Database) GetPipeline(config *setting.Config, docID string) (agenttasks.Pipeline, error) {
	doc, err := d.GetDoc(PipelinesColl, docID)
	if err != nil {
		return agenttasks.Pipeline{}, err
	}
	t := agenttasks.NewPipelineFromMap(doc)
	t.ID = docID
	//err = d.UpdateLivePipelineData(config, &t)
	return t, err
}

func (d *Database) ListPipelines() []dbcommon.DocItem {
	return d.ListDocs(PipelinesColl)
}

func (d *Database) AllPipelines(config *setting.Config) []agenttasks.Pipeline {
	tasks := d.DB().Use(PipelinesColl)
	tasks_id := make([]agenttasks.Pipeline, 0)

	tasks.ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		t := agenttasks.NewPipelineFromJson(docContent)
		t.ID = strconv.Itoa(id)
		tasks_id = append(tasks_id, t)
		return true
	})
	return tasks_id
}
