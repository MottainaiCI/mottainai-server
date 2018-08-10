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

	token "github.com/MottainaiCI/mottainai-server/pkg/token"
)

var TokenColl = "Tokens"

func (d *Database) IndexToken() {
	d.AddIndex(TokenColl, []string{"key"})
	d.AddIndex(TokenColl, []string{"user_id"})
}

func (d *Database) InsertToken(t *token.Token) (int, error) {
	return d.CreateToken(t.ToMap())
}

func (d *Database) CreateToken(t map[string]interface{}) (int, error) {
	return d.InsertDoc(TokenColl, t)
}

func (d *Database) DeleteToken(docID int) error {
	t, err := d.GetToken(docID)
	if err != nil {
		return err
	}

	t.Clear()
	return d.DeleteDoc(TokenColl, docID)
}

func (d *Database) UpdateToken(docID int, t map[string]interface{}) error {
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

func (d *Database) GetTokenByUserID(id int) (token.Token, error) {
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

	queryResult, err := d.FindDoc(TokenColl, `[{"eq": "`+name+`", "in": ["`+field+`"]}]`)
	if err != nil {
		return res, err
	}

	for docid := range queryResult {

		u, err := d.GetToken(docid)
		u.ID = docid
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

func (d *Database) GetTokensByUserID(id int) ([]token.Token, error) {
	return d.GetTokensByField("user_id", strconv.Itoa(id))
}

func (d *Database) GetToken(docID int) (token.Token, error) {
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
	Tokens := d.DB().Use(TokenColl)
	Tokens_id := make([]token.Token, 0)

	Tokens.ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		t := token.NewTokenFromJson(docContent)
		t.ID = id
		Tokens_id = append(Tokens_id, t)
		return true
	})
	return Tokens_id
}
