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

package client

import (
	"github.com/mxk/go-flowrate/flowrate"

	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	storageci "github.com/MottainaiCI/mottainai-server/pkg/storage"
)

func (d *Fetcher) NamespaceFileList(namespace string) ([]string, error) {
	var fileList []string
	url := d.Config.GetWeb().BuildURI(
		fmt.Sprintf("/api/namespace/%s/list", namespace))
	err := d.GetJSONOptions(url, map[string]string{}, &fileList)
	if err != nil {
		return []string{}, err
	}

	return fileList, nil
}

func (d *Fetcher) StorageFileList(storage string) ([]string, error) {
	var fileList []string
	url := d.Config.GetWeb().BuildURI(
		fmt.Sprintf("/api/storage/%s/list", storage))
	err := d.GetJSONOptions(url, map[string]string{}, &fileList)
	if err != nil {
		return []string{}, err
	}

	return fileList, nil
}

func (d *Fetcher) TaskFileList(task string) ([]string, error) {
	var fileList []string
	url := d.Config.GetWeb().BuildURI(
		fmt.Sprintf("/api/tasks/%s/artefacts", task))
	err := d.GetJSONOptions(url, map[string]string{}, &fileList)
	if err != nil {
		return []string{}, err
	}

	return fileList, nil
}

func (d *Fetcher) DownloadArtefactsFromTask(taskid, target string) error {
	return d.DownloadArtefactsGeneric(taskid, target, "artefact")
}

func (fetcher *Fetcher) UploadFile(path, art string) error {
	_, file := filepath.Split(path)
	rel := strings.Replace(path, art, "", 1)
	rel = strings.Replace(rel, file, "", 1)

	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		// do directory stuff
		return err
	case mode.IsRegular():
		fetcher.AppendTaskOutput("Uploading " + path + " to " + rel)
		err = fetcher.UploadArtefactRetry(path, rel, 5)
	}

	return err
}

func (d *Fetcher) DownloadArtefactsFromStorage(storage, target string) error {
	return d.DownloadArtefactsGeneric(storage, target, "storage")
}

func (d *Fetcher) DownloadArtefactsGeneric(id, target, artefact_type string) error {
	var list []string
	var err error
	var to_download string
	if artefact_type == "namespace" {
		list, err = d.NamespaceFileList(id)
		if err != nil {
			d.AppendTaskOutput("[Download] Failed getting namespace artefact list")
			return err
		}
		to_download = id
	} else if artefact_type == "storage" {
		list, err = d.StorageFileList(id)
		if err != nil {
			d.AppendTaskOutput("[Download] Failed getting storage file list")
			return err
		}
		var storage_data storageci.Storage
		url := d.Config.GetWeb().BuildURI(
			fmt.Sprintf("/api/%s/%s/show", artefact_type, id))
		err = d.GetJSONOptions(url, map[string]string{}, &storage_data)
		if err != nil {
			d.AppendTaskOutput("Downloading failed during retrieveing storage data : " + err.Error())
			return err
		}
		to_download = storage_data.Path

	} else if artefact_type == "artefact" {
		list, err = d.TaskFileList(id)
		if err != nil {
			d.AppendTaskOutput("[Download] Failed getting task artefacts list")
			return err
		}
		to_download = id
	}

	err = os.MkdirAll(target, os.ModePerm)
	if err != nil {
		d.AppendTaskOutput("[Download] Error: " + err.Error())
		return err
	}
	d.AppendTaskOutput("[Download] " + artefact_type + " artefacts from " + id)
	success := true
	for _, file := range list {
		trials := 5
		done := true

		reldir, _ := filepath.Split(file)
		for done {

			if trials < 0 {
				done = false
				success = false
			}
			err := os.MkdirAll(filepath.Join(target, reldir), os.ModePerm)
			if err != nil {
				d.AppendTaskOutput("[Download] Error: " + err.Error())
				return err
			}
			d.AppendTaskOutput("[Download]  " + d.BaseURL + "/" + artefact_type + "/" + to_download + file + " to " + filepath.Join(target, file))
			if ok, err := d.Download(d.BaseURL+"/"+artefact_type+"/"+to_download+file, filepath.Join(target, file)); !ok {
				d.AppendTaskOutput("[Download] failed : " + err.Error())
				trials--
			} else {
				done = false
				d.AppendTaskOutput("[Download] succeeded ")
			}

		}

	}

	if !success {
		return errors.New("Download failed")
	}

	return nil
}

