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
	"errors"
	"os"
	"os/exec"
	"strings"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"

	log "gopkg.in/clog.v1"
)

//TODO: Git* Can go in a separate object
func GitClone(url, dir string) (*git.Repository, error) {
	//os.RemoveAll(dir)
	r, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: url,
		//Progress: os.Stdout,
	})
	if err != nil {
		os.RemoveAll(dir)
		return nil, errors.New("Failed cloning repo: " + url + " " + dir + " " + err.Error())
	}
	return r, nil
}

func GitCheckoutCommit(r *git.Repository, commit string) error {
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(commit),
	})
	if err != nil {
		return err
	}
	return nil
}

func GitFetch(r *git.Repository, remote string, args []string) error {
	var refs []config.RefSpec
	for _, ref := range args {
		refs = append(refs, config.RefSpec(ref))
	}
	err := r.Fetch(&git.FetchOptions{
		RemoteName: remote,
		RefSpecs:   refs,
	})
	if err != nil {
		return err
	}
	return nil
}

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
