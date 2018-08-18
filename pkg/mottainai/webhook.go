/*

Copyright (C) 2018  Ettore Di Giacinto <mudler@gentoo.org>
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

package mottainai

import (
	"fmt"
	"log"

	"net/http"

	"github.com/google/go-github/github"
	anagent "github.com/mudler/anagent"
	"golang.org/x/oauth2"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
)

type WebHookServer struct {
	*Mottainai
}

func NewWebHookServer() *WebHookServer {
	return &WebHookServer{Mottainai: New()}
}

func ClassicWebHookServer() *WebHookServer {
	return &WebHookServer{Mottainai: Classic()}
}

func SetupWebHook(m *Mottainai) {

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: setting.Configuration.WebHookGitHubTokenUser})
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	// Get a client instance from github
	client := github.NewClient(tc)

	m.Map(tc)
	m.Map(client)
	a := anagent.New()

	a.TimerSeconds(int64(5), true, func() {})
	m.Map(a)
	a.Map(m)

}

func SetupWebHookAgent(m *Mottainai) {
	m.Invoke(func(a *anagent.Anagent) {
		go a.Start()
	})
	//go func(a *anagent.Anagent) {
	//	a.Start()
	//}(a)
}

func (m *WebHookServer) Start() error {

	m.Map(m)
	m.Map(m.Mottainai)

	SetupWebHook(m.Mottainai)
	SetupWebHookAgent(m.Mottainai)

	var listenAddr = fmt.Sprintf("%s:%s", setting.Configuration.HTTPAddr, setting.Configuration.HTTPPort)
	log.Printf("Listen: %v://%s", setting.Configuration.Protocol, listenAddr)

	//m.Run()
	err := http.ListenAndServe(listenAddr, m)

	if err != nil {
		log.Fatal(4, "Fail to start server: %v", err)
	}
	return nil

}
