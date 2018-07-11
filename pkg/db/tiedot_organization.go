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
	"errors"

	organization "github.com/MottainaiCI/mottainai-server/pkg/organization"
)

var OrganizationColl = "Organizations"

func (d *Database) IndexOrganization() {
	d.AddIndex(OrganizationColl, []string{"name"})
}

func (d *Database) InsertOrganization(t *organization.Organization) (int, error) {
	return d.CreateOrganization(t.ToMap())
}

func (d *Database) CreateOrganization(t map[string]interface{}) (int, error) {

	return d.InsertDoc(OrganizationColl, t)
}

func (d *Database) DeleteOrganization(docID int) error {

	t, err := d.GetOrganization(docID)
	if err != nil {
		return err
	}

	t.Clear()
	return d.DeleteDoc(OrganizationColl, docID)
}

func (d *Database) UpdateOrganization(docID int, t map[string]interface{}) error {
	return d.UpdateDoc(OrganizationColl, docID, t)
}

func (d *Database) GetOrganizationByName(name string) (organization.Organization, error) {
	res, err := d.GetOrganizationsByName(name)
	if err != nil {
		return organization.Organization{}, err
	} else if len(res) == 0 {
		return organization.Organization{}, errors.New("No organization name found")
	} else {
		return res[0], nil
	}
}

func (d *Database) GetOrganizationsByName(name string) ([]organization.Organization, error) {
	var res []organization.Organization

	queryResult, err := d.FindDoc(OrganizationColl, `[{"eq": "`+name+`", "in": ["name"]}]`)
	if err != nil {
		return res, err
	}

	for docid := range queryResult {

		u, err := d.GetOrganization(docid)
		u.ID = docid
		if err != nil {
			return res, err
		}
		res = append(res, u)
	}
	return res, nil
}

func (d *Database) GetOrganization(docID int) (organization.Organization, error) {
	doc, err := d.GetDoc(OrganizationColl, docID)
	if err != nil {
		return organization.Organization{}, err
	}
	t := organization.NewOrganizationFromMap(doc)
	t.ID = docID
	return t, err
}

func (d *Database) ListOrganizations() []DocItem {
	return d.ListDocs(OrganizationColl)
}

// TODO: Change it, expensive for now
func (d *Database) CountOrganizations() int {
	return len(d.ListOrganizations())
}

func (d *Database) AllOrganizations() []organization.Organization {
	Organizations := d.DB().Use(OrganizationColl)
	Organizations_id := make([]organization.Organization, 0)

	Organizations.ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		t := organization.NewOrganizationFromJson(docContent)
		t.ID = id
		Organizations_id = append(Organizations_id, t)
		return true
	})
	return Organizations_id
}
