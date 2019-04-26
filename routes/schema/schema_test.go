/*

Copyright (C) 2019  Ettore Di Giacinto <mudler@gentoo.org>
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
package schema_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/MottainaiCI/mottainai-server/routes/schema"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	macaron "gopkg.in/macaron.v1"
)

var _ = Describe("RouteGenerator", func() {

	Context("Task routes", func() {
		var Schema RouteGenerator = APIRouteGenerator{
			Task: map[string]Route{
				"test2": Route{Path: "/foo/bar/:baz", Type: "get"},
				"test":  Route{Path: "/foo/bar/", Type: "get"},
			},
		}
		m := macaron.Classic()

		It("resolves correctly", func() {
			Expect(Schema.GetTaskRoute("test").Path).To(Equal("/foo/bar/"))
			Expect(Schema.GetTaskRoute("test").Type).To(Equal("get"))
			Expect(func() { var _ string = Schema.GetTaskRoute("ff").Path }).To(Panic())

		})
		It("successfully writes a GET route to macaron", func() {
			result := ""
			result2 := ""
			Schema.GetTaskRoute("test").ToMacaron(m, func() { result = "bat" }, func() { result2 = "baz" })
			resp := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/foo/bar/", nil)
			Expect(err).ToNot(HaveOccurred())
			m.ServeHTTP(resp, req)
			Expect(result).To(Equal("bat"))
			Expect(result2).To(Equal("baz"))
			Expect(resp.Code).To(Equal(http.StatusOK))
		})

		It("successfully interpolates parameters", func() {
			Expect(Schema.GetTaskRoute("test2").InterpolatePath(map[string]string{":baz": "test"})).To(Equal("/foo/bar/test"))
		})
	})

})
