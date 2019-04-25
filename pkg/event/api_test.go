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

package event_test

import (
	. "github.com/MottainaiCI/mottainai-server/pkg/event"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("APIResponse event", func() {

	Describe("Decoder", func() {
		Context("API Response byte payload", func() {
			It("Decodes it successfully", func() {

				resp, err := DecodeAPIResponse([]byte(`{ "type": "test", "event": "foo", "status": "nope" }`))
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.ObjType).To(Equal("test"))
			})
		})

		Context("Invaled API Response byte payload", func() {
			It("Fails", func() {

				_, err := DecodeAPIResponse([]byte(`:(`))
				Expect(err).To(HaveOccurred())
			})
		})
	})

})
