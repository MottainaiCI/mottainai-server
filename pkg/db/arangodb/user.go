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

	user "github.com/MottainaiCI/mottainai-server/pkg/user"
)

var UserColl = "Users"

func (d *Database) IndexUser() {
	d.AddIndex(UserColl, []string{"name"})
	d.AddIndex(UserColl, []string{"email"})
	d.AddIndex(UserColl, []string{"ownerid"}) // TODO: Display MY tasks and public (global) if admin0
	d.AddIndex(UserColl, []string{"is_admin"})
	d.AddIndex(UserColl, []string{"is_manager"})
}

func (d *Database) InsertAndSaltUser(t *user.User) (string, error) {
	if err := t.SaltPassword(); err != nil {
		return "", err
	}

	return d.InsertUser(t)
}

func (d *Database) InsertUser(t *user.User) (string, error) {
	if len(t.Name) == 0 || len(t.Password) == 0 {
		return "", errors.New("No username or password for user")
	}
	if u, e := d.GetUserByName(t.Name); e == nil && len(u.Name) != 0 {
		return "", errors.New("User already exists")
	}

	if u, e := d.GetUserByEmail(t.Email); e == nil && len(u.Name) != 0 {
		return "", errors.New("E-mail belongs to another account")
	}

	return d.CreateUser(t.ToMap())
}

func (d *Database) CreateUser(t map[string]interface{}) (string, error) {

	return d.InsertDoc(UserColl, t)
}

func (d *Database) DeleteUser(docID string) error {

	t, err := d.GetUser(docID)
	if err != nil {
		return err
	}

	t.Clear()
	return d.DeleteDoc(UserColl, docID)
}

func (d *Database) UpdateUser(docID string, t map[string]interface{}) error {
	return d.UpdateDoc(UserColl, docID, t)
}

func (d *Database) SignIn(name, password string) (user.User, error) {
	res, err := d.GetUsersByName(name)
	if err != nil {
		return user.User{}, err
	}
	if len(res) == 0 {
		return user.User{}, errors.New("Wrong username or password")
	}

	u := res[0]
	if ok, newSalt := u.VerifyPassword(password); ok {
		if len(newSalt) > 0 {
			err := d.UpdateUser(u.ID, map[string]interface{}{"password": newSalt})
			if err != nil {
				return u, errors.New("Error while updating salt:" + newSalt)
			}
		}
		return u, nil
	} else {
		return user.User{}, errors.New("Wrong username or password")
	}
}

func (d *Database) GetUsersByField(field, name string) ([]user.User, error) {
	var res []user.User

	queryResult, err := d.FindDoc("", `FOR c IN `+UserColl+`
		FILTER c.`+field+` == "`+name+`"
		RETURN c`)
	if err != nil {
		return res, err
	}

	// Query result are document IDs
	for id, _ := range queryResult {

		// Read document
		u, err := d.GetUser(id)
		if err != nil {
			return res, err
		}
		res = append(res, u)
	}
	return res, nil
}

func (d *Database) GetUserByName(name string) (user.User, error) {
	res, err := d.GetUsersByName(name)
	if err != nil {
		return user.User{}, err
	} else if len(res) == 0 {
		return user.User{}, errors.New("No username found")
	} else {
		return res[0], nil
	}
}

func (d *Database) GetUserByEmail(email string) (user.User, error) {
	res, err := d.GetUsersByEmail(email)
	if err != nil {
		return user.User{}, err
	} else if len(res) == 0 {
		return user.User{}, errors.New("No username found")
	} else {
		return res[0], nil
	}
}

func (d *Database) GetUsersByEmail(email string) ([]user.User, error) {
	return d.GetUsersByField("email", email)
}

func (d *Database) GetUsersByName(name string) ([]user.User, error) {
	return d.GetUsersByField("name", name)
}

// TODO: To replace with a specific collection to index search
func (d *Database) GetUserByIdentity(identity_type, id string) (user.User, error) {
	all := d.AllUsers()
	var res []user.User
	for _, u := range all {
		if i, ok := u.Identities[identity_type]; ok && i.ID == id {
			res = append(res, u)
		}
	}
	if len(res) > 1 {
		return user.User{}, errors.New("More than one user match with same id")
	}
	if len(res) == 0 {
		return user.User{}, errors.New("No user id found")
	}
	return res[0], nil
}

func (d *Database) GetUser(docID string) (user.User, error) {
	doc, err := d.GetDoc(UserColl, docID)
	if err != nil {
		return user.User{}, err
	}
	t := user.NewUserFromMap(doc)
	t.ID = docID
	return t, err
}

func (d *Database) ListUsers() []dbcommon.DocItem {
	return d.ListDocs(UserColl)
}

// TODO: Change it, expensive for now
func (d *Database) CountUsers() int {
	return len(d.ListUsers())
}

func (d *Database) AllUsers() []user.User {
	Users_id := make([]user.User, 0)

	docs, err := d.FindDoc("", "FOR c IN "+UserColl+" return c")
	if err != nil {
		return Users_id
	}

	for k, _ := range docs {
		t, err := d.GetUser(k)
		if err != nil {
			return Users_id
		}
		Users_id = append(Users_id, t)
	}

	return Users_id
}
