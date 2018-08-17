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

package token

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/sethvargo/go-password/password"

	"reflect"
)

type Token struct {
	ID  string `json:"id" form:"id"`
	Key string `json:"key" form:"key"`

	UserId string `json:"user_id" form:"user_id"`
}

func GenerateUserToken(id string) (*Token, error) {
	t, err := GenerateToken()
	if err != nil {
		return t, err
	}
	t.UserId = id

	return t, nil
}

func GenerateToken() (*Token, error) {
	t := NewToken()
	res, err := password.Generate(64, 10, 10, false, false)
	if err != nil {
		return t, err
	}
	h := sha256.New()
	h.Write([]byte(res))
	t.Key = hex.EncodeToString(h.Sum(nil))

	return t, nil
}

func NewToken() *Token {
	return &Token{}
}

// TODO: Port NewTokenFromMap Task to same or make it common func
func NewTokenFromMap(t map[string]interface{}) Token {
	u := &Token{}
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
		if typeField.Type.Name() == "int" {

			if i, ok := t[tag.Get("form")].(int); ok {
				valueField.SetInt(int64(i))
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

func NewTokenFromJson(data []byte) Token {
	var t Token
	json.Unmarshal(data, &t)
	return t
}

func (t *Token) Clear() {
}

func (t *Token) ToMap() map[string]interface{} {

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
