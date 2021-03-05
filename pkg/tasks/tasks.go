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

package agenttasks

import (
	"encoding/json"

	"github.com/ghodss/yaml"

	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"reflect"
	"strconv"
	"time"

	"github.com/MottainaiCI/mottainai-server/pkg/client"
	"github.com/MottainaiCI/mottainai-server/pkg/namespace"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/MottainaiCI/mottainai-server/pkg/utils"
	flock "github.com/theckman/go-flock"
)

type Task struct {
	ID string `json:"ID" form:"ID"` // ARMv7l overflows :(

	Name                string   `json:"name" form:"name"`
	Source              string   `json:"source" form:"source"`
	Script              []string `json:"script" form:"script"`
	Directory           string   `json:"directory" form:"directory"`
	Type                string   `json:"type" form:"type"`
	Status              string   `json:"status" form:"status"`
	Output              string   `json:"output" form:"output"`
	Result              string   `json:"result" form:"result"`
	Entrypoint          []string `json:"entrypoint" form:"entrypoint"`
	Namespace           string   `json:"namespace" form:"namespace"`
	Commit              string   `json:"commit" form:"commit"`
	PrivKey             string   `json:"privkey" form:"privkey"`
	AuthHosts           string   `json:"authhosts" form:"authhosts"`
	Node                string   `json:"node_id" form:"node_id"`
	Owner               string   `json:"owner_id" form:"owner_id"`
	Image               string   `json:"image" form:"image"`
	ExitStatus          string   `json:"exit_status" form:"exit_status"`
	Storage             string   `json:"storage" form:"storage"`
	ArtefactPath        string   `json:"artefact_path" form:"artefact_path"`
	ArtefactPushFilters []string `json:"artefact_push_filters" form:"artefact_push_filters"`
	StoragePath         string   `json:"storage_path" form:"storage_path"`
	RootTask            string   `json:"root_task" form:"root_task"`
	Prune               string   `json:"prune" form:"prune"`
	CacheImage          string   `json:"cache_image" form:"cache_image"`
	CacheClean          string   `json:"cache_clean" form:"cache_clean"`
	PublishMode         string   `json:"publish_mode" form:"publish_mode"`
	PipelineID          string   `json:"pipeline_id" form:"pipeline_id"`

	NamespaceMerged  string   `json:"namespace_merged" form:"namespace_merged"`
	NamespaceFilters []string `json:"namespace_filters" form:"namespace_filters"`
	TagNamespace     string   `json:"tag_namespace" form:"tag_namespace"`

	CreatedTime string `json:"created_time" form:"created_time"`
	StartTime   string `json:"start_time" form:"start_time"`
	EndTime     string `json:"end_time" form:"end_time"`
	UpdatedTime string `json:"last_update_time" form:"last_update_time"`
	Queue       string `json:"queue" form:"queue"`
	Retry       string `json:"retry" form:"retry"`

	Delayed     string   `json:"eta" form:"eta"`
	TimeOut     float64  `json:"timeout" form:"timeout"`
	Binds       []string `json:"binds" form:"binds"`
	Environment []string `json:"environment" form:"environment"`

	Quota string `json:"quota" form:"quota"`
}

type Plan struct {
	*Task
	Planned string `json:"planned" form:"planned"`
}

func NewPlanFromMap(t map[string]interface{}) Plan {
	tk := NewTaskFromMap(t)
	var planned string
	if str, ok := t["planned"].(string); ok {
		planned = str
	}
	pl := Plan{Task: &tk}
	pl.Planned = planned
	return pl
}

func (t *Plan) ToMap() map[string]interface{} {

	ts := make(map[string]interface{})
	val := reflect.ValueOf(t).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		ts[tag.Get("form")] = valueField.Interface()
	}
	val = reflect.ValueOf(t.Task).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		ts[tag.Get("form")] = valueField.Interface()
		//fmt.Printf("Field Name: %s,\t Field Value: %v,\t Tag Value: %s\n", typeField.Name, valueField.Interface(), tag.Get("tag_name"))
	}
	return ts
}
func (t *Task) ToMap() map[string]interface{} {

	ts := make(map[string]interface{})
	val := reflect.ValueOf(t).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		ts[tag.Get("form")] = valueField.Interface()
		//fmt.Printf("Field Name: %s,\t Field Value: %v,\t Tag Value: %s\n", typeField.Name, valueField.Interface(), tag.Get("tag_name"))
	}
	return ts
}

func (t *Task) Trials() int {

	ret, err := strconv.Atoi(t.Retry)
	if err != nil {
		return 0
	}

	return ret
}

