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

	. "github.com/MottainaiCI/mottainai-server/pkg/tasks/executors"
)

var _ = Describe("Context", func() {

	ctx := NewExecutorContext()
	i := NewDefaultInstruction([]string{}, []string{})

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

		Context("When container paths are relative", func() {
			It("resolves them to the git folder", func() {
				Expect(ctx.ContainerPath("foo", "bar")).Should(Equal(path.Join(ctx.RootTaskDir, ctx.TaskRelativeDir, "foo", "bar")))
			})
		})

		Context("When host paths are absolute", func() {
			It("keeps them absolute", func() {
				Expect(ctx.ContainerPath("/foo", "bar")).Should(Equal(path.Join("/foo", "bar")))
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

		Context("When host paths are relative", func() {
			It("resolves them to the root task folder", func() {
				Expect(ctx.HostPath("foo", "bar")).Should(Equal(path.Join(ctx.RootTaskDir, "foo", "bar")))
			})
		})

	})

	BeforeEach(func() {
		ctx = NewExecutorContext()
		ctx.RootTaskDir = "/c"
		ctx.TaskRelativeDir = "d"
		ctx.ArtefactDir = "/a"
		ctx.StorageDir = "/b"
		i = NewDefaultInstruction([]string{}, []string{})
	})

	Describe("Context path resolution", func() {
		Context("With host mapping", func() {
			It("resolves the source repository mount", func() {
				mapping := ctx.ResolveArtefactsMounts(ArtefactMapping{}, i, true)

				Expect(i.MountsList()[0]).Should(Equal("/c/d/artefacts:/c/d/artefacts"))
				Expect(i.MountsList()[1]).Should(Equal("/c/d/storage:/c/d/storage"))
				Expect(mapping.ArtefactPath).Should(Equal("/c/d/artefacts"))
				Expect(mapping.StoragePath).Should(Equal("/c/d/storage"))
			})

			It("resolves the source repository mount with absolute path", func() {
				mapping := ctx.ResolveArtefactsMounts(ArtefactMapping{ArtefactPath: "/c", StoragePath: "/z"}, i, true)

				Expect(i.MountsList()[0]).Should(Equal("/c/d/c:/c"))
				Expect(i.MountsList()[1]).Should(Equal("/c/d/z:/z"))
				Expect(mapping.ArtefactPath).Should(Equal("/c/d/c"))
				Expect(mapping.StoragePath).Should(Equal("/c/d/z"))
			})

			It("resolves the source repository mount with relative path", func() {
				mapping := ctx.ResolveArtefactsMounts(ArtefactMapping{ArtefactPath: "c", StoragePath: "z"}, i, true)

				Expect(i.MountsList()[0]).Should(Equal("/c/d/c:/c/d/c"))
				Expect(i.MountsList()[1]).Should(Equal("/c/d/z:/c/d/z"))
				Expect(mapping.ArtefactPath).Should(Equal("/c/d/c"))
				Expect(mapping.StoragePath).Should(Equal("/c/d/z"))
			})
		})

		Context("Without host mapping", func() {
			It("resolves the source repository mount", func() {
				mapping := ctx.ResolveArtefactsMounts(ArtefactMapping{}, i, false)

				Expect(i.MountsList()[0]).Should(Equal("/a:/c/d/artefacts"))
				Expect(i.MountsList()[1]).Should(Equal("/b:/c/d/storage"))
				Expect(mapping.ArtefactPath).Should(Equal("/a"))
				Expect(mapping.StoragePath).Should(Equal("/b"))
			})

			It("resolves the source repository mount with custom paths (abs)", func() {
				mapping := ctx.ResolveArtefactsMounts(ArtefactMapping{ArtefactPath: "/blob/", StoragePath: "/blab/"}, i, false)

				Expect(i.MountsList()[0]).Should(Equal("/a:/blob"))
				Expect(i.MountsList()[1]).Should(Equal("/b:/blab"))
				Expect(mapping.ArtefactPath).Should(Equal("/a"))
				Expect(mapping.StoragePath).Should(Equal("/b"))
			})

			It("resolves the source repository mount with custom paths (relative)", func() {
				mapping := ctx.ResolveArtefactsMounts(ArtefactMapping{ArtefactPath: "blob/", StoragePath: "blab/"}, i, false)

				Expect(i.MountsList()[0]).Should(Equal("/a:/c/d/blob"))
				Expect(i.MountsList()[1]).Should(Equal("/b:/c/d/blab"))
				Expect(mapping.ArtefactPath).Should(Equal("/a"))
				Expect(mapping.StoragePath).Should(Equal("/b"))
			})
		})
	})
})
