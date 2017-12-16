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

package utils

import (
	"os"
	"os/exec"
	"strings"

	log "gopkg.in/clog.v1"
)

// Git executes a git command with the given args as a []string, outputs as a string
func Git(cmdArgs []string, dir string) (string, error) {
	var (
		cmdOut []byte
		err    error
	)
	cwd, _ := os.Getwd()
	os.Chdir(dir)
	cmdName := "git"
	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		log.Error(2, "There was an error running git command: ", err)
		log.Info(strings.Join(cmdArgs, " "))
		log.Error(2, string(cmdOut))
		return "", err
	}
	os.Chdir(cwd)
	result := string(cmdOut)
	return strings.TrimSpace(result), err
}

// GitAlignToUpstream executes a fetch --all and reset --hard to origin/master on the given git repository
func GitAlignToUpstream(workdir string) {
	GitAlignTo(workdir, "origin/master")
}

func GitAlignTo(workdir, to string) {
	log.Info(Git([]string{"fetch", "--all"}, workdir))
	log.Info(Git([]string{"reset", "--hard", to}, workdir))
}

func GitPrevCommit(workdir string) (string, error) {
	result, err := Git([]string{"log", "-2", `--pretty=format:"%h"`}, workdir)
	temp := strings.Split(result, "\n")
	return temp[1], err
}

// GitHead returns the Head of the given repository
func GitHead(workdir string) string {
	head, _ := Git([]string{"rev-parse", "HEAD"}, workdir)
	return head
}
