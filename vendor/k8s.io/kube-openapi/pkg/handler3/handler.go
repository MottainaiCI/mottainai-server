/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package handler3

import (
	"bytes"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
<<<<<<< HEAD
	openapi_v3 "github.com/google/gnostic-models/openapiv3"
	"github.com/google/uuid"
	"github.com/munnerz/goautoneg"

=======
	openapi_v3 "github.com/google/gnostic/openapiv3"
	"github.com/google/uuid"
	"github.com/munnerz/goautoneg"
>>>>>>> fe31cef4 (Update vendor github.com/MottainaiCI/lxd-compose@d928eed0eddfde18d58fe3a8ae780328c1b0d55c)
	"k8s.io/klog/v2"
	"k8s.io/kube-openapi/pkg/cached"
	"k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/spec3"
)

const (
	subTypeProtobufDeprecated = "com.github.proto-openapi.spec.v3@v1.0+protobuf"
	subTypeProtobuf           = "com.github.proto-openapi.spec.v3.v1.0+protobuf"
	subTypeJSON               = "json"
)

// OpenAPIV3Discovery is the format of the Discovery document for OpenAPI V3
// It maps Discovery paths to their corresponding URLs with a hash parameter included
type OpenAPIV3Discovery struct {
	Paths map[string]OpenAPIV3DiscoveryGroupVersion `json:"paths"`
}

// OpenAPIV3DiscoveryGroupVersion includes information about a group version and URL
// for accessing the OpenAPI. The URL includes a hash parameter to support client side caching
type OpenAPIV3DiscoveryGroupVersion struct {
	// Path is an absolute path of an OpenAPI V3 document in the form of /openapi/v3/apis/apps/v1?hash=014fbff9a07c
	ServerRelativeURL string `json:"serverRelativeURL"`
}

func ToV3ProtoBinary(json []byte) ([]byte, error) {
	document, err := openapi_v3.ParseDocument(json)
	if err != nil {
		return nil, err
	}
	return proto.Marshal(document)
}

type timedSpec struct {
	spec         []byte
	lastModified time.Time
}

// This type is protected by the lock on OpenAPIService.
type openAPIV3Group struct {
<<<<<<< HEAD
	specCache cached.LastSuccess[*spec3.OpenAPI]
	pbCache   cached.Value[timedSpec]
	jsonCache cached.Value[timedSpec]
=======
	specCache cached.Replaceable[*spec3.OpenAPI]
	pbCache   cached.Data[timedSpec]
	jsonCache cached.Data[timedSpec]
>>>>>>> fe31cef4 (Update vendor github.com/MottainaiCI/lxd-compose@d928eed0eddfde18d58fe3a8ae780328c1b0d55c)
}

