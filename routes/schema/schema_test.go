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
	"bytes"
	"net/http"
	"net/http/httptest"

	. "github.com/MottainaiCI/mottainai-server/routes/schema"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	macaron "gopkg.in/macaron.v1"
)

var _ = Describe("RouteGenerator", func() {

	Context("Task routes", func() {
		var Schema RouteGenerator = &APIRouteGenerator{
			Task: map[string]Route{
				"test2": &APIRoute{Path: "/foo/bar/:baz", Type: "get"},
				"test4": &APIRoute{Path: "/foo/bar/:baz/:baz.log", Type: "get"},
				"test":  &APIRoute{Path: "/foo/bar/", Type: "get"},
				"test3": &APIRoute{Path: "/foo/bar2/", Type: "post"},
			},
		}
		m := macaron.Classic()

		It("resolves correctly", func() {
			Expect(Schema.GetTaskRoute("test").GetPath()).To(Equal("/foo/bar/"))
			Expect(Schema.GetTaskRoute("test").GetType()).To(Equal("get"))
			Expect(Schema.GetTaskRoute("test").RequireFormEncode()).To(Equal(false))
			Expect(Schema.GetTaskRoute("test3").RequireFormEncode()).To(Equal(true))
			Expect(func() { var _ string = Schema.GetTaskRoute("ff").GetPath() }).To(Panic())
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
			Expect(Schema.GetTaskRoute("test2").InterpolatePath(map[string]interface{}{":baz": "test"})).To(Equal("/foo/bar/test"))
			Expect(Schema.GetTaskRoute("test4").InterpolatePath(map[string]interface{}{":baz": "test"})).To(Equal("/foo/bar/test/test.log"))

		})
		It("successfully interpolates parameters", func() {
			Expect(Schema.GetTaskRoute("test2").InterpolatePath(map[string]interface{}{"baz": "test"})).To(Equal("/foo/bar/test"))
			Expect(Schema.GetTaskRoute("test4").InterpolatePath(map[string]interface{}{"baz": "test"})).To(Equal("/foo/bar/test/test.log"))

		})
		It("successfully removes interpolation", func() {
			Expect(Schema.GetTaskRoute("test2").RemoveInterpolations(map[string]interface{}{"baz": "test", "foo": "bar"})).To(Equal(map[string]interface{}{"foo": "bar"}))
			Expect(Schema.GetTaskRoute("test4").RemoveInterpolations(map[string]interface{}{"baz": "test", "foo": "bar"})).To(Equal(map[string]interface{}{"foo": "bar"}))
		})

		It("successfully generates interpolated http requests", func() {
			body := new(bytes.Buffer)
			var req *http.Request
			req, err := Schema.GetTaskRoute("test2").NewRequest("http://example.com", map[string]string{":baz": "test"}, body)
			Expect(err).ToNot(HaveOccurred())
			Expect(req.Method).To(Equal("GET"))
			Expect(req.URL.Path).To(Equal("/foo/bar/test"))
			Expect(req.Host).To(Equal("example.com"))
		})
	})

})
