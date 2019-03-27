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
package agenttasks_test

import (
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/MottainaiCI/mottainai-server/pkg/tasks"
)

var _ = Describe("Context", func() {
	ctx := NewExecutorContext()
	ctx.RootTaskDir = "a"
	ctx.TaskRelativeDir = "b"

	Describe("Context Structure", func() {
		Context("When container paths are absolute", func() {
			It("keeps them", func() {
				Expect(ctx.ContainerPath("/foo")).Should(Equal("/foo"))
			})
		})
		Context("When container paths are relative", func() {
			It("resolves them to the git folder", func() {
				Expect(ctx.ContainerPath("foo")).Should(Equal(path.Join(ctx.RootTaskDir, ctx.TaskRelativeDir, "foo")))
			})
		})

		Context("When host paths are absolute", func() {
			It("resolves them to the git folder", func() {
				Expect(ctx.HostPath("/foo")).Should(Equal(path.Join(ctx.RootTaskDir, "foo")))
			})
		})
		Context("When host paths are relative", func() {
			It("resolves them to the git folder", func() {
				Expect(ctx.HostPath("foo")).Should(Equal(path.Join(ctx.RootTaskDir, "foo")))
			})
		})
	})

})
