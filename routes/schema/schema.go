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
	"strings"

	macaron "gopkg.in/macaron.v1"
)

type RouteGenerator interface {
	GetTaskRoute(s string) *Route
	GetNodeRoute(s string) *Route
	GetWebHookRoute(s string) *Route
	GetNamespaceRoute(s string) *Route
	GetUserRoute(s string) *Route
	GetTokenRoute(s string) *Route
	GetStorageRoute(s string) *Route
	GetStatsRoute(s string) *Route
	GetSettingRoute(s string) *Route
}

type APIRouteGenerator struct {
	Task      map[string]Route
	Node      map[string]Route
	WebHook   map[string]Route
	Namespace map[string]Route
	User      map[string]Route
	Token     map[string]Route
	Storage   map[string]Route
	Stats     map[string]Route
	Setting   map[string]Route
}

func (g APIRouteGenerator) GetTaskRoute(s string) *Route {
	r, ok := g.Task[s]
	if ok {
		return &r
	}

	return nil
}

func (g APIRouteGenerator) GetNodeRoute(s string) *Route {
	r, ok := g.Node[s]
	if ok {
		return &r
	}

	return nil
}

func (g APIRouteGenerator) GetWebHookRoute(s string) *Route {
	r, ok := g.WebHook[s]
	if ok {
		return &r
	}

	return nil
}
func (g APIRouteGenerator) GetNamespaceRoute(s string) *Route {
	r, ok := g.Namespace[s]
	if ok {
		return &r
	}

	return nil
}
func (g APIRouteGenerator) GetUserRoute(s string) *Route {
	r, ok := g.User[s]
	if ok {
		return &r
	}

	return nil
}
func (g APIRouteGenerator) GetTokenRoute(s string) *Route {
	r, ok := g.Token[s]
	if ok {
		return &r
	}

	return nil
}
func (g APIRouteGenerator) GetStorageRoute(s string) *Route {
	r, ok := g.Storage[s]
	if ok {
		return &r
	}

	return nil
}
func (g APIRouteGenerator) GetStatsRoute(s string) *Route {
	r, ok := g.Stats[s]
	if ok {
		return &r
	}

	return nil
}
func (g APIRouteGenerator) GetSettingRoute(s string) *Route {
	r, ok := g.Setting[s]
	if ok {
		return &r
	}

	return nil
}

type Route struct {
	Path string
	Type string
}

func (r Route) InterpolatePath(opts map[string]string) string {
	res := r.Path
	for k, v := range opts {
		res = strings.Replace(res, k, v, 1)
	}
	return res
}

func (r Route) ToMacaron(m *macaron.Macaron, v ...macaron.Handler) {
	switch r.Type {
	case "get":
		m.Get(r.Path, v...)
	case "post":
		m.Post(r.Path, v...)
	case "patch":
		m.Patch(r.Path, v...)
	case "put":
		m.Put(r.Path, v...)
	case "delete":
		m.Delete(r.Path, v...)
	case "any":
		m.Any(r.Path, v...)
	case "options":
		m.Options(r.Path, v...)
	}
}