func NewPlanFromJson(data []byte) Plan {
	var t Plan
	json.Unmarshal(data, &t)
	return t
}

func NewTaskFromJson(data []byte) Task {
	var t Task
	json.Unmarshal(data, &t)
	return t
}

func FetchTask(fetcher client.HttpClient) (Task, error) {
	task_data, err := fetcher.GetTask()

	if err != nil {
		return Task{}, err
	}
	return NewTaskFromJson(task_data), nil
}

func NewTaskFromMap(t map[string]interface{}) Task {

	var (
		source            string
		script            []string
		directory         string
		namespace         string
		commit            string
		tasktype          string
		output            string
		image             string
		status            string
		result            string
		exit_status       string
		created_time      string
		start_time        string
		end_time          string
		last_update_time  string
		storage           string
		storage_path      string
		artefact_path     string
		quota             string
		root_task         string
		prune             string
		namespace_merged  string
		tag_namespace     string
		name              string
		cache_image       string
		cache_clean       string
		queue             string
		owner, node       string
		privkey           string
		environment       []string
		binds             []string
		namespace_filters []string
		artefact_pfilters []string
	)

	binds = make([]string, 0)
	environment = make([]string, 0)
	script = make([]string, 0)
	namespace_filters = make([]string, 0)
	artefact_pfilters = make([]string, 0)
	// Default mode maintains compatibility with first
	// implementation where merged namespace was the
	// logic
	namespace_merged = "true"

	if arr, ok := t["binds"].([]interface{}); ok {
		for _, v := range arr {
			binds = append(binds, v.(string))
		}
	}

	if arr, ok := t["environment"].([]interface{}); ok {
		for _, v := range arr {
			environment = append(environment, v.(string))
		}
	}

	if arr, ok := t["script"].([]interface{}); ok {
		for _, v := range arr {
			script = append(script, v.(string))
		}
	}
	if arr, ok := t["namespace_filters"].([]interface{}); ok {
		for _, v := range arr {
			namespace_filters = append(namespace_filters, v.(string))
		}
	}
	if arr, ok := t["artefact_push_filters"].([]interface{}); ok {
		for _, v := range arr {
			artefact_pfilters = append(artefact_pfilters, v.(string))
		}
	}

	if i, ok := t["name"].(string); ok {
		name = i
	}
	if i, ok := t["owner_id"].(string); ok {
		owner = i
	}
	if i, ok := t["node_id"].(string); ok {
		node = i
	}
	if str, ok := t["queue"].(string); ok {
		queue = str
	}
	if str, ok := t["root_task"].(string); ok {
		root_task = str
	}
	if str, ok := t["exit_status"].(string); ok {
		exit_status = str
	}
	if str, ok := t["source"].(string); ok {
		source = str
	}
	if str, ok := t["privkey"].(string); ok {
		privkey = str
	}
	if str, ok := t["directory"].(string); ok {
		directory = str
	}
	if str, ok := t["type"].(string); ok {
		tasktype = str
	}
	if str, ok := t["quota"].(string); ok {
		quota = str
	}
	if str, ok := t["task"].(string); ok {
		tasktype = str
	}
	if str, ok := t["namespace"].(string); ok {
		namespace = str
	}
	if str, ok := t["commit"].(string); ok {
		commit = str
	}
	if str, ok := t["output"].(string); ok {
		output = str
	}
	if str, ok := t["result"].(string); ok {
		result = str
	}
	if str, ok := t["cache_clean"].(string); ok {
		cache_clean = str
	}
	if str, ok := t["tag_namespace"].(string); ok {
		tag_namespace = str
	}
	if str, ok := t["namespace_merged"].(string); ok {
		namespace_merged = str
	}
	if str, ok := t["status"].(string); ok {
		status = str
	}
	if str, ok := t["image"].(string); ok {
		image = str
	}
	if str, ok := t["storage"].(string); ok {
		storage = str
	}
	if str, ok := t["created_time"].(string); ok {
		created_time = str
	}
	if str, ok := t["start_time"].(string); ok {
		start_time = str
	}
	if str, ok := t["last_update_time"].(string); ok {
		last_update_time = str
	}
	if str, ok := t["end_time"].(string); ok {
		end_time = str
	}
	if str, ok := t["storage_path"].(string); ok {
		storage_path = str
	}
	if str, ok := t["artefact_path"].(string); ok {
		artefact_path = str
	}
	if str, ok := t["prune"].(string); ok {
		prune = str
	}
	if str, ok := t["cache_image"].(string); ok {
		cache_image = str
	}

	var timeout float64
	if str, ok := t["timeout"].(float64); ok {
		timeout = str
	}
	var delayed string
	if str, ok := t["string"].(string); ok {
		delayed = str
	}
	var publish string
	if str, ok := t["publish_mode"].(string); ok {
		publish = str
	}
	var entrypoint []string
	entrypoint = make([]string, 0)

	if arr, ok := t["entrypoint"].([]interface{}); ok {
		for _, v := range arr {
			entrypoint = append(entrypoint, v.(string))
		}
	}

	var id string
	if str, ok := t["ID"].(string); ok {
		id = str
	}
	var pipelineId string
	if str, ok := t["pipeline_id"].(string); ok {
		pipelineId = str
	}

	var retry string
	if str, ok := t["retry"].(string); ok {
		retry = str
	}
	task := Task{
		Retry:               retry,
		ID:                  id,
		PipelineID:          pipelineId,
		Queue:               queue,
		Source:              source,
		PrivKey:             privkey,
		Script:              script,
		Quota:               quota,
		Delayed:             delayed,
		Directory:           directory,
		Type:                tasktype,
		Namespace:           namespace,
		NamespaceFilters:    namespace_filters,
		Commit:              commit,
		Name:                name,
		Entrypoint:          entrypoint,
		Output:              output,
		PublishMode:         publish,
		Result:              result,
		Status:              status,
		Storage:             storage,
		StoragePath:         storage_path,
		ArtefactPath:        artefact_path,
		ArtefactPushFilters: artefact_pfilters,
		Image:               image,
		ExitStatus:          exit_status,
		CreatedTime:         created_time,
		StartTime:           start_time,
		EndTime:             end_time,
		UpdatedTime:         last_update_time,
		RootTask:            root_task,
		NamespaceMerged:     namespace_merged,
		TagNamespace:        tag_namespace,
		Node:                node,
		Prune:               prune,
		CacheImage:          cache_image,
		Environment:         environment,
		Binds:               binds,
		CacheClean:          cache_clean,
		Owner:               owner,
		TimeOut:             timeout,
	}
	return task
}

