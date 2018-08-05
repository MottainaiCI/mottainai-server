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

package user

import (
	"encoding/json"
	"reflect"

	passlib "github.com/MottainaiCI/passlib"
)

type Identity struct {
	ID        string `json:"identity_id" form:"identity_id"`
	Provider  string `json:"provider" form:"provider"`
	AvatarURL string `json:"avatar_url" form:"avatar_url"`
}

type User struct {
	ID   int    `json:"id" form:"id"`
	Name string `json:"name" form:"name"`

	Identities map[string]Identity `json:"identities" form:"identities"`
	// Auth
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
	Admin    string `json:"is_admin" form:"is_admin"`
	Manager  string `json:"is_manager" form:"is_manager"`
}

func (u *User) AddIdentity(t string, i *Identity) {
	if len(u.Identities) == 0 {
		u.Identities = make(map[string]Identity)
	}

	u.Identities[t] = *i
}

func (u *User) RemoveIdentity(t string) {
	if len(u.Identities) == 0 {
		u.Identities = make(map[string]Identity)
		return
	}
	delete(u.Identities, t)
}

func (u *User) IsManager() bool {
	if u.Manager == "yes" {
		return true
	}
	return false
}

func (u *User) MakeManager() {
	u.Manager = "yes"
}

func (u *User) RemoveManager() {
	u.Manager = "no"
}

func (u *User) IsAdmin() bool {
	if len(u.Admin) > 0 && u.Admin == "yes" {
		return true
	}
	return false
}

func (u *User) IsManagerOrAdmin() bool {
	if u.IsAdmin() || u.IsManager() {
		return true
	}
	return false
}

func (u *User) MakeAdmin() {
	u.Admin = "yes"
}

func (u *User) RemoveAdmin() {
	u.Admin = "no"
}

func (u *User) SaltPassword() error {
	hash, err := passlib.Hash(u.Password)
	if err != nil {
		// couldn't hash password for some reason
		return err
	}
	u.Password = hash
	return nil
}

func (u *User) VerifyPassword(pass string) (bool, string) {
	newHash, err := passlib.Verify(pass, u.Password)
	if err != nil {
		// incorrect password, malformed hash, etc.
		// either way, reject
		return false, ""
	}

	return true, newHash
}

// TODO: Port NewUserFromMap Task to same or make it common func
func NewUserFromMap(t map[string]interface{}) User {
	u := &User{}
	val := reflect.ValueOf(u).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		if typeField.Type.Name() == "string" {
			if str, ok := t[tag.Get("form")].(string); ok {
				valueField.SetString(str)
			}
		}
		if typeField.Type.Name() == "int64" {

			if i, ok := t[tag.Get("form")].(int64); ok {
				valueField.SetInt(i)
			}
		}

		if typeField.Type.Name() == "bool" {
			if b, ok := t[tag.Get("form")].(bool); ok {
				valueField.SetBool(b)
			}
		}

		if typeField.Type.Kind() == reflect.ValueOf(u.Identities).Kind() {
			if b, ok := t[tag.Get("form")].(map[string]interface{}); ok {

				m := make(map[string]Identity)
				for k, v := range b {
					//fmt.Println("TIPO", , k)
					m[k] = NewIdentityFromMap(v.(map[string]interface{}))
				}

				valueField.Set(reflect.ValueOf(m))
			} else if b, ok := t[tag.Get("form")].([]interface{}); ok {
				// convert all to string before set
				var r []string
				for _, f := range b {
					r = append(r, f.(string))
				}
				valueField.Set(reflect.ValueOf(r))
			}
		}
		//fmt.Printf("Field Name: %s,\t Field Value: %v,\t Tag Value: %s\n", typeField.Name, valueField.Interface(), tag.Get("tag_name"))
	}
	return *u
}

// TODO: Port NewUserFromMap Task to same or make it common func
func NewIdentityFromMap(t map[string]interface{}) Identity {
	u := &Identity{}
	val := reflect.ValueOf(u).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		if typeField.Type.Name() == "string" {
			if str, ok := t[tag.Get("form")].(string); ok {
				valueField.SetString(str)
			}
		}
		if typeField.Type.Name() == "int64" {

			if i, ok := t[tag.Get("form")].(int64); ok {
				valueField.SetInt(i)
			}
		}

		if typeField.Type.Name() == "bool" {
			if b, ok := t[tag.Get("form")].(bool); ok {
				valueField.SetBool(b)
			}
		}

		//fmt.Printf("Field Name: %s,\t Field Value: %v,\t Tag Value: %s\n", typeField.Name, valueField.Interface(), tag.Get("tag_name"))
	}
	return *u
}

func NewUserFromJson(data []byte) User {
	var t User
	json.Unmarshal(data, &t)
	return t
}

func (t *User) Clear() {
}

func (t *User) ToMap() map[string]interface{} {

	ts := make(map[string]interface{})
	val := reflect.ValueOf(t).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		tag := typeField.Tag

		ts[tag.Get("form")] = valueField.Interface()
		//fmt.Printf("Field Name: %s,\t Field Value: %v,\t Tag Value: %s\n", typeField.Name, valueField.Interface(), tag.Get("tag_name"))
	}
	return ts
}
