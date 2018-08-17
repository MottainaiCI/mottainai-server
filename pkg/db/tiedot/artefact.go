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
	"strconv"

	"github.com/MottainaiCI/mottainai-server/pkg/artefact"
	dbcommon "github.com/MottainaiCI/mottainai-server/pkg/db/common"
)

var ArtefactColl = "Artefacts"

func (d *Database) IndexArtefacts() {
	d.AddIndex(ArtefactColl, []string{"task"})
	d.AddIndex(ArtefactColl, []string{"namespace"})
}

func (d *Database) CreateArtefact(t map[string]interface{}) (string, error) {
	return d.InsertDoc(ArtefactColl, t)
}

func (d *Database) DeleteArtefact(docID string) error {
	return d.DeleteDoc(ArtefactColl, docID)
}

func (d *Database) UpdateArtefact(docID string, t map[string]interface{}) error {
	return d.UpdateDoc(ArtefactColl, docID, t)
}

func (d *Database) GetArtefact(docID string) (artefact.Artefact, error) {
	doc, err := d.GetDoc(ArtefactColl, docID)
	if err != nil {
		return artefact.Artefact{}, err
	}
	t := artefact.NewFromMap(doc)
	t.ID = docID
	return t, err
}

func (d *Database) SearchArtefact(name string) (artefact.Artefact, error) {
	queryResult, err := d.FindDoc(ArtefactColl, `[{"eq": "`+name+`", "in": ["name"]}]`)
	var res []artefact.Artefact
	if err != nil {
		return artefact.Artefact{}, err
	}

	// Query result are document IDs
	for id := range queryResult {

		// Read document
		art, err := d.GetArtefact(id)
		if err != nil {
			return artefact.Artefact{}, err
		}
		res = append(res, art)
	}
	return res[0], nil
}

func (d *Database) ListArtefacts() []dbcommon.DocItem {
	return d.ListDocs(ArtefactColl)
}

func (d *Database) AllArtefacts() []artefact.Artefact {
	Artefacts := d.DB().Use(ArtefactColl)
	Artefacts_id := make([]artefact.Artefact, 0)

	Artefacts.ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		t := artefact.NewFromJson(docContent)
		t.ID = strconv.Itoa(id)
		Artefacts_id = append(Artefacts_id, t)
		return true
	})
	return Artefacts_id
}
