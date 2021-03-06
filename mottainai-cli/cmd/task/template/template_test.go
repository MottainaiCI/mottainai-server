// Copyright © 2019 Ettore Di Giacinto <mudler@gentoo.org>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package template_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/MottainaiCI/mottainai-server/mottainai-cli/cmd/task/template"
)

func TestTemplates(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Template tests")
}

var _ = Describe("Template", func() {

	Describe("Draw", func() {
		Context("Using a simple template", func() {
			It("renders a specfile", func() {
				raw := `{{.EmailFrom}}`
				t := New()
				t.Values["EmailFrom"] = "test"
				res, err := t.Draw(raw)
				Expect(err).ToNot(HaveOccurred())
				Expect(res).To(Equal("test"))
			})
		})
	})

	Describe("LoadValues", func() {
		Context("Using a simple template", func() {
			It("renders a specfile", func() {
				raw := `
values:
  image: 1

`
				t := New()
				err := t.LoadValues(raw)
				Expect(err).ToNot(HaveOccurred())
				Expect(t.Values["image"]).To(Equal(1))
			})
		})
	})

	Describe("LoadArray", func() {
		Context("Load array values", func() {
			It("renders a specfile", func() {
				raw := `
values:
  images:
    - "image1"
    - "image2"

`

				t := New()
				err := t.LoadValues(raw)
				Expect(err).ToNot(HaveOccurred())
				Expect(t.Values["images"]).To(Equal([]interface{}{"image1", "image2"}))
			})
		})
	})

})
