/*

Copyright (C) 2018-2019  Ettore Di Giacinto <mudler@gentoo.org>
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
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	schema "github.com/MottainaiCI/mottainai-server/routes/schema"
	v1 "github.com/MottainaiCI/mottainai-server/routes/schema/v1"

	utils "github.com/MottainaiCI/mottainai-server/pkg/utils"

	storageci "github.com/MottainaiCI/mottainai-server/pkg/storage"
)

func (d *Fetcher) NamespaceFileList(namespace string) ([]string, error) {
	var fileList []string

	req := schema.Request{
		Route:   v1.Schema.GetNamespaceRoute("show_artefacts"),
		Options: map[string]interface{}{":name": namespace},
		Target:  &fileList,
	}

	err := d.Handle(req)
	if err != nil {
		return []string{}, err
	}

	return fileList, nil
}

func (d *Fetcher) StorageFileList(storage string) ([]string, error) {
	var fileList []string

	req := schema.Request{
		Route:   v1.Schema.GetStorageRoute("show_artefacts"),
		Options: map[string]interface{}{":id": storage},
		Target:  &fileList,
	}

	err := d.Handle(req)
	if err != nil {
		return []string{}, err
	}

	return fileList, nil
}

func (d *Fetcher) TaskFileList(task string) ([]string, error) {
	var fileList []string

	req := schema.Request{
		Route:   v1.Schema.GetTaskRoute("artefact_list"),
		Options: map[string]interface{}{":id": task},
		Target:  &fileList,
	}

	err := d.Handle(req)
	if err != nil {
		return []string{}, err
	}

	return fileList, nil
}

func (d *Fetcher) DownloadArtefactsFromTask(taskid, target string, filters []string) error {
	return d.DownloadArtefactsGeneric(taskid, target, "artefact", filters)
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
		fetcher.AppendTaskOutput("Uploading " + path + " to " + rel + "\n")
		err = fetcher.UploadArtefactRetry(path, rel, 5)
	}

	return err
}

func (d *Fetcher) DownloadArtefactsFromStorage(storage, target string) error {
	return d.DownloadArtefactsGeneric(storage, target, "storage", []string{})
}

func (d *Fetcher) DownloadArtefactsGeneric(id, target, artefact_type string, filters []string) error {
	var list []string
	var err error
	var to_download string
	var filterRegexp []*regexp.Regexp = make([]*regexp.Regexp, 0)

	for _, filter := range filters {
		r, e := regexp.Compile(filter)
		if e != nil {
			d.AppendTaskOutput("Failed compiling regex (" + filter + "):" + e.Error() + "\n")
			return err
		}
		filterRegexp = append(filterRegexp, r)
	}

	if artefact_type == "namespace" {
		list, err = d.NamespaceFileList(id)
		if err != nil {
			d.AppendTaskOutput("[Download] Failed getting namespace artefact list" + "\n")
			return err
		}
		to_download = id
	} else if artefact_type == "storage" {
		list, err = d.StorageFileList(id)
		if err != nil {
			d.AppendTaskOutput("[Download] Failed getting storage file list" + "\n")
			return err
		}
		var storage_data storageci.Storage

		req := schema.Request{
			Route:   v1.Schema.GetStorageRoute("show"),
			Options: map[string]interface{}{":id": id},
			Target:  &storage_data,
		}

		err := d.Handle(req)
		if err != nil {
			d.AppendTaskOutput("Downloading failed during retrieveing storage data : " + err.Error() + "\n")
			return err
		}

		to_download = storage_data.Path

	} else if artefact_type == "artefact" {
		list, err = d.TaskFileList(id)
		if err != nil {
			d.AppendTaskOutput("[Download] Failed getting task artefacts list" + "\n")
			return err
		}
		to_download = id
	}

	err = os.MkdirAll(target, os.ModePerm)
	if err != nil {
		d.AppendTaskOutput("[Download] Error: " + err.Error() + "\n")
		return err
	}
	d.AppendTaskOutput("[Download] " + artefact_type + " artefacts from " + id + "\n")
	success := true
	for _, file := range list {
		trials := 5
		done := true
		skipped := false
		if len(filters) > 0 {
			skipped = true
			// Check if artefacts match with filter
			for _, filter := range filterRegexp {
				if filter.MatchString(file) {
					skipped = false
					break
				}
			}
		}

		if skipped {
			d.AppendTaskOutput("[Download] File " + file + " filtered." + "\n")
			continue
		}

		reldir, _ := filepath.Split(file)
		for done {

			if trials < 0 {
				done = false
				success = false
			}
			err := os.MkdirAll(filepath.Join(target, reldir), os.ModePerm)
			if err != nil {
				d.AppendTaskOutput("[Download] Error: " + err.Error() + "\n")
				return err
			}
			location := d.BaseURL + "/" + artefact_type + "/" + to_download + utils.PathEscape(file)

			d.AppendTaskOutput("[Download]  " + location + " to " + filepath.Join(target, file) + "\n")
			if ok, err := d.Download(location, filepath.Join(target, file)); !ok {
				d.AppendTaskOutput("[Download] failed : " + err.Error() + "\n")
				trials--
			} else {
				done = false
				d.AppendTaskOutput("[Download] succeeded " + "\n")
			}

		}

	}

	if !success {
		return errors.New("Download failed")
	}

	return nil
}

func (d *Fetcher) DownloadArtefactsFromNamespace(namespace, target string, filters []string) error {
	return d.DownloadArtefactsGeneric(namespace, target, "namespace", filters)
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
		d.AppendTaskOutput("Download with bandwidth limit of: " + strconv.FormatInt(1024*d.Config.GetAgent().DownloadRateLimit, 10) + "\n")
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

	req := schema.Request{
		Route: v1.Schema.GetStorageRoute("upload"),
		Options: map[string]interface{}{
			"name":      file,
			"path":      relativepath,
			"storageid": storageid,
		},
	}

	if err := f.HandleUploadLargeFile(req, "file", fullpath, f.ChunkSize); err != nil {
		f.AppendTaskOutput("[Upload] Error while uploading artefact " + file + ": " + err.Error() + "\n")
		return err
	}
	return nil
}

func (f *Fetcher) UploadArtefactRetry(fullpath, relativepath string, trials int) error {
	trial := 1
	err := f.UploadArtefact(fullpath, relativepath)
	for err != nil && trial < trials {
		trial++
		f.AppendTaskOutput("[Upload] Trial " + strconv.Itoa(trial) + "\n")
		err = f.UploadArtefact(fullpath, relativepath)
	}
	return err
}

func (f *Fetcher) UploadArtefact(fullpath, relativepath string) error {
	_, file := filepath.Split(fullpath)

	req := schema.Request{
		Route: v1.Schema.GetTaskRoute("artefact_upload"),
		Options: map[string]interface{}{
			"name":   file,
			"path":   relativepath,
			"taskid": f.docID,
		},
	}

	if err := f.HandleUploadLargeFile(req, "file", fullpath, f.ChunkSize); err != nil {
		f.AppendTaskOutput("[Upload] Error while uploading artefact " + file + ": " + err.Error() + "\n")
		return err
	}
	return nil
}

func (f *Fetcher) UploadNamespaceFile(namespace, fullpath, relativepath string) error {
	_, file := filepath.Split(fullpath)

	req := schema.Request{
		Route: v1.Schema.GetNamespaceRoute("upload"),
		Options: map[string]interface{}{
			"name":      file,
			"path":      relativepath,
			"namespace": namespace,
		},
	}

	if err := f.HandleUploadLargeFile(req, "file", fullpath, f.ChunkSize); err != nil {
		return err
	}
	return nil
}
