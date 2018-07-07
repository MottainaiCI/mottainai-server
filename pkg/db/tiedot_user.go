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

func (d *Database) InsertAndSaltUser(t *user.User) (int, error) {
	if err := t.SaltPassword(); err != nil {
		return 0, err
	}

	return d.InsertUser(t)
}

func (d *Database) InsertUser(t *user.User) (int, error) {
	return d.CreateUser(t.ToMap())
}

func (d *Database) CreateUser(t map[string]interface{}) (int, error) {

	return d.InsertDoc(UserColl, t)
}

func (d *Database) DeleteUser(docID int) error {

	t, err := d.GetUser(docID)
	if err != nil {
		return err
	}

	t.Clear()
	return d.DeleteDoc(UserColl, docID)
}

func (d *Database) UpdateUser(docID int, t map[string]interface{}) error {
	return d.UpdateDoc(UserColl, docID, t)
}

func (d *Database) SignIn(name, password string) (user.User, error) {
	res, err := d.GetUsersByName(name)
	if err != nil {
		return user.User{}, err
	}
	if len(res) == 0 {
		return user.User{}, errors.New("No username found")
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

func (d *Database) GetUsersByName(name string) ([]user.User, error) {
	var res []user.User

	queryResult, err := d.FindDoc(UserColl, `[{"eq": "`+name+`", "in": ["name"]}]`)
	if err != nil {
		return res, err
	}

	for docid := range queryResult {

		u, err := d.GetUser(docid)
		u.ID = docid
		if err != nil {
			return res, err
		}
		res = append(res, u)
	}
	return res, nil
}

func (d *Database) GetUser(docID int) (user.User, error) {
	doc, err := d.GetDoc(UserColl, docID)
	if err != nil {
		return user.User{}, err
	}
	t := user.NewUserFromMap(doc)
	t.ID = docID
	return t, err
}

func (d *Database) ListUsers() []DocItem {
	return d.ListDocs(UserColl)
}

// TODO: Change it, expensive for now
func (d *Database) CountUsers() int {
	return len(d.ListUsers())
}

func (d *Database) AllUsers() []user.User {
	Users := d.DB().Use(UserColl)
	Users_id := make([]user.User, 0)

	Users.ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		t := user.NewUserFromJson(docContent)
		t.ID = id
		Users_id = append(Users_id, t)
		return true
	})
	return Users_id
}
