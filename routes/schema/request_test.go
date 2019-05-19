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
	. "github.com/MottainaiCI/mottainai-server/routes/schema"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Request", func() {

	Context("Task routes", func() {
		var Schema RouteGenerator = &APIRouteGenerator{
			Task: map[string]Route{

				"test": &APIRoute{Path: "/foo/bar2/:baz", Type: "post"},
			},
		}

		It("successfully generates interpolated http requests", func() {
			r := &Request{Route: Schema.GetTaskRoute("test"), Options: map[string]interface{}{"baz": "test", "barbarbar": "ok"}}
			req, err := r.NewAPIHTTPRequest("http://example.com")
			Expect(err).ToNot(HaveOccurred())
			Expect(req.Method).To(Equal("POST"))
			Expect(req.URL.Path).To(Equal("/foo/bar2/test"))
			Expect(req.Host).To(Equal("example.com"))
			Expect(r.Options).To(Equal(map[string]interface{}{"barbarbar": "ok"}))
			Expect(req.PostFormValue("barbarbar")).To(Equal("ok"))
		})
	})

})
