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

package webhook

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/sethvargo/go-password/password"

	"reflect"
)

type WebHook struct {
	ID      string `json:"id" form:"id"`
	Key     string `json:"key" form:"key"`
	Type    string `json:"type" form:"type"`
	URL     string `json:"url" form:"url"`
	OwnerId string `json:"owner_id" form:"owner_id"`
}

func GenerateUserWebHook(id string) (*WebHook, error) {
	t, err := GenerateWebHook()
	if err != nil {
		return t, err
	}
	t.OwnerId = id

	return t, nil
}

func Gen() (string, error) {
	res, err := password.Generate(64, 10, 10, false, false)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	h.Write([]byte(res))

	return hex.EncodeToString(h.Sum(nil)), nil
}
func GenerateWebHook() (*WebHook, error) {
	t := NewWebHook()
	res, err := Gen()
	if err != nil {
		return t, err
	}
	t.Key = res

	res, err = Gen()
	if err != nil {
		return t, err
	}
	t.URL = res

	return t, nil
}

func NewWebHook() *WebHook {
	return &WebHook{}
}

// TODO: Port NewWebHookFromMap Task to same or make it common func
func NewWebHookFromMap(t map[string]interface{}) WebHook {
	u := &WebHook{}
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

func NewWebHookFromJson(data []byte) WebHook {
	var t WebHook
	json.Unmarshal(data, &t)
	return t
}

func (t *WebHook) Clear() {
}

func (t *WebHook) ToMap() map[string]interface{} {

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
