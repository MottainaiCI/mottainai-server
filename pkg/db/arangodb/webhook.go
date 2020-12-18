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

	webhook "github.com/MottainaiCI/mottainai-server/pkg/webhook"
)

var WebHookColl = "WebHooks"

func (d *Database) IndexWebHook() {
	d.AddIndex(WebHookColl, []string{"key"})
	d.AddIndex(WebHookColl, []string{"url"})
	d.AddIndex(WebHookColl, []string{"owner_id"})
}

func (d *Database) InsertWebHook(t *webhook.WebHook) (string, error) {
	return d.CreateWebHook(t.ToMap())
}

func (d *Database) CreateWebHook(t map[string]interface{}) (string, error) {
	return d.InsertDoc(WebHookColl, t)
}

func (d *Database) DeleteWebHook(docID string) error {
	t, err := d.GetWebHook(docID)
	if err != nil {
		return err
	}

	t.Clear()
	return d.DeleteDoc(WebHookColl, docID)
}

func (d *Database) UpdateWebHook(docID string, t map[string]interface{}) error {
	return d.UpdateDoc(WebHookColl, docID, t)
}

func (d *Database) GetWebHookByKey(name string) (webhook.WebHook, error) {
	res, err := d.GetWebHooksByKey(name)
	if err != nil {
		return webhook.WebHook{}, err
	} else if len(res) == 0 {
		return webhook.WebHook{}, errors.New("No webhookname found")
	} else {
		return res[0], nil
	}
}

func (d *Database) GetWebHookByURL(url string) (webhook.WebHook, error) {
	res, err := d.GetWebHooksByURL(url)
	if err != nil {
		return webhook.WebHook{}, err
	} else if len(res) == 0 {
		return webhook.WebHook{}, errors.New("No webhookname found")
	} else {
		return res[0], nil
	}
}

func (d *Database) GetWebHookByUserID(id string) (webhook.WebHook, error) {
	res, err := d.GetWebHooksByUserID(id)
	if err != nil {
		return webhook.WebHook{}, err
	} else if len(res) == 0 {
		return webhook.WebHook{}, errors.New("No webhookname found")
	} else {
		return res[0], nil
	}
}

func (d *Database) GetWebHooksByField(field, name string) ([]webhook.WebHook, error) {

	var res []webhook.WebHook

	queryResult, err := d.FindDoc("", `FOR c IN `+WebHookColl+`
		FILTER c.`+field+` == "`+name+`"
		RETURN c`)
	if err != nil {
		return res, err
	}

	// Query result are document IDs
	for id, doc := range queryResult {
		t := webhook.NewWebHookFromMap(doc.(map[string]interface{}))
		t.ID = id
		res = append(res, t)
	}
	return res, nil
}

func (d *Database) GetWebHooksByKey(name string) ([]webhook.WebHook, error) {
	return d.GetWebHooksByField("key", name)
}

func (d *Database) GetWebHooksByURL(name string) ([]webhook.WebHook, error) {
	return d.GetWebHooksByField("url", name)
}

func (d *Database) GetWebHooksByUserID(id string) ([]webhook.WebHook, error) {
	return d.GetWebHooksByField("owner_id", id)
}

func (d *Database) GetWebHook(docID string) (webhook.WebHook, error) {
	doc, err := d.GetDoc(WebHookColl, docID)
	if err != nil {
		return webhook.WebHook{}, err
	}
	t := webhook.NewWebHookFromMap(doc)
	t.ID = docID
	return t, err
}

func (d *Database) ListWebHooks() []dbcommon.DocItem {
	return d.ListDocs(WebHookColl)
}

// TODO: Change it, expensive for now
func (d *Database) CountWebHooks() int {
	return len(d.ListWebHooks())
}

func (d *Database) AllWebHooks() []webhook.WebHook {

	WebHooks_id := make([]webhook.WebHook, 0)

	docs, err := d.FindDoc("", "FOR c IN "+WebHookColl+" return c")
	if err != nil {
		return WebHooks_id
	}

	for k, doc := range docs {
		t := webhook.NewWebHookFromMap(doc.(map[string]interface{}))
		t.ID = k
		WebHooks_id = append(WebHooks_id, t)
	}

	return WebHooks_id
}
