/*

Copyright (C) 2017-2018  Ettore Di Giacinto <mudler@gentoo.org>
Some code portions and re-implemented design are also coming
from the Gogs project, which is using the go-macaron framework and was
really source of ispiration. Kudos to them!

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

package client

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/MottainaiCI/mottainai-server/pkg/settings"
)

type Fetcher struct {
	BaseURL string
	docID   string
}

func NewClient() *Fetcher {
	var f Fetcher
	f.BaseURL = setting.Configuration.AppURL
	return &f
}

func NewFetcher(docID string) *Fetcher {
	f := NewClient()
	f.docID = docID
	return f
}

func (f *Fetcher) GetOptions(url string, option map[string]string) ([]byte, error) {
	hclient := &http.Client{}
	request, err := http.NewRequest("GET", f.BaseURL+url, nil)
	if err != nil {
		return []byte{}, err
	}

	q := request.URL.Query()
	for k, v := range option {
		q.Add(k, v)
	}
	request.URL.RawQuery = q.Encode()
	if err != nil {
		return []byte{}, err
	}

	response, err := hclient.Do(request)
	if err != nil {
		return []byte{}, err
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	return contents, err
}

func (f *Fetcher) PostOptions(URL string, option map[string]string) ([]byte, error) {
	hclient := &http.Client{}

	form := url.Values{}
	for k, v := range option {
		form.Add(k, v)
	}

	request, err := http.NewRequest("POST", f.BaseURL+URL, strings.NewReader(form.Encode()))
	if err != nil {
		return []byte{}, err
	}

	//req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// q := request.URL.Query()
	// for k, v := range option {
	// 	q.Add(k, v)
	// }
	// request.URL.RawQuery = q.Encode()
	// if err != nil {
	// 	return []byte{}, err
	// }

	response, err := hclient.Do(request)
	if err != nil {
		return []byte{}, err
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	return contents, err
}
