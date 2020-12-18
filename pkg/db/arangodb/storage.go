/*

Copyright (C) 2018  Ettore Di Giacinto <mudler@gentoo.org>
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
	"os"
	"path/filepath"

	dbcommon "github.com/MottainaiCI/mottainai-server/pkg/db/common"

	"github.com/MottainaiCI/mottainai-server/pkg/artefact"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/MottainaiCI/mottainai-server/pkg/storage"
)

var StorageColl = "Storages"

func (d *Database) IndexStorage() {
	d.AddIndex(StorageColl, []string{"name"})
	d.AddIndex(StorageColl, []string{"path"})
	d.AddIndex(StorageColl, []string{"owner_id"})
}

func (d *Database) CreateStorage(t map[string]interface{}) (string, error) {
	return d.InsertDoc(StorageColl, t)
}

func (d *Database) DeleteStorage(docID string) error {

	ns, err := d.GetStorage(docID)
	if err != nil {
		return err
	}

	d.Invoke(func(config *setting.Config) {
		os.RemoveAll(filepath.Join(config.GetStorage().StoragePath, ns.Path))
	})

	return d.DeleteDoc(StorageColl, docID)
}

func (d *Database) UpdateStorage(docID string, t map[string]interface{}) error {
	return d.UpdateDoc(StorageColl, docID, t)
}

func (d *Database) SearchStorage(name string) (storage.Storage, error) {

	queryResult, err := d.FindDoc("", `FOR c IN `+StorageColl+`
		FILTER c.name == "`+name+`"
		RETURN c`)
	if err != nil {
		return storage.Storage{}, err
	}
	var res []storage.Storage
	if err != nil {
		return storage.Storage{}, err
	}
	// Query result are document IDs
	for id, doc := range queryResult {
		t := storage.NewFromMap(doc.(map[string]interface{}))
		t.ID = id
		res = append(res, t)
	}
	if len(res) == 0 {
		return storage.Storage{}, errors.New("No storages found")
	}
	return res[0], nil
}

func (d *Database) GetStorage(docID string) (storage.Storage, error) {
	doc, err := d.GetDoc(StorageColl, docID)
	if err != nil {
		return storage.Storage{}, err
	}
	t := storage.NewFromMap(doc)
	id := docID
	t.ID = id
	return t, err
}

func (d *Database) ListStorages() []dbcommon.DocItem {
	return d.ListDocs(StorageColl)
}

func (d *Database) GetStorageArtefacts(id string) ([]artefact.Artefact, error) {

	queryResult, err := d.FindDoc("", `FOR c IN `+ArtefactColl+`
		FILTER c.storage == "`+id+`"
		RETURN c`)

	var res []artefact.Artefact
	if err != nil {
		return res, err
	}

	// Query result are document IDs
	for id, doc := range queryResult {
		art := artefact.NewFromMap(doc.(map[string]interface{}))
		art.ID = id
		res = append(res, art)
	}
	return res, nil
}

func (d *Database) AllStorages() []storage.Storage {
	Storages_id := make([]storage.Storage, 0)

	docs, err := d.FindDoc("", "FOR c IN "+StorageColl+" return c")
	if err != nil {
		return Storages_id
	}

	for k, doc := range docs {
		t := storage.NewFromMap(doc.(map[string]interface{}))
		t.ID = k
		Storages_id = append(Storages_id, t)
	}

	return Storages_id
}
