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
	"errors"

	dbcommon "github.com/MottainaiCI/mottainai-server/pkg/db/common"

	token "github.com/MottainaiCI/mottainai-server/pkg/token"
)

var TokenColl = "Tokens"

func (d *Database) IndexToken() {
	d.AddIndex(TokenColl, []string{"key"})
	d.AddIndex(TokenColl, []string{"user_id"})
}
func (d *Database) InsertToken(t *token.Token) (string, error) {
	return d.CreateToken(t.ToMap())
}

func (d *Database) CreateToken(t map[string]interface{}) (string, error) {
	return d.InsertDoc(TokenColl, t)
}

func (d *Database) DeleteToken(docID string) error {
	t, err := d.GetToken(docID)
	if err != nil {
		return err
	}

	t.Clear()
	return d.DeleteDoc(TokenColl, docID)
}

func (d *Database) UpdateToken(docID string, t map[string]interface{}) error {
	return d.UpdateDoc(TokenColl, docID, t)
}

func (d *Database) GetTokenByKey(name string) (token.Token, error) {
	res, err := d.GetTokensByKey(name)
	if err != nil {
		return token.Token{}, err
	} else if len(res) == 0 {
		return token.Token{}, errors.New("No tokenname found")
	} else {
		return res[0], nil
	}
}

func (d *Database) GetTokenByUserID(id string) (token.Token, error) {
	res, err := d.GetTokensByUserID(id)
	if err != nil {
		return token.Token{}, err
	} else if len(res) == 0 {
		return token.Token{}, errors.New("No tokenname found")
	} else {
		return res[0], nil
	}
}

func (d *Database) GetTokensByField(field, name string) ([]token.Token, error) {

	var res []token.Token

	queryResult, err := d.FindDoc("", `FOR c IN `+TokenColl+`
		FILTER c.`+field+` == "`+name+`"
		RETURN c`)
	if err != nil {
		return res, err
	}

	// Query result are document IDs
	for id, _ := range queryResult {

		// Read document
		u, err := d.GetToken(id)
		if err != nil {
			return res, err
		}
		res = append(res, u)
	}
	return res, nil

}

func (d *Database) GetTokensByKey(name string) ([]token.Token, error) {
	return d.GetTokensByField("key", name)
}

func (d *Database) GetTokensByUserID(id string) ([]token.Token, error) {
	return d.GetTokensByField("user_id", id)
}

func (d *Database) GetToken(docID string) (token.Token, error) {
	doc, err := d.GetDoc(TokenColl, docID)
	if err != nil {
		return token.Token{}, err
	}
	t := token.NewTokenFromMap(doc)
	t.ID = docID
	return t, err
}

func (d *Database) ListTokens() []dbcommon.DocItem {
	return d.ListDocs(TokenColl)
}

// TODO: Change it, expensive for now
func (d *Database) CountTokens() int {
	return len(d.ListTokens())
}

func (d *Database) AllTokens() []token.Token {
	Tokens_id := make([]token.Token, 0)

	docs, err := d.FindDoc("", "FOR c IN "+TokenColl+" return c")
	if err != nil {
		return Tokens_id
	}

	for k, doc := range docs {
		t := token.NewTokenFromMap(doc.(map[string]interface{}))
		t.ID = k
		Tokens_id = append(Tokens_id, t)
	}

	return Tokens_id
}
