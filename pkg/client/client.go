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
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/gob"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"

	"github.com/mudler/anagent"
)

type HttpClient interface {
	AppendTaskOutput(string) ([]byte, error)

	GetTask() ([]byte, error)
	AbortTask()
	DownloadArtefactsFromTask(string, string)
	DownloadArtefactsFromNamespace(string, string)
	DownloadArtefactsFromStorage(string, string)
	UploadFile(string, string) error
	FailTask(string)
	SetTaskField(string, string) ([]byte, error)
	RegisterNode(string, string) ([]byte, error)
	Doc(string)
	SetupTask()
	FinishTask()
	ErrorTask()
	SuccessTask()
	StreamOutput(io.Reader)
	RunTask()
}

type Fetcher struct {
	BaseURL       string
	docID         string
	Token         string
	TrustedCert   string
	Jar           *http.CookieJar
	Agent         *anagent.Anagent
	ActiveReports bool
}

func NewTokenClient(host, token string) *Fetcher {
	f := NewBasicClient()
	f.BaseURL = host
	f.Token = token
	return f
}

func NewClient(host string) *Fetcher {
	f := NewBasicClient()
	f.BaseURL = host
	return f
}

func NewFetcher(docID string) *Fetcher {
	f := NewClient(setting.Configuration.AppURL)
	f.docID = docID
	return f
}

func NewBasicClient() *Fetcher {
	// Basic constructor
	f := &Fetcher{}
	if len(setting.Configuration.TLSCert) > 0 {
		f.TrustedCert = setting.Configuration.TLSCert
	}
	return f
}

func New(docID string, a *anagent.Anagent) *Fetcher {
	f := NewClient(setting.Configuration.AppURL)
	f.docID = docID
	f.Agent = a
	return f
}

func (f *Fetcher) Doc(id string) {
	f.docID = id
}

func (f *Fetcher) newHttpClient() *http.Client {

	c := &http.Client{}

	if len(f.TrustedCert) > 0 {
		rootCAs, _ := x509.SystemCertPool()

		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		// Read in the cert file
		certs, err := ioutil.ReadFile(f.TrustedCert)
		if err != nil {
			log.Fatalf("Failed to append %q to RootCAs: %v", f.TrustedCert, err)
		}

		// Append our cert to the system pool
		if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
			log.Println("No certs appended, using system certs only")
		}

		// Trust the augmented cert pool in our client
		config := &tls.Config{
			RootCAs: rootCAs,
		}
		tr := &http.Transport{TLSClientConfig: config}
		c.Transport = tr
	}

	if f.Jar != nil {
		c.Jar = *f.Jar
	}
	return c
}

func (f *Fetcher) setAuthHeader(r *http.Request) *http.Request {
	if len(f.Token) > 0 {
		r.Header.Add("Authorization", "token "+f.Token)
	}
	return r
}

func (f *Fetcher) GetJSONOptions(url string, option map[string]string, target interface{}) error {
	hclient := f.newHttpClient()
	request, err := http.NewRequest("GET", f.BaseURL+url, nil)
	f.setAuthHeader(request)

	if err != nil {
		return err
	}

	q := request.URL.Query()
	for k, v := range option {
		q.Add(k, v)
	}
	request.URL.RawQuery = q.Encode()
	if err != nil {
		return err
	}

	response, err := hclient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return json.NewDecoder(response.Body).Decode(target)
}