func FromFile(file string) (*Task, error) {
	var t *Task
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return t, err
	}
	if err := json.Unmarshal(content, &t); err != nil {
		return t, err
	}
	return t, nil
}
func PlanFromYaml(file string) (*Plan, error) {
	var t *Plan
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return t, err
	}
	if err := yaml.Unmarshal(content, &t); err != nil {
		return t, err
	}
	return t, nil
}
func PlanFromJSON(file string) (*Plan, error) {
	var t *Plan
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return t, err
	}
	if err := json.Unmarshal(content, &t); err != nil {
		return t, err
	}
	return t, nil
}
func PipelineFromJSON(file string) (*Pipeline, error) {
	var t *Pipeline
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return t, err
	}
	if err := json.Unmarshal(content, &t); err != nil {
		return t, err
	}
	return t, nil
}

func FromYamlFile(file string) (*Task, error) {
	var t *Task
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return t, err
	}
	if err := yaml.Unmarshal(content, &t); err != nil {
		return t, err
	}
	return t, nil
}

func PipelineFromYaml(file string) (*Pipeline, error) {
	var t *Pipeline
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return t, err
	}
	if err := yaml.Unmarshal(content, &t); err != nil {
		return t, err
	}
	return t, nil
}

func (t *Task) Reset() {
	t.Output = ""
	t.Result = setting.TASK_RESULT_UNKNOWN
	t.Status = ""
	t.ExitStatus = ""
	t.CreatedTime = time.Now().Format("20060102150405")
	t.EndTime = ""
	t.UpdatedTime = ""
	t.Owner = ""
	t.Node = ""
	t.StartTime = ""
}

func (t *Task) IsOwner(id string) bool {

	if id == t.Owner {
		return true
	}

	return false
}
func (t *Task) Working() bool {

	if t.IsSetup() || t.IsRunning() {
		return true
	}
	return false
}
func (t *Task) IsSetup() bool {

	if t.Status == setting.TASK_STATE_SETUP {
		return true
	}
	return false
}

func (t *Task) IsRunning() bool {

	if t.Status == setting.TASK_STATE_RUNNING {
		return true
	}
	return false
}
func (t *Task) IsWaiting() bool {

	if t.Status == setting.TASK_STATE_WAIT {
		return true
	}
	return false
}

func (t *Task) ClearBuildLog(artefactPath string) {
	os.RemoveAll(path.Join(artefactPath, t.ID, "build_"+t.ID+".log"))
}

