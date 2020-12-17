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
	"io"
	"net/http"
	"strings"

	macaron "gopkg.in/macaron.v1"
)

type RouteGenerator interface {
	GetTaskRoute(s string) Route
	GetNodeRoute(s string) Route
	GetWebHookRoute(s string) Route
	GetSecretRoute(s string) Route
	GetNamespaceRoute(s string) Route
	GetUserRoute(s string) Route
	GetTokenRoute(s string) Route
	GetStorageRoute(s string) Route
	GetStatsRoute(s string) Route
	GetSettingRoute(s string) Route
	GetClientRoute(s string) Route
}

type APIRouteGenerator struct {
	Task      map[string]Route
	Node      map[string]Route
	WebHook   map[string]Route
	Secret    map[string]Route
	Namespace map[string]Route
	User      map[string]Route
	Token     map[string]Route
	Storage   map[string]Route
	Stats     map[string]Route
	Setting   map[string]Route
	Client    map[string]Route
}

func (g *APIRouteGenerator) GetSecretRoute(s string) Route {
	r, ok := g.Secret[s]
	if ok {
		return r
	}

	return nil
}

func (g *APIRouteGenerator) GetTaskRoute(s string) Route {
	r, ok := g.Task[s]
	if ok {
		return r
	}

	return nil
}

func (g *APIRouteGenerator) GetNodeRoute(s string) Route {
	r, ok := g.Node[s]
	if ok {
		return r
	}

	return nil
}

func (g APIRouteGenerator) GetWebHookRoute(s string) Route {
	r, ok := g.WebHook[s]
	if ok {
		return r
	}

	return nil
}
func (g *APIRouteGenerator) GetNamespaceRoute(s string) Route {
	r, ok := g.Namespace[s]
	if ok {
		return r
	}

	return nil
}
func (g *APIRouteGenerator) GetUserRoute(s string) Route {
	r, ok := g.User[s]
	if ok {
		return r
	}

	return nil
}
func (g *APIRouteGenerator) GetTokenRoute(s string) Route {
	r, ok := g.Token[s]
	if ok {
		return r
	}

	return nil
}
func (g *APIRouteGenerator) GetStorageRoute(s string) Route {
	r, ok := g.Storage[s]
	if ok {
		return r
	}

	return nil
}
func (g *APIRouteGenerator) GetStatsRoute(s string) Route {
	r, ok := g.Stats[s]
	if ok {
		return r
	}

	return nil
}
func (g *APIRouteGenerator) GetSettingRoute(s string) Route {
	r, ok := g.Setting[s]
	if ok {
		return r
	}

	return nil
}
func (g *APIRouteGenerator) GetClientRoute(s string) Route {
	r, ok := g.Client[s]
	if ok {
		return r
	}

	return nil
}

type Route interface {
	InterpolatePath(map[string]interface{}) string
	NewRequest(string, map[string]string, io.Reader) (*http.Request, error)
	NewAPIRequest(string, map[string]interface{}, io.Reader) (*http.Request, error)

	ToMacaron(*macaron.Macaron, ...macaron.Handler)
	GetPath() string
	GetType() string
	RequireFormEncode() bool
	RemoveInterpolations(map[string]interface{}) map[string]interface{}
}
type APIRoute struct {
	Path string
	Type string
}

func (r *APIRoute) GetPath() string {
	return r.Path
}

func (r *APIRoute) GetType() string {
	return r.Type
}

func (r *APIRoute) InterpolatePath(opts map[string]interface{}) string {
	res := r.Path
	for k, v := range opts {
		rv, ok := v.(string)
		if !ok {
			continue
		}
		relaxedK := strings.TrimPrefix(k, ":")
		res = strings.Replace(res, ":"+relaxedK, rv, -1)
	}
	return res
}

func (r *APIRoute) RemoveInterpolations(opts map[string]interface{}) map[string]interface{} {

	res := make(map[string]interface{})
	for k, v := range opts {
		relaxedK := strings.TrimPrefix(k, ":")

		if !strings.Contains(r.Path, ":"+relaxedK) {
			res[k] = v
		}

	}
	return res
}

func (r *APIRoute) NewAPIRequest(baseURL string, interpolate map[string]interface{}, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(strings.ToUpper(r.GetType()), baseURL+r.InterpolatePath(interpolate), body)
	if err != nil {
		return req, err
	}
	if r.RequireFormEncode() {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	// t := strings.ToUpper(r.GetType())
	// switch t {
	// case "POST":
	// 	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// case "PUT":
	// 	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// case "PATCH":
	// 	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// }

	return req, nil
}

// DUP for now, it will be removed once we remove all the calls with Interpolations
func (r *APIRoute) NewRequest(baseURL string, interpolate map[string]string, body io.Reader) (*http.Request, error) {
	m := make(map[string]interface{})
	for k, v := range interpolate {
		m[k] = interface{}(v)
	}
	return r.NewAPIRequest(baseURL, m, body)
}

func (r *APIRoute) RequireFormEncode() bool {
	t := strings.ToUpper(r.GetType())
	switch t {
	case "POST":
		return true
	case "PUT":
		return true
	case "PATCH":
		return true
	default:
		return false
	}
}

func (r *APIRoute) ToMacaron(m *macaron.Macaron, v ...macaron.Handler) {
	t := strings.ToUpper(r.GetType())
	p := r.GetPath()
	switch t {
	case "GET":
		m.Get(p, v...)
	case "POST":
		m.Post(p, v...)
	case "PATCH":
		m.Patch(p, v...)
	case "PUT":
		m.Put(p, v...)
	case "DELETE":
		m.Delete(p, v...)
	case "ANY":
		m.Any(p, v...)
	case "OPTIONS":
		m.Options(p, v...)
	}
}