func (d *Fetcher) DownloadArtefactsFromNamespace(namespace, target string) error {
	return d.DownloadArtefactsGeneric(namespace, target, "namespace")
}

func responseSuccess(resp *http.Response) bool {
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return true
	} else {
		return false
	}
}
func (d *Fetcher) Download(url, where string) (bool, error) {
	fileName := where

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	d.setAuthHeader(request)

	client := d.newHttpClient()
	response, err := client.Do(request)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()
	body := response.Body
	if d.Config.GetAgent().DownloadRateLimit != 0 {
		// KB
		d.AppendTaskOutput("Download with bandwidth limit of: " + strconv.FormatInt(1024*d.Config.GetAgent().DownloadRateLimit, 10))
		body = flowrate.NewReader(response.Body, 1024*d.Config.GetAgent().DownloadRateLimit)
	}
	if !responseSuccess(response) {
		return false, errors.New("Error: " + response.Status)
	}

	output, err := os.Create(fileName)
	if err != nil {
		return false, err
	}
	defer output.Close()

	_, err = io.Copy(output, body)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (f *Fetcher) UploadStorageFile(storageid, fullpath, relativepath string) error {
	_, file := filepath.Split(fullpath)

	opts := map[string]string{
		"name":      file,
		"path":      relativepath,
		"storageid": storageid,
		//	"namespace": dir,
	}

	url := f.Config.GetWeb().BuildURI("/api/storage/upload")
	request, err := f.Upload(url, opts, "file", fullpath)
	if err != nil {
		panic(err)
	}
	f.setAuthHeader(request)

	client := f.newHttpClient()
	resp, err := client.Do(request)
	if err != nil {
		panic(err)
	} else {
		var bodyContent []byte
		resp.Body.Read(bodyContent)
		resp.Body.Close()
	}
	return nil
}

func (f *Fetcher) UploadArtefactRetry(fullpath, relativepath string, trials int) error {
	trial := 1
	err := f.UploadArtefact(fullpath, relativepath)
	for err != nil && trial < trials {
		trial++
		f.AppendTaskOutput("[Upload] Trial " + strconv.Itoa(trial))
		err = f.UploadArtefact(fullpath, relativepath)
	}
	return err
}

func (f *Fetcher) UploadArtefact(fullpath, relativepath string) error {
	_, file := filepath.Split(fullpath)

	opts := map[string]string{
		"name":   file,
		"path":   relativepath,
		"taskid": f.docID,
		//	"namespace": dir,
	}

	url := f.Config.GetWeb().BuildURI("/api/tasks/artefact/upload")
	if err := f.UploadLargeFile(url, opts, "file", fullpath, f.ChunkSize); err != nil {
		f.AppendTaskOutput("[Upload] Error while uploading artefact " + file + ": " + err.Error())
		return err
	}
	return nil
}

func (f *Fetcher) UploadNamespaceFile(namespace, fullpath, relativepath string) error {
	_, file := filepath.Split(fullpath)

	opts := map[string]string{
		"name":      file,
		"path":      relativepath,
		"namespace": namespace,
		//	"namespace": dir,
	}

	url := f.Config.GetWeb().BuildURI("/api/namespace/upload")
	if err := f.UploadLargeFile(url, opts, "file", fullpath, f.ChunkSize); err != nil {
		return err
	}
	return nil
}