func (t *Task) Clear(artefactPath string, lockPath string) {
	os.RemoveAll(path.Join(artefactPath, t.ID))
	os.RemoveAll(path.Join(lockPath, t.ID+".lock"))
}

func (t *Task) GetLogPart(pos int, artefactPath string, lockPath string) string {
	var b3 []byte
	err := t.LockSection(func() error {
		file, err := os.Open(path.Join(artefactPath, t.ID, "build_"+t.ID+".log"))
		if err != nil {
			return err
		}
		_, err = file.Seek(int64(pos), 0)
		if err != nil {
			return err
		}
		fi, err := file.Stat()
		if err != nil {
			return err
		}

		b3 = make([]byte, fi.Size()-int64(pos))
		_, err = file.Read(b3)
		if err != nil {
			return err
		}

		return nil
	}, lockPath)

	if err != nil {
		return ""
	}
	return string(b3)
}

func (t *Task) TailLog(pos int, artefactPath string, lockPath string) string {
	var b3 []byte
	err := t.LockSection(func() error {
		file, err := os.Open(path.Join(artefactPath, t.ID, "build_"+t.ID+".log"))
		if err != nil {
			return err
		}

		fi, err := file.Stat()
		if err != nil {
			return err
		}

		where := fi.Size() - int64(pos)
		if int64(pos) > fi.Size() {
			where = 0
		}

		_, err = file.Seek(where, 0)
		if err != nil {
			return err
		}

		b3 = make([]byte, int64(pos))
		_, err = file.Read(b3)
		if err != nil {
			return err
		}

		return nil
	}, lockPath)

	if err != nil {
		return ""
	}
	return string(b3)
}

func (t *Task) LockSection(f func() error, lockPath string) error {
	os.MkdirAll(lockPath, os.ModePerm)
	lockfile := path.Join(lockPath, t.ID+".lock")
	fileLock := flock.NewFlock(lockfile)

	locked, err := fileLock.TryLock()
	if err != nil {
		return err
	}

	if locked {
		err := f()
		fileLock.Unlock()
		return err
	}
	return nil
}

func (t *Task) AppendBuildLog(s string, artefactPath string, lockPath string) error {

	os.MkdirAll(path.Join(artefactPath, t.ID), os.ModePerm)
	return t.LockSection(func() error {

		file, err := os.OpenFile(path.Join(artefactPath, t.ID, "build_"+t.ID+".log"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			return err
		}

		defer file.Close()

		if _, err = file.WriteString(s); err != nil {
			return err
		}
		return nil
	}, lockPath)

}

func (t *Task) IsStopped() bool {
	if t.Status == setting.TASK_STATE_STOPPED || t.Status == setting.TASK_STATE_ASK_STOP {
		return true
	}

	return false
}
func (t *Task) IsDone() bool {
	if t.Status == setting.TASK_STATE_DONE {
		return true
	}

	return false
}

func (t *Task) WantsClean() bool {
	if len(t.CacheClean) > 0 {
		return true
	}

	return false
}

func (t *Task) IsSuccess() bool {
	if t.ExitStatus == "0" {
		return true
	}

	return false
}

func (t *Task) IsNamespaceMerged() bool {
	if t.NamespaceMerged == "" || t.NamespaceMerged == "true" || t.NamespaceMerged == "yes" {
		return true
	}
	return false
}

func (t *Task) IsPublishAppendMode() bool {
	if t.PublishMode == setting.TASK_PUBLISH_MODE_APPEND {
		return true
	}

	return false
}

func (t *Task) HandleStatus(namespacePath string, artefactPath string) {
	if t.Status == setting.TASK_STATE_DONE {
		if t.ExitStatus == "0" {
			t.OnSuccess(namespacePath, artefactPath)
		} else {
			t.OnFailure()
		}
		t.Done()
	}
}

func (t *Task) DecodeStatus(state string) string {
	if state == "0" {
		return setting.TASK_RESULT_SUCCESS
	}

	return setting.TASK_RESULT_FAILED
}

func (t *Task) Artefacts(artefactPath string) []string {
	return utils.TreeList(filepath.Join(artefactPath, t.ID))
}

func (t *Task) Done() {
}

func (t *Task) OnFailure() {
}

func (t *Task) OnSuccess(namespacePath string, artefactPath string) {
	if len(t.TagNamespace) > 0 {
		ns := namespace.NewFromMap(map[string]interface{}{"name": t.TagNamespace, "path": t.TagNamespace})
		if t.IsPublishAppendMode() {
			ns.Append(t.ID, namespacePath, artefactPath)
		} else {
			ns.Tag(t.ID, namespacePath, artefactPath)
		}
	}
}
