/*

Copyright (C) 2019  Ettore Di Giacinto <mudler@gentoo.org>

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
package schema

import (
	"bytes"
	"encoding/gob"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/MottainaiCI/mottainai-server/pkg/utils"
)

type Request struct {
	Route   Route
	Options map[string]interface{}
	Target  interface{}
	Body    io.Reader
}

func (req *Request) NewAPIHTTPRequest(endpoint string) (*http.Request, error) {
	interpolations := req.Options
	relaxedInterpolations := req.Route.RemoveInterpolations(req.Options)
	req.Options = relaxedInterpolations
	httpRequest, err := req.Route.NewAPIRequest(endpoint, interpolations, req.Body)
	if err != nil {
		return nil, err
	}

	var InterfaceList []interface{}
	var Strings []string
	var String string

	if req.Route.RequireFormEncode() && req.Body == nil {
		form := url.Values{}

		for k, v := range req.Options {
			if reflect.TypeOf(v) == reflect.TypeOf(InterfaceList) {
				for _, el := range v.([]interface{}) {
					form.Add(k, el.(string))
				}
			} else if reflect.TypeOf(v) == reflect.TypeOf(Strings) {
				for _, el := range v.([]string) {
					form.Add(k, el)
				}

			} else if reflect.TypeOf(v) == reflect.TypeOf(float64(0)) {
				form.Add(k, utils.FloatToString(v.(float64)))

			} else if reflect.TypeOf(v) == reflect.TypeOf(String) {
				form.Add(k, v.(string))
			} else {
				var b bytes.Buffer
				e := gob.NewEncoder(&b)
				if err := e.Encode(v); err != nil {
					return nil, err
				}
				form.Add(k, b.String())
			}
		}

		httpRequest, err = req.Route.NewAPIRequest(endpoint, interpolations, strings.NewReader(form.Encode()))
	} else {
		q := httpRequest.URL.Query()
		for k, v := range req.Options {
			if reflect.TypeOf(v) == reflect.TypeOf(InterfaceList) {
				for _, el := range v.([]interface{}) {
					q.Add(k, el.(string))
				}
			} else if reflect.TypeOf(v) == reflect.TypeOf(Strings) {
				for _, el := range v.([]string) {
					q.Add(k, el)
				}

			} else if reflect.TypeOf(v) == reflect.TypeOf(float64(0)) {
				q.Add(k, utils.FloatToString(v.(float64)))

			} else if reflect.TypeOf(v) == reflect.TypeOf(String) {
				q.Add(k, v.(string))
			} else {
				var b bytes.Buffer
				e := gob.NewEncoder(&b)
				if err := e.Encode(v); err != nil {
					return nil, err
				}
				q.Add(k, b.String())
			}
		}
		httpRequest.URL.RawQuery = q.Encode()
	}
	return httpRequest, nil
}
