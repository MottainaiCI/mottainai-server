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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/MottainaiCI/mottainai-server/pkg/tasks"
)

var _ = Describe("Instruction", func() {
	i := NewDefaultInstruction([]string{"/bin/bash", "-c"}, []string{"echo 'bar'", "echo 'foo'"})

	Describe("DefaultInstruction ToScript()", func() {
		Context("Converts them to one string", func() {
			It("simply joins them", func() {
				Expect(i.ToScript()).Should(Equal("echo 'bar' && echo 'foo'"))
			})
		})
		Context("empty script", func() {
			It("returns empty", func() {
				Expect(NewDefaultInstruction([]string{}, []string{}).ToScript()).Should(Equal(""))
			})
		})
	})

	Describe("NewDebugInstruction", func() {
		instruction := NewDebugInstruction([]string{"echo 'bar'", "echo 'foo'"})

		Context("ToScript() Converts them to one string", func() {
			It("simply joins them", func() {
				Expect(instruction.ToScript()).Should(Equal("-c && pwd && ls -liah && echo 'bar' && echo 'foo'"))
			})
		})
		Context("CommandList()", func() {
			It("return all the commands", func() {
				Expect(instruction.CommandList()).Should(Equal([]string{"-c", "pwd", "ls -liah", "echo 'bar' && echo 'foo'"}))
			})
		})
	})

	Describe("NewBashInstruction", func() {
		instruction := NewBashInstruction([]string{"echo 'bar'", "echo 'foo'"})

		Context("ToScript() Converts them to one string", func() {
			It("simply joins them", func() {
				Expect(instruction.ToScript()).Should(Equal("-c && echo 'bar' && echo 'foo'"))
			})
		})
		Context("CommandList()", func() {
			It("return all the commands", func() {
				Expect(instruction.CommandList()).Should(Equal([]string{"-c", "echo 'bar'", "echo 'foo'"}))
			})
		})
	})

})
