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

	queryResult, err := d.FindDoc("", `FOR c IN `+PipelinesColl+`
		FILTER c.owner_id == "`+id+`"
		RETURN c`)
	if err != nil {
		return []agenttasks.Pipeline{}, err
	}
	var res []agenttasks.Pipeline

	// Query result are document IDs
	for id, _ := range queryResult {

		// Read document
		art, err := d.GetPipeline(config, id)
		if err != nil {
			return []agenttasks.Pipeline{}, err
		}
		res = append(res, art)
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

	tasks_id := make([]agenttasks.Pipeline, 0)

	docs, err := d.FindDoc("", "FOR c IN "+PipelinesColl+" return c")
	if err != nil {
		return tasks_id
	}

	for k, _ := range docs {
		t, err := d.GetPipeline(config, k)
		if err != nil {
			return tasks_id
		}
		tasks_id = append(tasks_id, t)
	}

	return tasks_id
}
