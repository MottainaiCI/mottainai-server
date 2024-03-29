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
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	tiedot_data "github.com/HouzuoGuo/tiedot/data"
	"github.com/HouzuoGuo/tiedot/db"
	"github.com/mudler/anagent"

	dbcommon "github.com/MottainaiCI/mottainai-server/pkg/db/common"
	"github.com/MottainaiCI/mottainai-server/pkg/entities"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"
)

type Database struct {
	*anagent.Anagent
	DBPath string
	DBName string
}

var Collections = []string{
	WebHookColl, TaskColl, SecretColl,
	UserColl, PlansColl, PipelinesColl,
	NodeColl, NamespaceColl, TokenColl,
	ArtefactColl, StorageColl, OrganizationColl,
	SettingColl, QueueColl, NodeQueuesColl,
}

func (d *Database) GetCollectionName(entity entities.MottainaiEntity) (ans string) {
	switch entity {
	case entities.Webhooks:
		ans = WebHookColl
	case entities.Tasks:
		ans = TaskColl
	case entities.Secrets:
		ans = SecretColl
	case entities.Users:
		ans = UserColl
	case entities.Plans:
		ans = PlansColl
	case entities.Pipelines:
		ans = PipelinesColl
	case entities.Nodes:
		ans = NodeColl
	case entities.Namespaces:
		ans = NamespaceColl
	case entities.Tokens:
		ans = TokenColl
	case entities.Artefacts:
		ans = ArtefactColl
	case entities.Storages:
		ans = StorageColl
	case entities.Organizations:
		ans = OrganizationColl
	case entities.Settings:
		ans = SettingColl
	case entities.Queues:
		ans = QueueColl
	case entities.NodeQueues:
		ans = NodeQueuesColl
	}
	return
}

func New(path string) *Database {
	return &Database{Anagent: anagent.New(), DBPath: path}
}

func (d *Database) GetAgent() *anagent.Anagent {
	return d.Anagent
}

func (d *Database) Init() {
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
	d.IndexNodeQueue()
	d.IndexQueue()
	d.IndexNamespace()
	d.IndexArtefacts()
	d.IndexStorage()
	d.IndexUser()
	d.IndexToken()
	d.IndexOrganization()
	d.IndexSetting()
	d.IndexPipeline()
	d.IndexWebHook()
	d.IndexSecret()
}

var MyDbInstance *db.DB

func (d *Database) DB() *db.DB {
	dbconf := filepath.Join(d.DBPath, "data-config.json")

	if MyDbInstance != nil {
		return MyDbInstance
	}

	exists, err := utils.Exists(dbconf)
	if err != nil {
		panic(err)
	}

	if !exists {
		d.Invoke(func(config *setting.Config) {
			// POST: Create config file with our settings.
			conf := tiedot_data.Config{
				DocMaxRoom:    config.GetDatabase().TiedotDocMaxRoom,
				ColFileGrowth: config.GetDatabase().TiedotColFileGrowth,
				PerBucket:     config.GetDatabase().TiedotPerBucket,
				HTFileGrowth:  config.GetDatabase().TiedotHTFileGrowth,
				HashBits:      config.GetDatabase().TiedotHashBits,
			}

			confData, err := json.Marshal(conf)
			if err != nil {
				panic(err)
			}

			if err := os.MkdirAll(d.DBPath, 0700); err != nil {
				panic(err)
			}

			f, err := os.OpenFile(dbconf, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				panic(err)
			}
			defer f.Close()

			_, err = f.Write(confData)
			if err != nil {
				panic(err)
			}

		})
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

func (d *Database) InsertDoc(coll string, t map[string]interface{}) (string, error) {
	// Insert document (afterwards the docID uniquely identifies the document and will never change)

	id, err := d.DB().Use(coll).Insert(t)
	return strconv.Itoa(id), err
}

func (d *Database) RestoreDoc(coll, id string, t map[string]interface{}) error {
	idi, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	return d.DB().Use(coll).InsertRecovery(idi, t)
}

func (d *Database) FindDoc(coll string, searchquery string) (map[string]interface{}, error) {

	var query interface{}
	json.Unmarshal([]byte(searchquery), &query)

	queryResult := make(map[int]struct{}) // query result (document IDs) goes into map keys
	res := make(map[string]interface{})   // query result (document IDs) goes into map keys

	err := db.EvalQuery(query, d.DB().Use(coll), &queryResult)

	for k, v := range queryResult {
		res[strconv.Itoa(k)] = v
	}

	return res, err
}

func (d *Database) DeleteDoc(coll string, docID string) error {
	uuid, err := strconv.Atoi(docID)
	if err != nil {
		return err
	}
	return d.DB().Use(coll).Delete(uuid)
}

func (d *Database) UpdateDoc(coll string, docID string, t map[string]interface{}) error {
	uuid, err := strconv.Atoi(docID)
	if err != nil {
		return err
	}
	old, _ := d.GetDoc(coll, docID)
	for k, v := range t {
		old[k] = v
	}
	return d.DB().Use(coll).Update(uuid, old)
}

func (d *Database) ReplaceDoc(coll string, docID string, t map[string]interface{}) error {
	uuid, err := strconv.Atoi(docID)
	if err != nil {
		return err
	}
	return d.DB().Use(coll).Update(uuid, t)
}

func (d *Database) GetDoc(coll string, docID string) (map[string]interface{}, error) {
	uuid, err := strconv.Atoi(docID)
	if err != nil {
		return map[string]interface{}{}, err
	}
	return d.DB().Use(coll).Read(uuid)
}

func (d *Database) ListDocs(coll string) []dbcommon.DocItem {
	tasks := d.DB().Use(coll)
	tasks_id := make([]dbcommon.DocItem, 0)
	tasks.ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		tasks_id = append(tasks_id, dbcommon.DocItem{Id: strconv.Itoa(id), Content: string(docContent)})
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
