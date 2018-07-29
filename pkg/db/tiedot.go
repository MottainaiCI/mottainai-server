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
	"encoding/json"

	"github.com/HouzuoGuo/tiedot/db"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"

	"github.com/MottainaiCI/mottainai-server/pkg/utils"
)

var Collections = []string{TaskColl,
	UserColl, PlansColl, PipelinesColl, NodeColl, NamespaceColl, TokenColl, ArtefactColl, StorageColl, OrganizationColl, SettingColl}

//var DBInstance *Interface{}
/// POC
func (d *Database) Init() {
	d.DBPath = setting.Configuration.DBPath

	colls := d.DB().AllCols()
	for _, c := range Collections {
		if !utils.ArrayContainsString(colls, c) {
			if err := d.DB().Create(c); err != nil {
				return
			}
		}
	}

	d.IndexPlan()
	d.IndexTask()
	d.IndexNode()
	d.IndexNamespace()
	d.IndexArtefacts()
	d.IndexStorage()
	d.IndexUser()
	d.IndexToken()
	d.IndexOrganization()
	d.IndexSetting()
	d.IndexPipeline()
}

var MyDbInstance *db.DB

func (d *Database) DB() *db.DB {
	if MyDbInstance != nil {
		return MyDbInstance
	}

	myDB, err := db.OpenDB(d.DBPath)
	if err != nil {
		panic(err)
	}
	MyDbInstance = myDB
	return myDB
}

func (d *Database) AddIndex(coll string, i []string) error {
	return d.DB().Use(coll).Index(i)
}

func (d *Database) AllIndex(coll string) [][]string {
	return d.DB().Use(coll).AllIndexes()
}

func (d *Database) RemoveIndex(coll string, i []string) error {
	return d.DB().Use(coll).Unindex(i)
}

func (d *Database) InsertDoc(coll string, t map[string]interface{}) (int, error) {
	// Insert document (afterwards the docID uniquely identifies the document and will never change)
	return d.DB().Use(coll).Insert(t)
}

func (d *Database) FindDoc(coll string, searchquery string) (map[int]struct{}, error) {

	var query interface{}
	json.Unmarshal([]byte(searchquery), &query)

	queryResult := make(map[int]struct{}) // query result (document IDs) goes into map keys
	err := db.EvalQuery(query, d.DB().Use(coll), &queryResult)

	return queryResult, err
}

func (d *Database) DeleteDoc(coll string, docID int) error {
	return d.DB().Use(coll).Delete(docID)
}

func (d *Database) UpdateDoc(coll string, docID int, t map[string]interface{}) error {

	old, _ := d.GetDoc(coll, docID)
	for k, v := range t {
		old[k] = v
	}
	return d.DB().Use(coll).Update(docID, old)
}

func (d *Database) ReplaceDoc(coll string, docID int, t map[string]interface{}) error {
	return d.DB().Use(coll).Update(docID, t)
}

func (d *Database) GetDoc(coll string, docID int) (map[string]interface{}, error) {
	return d.DB().Use(coll).Read(docID)
}

type DocItem struct {
	Id      int
	Content interface{}
}

func (d *Database) ListDocs(coll string) []DocItem {
	tasks := d.DB().Use(coll)
	tasks_id := make([]DocItem, 0)
	tasks.ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		tasks_id = append(tasks_id, DocItem{Id: id, Content: string(docContent)})
		return true
	})
	return tasks_id
}

func (d *Database) RenameColl(coll, coll2 string) error {
	err := d.DB().Rename(coll, coll2)
	return err
}

func (d *Database) DropColl(coll string) error {
	err := d.DB().Drop(coll)
	if err != nil {
		panic(err)
	}
	return err
}

func (d *Database) ScrubColl(coll string) error {
	err := d.DB().Scrub(coll)
	if err != nil {
		panic(err)
	}
	return err
}