func newOpenAPIV3Group() *openAPIV3Group {
	o := &openAPIV3Group{}
<<<<<<< HEAD
	o.jsonCache = cached.Transform[*spec3.OpenAPI](func(spec *spec3.OpenAPI, etag string, err error) (timedSpec, string, error) {
		if err != nil {
			return timedSpec{}, "", err
		}
		json, err := json.Marshal(spec)
		if err != nil {
			return timedSpec{}, "", err
		}
		return timedSpec{spec: json, lastModified: time.Now()}, computeETag(json), nil
	}, &o.specCache)
	o.pbCache = cached.Transform(func(ts timedSpec, etag string, err error) (timedSpec, string, error) {
		if err != nil {
			return timedSpec{}, "", err
		}
		proto, err := ToV3ProtoBinary(ts.spec)
		if err != nil {
			return timedSpec{}, "", err
		}
		return timedSpec{spec: proto, lastModified: ts.lastModified}, etag, nil
=======
	o.jsonCache = cached.NewTransformer[*spec3.OpenAPI](func(result cached.Result[*spec3.OpenAPI]) cached.Result[timedSpec] {
		if result.Err != nil {
			return cached.NewResultErr[timedSpec](result.Err)
		}
		json, err := json.Marshal(result.Data)
		if err != nil {
			return cached.NewResultErr[timedSpec](err)
		}
		return cached.NewResultOK(timedSpec{spec: json, lastModified: time.Now()}, computeETag(json))
	}, &o.specCache)
	o.pbCache = cached.NewTransformer(func(result cached.Result[timedSpec]) cached.Result[timedSpec] {
		if result.Err != nil {
			return cached.NewResultErr[timedSpec](result.Err)
		}
		proto, err := ToV3ProtoBinary(result.Data.spec)
		if err != nil {
			return cached.NewResultErr[timedSpec](err)
		}
		return cached.NewResultOK(timedSpec{spec: proto, lastModified: result.Data.lastModified}, result.Etag)
>>>>>>> fe31cef4 (Update vendor github.com/MottainaiCI/lxd-compose@d928eed0eddfde18d58fe3a8ae780328c1b0d55c)
	}, o.jsonCache)
	return o
}

<<<<<<< HEAD
func (o *openAPIV3Group) UpdateSpec(openapi cached.Value[*spec3.OpenAPI]) {
	o.specCache.Store(openapi)
=======
func (o *openAPIV3Group) UpdateSpec(openapi cached.Data[*spec3.OpenAPI]) {
	o.specCache.Replace(openapi)
>>>>>>> fe31cef4 (Update vendor github.com/MottainaiCI/lxd-compose@d928eed0eddfde18d58fe3a8ae780328c1b0d55c)
}

// OpenAPIService is the service responsible for serving OpenAPI spec. It has
// the ability to safely change the spec while serving it.
type OpenAPIService struct {
	// Mutex protects the schema map.
	mutex    sync.Mutex
	v3Schema map[string]*openAPIV3Group

<<<<<<< HEAD
	discoveryCache cached.LastSuccess[timedSpec]
=======
	discoveryCache cached.Replaceable[timedSpec]
>>>>>>> fe31cef4 (Update vendor github.com/MottainaiCI/lxd-compose@d928eed0eddfde18d58fe3a8ae780328c1b0d55c)
}

func computeETag(data []byte) string {
	if data == nil {
		return ""
	}
	return fmt.Sprintf("%X", sha512.Sum512(data))
}

func constructServerRelativeURL(gvString, etag string) string {
	u := url.URL{Path: path.Join("/openapi/v3", gvString)}
	query := url.Values{}
	query.Set("hash", etag)
	u.RawQuery = query.Encode()
	return u.String()
}

// NewOpenAPIService builds an OpenAPIService starting with the given spec.
func NewOpenAPIService() *OpenAPIService {
	o := &OpenAPIService{}
	o.v3Schema = make(map[string]*openAPIV3Group)
	// We're not locked because we haven't shared the structure yet.
<<<<<<< HEAD
	o.discoveryCache.Store(o.buildDiscoveryCacheLocked())
	return o
}

func (o *OpenAPIService) buildDiscoveryCacheLocked() cached.Value[timedSpec] {
	caches := make(map[string]cached.Value[timedSpec], len(o.v3Schema))
	for gvName, group := range o.v3Schema {
		caches[gvName] = group.jsonCache
	}
	return cached.Merge(func(results map[string]cached.Result[timedSpec]) (timedSpec, string, error) {
		discovery := &OpenAPIV3Discovery{Paths: make(map[string]OpenAPIV3DiscoveryGroupVersion)}
		for gvName, result := range results {
			if result.Err != nil {
				return timedSpec{}, "", result.Err
=======
	o.discoveryCache.Replace(o.buildDiscoveryCacheLocked())
	return o
}

func (o *OpenAPIService) buildDiscoveryCacheLocked() cached.Data[timedSpec] {
	caches := make(map[string]cached.Data[timedSpec], len(o.v3Schema))
	for gvName, group := range o.v3Schema {
		caches[gvName] = group.jsonCache
	}
	return cached.NewMerger(func(results map[string]cached.Result[timedSpec]) cached.Result[timedSpec] {
		discovery := &OpenAPIV3Discovery{Paths: make(map[string]OpenAPIV3DiscoveryGroupVersion)}
		for gvName, result := range results {
			if result.Err != nil {
				return cached.NewResultErr[timedSpec](result.Err)
>>>>>>> fe31cef4 (Update vendor github.com/MottainaiCI/lxd-compose@d928eed0eddfde18d58fe3a8ae780328c1b0d55c)
			}
			discovery.Paths[gvName] = OpenAPIV3DiscoveryGroupVersion{
				ServerRelativeURL: constructServerRelativeURL(gvName, result.Etag),
			}
		}
		j, err := json.Marshal(discovery)
		if err != nil {
<<<<<<< HEAD
			return timedSpec{}, "", err
		}
		return timedSpec{spec: j, lastModified: time.Now()}, computeETag(j), nil
=======
			return cached.NewResultErr[timedSpec](err)
		}
		return cached.NewResultOK(timedSpec{spec: j, lastModified: time.Now()}, computeETag(j))
>>>>>>> fe31cef4 (Update vendor github.com/MottainaiCI/lxd-compose@d928eed0eddfde18d58fe3a8ae780328c1b0d55c)
	}, caches)
}

func (o *OpenAPIService) getSingleGroupBytes(getType string, group string) ([]byte, string, time.Time, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	v, ok := o.v3Schema[group]
	if !ok {
		return nil, "", time.Now(), fmt.Errorf("Cannot find CRD group %s", group)
	}
<<<<<<< HEAD
	switch getType {
	case subTypeJSON:
		ts, etag, err := v.jsonCache.Get()
		return ts.spec, etag, ts.lastModified, err
	case subTypeProtobuf, subTypeProtobufDeprecated:
		ts, etag, err := v.pbCache.Get()
		return ts.spec, etag, ts.lastModified, err
	default:
		return nil, "", time.Now(), fmt.Errorf("Invalid accept clause %s", getType)
	}
}

// UpdateGroupVersionLazy adds or updates an existing group with the new cached.
func (o *OpenAPIService) UpdateGroupVersionLazy(group string, openapi cached.Value[*spec3.OpenAPI]) {
=======
	result := cached.Result[timedSpec]{}
	switch getType {
	case subTypeJSON:
		result = v.jsonCache.Get()
	case subTypeProtobuf, subTypeProtobufDeprecated:
		result = v.pbCache.Get()
	default:
		return nil, "", time.Now(), fmt.Errorf("Invalid accept clause %s", getType)
	}
	return result.Data.spec, result.Etag, result.Data.lastModified, result.Err
}

// UpdateGroupVersionLazy adds or updates an existing group with the new cached.
func (o *OpenAPIService) UpdateGroupVersionLazy(group string, openapi cached.Data[*spec3.OpenAPI]) {
>>>>>>> fe31cef4 (Update vendor github.com/MottainaiCI/lxd-compose@d928eed0eddfde18d58fe3a8ae780328c1b0d55c)
	o.mutex.Lock()
	defer o.mutex.Unlock()
	if _, ok := o.v3Schema[group]; !ok {
		o.v3Schema[group] = newOpenAPIV3Group()
		// Since there is a new item, we need to re-build the cache map.
<<<<<<< HEAD
		o.discoveryCache.Store(o.buildDiscoveryCacheLocked())
=======
		o.discoveryCache.Replace(o.buildDiscoveryCacheLocked())
>>>>>>> fe31cef4 (Update vendor github.com/MottainaiCI/lxd-compose@d928eed0eddfde18d58fe3a8ae780328c1b0d55c)
	}
	o.v3Schema[group].UpdateSpec(openapi)
}

func (o *OpenAPIService) UpdateGroupVersion(group string, openapi *spec3.OpenAPI) {
<<<<<<< HEAD
	o.UpdateGroupVersionLazy(group, cached.Static(openapi, uuid.New().String()))
=======
	o.UpdateGroupVersionLazy(group, cached.NewResultOK(openapi, uuid.New().String()))
>>>>>>> fe31cef4 (Update vendor github.com/MottainaiCI/lxd-compose@d928eed0eddfde18d58fe3a8ae780328c1b0d55c)
}

func (o *OpenAPIService) DeleteGroupVersion(group string) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	delete(o.v3Schema, group)
	// Rebuild the merge cache map since the items have changed.
<<<<<<< HEAD
	o.discoveryCache.Store(o.buildDiscoveryCacheLocked())
}

func (o *OpenAPIService) HandleDiscovery(w http.ResponseWriter, r *http.Request) {
	ts, etag, err := o.discoveryCache.Get()
	if err != nil {
		klog.Errorf("Error serving discovery: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Etag", strconv.Quote(etag))
	w.Header().Set("Content-Type", "application/json")
	http.ServeContent(w, r, "/openapi/v3", ts.lastModified, bytes.NewReader(ts.spec))
=======
	o.discoveryCache.Replace(o.buildDiscoveryCacheLocked())
}

func (o *OpenAPIService) HandleDiscovery(w http.ResponseWriter, r *http.Request) {
	result := o.discoveryCache.Get()
	if result.Err != nil {
		klog.Errorf("Error serving discovery: %s", result.Err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Etag", strconv.Quote(result.Etag))
	w.Header().Set("Content-Type", "application/json")
	http.ServeContent(w, r, "/openapi/v3", result.Data.lastModified, bytes.NewReader(result.Data.spec))
>>>>>>> fe31cef4 (Update vendor github.com/MottainaiCI/lxd-compose@d928eed0eddfde18d58fe3a8ae780328c1b0d55c)
}

func (o *OpenAPIService) HandleGroupVersion(w http.ResponseWriter, r *http.Request) {
	url := strings.SplitAfterN(r.URL.Path, "/", 4)
	group := url[3]

	decipherableFormats := r.Header.Get("Accept")
	if decipherableFormats == "" {
		decipherableFormats = "*/*"
	}
	clauses := goautoneg.ParseAccept(decipherableFormats)
	w.Header().Add("Vary", "Accept")

	if len(clauses) == 0 {
		return
	}

	accepted := []struct {
		Type                string
		SubType             string
		ReturnedContentType string
	}{
		{"application", subTypeJSON, "application/" + subTypeJSON},
		{"application", subTypeProtobuf, "application/" + subTypeProtobuf},
		{"application", subTypeProtobufDeprecated, "application/" + subTypeProtobuf},
	}

	for _, clause := range clauses {
		for _, accepts := range accepted {
			if clause.Type != accepts.Type && clause.Type != "*" {
				continue
			}
			if clause.SubType != accepts.SubType && clause.SubType != "*" {
				continue
			}
			data, etag, lastModified, err := o.getSingleGroupBytes(accepts.SubType, group)
			if err != nil {
				return
			}
			// Set Content-Type header in the reponse
			w.Header().Set("Content-Type", accepts.ReturnedContentType)

			// ETag must be enclosed in double quotes: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag
			w.Header().Set("Etag", strconv.Quote(etag))

			if hash := r.URL.Query().Get("hash"); hash != "" {
				if hash != etag {
					u := constructServerRelativeURL(group, etag)
					http.Redirect(w, r, u, 301)
					return
				}
				// The Vary header is required because the Accept header can
				// change the contents returned. This prevents clients from caching
				// protobuf as JSON and vice versa.
				w.Header().Set("Vary", "Accept")

				// Only set these headers when a hash is given.
				w.Header().Set("Cache-Control", "public, immutable")
				// Set the Expires directive to the maximum value of one year from the request,
				// effectively indicating that the cache never expires.
				w.Header().Set("Expires", time.Now().AddDate(1, 0, 0).Format(time.RFC1123))
			}
			http.ServeContent(w, r, "", lastModified, bytes.NewReader(data))
			return
		}
	}
	w.WriteHeader(406)
	return
}

func (o *OpenAPIService) RegisterOpenAPIV3VersionedService(servePath string, handler common.PathHandlerByGroupVersion) error {
	handler.Handle(servePath, http.HandlerFunc(o.HandleDiscovery))
	handler.HandlePrefix(servePath+"/", http.HandlerFunc(o.HandleGroupVersion))
	return nil
}
