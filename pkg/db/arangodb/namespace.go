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
	"os"
	"path/filepath"

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

func (d *Database) CreateNamespace(t map[string]interface{}) (string, error) {
	return d.InsertDoc(NamespaceColl, t)
}

func (d *Database) DeleteNamespace(docID string) error {

	ns, err := d.GetNamespace(docID)
	if err != nil {
		return err
	}
	artefacts, err := d.GetNamespaceArtefacts(docID)
	if err != nil {
		return err
	}
	d.Invoke(func(config *setting.Config) {
		for _, artefact := range artefacts {
			artefact.CleanFromNamespace(ns.Path, config)
			d.DeleteArtefact(artefact.ID)
		}

		os.RemoveAll(filepath.Join(config.GetStorage().NamespacePath, ns.Path))
	})

	return d.DeleteDoc(NamespaceColl, docID)
}

func (d *Database) UpdateNamespace(docID string, t map[string]interface{}) error {
	return d.UpdateDoc(NamespaceColl, docID, t)
}

func (d *Database) SearchNamespace(name string) (namespace.Namespace, error) {
	queryResult, err := d.FindDoc("", `FOR c IN `+NamespaceColl+`
		FILTER c.name == "`+name+`"
		RETURN c`)
	if err != nil {
		return namespace.Namespace{}, err
	}
	var res []namespace.Namespace
	if err != nil {
		return namespace.Namespace{}, err
	}
	// Query result are document IDs
	for id, doc := range queryResult {
		t := namespace.NewFromMap(doc.(map[string]interface{}))
		t.ID = id
		res = append(res, t)
	}
	return res[0], nil
}

func (d *Database) GetNamespace(docID string) (namespace.Namespace, error) {
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

func (d *Database) GetNamespaceArtefacts(id string) ([]artefact.Artefact, error) {
	queryResult, err := d.FindDoc("", `FOR c IN `+ArtefactColl+`
    FILTER c.namespace == "`+id+`"
    RETURN c`)
	if err != nil {
		return []artefact.Artefact{}, err
	}
	var res []artefact.Artefact

	// Query result are document IDs
	for id, doc := range queryResult {
		art := artefact.NewFromMap(doc.(map[string]interface{}))
		art.ID = id
		res = append(res, art)
	}
	return res, nil
}

func (d *Database) AllNamespaces() []namespace.Namespace {
	Namespaces_id := make([]namespace.Namespace, 0)

	docs, err := d.FindDoc("", "FOR c IN "+NamespaceColl+" return c")
	if err != nil {
		return []namespace.Namespace{}
	}

	for k, doc := range docs {
		t := namespace.NewFromMap(doc.(map[string]interface{}))
		t.ID = k
		Namespaces_id = append(Namespaces_id, t)
	}

	return Namespaces_id
}
