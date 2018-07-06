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

package database

import (
	agenttasks "github.com/MottainaiCI/mottainai-server/pkg/tasks"
)

var PlansColl = "Plans"

func (d *Database) IndexPlan() {
	d.AddIndex(PlansColl, []string{"status"})
	d.AddIndex(PlansColl, []string{"result"})
	d.AddIndex(PlansColl, []string{"result", "status"})
}

func (d *Database) InsertPlan(t *agenttasks.Plan) (int, error) {
	return d.CreatePlan(t.ToMap())
}

func (d *Database) CreatePlan(t map[string]interface{}) (int, error) {
	return d.InsertDoc(PlansColl, t)
}

func (d *Database) ClonePlan(t int) (int, error) {
	task, err := d.GetPlan(t)
	if err != nil {
		return 0, err
	}

	return d.InsertPlan(&task)
}

func (d *Database) DeletePlan(docID int) error {
	return d.DeleteDoc(PlansColl, docID)
}

func (d *Database) UpdatePlan(docID int, t map[string]interface{}) error {
	return d.UpdateDoc(PlansColl, docID, t)
}

func (d *Database) GetPlan(docID int) (agenttasks.Plan, error) {
	doc, err := d.GetDoc(PlansColl, docID)
	if err != nil {
		return agenttasks.Plan{}, err
	}
	th := agenttasks.DefaultTaskHandler()
	t := th.NewPlanFromMap(doc)
	t.ID = docID
	return t, err
}

func (d *Database) ListPlans() []DocItem {
	return d.ListDocs(PlansColl)
}

func (d *Database) AllPlans() []agenttasks.Plan {
	tasks := d.DB().Use(PlansColl)
	tasks_id := make([]agenttasks.Plan, 0)
	th := agenttasks.DefaultTaskHandler()

	tasks.ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		t := th.NewPlanFromJson(docContent)
		if t.Task == nil {
			t.Task = &agenttasks.Task{}
		}
		t.ID = id
		tasks_id = append(tasks_id, t)
		return true
	})
	return tasks_id
}
