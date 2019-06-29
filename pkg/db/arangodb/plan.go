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

var PlansColl = "Plans"

func (d *Database) IndexPlan() {
	d.AddIndex(PlansColl, []string{"status"})
	d.AddIndex(PlansColl, []string{"result"})
	d.AddIndex(PlansColl, []string{"result", "status"})
}

func (d *Database) InsertPlan(t *agenttasks.Plan) (string, error) {
	return d.CreatePlan(t.ToMap())
}

func (d *Database) CreatePlan(t map[string]interface{}) (string, error) {
	return d.InsertDoc(PlansColl, t)
}

func (d *Database) ClonePlan(config *setting.Config, t string) (string, error) {
	task, err := d.GetPlan(config, t)
	if err != nil {
		return "", err
	}

	return d.InsertPlan(&task)
}

func (d *Database) DeletePlan(docID string) error {
	return d.DeleteDoc(PlansColl, docID)
}

func (d *Database) UpdatePlan(docID string, t map[string]interface{}) error {
	return d.UpdateDoc(PlansColl, docID, t)
}

func (d *Database) GetPlan(config *setting.Config, docID string) (agenttasks.Plan, error) {
	doc, err := d.GetDoc(PlansColl, docID)
	if err != nil {
		return agenttasks.Plan{}, err
	}
	t := agenttasks.NewPlanFromMap(doc)
	t.ID = docID
	return t, err
}

func (d *Database) ListPlans() []dbcommon.DocItem {
	return d.ListDocs(PlansColl)
}

func (d *Database) AllPlans(config *setting.Config) []agenttasks.Plan {
	tasks_id := make([]agenttasks.Plan, 0)

	docs, err := d.FindDoc("", "FOR c IN "+PlansColl+" return c")
	if err != nil {
		return tasks_id
	}

	for k, _ := range docs {
		t, err := d.GetPlan(config, k)
		if err != nil {
			return tasks_id
		}
		tasks_id = append(tasks_id, t)
	}

	return tasks_id
}
