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
	"context"
	"crypto/tls"
	"fmt"

	dbcommon "github.com/MottainaiCI/mottainai-server/pkg/db/common"

	arango "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/mudler/anagent"
)

type Database struct {
	*anagent.Anagent
	Endpoints                []string
	Database, DBUser, DBPass string
	CertPath, KeyPath        string
}

var Collections = []string{WebHookColl, TaskColl, SecretColl,
	UserColl, PlansColl, PipelinesColl, NodeColl, NamespaceColl, TokenColl, ArtefactColl, StorageColl, OrganizationColl, SettingColl}

func New(db, u, p, cp, kp string, e []string) *Database {
	return &Database{Anagent: anagent.New(), Database: db, Endpoints: e, CertPath: cp, KeyPath: kp, DBUser: u, DBPass: p}
}

func (d *Database) GetAgent() *anagent.Anagent {
	return d.Anagent
}

func (d *Database) createCollections() error {
	for _, c := range Collections {
		ctx := context.Background()
		found, err := d.DB().CollectionExists(ctx, c)
		if err != nil {
			return err
		}

		if !found {
			ctx := context.Background()
			// TODO: More options
			options := &arango.CreateCollectionOptions{ /* ... */ }
			_, err := d.DB().CreateCollection(ctx, c, options)
			if err != nil {
				return err
			}
		}

	}
	return nil
}
func (d *Database) Init() {

	d.createCollections()

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
	d.IndexSecret()
	d.IndexWebHook()
}

func (d *Database) AddIndex(coll string, i []string) error {
	ctx := context.Background()

	col, err := d.UseCol(coll)
	if err != nil {
		return err
	}
	// TODO: More options
	_, _, err = col.EnsureSkipListIndex(ctx, i, &arango.EnsureSkipListIndexOptions{})

	return err
}

var Instance arango.Database
var Client arango.Client

func (d *Database) GetClient() arango.Client {
	if Client != nil {
		return Client
	}
	config := &tls.Config{}

	if len(d.CertPath) > 0 && len(d.KeyPath) > 0 {
		cer, err := tls.LoadX509KeyPair(d.CertPath, d.KeyPath)
		if err != nil {
			panic(err)
		}
		config = &tls.Config{Certificates: []tls.Certificate{cer}}
	}
	// for now be fatal with DB errors
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: d.Endpoints,
		TLSConfig: config,
	})
	if err != nil {
		panic(err)
	}

	arangoConfig := arango.ClientConfig{Connection: conn}

	if len(d.DBUser) > 0 && len(d.DBPass) > 0 {
		arangoConfig.Authentication = arango.BasicAuthentication(d.DBUser, d.DBPass)
	}

	c, err := arango.NewClient(arangoConfig)
	if err != nil {
		panic(err)
	}

	Client = c
	return c
}

func (d *Database) DB() arango.Database {
	if Instance != nil {
		return Instance
	}
	c := d.GetClient()
	ctx := context.Background()

	ok, err := c.DatabaseExists(ctx, d.Database)
	if err != nil {
		panic(err)
	}

	if !ok {
		_, err = c.CreateDatabase(ctx, d.Database, &arango.CreateDatabaseOptions{})
		if err != nil {
			panic(err)
		}
	}

	db, err := c.Database(ctx, d.Database)
	if err != nil {
		panic(err)
	}
	Instance = db
	return db
}

func (d *Database) UseCol(coll string) (arango.Collection, error) {
	ctx := context.Background()
	col, err := d.DB().Collection(ctx, coll)
	if err != nil {
		return nil, err
	}
	return col, err
}

func (d *Database) InsertDoc(coll string, t map[string]interface{}) (string, error) {
	// Insert document (afterwards the docID uniquely identifies the document and will never change)
	col, err := d.UseCol(coll)
	if err != nil {
		return "", err
	}

	// FIXME: We need to get rid of this from the schema e.g. with a factory that builds object consumed here
	delete(t, "ID")
	delete(t, "id")

	ctx := context.Background()
	meta, err := col.CreateDocument(ctx, t)
	if err != nil {
		return "", err
	}
	return meta.Key, err
}

func (d *Database) FindDoc(coll string, searchquery string) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	ctx := context.Background()
	cursor, err := d.DB().Query(ctx, searchquery, nil)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()
	for {
		var doc interface{}
		meta, err := cursor.ReadDocument(ctx, &doc)
		if arango.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}
		res[meta.Key] = doc
	}

	return res, nil
}

func (d *Database) DeleteDoc(coll string, docID string) error {
	col, err := d.UseCol(coll)
	if err != nil {
		return err
	}

	ctx := context.Background()
	_, err = col.RemoveDocument(ctx, docID)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) GetDoc(coll string, docID string) (map[string]interface{}, error) {
	var doc map[string]interface{}
	col, err := d.UseCol(coll)
	if err != nil {
		return doc, err
	}

	ctx := context.Background()
	_, err = col.ReadDocument(ctx, docID, &doc)
	if err != nil {
		return doc, err
	}

	return doc, nil

}

func (d *Database) UpdateDoc(coll string, docID string, t map[string]interface{}) error {

	old, _ := d.GetDoc(coll, docID)
	for k, v := range t {
		old[k] = v
	}
	return d.ReplaceDoc(coll, docID, old)
}

func (d *Database) ReplaceDoc(coll string, docID string, t map[string]interface{}) error {
	col, err := d.UseCol(coll)
	if err != nil {
		return err
	}

	ctx := context.Background()

	_, err = col.UpdateDocument(ctx, docID, t)
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) ListDocs(coll string) []dbcommon.DocItem {
	tasks_id := make([]dbcommon.DocItem, 0)

	docs, err := d.FindDoc("", "FOR c IN "+coll+" return c")
	if err != nil {
		return []dbcommon.DocItem{}
	}

	for k, v := range docs {
		tasks_id = append(tasks_id, dbcommon.DocItem{Id: k, Content: fmt.Sprintf("%v", v)})
	}

	return tasks_id
}

func (d *Database) DropColl(coll string) error {
	col, err := d.UseCol(coll)
	if err != nil {
		return err
	}
	err = col.Remove(nil)
	if err != nil {
		return err
	}
	return nil
}
