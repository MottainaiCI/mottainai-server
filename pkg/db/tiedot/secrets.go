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
	"errors"
	"strconv"

	dbcommon "github.com/MottainaiCI/mottainai-server/pkg/db/common"

	secret "github.com/MottainaiCI/mottainai-server/pkg/secret"
)

var SecretColl = "Secrets"

func (d *Database) IndexSecret() {
	d.AddIndex(SecretColl, []string{"name"})
	d.AddIndex(SecretColl, []string{"owner_id"})
}

func (d *Database) InsertSecret(t *secret.Secret) (string, error) {
	return d.CreateSecret(t.ToMap())
}

func (d *Database) CreateSecret(t map[string]interface{}) (string, error) {
	return d.InsertDoc(SecretColl, t)
}

func (d *Database) DeleteSecret(docID string) error {
	t, err := d.GetSecret(docID)
	if err != nil {
		return err
	}

	t.Clear()
	return d.DeleteDoc(SecretColl, docID)
}

func (d *Database) UpdateSecret(docID string, t map[string]interface{}) error {
	return d.UpdateDoc(SecretColl, docID, t)
}

func (d *Database) GetSecretByUserID(id string) (secret.Secret, error) {
	res, err := d.GetSecretsByUserID(id)
	if err != nil {
		return secret.Secret{}, err
	} else if len(res) == 0 {
		return secret.Secret{}, errors.New("No secretname found")
	} else {
		return res[0], nil
	}
}

func (d *Database) GetSecretByName(name string) (secret.Secret, error) {
	res, err := d.GetSecretsByName(name)
	if err != nil {
		return secret.Secret{}, err
	} else if len(res) == 0 {
		return secret.Secret{}, errors.New("No secretname found")
	} else {
		return res[0], nil
	}
}

func (d *Database) GetSecretsByField(field, name string) ([]secret.Secret, error) {
	var res []secret.Secret

	queryResult, err := d.FindDoc(SecretColl, `[{"eq": "`+name+`", "in": ["`+field+`"]}]`)
	if err != nil {
		return res, err
	}

	for docid := range queryResult {

		u, err := d.GetSecret(docid)
		u.ID = docid
		if err != nil {
			return res, err
		}
		res = append(res, u)
	}
	return res, nil
}

func (d *Database) GetSecretsByUserID(id string) ([]secret.Secret, error) {
	return d.GetSecretsByField("owner_id", id)
}

func (d *Database) GetSecretsByName(name string) ([]secret.Secret, error) {
	return d.GetSecretsByField("name", name)
}

func (d *Database) GetSecret(docID string) (secret.Secret, error) {
	doc, err := d.GetDoc(SecretColl, docID)
	if err != nil {
		return secret.Secret{}, err
	}
	t := secret.NewSecretFromMap(doc)
	t.ID = docID
	return t, err
}

func (d *Database) ListSecrets() []dbcommon.DocItem {
	return d.ListDocs(SecretColl)
}

// TODO: Change it, expensive for now
func (d *Database) CountSecrets() int {
	return len(d.ListSecrets())
}

func (d *Database) AllSecrets() []secret.Secret {
	Secrets := d.DB().Use(SecretColl)
	Secrets_id := make([]secret.Secret, 0)

	Secrets.ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		t := secret.NewSecretFromJson(docContent)
		t.ID = strconv.Itoa(id)
		Secrets_id = append(Secrets_id, t)
		return true
	})
	return Secrets_id
}
