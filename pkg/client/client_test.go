/*

Copyright (C) 2017-2018  Ettore Di Giacinto <mudler@gentoo.org>
Some code portions and re-implemented design are also coming
from the Gogs project, which is using the go-macaron framework and was
really source of ispiration. Kudos to them!

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

package client

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/mudler/anagent"
)

func TestNewTokenClient(t *testing.T) {
	type args struct {
		host  string
		token string
	}
	tests := []struct {
		name string
		args args
		want *Fetcher
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTokenClient(tt.args.host, tt.args.token); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTokenClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	type args struct {
		host string
	}
	tests := []struct {
		name string
		args args
		want *Fetcher
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewClient(tt.args.host); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewFetcher(t *testing.T) {
	type args struct {
		docID string
	}
	tests := []struct {
		name string
		args args
		want *Fetcher
	}{
		{"Create", args{"20"}, &Fetcher{docID: "20"}},
		{"Create2", args{"String"}, &Fetcher{docID: "String"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFetcher(tt.args.docID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFetcher() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewBasicClient(t *testing.T) {
	tests := []struct {
		name string
		want *Fetcher
	}{
		{"Basic", &Fetcher{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBasicClient(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBasicClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		docID string
		a     *anagent.Anagent
	}
	tests := []struct {
		name string
		args args
		want *Fetcher
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.docID, tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetcher_Doc(t *testing.T) {
	type fields struct {
		BaseURL       string
		docID         string
		Token         string
		TrustedCert   string
		Jar           *http.CookieJar
		Agent         *anagent.Anagent
		ActiveReports bool
	}
	type args struct {
		id string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Fetcher{
				BaseURL:       tt.fields.BaseURL,
				docID:         tt.fields.docID,
				Token:         tt.fields.Token,
				TrustedCert:   tt.fields.TrustedCert,
				Jar:           tt.fields.Jar,
				Agent:         tt.fields.Agent,
				ActiveReports: tt.fields.ActiveReports,
			}
			f.Doc(tt.args.id)
		})
	}
}

func TestFetcher_newHttpClient(t *testing.T) {
	type fields struct {
		BaseURL       string
		docID         string
		Token         string
		TrustedCert   string
		Jar           *http.CookieJar
		Agent         *anagent.Anagent
		ActiveReports bool
	}
	tests := []struct {
		name   string
		fields fields
		want   *http.Client
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Fetcher{
				BaseURL:       tt.fields.BaseURL,
				docID:         tt.fields.docID,
				Token:         tt.fields.Token,
				TrustedCert:   tt.fields.TrustedCert,
				Jar:           tt.fields.Jar,
				Agent:         tt.fields.Agent,
				ActiveReports: tt.fields.ActiveReports,
			}
			if got := f.newHttpClient(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Fetcher.newHttpClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetcher_setAuthHeader(t *testing.T) {
	type fields struct {
		BaseURL       string
		docID         string
		Token         string
		TrustedCert   string
		Jar           *http.CookieJar
		Agent         *anagent.Anagent
		ActiveReports bool
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *http.Request
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Fetcher{
				BaseURL:       tt.fields.BaseURL,
				docID:         tt.fields.docID,
				Token:         tt.fields.Token,
				TrustedCert:   tt.fields.TrustedCert,
				Jar:           tt.fields.Jar,
				Agent:         tt.fields.Agent,
				ActiveReports: tt.fields.ActiveReports,
			}
			if got := f.setAuthHeader(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Fetcher.setAuthHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetcher_GetJSONOptions(t *testing.T) {
	type fields struct {
		BaseURL       string
		docID         string
		Token         string
		TrustedCert   string
		Jar           *http.CookieJar
		Agent         *anagent.Anagent
		ActiveReports bool
	}
	type args struct {
		url    string
		option map[string]string
		target interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Fetcher{
				BaseURL:       tt.fields.BaseURL,
				docID:         tt.fields.docID,
				Token:         tt.fields.Token,
				TrustedCert:   tt.fields.TrustedCert,
				Jar:           tt.fields.Jar,
				Agent:         tt.fields.Agent,
				ActiveReports: tt.fields.ActiveReports,
			}
			if err := f.GetJSONOptions(tt.args.url, tt.args.option, tt.args.target); (err != nil) != tt.wantErr {
				t.Errorf("Fetcher.GetJSONOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFetcher_GetOptions(t *testing.T) {
	type fields struct {
		BaseURL       string
		docID         string
		Token         string
		TrustedCert   string
		Jar           *http.CookieJar
		Agent         *anagent.Anagent
		ActiveReports bool
	}
	type args struct {
		url    string
		option map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Fetcher{
				BaseURL:       tt.fields.BaseURL,
				docID:         tt.fields.docID,
				Token:         tt.fields.Token,
				TrustedCert:   tt.fields.TrustedCert,
				Jar:           tt.fields.Jar,
				Agent:         tt.fields.Agent,
				ActiveReports: tt.fields.ActiveReports,
			}
			got, err := f.GetOptions(tt.args.url, tt.args.option)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fetcher.GetOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Fetcher.GetOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetcher_GenericForm(t *testing.T) {
	type fields struct {
		BaseURL       string
		docID         string
		Token         string
		TrustedCert   string
		Jar           *http.CookieJar
		Agent         *anagent.Anagent
		ActiveReports bool
	}
	type args struct {
		URL    string
		option map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Fetcher{
				BaseURL:       tt.fields.BaseURL,
				docID:         tt.fields.docID,
				Token:         tt.fields.Token,
				TrustedCert:   tt.fields.TrustedCert,
				Jar:           tt.fields.Jar,
				Agent:         tt.fields.Agent,
				ActiveReports: tt.fields.ActiveReports,
			}
			got, err := f.GenericForm(tt.args.URL, tt.args.option)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fetcher.GenericForm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Fetcher.GenericForm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetcher_Form(t *testing.T) {
	type fields struct {
		BaseURL       string
		docID         string
		Token         string
		TrustedCert   string
		Jar           *http.CookieJar
		Agent         *anagent.Anagent
		ActiveReports bool
	}
	type args struct {
		URL    string
		option map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Fetcher{
				BaseURL:       tt.fields.BaseURL,
				docID:         tt.fields.docID,
				Token:         tt.fields.Token,
				TrustedCert:   tt.fields.TrustedCert,
				Jar:           tt.fields.Jar,
				Agent:         tt.fields.Agent,
				ActiveReports: tt.fields.ActiveReports,
			}
			got, err := f.Form(tt.args.URL, tt.args.option)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fetcher.Form() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Fetcher.Form() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetcher_PostOptions(t *testing.T) {
	type fields struct {
		BaseURL       string
		docID         string
		Token         string
		TrustedCert   string
		Jar           *http.CookieJar
		Agent         *anagent.Anagent
		ActiveReports bool
	}
	type args struct {
		URL    string
		option map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Fetcher{
				BaseURL:       tt.fields.BaseURL,
				docID:         tt.fields.docID,
				Token:         tt.fields.Token,
				TrustedCert:   tt.fields.TrustedCert,
				Jar:           tt.fields.Jar,
				Agent:         tt.fields.Agent,
				ActiveReports: tt.fields.ActiveReports,
			}
			got, err := f.PostOptions(tt.args.URL, tt.args.option)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fetcher.PostOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Fetcher.PostOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetcher_UploadLargeFile(t *testing.T) {
	type fields struct {
		BaseURL       string
		docID         string
		Token         string
		TrustedCert   string
		Jar           *http.CookieJar
		Agent         *anagent.Anagent
		ActiveReports bool
	}
	type args struct {
		uri       string
		params    map[string]string
		paramName string
		filePath  string
		chunkSize int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Fetcher{
				BaseURL:       tt.fields.BaseURL,
				docID:         tt.fields.docID,
				Token:         tt.fields.Token,
				TrustedCert:   tt.fields.TrustedCert,
				Jar:           tt.fields.Jar,
				Agent:         tt.fields.Agent,
				ActiveReports: tt.fields.ActiveReports,
			}
			if err := f.UploadLargeFile(tt.args.uri, tt.args.params, tt.args.paramName, tt.args.filePath, tt.args.chunkSize); (err != nil) != tt.wantErr {
				t.Errorf("Fetcher.UploadLargeFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFetcher_Upload(t *testing.T) {
	type fields struct {
		BaseURL       string
		docID         string
		Token         string
		TrustedCert   string
		Jar           *http.CookieJar
		Agent         *anagent.Anagent
		ActiveReports bool
	}
	type args struct {
		uri       string
		params    map[string]string
		paramName string
		path      string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *http.Request
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Fetcher{
				BaseURL:       tt.fields.BaseURL,
				docID:         tt.fields.docID,
				Token:         tt.fields.Token,
				TrustedCert:   tt.fields.TrustedCert,
				Jar:           tt.fields.Jar,
				Agent:         tt.fields.Agent,
				ActiveReports: tt.fields.ActiveReports,
			}
			got, err := f.Upload(tt.args.uri, tt.args.params, tt.args.paramName, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fetcher.Upload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Fetcher.Upload() = %v, want %v", got, tt.want)
			}
		})
	}
}
