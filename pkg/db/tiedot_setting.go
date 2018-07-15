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
	"strconv"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
)

var SettingColl = "Settings"

func (d *Database) IndexSetting() {
	d.AddIndex(SettingColl, []string{"key"})
	d.AddIndex(SettingColl, []string{"value"})
}

func (d *Database) InsertSetting(t *setting.Setting) (int, error) {
	if _, e := d.GetSettingByKey(t.Key); e == nil {
		return 0, errors.New("Setting already exist")
	}

	return d.CreateSetting(t.ToMap())
}

func (d *Database) CreateSetting(t map[string]interface{}) (int, error) {
	return d.InsertDoc(SettingColl, t)
}

func (d *Database) DeleteSetting(docID int) error {
	t, err := d.GetSetting(docID)
	if err != nil {
		return err
	}

	t.Clear()
	return d.DeleteDoc(SettingColl, docID)
}

func (d *Database) UpdateSetting(docID int, t map[string]interface{}) error {
	return d.UpdateDoc(SettingColl, docID, t)
}

func (d *Database) GetSettingByKey(name string) (setting.Setting, error) {
	res, err := d.GetSettingsByKey(name)
	if err != nil {
		return setting.Setting{}, err
	} else if len(res) == 0 {
		return setting.Setting{}, errors.New("No settingname found")
	} else {
		return res[0], nil
	}
}

func (d *Database) GetSettingByUserID(id int) (setting.Setting, error) {
	res, err := d.GetSettingsByUserID(id)
	if err != nil {
		return setting.Setting{}, err
	} else if len(res) == 0 {
		return setting.Setting{}, errors.New("No settingname found")
	} else {
		return res[0], nil
	}
}

func (d *Database) GetSettingsByField(field, name string) ([]setting.Setting, error) {
	var res []setting.Setting

	queryResult, err := d.FindDoc(SettingColl, `[{"eq": "`+name+`", "in": ["`+field+`"]}]`)
	if err != nil {
		return res, err
	}

	for docid := range queryResult {

		u, err := d.GetSetting(docid)
		u.ID = docid
		if err != nil {
			return res, err
		}
		res = append(res, u)
	}
	return res, nil
}

func (d *Database) GetSettingsByKey(name string) ([]setting.Setting, error) {
	return d.GetSettingsByField("key", name)
}

func (d *Database) GetSettingsByUserID(id int) ([]setting.Setting, error) {
	return d.GetSettingsByField("user_id", strconv.Itoa(id))
}

func (d *Database) GetSetting(docID int) (setting.Setting, error) {
	doc, err := d.GetDoc(SettingColl, docID)
	if err != nil {
		return setting.Setting{}, err
	}
	t := setting.NewSettingFromMap(doc)
	t.ID = docID
	return t, err
}

func (d *Database) ListSettings() []DocItem {
	return d.ListDocs(SettingColl)
}

// TODO: Change it, expensive for now
func (d *Database) CountSettings() int {
	return len(d.ListSettings())
}

func (d *Database) AllSettings() []setting.Setting {
	Settings := d.DB().Use(SettingColl)
	Settings_id := make([]setting.Setting, 0)

	Settings.ForEachDoc(func(id int, docContent []byte) (willMoveOn bool) {
		t := setting.NewSettingFromJson(docContent)
		t.ID = id
		Settings_id = append(Settings_id, t)
		return true
	})
	return Settings_id
}
