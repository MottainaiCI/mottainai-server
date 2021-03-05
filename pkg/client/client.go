/*

Copyright (C) 2017-2019  Ettore Di Giacinto <mudler@gentoo.org>
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
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/mxk/go-flowrate/flowrate"

	event "github.com/MottainaiCI/mottainai-server/pkg/event"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	schema "github.com/MottainaiCI/mottainai-server/routes/schema"

	"github.com/mudler/anagent"
)

var _ HttpClient = &Fetcher{}

type HttpClient interface {
	AppendTaskOutput(string) (event.APIResponse, error)

	GetTask() ([]byte, error)
	AbortTask()
	DownloadArtefactsFromTask(string, string, []string) error
	DownloadArtefactsFromNamespace(string, string, []string) error
	DownloadArtefactsFromStorage(string, string) error
	UploadFile(string, string) error
	FailTask(string)
	SetTaskField(string, string) (event.APIResponse, error)
	RegisterNode(string, string) (event.APIResponse, error)
	Doc(string)
	SetUploadChunkSize(int)
	SetupTask() (event.APIResponse, error)
	FinishTask()
	ErrorTask()
	SuccessTask()
	StreamOutput(io.Reader)
	RunTask()

	StorageDelete(id string) (event.APIResponse, error)
	StorageRemovePath(id, path string) (event.APIResponse, error)
	StorageCreate(t string) (event.APIResponse, error)
	SettingCreate(data map[string]interface{}) (event.APIResponse, error)
	SettingRemove(id string) (event.APIResponse, error)
	SettingUpdate(data map[string]interface{}) (event.APIResponse, error)
	PlanDelete(id string) (event.APIResponse, error)
	PlanCreate(taskdata map[string]interface{}) (event.APIResponse, error)
	SetBaseURL(url string)
	SetAgent(a *anagent.Anagent)
	SetActiveReport(b bool)
	SetToken(t string)
	HandleRaw(req schema.Request, fn func(io.ReadCloser) error) error
	Handle(req schema.Request) error
	HandleAPIResponse(req schema.Request) (event.APIResponse, error)
	HandleUploadLargeFile(request schema.Request, paramName string, filePath string, chunkSize int) error
	TaskLog(id string) ([]byte, error)
	TaskDelete(id string) (event.APIResponse, error)
	SetTaskStatus(status string) (event.APIResponse, error)
	StartTask(id string) (event.APIResponse, error)
	StopTask(id string) (event.APIResponse, error)
	CreateTask(taskdata map[string]interface{}) (event.APIResponse, error)
	CloneTask(id string) (event.APIResponse, error)
	TaskLogArtefact(id string) ([]byte, error)
	TaskStream(id, pos string) ([]byte, error)
	AllTasks() ([]byte, error)
	SetTaskResult(result string) (event.APIResponse, error)
	SetTaskOutput(output string) (event.APIResponse, error)
	WebHookTaskUpdate(id string, data map[string]interface{}) (event.APIResponse, error)
	WebHookPipelineUpdate(id string, data map[string]interface{}) (event.APIResponse, error)
	WebHookDelete(id string) (event.APIResponse, error)
	WebHookDeleteTask(id string) (event.APIResponse, error)
	WebHookDeletePipeline(id string) (event.APIResponse, error)
	WebHookEdit(data map[string]interface{}) (event.APIResponse, error)
	WebHookCreate(t string) (event.APIResponse, error)
	TokenDelete(id string) (event.APIResponse, error)
	TokenCreate() (event.APIResponse, error)
	UploadStorageFile(storageid, fullpath, relativepath string) error
	UploadArtefactRetry(fullpath, relativepath string, trials int) error
	UploadArtefact(fullpath, relativepath string) error
	UploadNamespaceFile(namespace, fullpath, relativepath string) error
	UserCreate(data map[string]interface{}) (event.APIResponse, error)
	UserRemove(id string) (event.APIResponse, error)
	UserUpdate(id string, data map[string]interface{}) (event.APIResponse, error)
	UserSet(id, t string) (event.APIResponse, error)
	UserUnset(id, t string) (event.APIResponse, error)
	PipelineDelete(id string) (event.APIResponse, error)
	PipelineCreate(taskdata map[string]interface{}) (event.APIResponse, error)
	NamespaceDelete(id string) (event.APIResponse, error)
	NamespaceRemovePath(id, path string) (event.APIResponse, error)
	NamespaceClone(from, to string) (event.APIResponse, error)
	NamespaceAppend(id, name string) (event.APIResponse, error)
	NamespaceTag(id, tag string) (event.APIResponse, error)
	NamespaceCreate(t string) (event.APIResponse, error)
	GetBaseURL() (url string)
	CreateNode() (event.APIResponse, error)
	RemoveNode(id string) (event.APIResponse, error)
	NodesTask(key string, target interface{}) error
	NamespaceFileList(namespace string) ([]string, error)
	StorageFileList(storage string) ([]string, error)
	TaskFileList(task string) ([]string, error)
	DownloadArtefactsGeneric(id, target, artefact_type string, filters []string) error
	Download(url, where string) (bool, error)

	SecretDelete(id string) (event.APIResponse, error)
	SecretEdit(data map[string]interface{}) (event.APIResponse, error)
	SecretCreate(t string) (event.APIResponse, error)
}

type Fetcher struct {
	ChunkSize int
	BaseURL   string
	docID     string
	// TODO: this could be handled directly from Config
	Token string
	// TODO: this could be handled directly from Config
	TrustedCert   string
	Jar           *http.CookieJar
	Agent         *anagent.Anagent
	ActiveReports bool
	Config        *setting.Config
}

func NewTokenClient(host, token string, config *setting.Config) HttpClient {
	f := NewBasicClient(config)
	f.SetBaseURL(host)
	f.SetToken(token)
	return f
}

func NewClient(host string, config *setting.Config) HttpClient {
	f := NewBasicClient(config)
	f.SetBaseURL(host)
	return f
}

func NewFetcher(docID string, config *setting.Config) HttpClient {
	f := NewClient(config.GetWeb().AppURL, config)
	f.Doc(docID)
	return f
}

func NewBasicClient(config *setting.Config) HttpClient {
	// Basic constructor
	f := &Fetcher{Config: config, ChunkSize: 512}
	if len(config.GetGeneral().TLSCert) > 0 {
		f.TrustedCert = config.GetGeneral().TLSCert
	}
	return f
}

func New(docID string, a *anagent.Anagent, config *setting.Config) HttpClient {
	f := NewClient(config.GetWeb().AppURL, config)
	f.Doc(docID)
	f.SetAgent(a)
	return f
}

func (f *Fetcher) GetBaseURL() (url string) {
	url = f.BaseURL
	return
}
func (f *Fetcher) SetBaseURL(url string) {
	f.BaseURL = url
}
func (f *Fetcher) SetAgent(a *anagent.Anagent) {
	f.Agent = a
}
func (f *Fetcher) SetActiveReport(b bool) {
	f.ActiveReports = b
}
func (f *Fetcher) SetToken(t string) {
	f.Token = t
}

func (f *Fetcher) Doc(id string) {
	f.docID = id
}

func (f *Fetcher) newHttpClient() *http.Client {

	c := &http.Client{Timeout: time.Second * time.Duration(f.Config.GetGeneral().ClientTimeout)}

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

func (f *Fetcher) SetUploadChunkSize(s int) {
	f.ChunkSize = s
}

func (f *Fetcher) setAuthHeader(r *http.Request) *http.Request {
	if len(f.Token) > 0 {
		r.Header.Add("Authorization", "token "+f.Token)
	}
	return r
}

func (f *Fetcher) HandleRaw(req schema.Request, fn func(io.ReadCloser) error) error {

	hclient := f.newHttpClient()
	baseurl := f.BaseURL + f.Config.GetWeb().BuildURI("")
	request, err := req.NewAPIHTTPRequest(baseurl)
	if err != nil {
		return err
	}

	f.setAuthHeader(request)

	response, err := hclient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return fn(response.Body)
}

func (f *Fetcher) Handle(req schema.Request) error {
	return f.HandleRaw(req, func(b io.ReadCloser) error {
		buf := new(bytes.Buffer)
		n, err := buf.ReadFrom(b)
		if n == 0 {
			return errors.New("Received empty response from server")
		} else if err != nil {
			return err
		}

		return json.Unmarshal(buf.Bytes(), req.Target)
	})
}

func (f *Fetcher) HandleAPIResponse(req schema.Request) (event.APIResponse, error) {
	resp := &event.APIResponse{}
	req.Target = resp
	err := f.Handle(req)
	if err != nil {
		return *resp, err
	}

	return *resp, nil
}

func (f *Fetcher) HandleUploadLargeFile(request schema.Request, paramName string, filePath string, chunkSize int) error {

	option := request.Options
	baseurl := f.BaseURL + f.Config.GetWeb().BuildURI("")

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

	for key, value := range option {
		err = mpWriter.WriteField(key, value.(string))
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
	request.Body = rd

	req, err := request.NewAPIHTTPRequest(baseurl)
	if err != nil {
		return err
	}

	// XXX: Yeah, this is just a fancier way of reading slowly from kernel buffers, i know.
	if f.Config.GetAgent().UploadRateLimit != 0 {
		f.AppendTaskOutput("Upload with bandwidth limit of: " + strconv.FormatInt(1024*f.Config.GetAgent().UploadRateLimit, 10) + "\n")
		reader := flowrate.NewReader(io.Reader(rd), 1024*f.Config.GetAgent().UploadRateLimit)
		request.Body = reader
		req, err = request.NewAPIHTTPRequest(baseurl)
		if err != nil {
			return err
		}
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
		return err
	}
	defer resp.Body.Close()

	body := &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("[Upload] Error while uploading " + filePath + ": " + strconv.Itoa(resp.StatusCode))
	}

	return nil
}