func (f *Fetcher) GetOptions(url string, option map[string]string) ([]byte, error) {
	hclient := f.newHttpClient()
	request, err := http.NewRequest("GET", f.BaseURL+url, nil)
	f.setAuthHeader(request)
	if err != nil {
		return []byte{}, err
	}

	q := request.URL.Query()
	for k, v := range option {
		q.Add(k, v)
	}
	request.URL.RawQuery = q.Encode()
	if err != nil {
		return []byte{}, err
	}

	response, err := hclient.Do(request)
	if err != nil {
		return []byte{}, err
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	return contents, err
}

func (f *Fetcher) GenericForm(URL string, option map[string]interface{}) ([]byte, error) {
	hclient := f.newHttpClient()
	form := url.Values{}
	var InterfaceList []interface{}
	var Strings []string
	var String string

	for k, v := range option {
		if reflect.TypeOf(v) == reflect.TypeOf(InterfaceList) {
			for _, el := range v.([]interface{}) {
				form.Add(k, el.(string))
			}
		} else if reflect.TypeOf(v) == reflect.TypeOf(Strings) {
			for _, el := range v.([]string) {
				form.Add(k, el)
			}

		} else if reflect.TypeOf(v) == reflect.TypeOf(float64(0)) {
			form.Add(k, utils.FloatToString(v.(float64)))

		} else if reflect.TypeOf(v) == reflect.TypeOf(String) {
			form.Add(k, v.(string))
		} else {
			var b bytes.Buffer
			e := gob.NewEncoder(&b)
			if err := e.Encode(v); err != nil {
				panic(err)
			}
			form.Add(k, b.String())
		}
	}

	request, err := http.NewRequest("POST", f.BaseURL+URL, strings.NewReader(form.Encode()))
	f.setAuthHeader(request)
	if err != nil {
		return []byte{}, err
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := hclient.Do(request)
	if err != nil {
		return []byte{}, err
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	return contents, err
}

func (f *Fetcher) Form(URL string, option map[string]string) ([]byte, error) {
	hclient := f.newHttpClient()

	form := url.Values{}
	for k, v := range option {
		form.Add(k, v)
	}

	request, err := http.NewRequest("POST", f.BaseURL+URL, strings.NewReader(form.Encode()))
	f.setAuthHeader(request)
	if err != nil {
		return []byte{}, err
	}
	//request.Header.Add("Content-Type", writer.FormDataContentType())

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// q := request.URL.Query()
	// for k, v := range option {
	// 	q.Add(k, v)
	// }
	// request.URL.RawQuery = q.Encode()
	// if err != nil {
	// 	return []byte{}, err
	// }

	response, err := hclient.Do(request)
	if err != nil {
		return []byte{}, err
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	return contents, err
}

func (f *Fetcher) PostOptions(URL string, option map[string]string) ([]byte, error) {
	hclient := f.newHttpClient()

	form := url.Values{}
	for k, v := range option {
		form.Add(k, v)
	}

	request, err := http.NewRequest("POST", f.BaseURL+URL, strings.NewReader(form.Encode()))
	f.setAuthHeader(request)

	if err != nil {
		return []byte{}, err
	}
	//request.Header.Add("Content-Type", writer.FormDataContentType())

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	q := request.URL.Query()
	for k, v := range option {
		q.Add(k, v)
	}
	request.URL.RawQuery = q.Encode()
	if err != nil {
		return []byte{}, err
	}

	response, err := hclient.Do(request)
	if err != nil {
		return []byte{}, err
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	return contents, err
}

func (f *Fetcher) UploadLargeFile(uri string, params map[string]string, paramName string, filePath string, chunkSize int) error {
	//open file and retrieve info
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	fi, err := file.Stat()
	if err != nil {
		return err
	}
	defer file.Close()

	//buffer for storing multipart data
	byteBuf := &bytes.Buffer{}

	//part: parameters
	mpWriter := multipart.NewWriter(byteBuf)
	for key, value := range params {
		err = mpWriter.WriteField(key, value)
		if err != nil {
			return err
		}
	}

	//part: file
	mpWriter.CreateFormFile(paramName, fi.Name())
	contentType := mpWriter.FormDataContentType()

	nmulti := byteBuf.Len()
	multi := make([]byte, nmulti)
	_, err = byteBuf.Read(multi)
	if err != nil {
		return err
	}
	//part: latest boundary
	//when multipart closed, latest boundary is added
	mpWriter.Close()
	nboundary := byteBuf.Len()
	lastBoundary := make([]byte, nboundary)
	_, err = byteBuf.Read(lastBoundary)
	if err != nil {
		return err
	}

	//use pipe to pass request
	rd, wr := io.Pipe()
	defer rd.Close()

	go func() {
		defer wr.Close()

		//write multipart
		_, _ = wr.Write(multi)

		//write file
		buf := make([]byte, chunkSize)
		for {
			n, err := file.Read(buf)
			if err != nil {
				break
			}
			_, _ = wr.Write(buf[:n])
		}
		//write boundary
		_, _ = wr.Write(lastBoundary)
	}()

	//construct request with rd
	req, err := http.NewRequest("POST", f.BaseURL+uri, rd)
	if err != nil {
		return err
	}
	f.setAuthHeader(req)

	req.TransferEncoding = []string{"chunked"}

	req.Header.Set("Content-Type", contentType)
	req.ContentLength = -1 //totalSize
	req.ProtoMajor = 1
	req.ProtoMinor = 1
	req.Header.Add("Connection", "keep-alive")

	//process request
	client := f.newHttpClient()
	client.Timeout = 0
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println(resp.StatusCode)
		log.Println(resp.Header)

		body := &bytes.Buffer{}
		_, _ = body.ReadFrom(resp.Body)
		resp.Body.Close()
		log.Println(body)
		if resp.StatusCode != 200 {
			return errors.New("[Upload] Error while uploading " + filePath + ": " + strconv.Itoa(resp.StatusCode))
		}
	}
	return err
}

// Creates a new file upload http request with optional extra params
func (f *Fetcher) Upload(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	file.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, fi.Name())
	if err != nil {
		return nil, err
	}
	part.Write(fileContents)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", f.BaseURL+uri, body)
	f.setAuthHeader(request)

	if err != nil {
		return request, nil
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())
	return request, nil
}
