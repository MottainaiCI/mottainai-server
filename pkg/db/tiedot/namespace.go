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

package tiedot

import (
	"os"
	"path/filepath"
	"strconv"

	dbcommon "github.com/MottainaiCI/mottainai-server/pkg/db/common"

	"github.com/MottainaiCI/mottainai-server/pkg/artefact"
	"github.com/MottainaiCI/mottainai-server/pkg/namespace"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
)

var NamespaceColl = "Namespaces"

func (d *Database) IndexNamespace() {
	d.AddIndex(NamespaceColl, []string{"name"})
	d.AddIndex(NamespaceColl, []string{"path"})
}
func (d *Database) CreateNamespace(t map[string]interface{}) (int, error) {
	return d.InsertDoc(NamespaceColl, t)
}

func (d *Database) DeleteNamespace(docID int) error {

	ns, err := d.GetNamespace(docID)
	if err != nil {
		return err
	}
	artefacts, err := d.GetNamespaceArtefacts(docID)
	if err != nil {
		return err
	}
	for _, artefact := range artefacts {
		artefact.CleanFromNamespace(ns.Path)
		d.DeleteArtefact(artefact.ID)
	}
	os.RemoveAll(filepath.Join(setting.Configuration.NamespacePath, ns.Path))

	return d.DeleteDoc(NamespaceColl, docID)
}

func (d *Database) UpdateNamespace(docID int, t map[string]interface{}) error {
	return d.UpdateDoc(NamespaceColl, docID, t)
}

func (d *Database) SearchNamespace(name string) (namespace.Namespace, error) {
	queryResult, err := d.FindDoc(NamespaceColl, `[{"eq": "`+name+`", "in": ["name"]}]`)
	var res []namespace.Namespace
	if err != nil {
		return namespace.Namespace{}, err
	}
	ns := d.DB().Use(NamespaceColl)

	// Query result are document IDs
	for id := range queryResult {
		// Read document
		readBack, err := ns.Read(id)
		if err != nil {
			return namespace.Namespace{}, err
		}
		res = append(res, namespace.NewFromMap(readBack))
	}
	return res[0], nil
}

func (d *Database) GetNamespace(docID int) (namespace.Namespace, error) {
	doc, err := d.GetDoc(NamespaceColl, docID)
	if err != nil {
		return namespace.Namespace{}, err
	}
	t := namespace.NewFromMap(doc)
	t.ID = docID
	return t, err
}

func (d *Database) ListNamespaces() []dbcommon.DocItem {
	return d.ListDocs(NamespaceColl)
}

func (d *Database) GetNamespaceArtefacts(id int) ([]artefact.Artefact, error) {
	queryResult, err := d.FindDoc(ArtefactColl, `[{"eq": "`+strconv.Itoa(id)+`", "in": ["namespace"]}]`)
	var res []artefact.Artefact
	if err != nil {
		return res, err
	}

	// Query result are document IDs
	for docid := range queryResult {
		// Read document
		art, err := d.GetArtefact(docid)
		if err != nil {
			return res, err
		}

		res = append(res, art)
	}
	return res, nil
}

func (d *Database) AllNamespaces() []namespace.Namespace {
	Namespaces := d.DB().Use(NamespaceColl)
	Namespaces_id := make([]namespace.Namespace, 0)

	Namespaces.ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		t := namespace.NewFromJson(docContent)
		t.ID = id
		Namespaces_id = append(Namespaces_id, t)
		return true
	})
	return Namespaces_id
}
