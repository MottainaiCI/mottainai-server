// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protodesc

import (
<<<<<<< HEAD
=======
	_ "embed"
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	"fmt"
	"os"
	"sync"

<<<<<<< HEAD
	"google.golang.org/protobuf/internal/editiondefaults"
	"google.golang.org/protobuf/internal/filedesc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	gofeaturespb "google.golang.org/protobuf/types/gofeaturespb"
)

=======
	"google.golang.org/protobuf/internal/filedesc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	SupportedEditionsMinimum = descriptorpb.Edition_EDITION_PROTO2
	SupportedEditionsMaximum = descriptorpb.Edition_EDITION_2023
)

//go:embed editions_defaults.binpb
var binaryEditionDefaults []byte
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
var defaults = &descriptorpb.FeatureSetDefaults{}
var defaultsCacheMu sync.Mutex
var defaultsCache = make(map[filedesc.Edition]*descriptorpb.FeatureSet)

func init() {
<<<<<<< HEAD
	err := proto.Unmarshal(editiondefaults.Defaults, defaults)
=======
	err := proto.Unmarshal(binaryEditionDefaults, defaults)
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unmarshal editions defaults: %v\n", err)
		os.Exit(1)
	}
}

func fromEditionProto(epb descriptorpb.Edition) filedesc.Edition {
	return filedesc.Edition(epb)
}

func toEditionProto(ed filedesc.Edition) descriptorpb.Edition {
	switch ed {
	case filedesc.EditionUnknown:
		return descriptorpb.Edition_EDITION_UNKNOWN
	case filedesc.EditionProto2:
		return descriptorpb.Edition_EDITION_PROTO2
	case filedesc.EditionProto3:
		return descriptorpb.Edition_EDITION_PROTO3
	case filedesc.Edition2023:
		return descriptorpb.Edition_EDITION_2023
	default:
		panic(fmt.Sprintf("unknown value for edition: %v", ed))
	}
}

func getFeatureSetFor(ed filedesc.Edition) *descriptorpb.FeatureSet {
	defaultsCacheMu.Lock()
	defer defaultsCacheMu.Unlock()
	if def, ok := defaultsCache[ed]; ok {
		return def
	}
	edpb := toEditionProto(ed)
	if defaults.GetMinimumEdition() > edpb || defaults.GetMaximumEdition() < edpb {
		// This should never happen protodesc.(FileOptions).New would fail when
		// initializing the file descriptor.
		// This most likely means the embedded defaults were not updated.
		fmt.Fprintf(os.Stderr, "internal error: unsupported edition %v (did you forget to update the embedded defaults (i.e. the bootstrap descriptor proto)?)\n", edpb)
		os.Exit(1)
	}
<<<<<<< HEAD
	fsed := defaults.GetDefaults()[0]
=======
	fs := defaults.GetDefaults()[0].GetFeatures()
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	// Using a linear search for now.
	// Editions are guaranteed to be sorted and thus we could use a binary search.
	// Given that there are only a handful of editions (with one more per year)
	// there is not much reason to use a binary search.
	for _, def := range defaults.GetDefaults() {
		if def.GetEdition() <= edpb {
<<<<<<< HEAD
			fsed = def
=======
			fs = def.GetFeatures()
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
		} else {
			break
		}
	}
<<<<<<< HEAD
	fs := proto.Clone(fsed.GetFixedFeatures()).(*descriptorpb.FeatureSet)
	proto.Merge(fs, fsed.GetOverridableFeatures())
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	defaultsCache[ed] = fs
	return fs
}

<<<<<<< HEAD
// mergeEditionFeatures merges the parent and child feature sets. This function
// should be used when initializing Go descriptors from descriptor protos which
// is why the parent is a filedesc.EditionsFeatures (Go representation) while
// the child is a descriptorproto.FeatureSet (protoc representation).
// Any feature set by the child overwrites what is set by the parent.
func mergeEditionFeatures(parentDesc protoreflect.Descriptor, child *descriptorpb.FeatureSet) filedesc.EditionFeatures {
	var parentFS filedesc.EditionFeatures
	switch p := parentDesc.(type) {
	case *filedesc.File:
		parentFS = p.L1.EditionFeatures
	case *filedesc.Message:
		parentFS = p.L1.EditionFeatures
	default:
		panic(fmt.Sprintf("unknown parent type %T", parentDesc))
	}
	if child == nil {
		return parentFS
	}
	if fp := child.FieldPresence; fp != nil {
		parentFS.IsFieldPresence = *fp == descriptorpb.FeatureSet_LEGACY_REQUIRED ||
			*fp == descriptorpb.FeatureSet_EXPLICIT
		parentFS.IsLegacyRequired = *fp == descriptorpb.FeatureSet_LEGACY_REQUIRED
	}
	if et := child.EnumType; et != nil {
		parentFS.IsOpenEnum = *et == descriptorpb.FeatureSet_OPEN
	}

	if rfe := child.RepeatedFieldEncoding; rfe != nil {
		parentFS.IsPacked = *rfe == descriptorpb.FeatureSet_PACKED
	}

	if utf8val := child.Utf8Validation; utf8val != nil {
		parentFS.IsUTF8Validated = *utf8val == descriptorpb.FeatureSet_VERIFY
	}

	if me := child.MessageEncoding; me != nil {
		parentFS.IsDelimitedEncoded = *me == descriptorpb.FeatureSet_DELIMITED
	}

	if jf := child.JsonFormat; jf != nil {
		parentFS.IsJSONCompliant = *jf == descriptorpb.FeatureSet_ALLOW
	}

	if goFeatures, ok := proto.GetExtension(child, gofeaturespb.E_Go).(*gofeaturespb.GoFeatures); ok && goFeatures != nil {
		if luje := goFeatures.LegacyUnmarshalJsonEnum; luje != nil {
			parentFS.GenerateLegacyUnmarshalJSON = *luje
		}
	}

	return parentFS
=======
func resolveFeatureHasFieldPresence(fileDesc *filedesc.File, fieldDesc *descriptorpb.FieldDescriptorProto) bool {
	fs := fieldDesc.GetOptions().GetFeatures()
	if fs == nil || fs.FieldPresence == nil {
		return fileDesc.L1.EditionFeatures.IsFieldPresence
	}
	return fs.GetFieldPresence() == descriptorpb.FeatureSet_LEGACY_REQUIRED ||
		fs.GetFieldPresence() == descriptorpb.FeatureSet_EXPLICIT
}

func resolveFeatureRepeatedFieldEncodingPacked(fileDesc *filedesc.File, fieldDesc *descriptorpb.FieldDescriptorProto) bool {
	fs := fieldDesc.GetOptions().GetFeatures()
	if fs == nil || fs.RepeatedFieldEncoding == nil {
		return fileDesc.L1.EditionFeatures.IsPacked
	}
	return fs.GetRepeatedFieldEncoding() == descriptorpb.FeatureSet_PACKED
}

func resolveFeatureEnforceUTF8(fileDesc *filedesc.File, fieldDesc *descriptorpb.FieldDescriptorProto) bool {
	fs := fieldDesc.GetOptions().GetFeatures()
	if fs == nil || fs.Utf8Validation == nil {
		return fileDesc.L1.EditionFeatures.IsUTF8Validated
	}
	return fs.GetUtf8Validation() == descriptorpb.FeatureSet_VERIFY
}

func resolveFeatureDelimitedEncoding(fileDesc *filedesc.File, fieldDesc *descriptorpb.FieldDescriptorProto) bool {
	fs := fieldDesc.GetOptions().GetFeatures()
	if fs == nil || fs.MessageEncoding == nil {
		return fileDesc.L1.EditionFeatures.IsDelimitedEncoded
	}
	return fs.GetMessageEncoding() == descriptorpb.FeatureSet_DELIMITED
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
}

// initFileDescFromFeatureSet initializes editions related fields in fd based
// on fs. If fs is nil it is assumed to be an empty featureset and all fields
// will be initialized with the appropriate default. fd.L1.Edition must be set
// before calling this function.
func initFileDescFromFeatureSet(fd *filedesc.File, fs *descriptorpb.FeatureSet) {
	dfs := getFeatureSetFor(fd.L1.Edition)
<<<<<<< HEAD
	// initialize the featureset with the defaults
	fd.L1.EditionFeatures = mergeEditionFeatures(fd, dfs)
	// overwrite any options explicitly specified
	fd.L1.EditionFeatures = mergeEditionFeatures(fd, fs)
=======
	if fs == nil {
		fs = &descriptorpb.FeatureSet{}
	}

	var fieldPresence descriptorpb.FeatureSet_FieldPresence
	if fp := fs.FieldPresence; fp != nil {
		fieldPresence = *fp
	} else {
		fieldPresence = *dfs.FieldPresence
	}
	fd.L1.EditionFeatures.IsFieldPresence = fieldPresence == descriptorpb.FeatureSet_LEGACY_REQUIRED ||
		fieldPresence == descriptorpb.FeatureSet_EXPLICIT

	var enumType descriptorpb.FeatureSet_EnumType
	if et := fs.EnumType; et != nil {
		enumType = *et
	} else {
		enumType = *dfs.EnumType
	}
	fd.L1.EditionFeatures.IsOpenEnum = enumType == descriptorpb.FeatureSet_OPEN

	var respeatedFieldEncoding descriptorpb.FeatureSet_RepeatedFieldEncoding
	if rfe := fs.RepeatedFieldEncoding; rfe != nil {
		respeatedFieldEncoding = *rfe
	} else {
		respeatedFieldEncoding = *dfs.RepeatedFieldEncoding
	}
	fd.L1.EditionFeatures.IsPacked = respeatedFieldEncoding == descriptorpb.FeatureSet_PACKED

	var isUTF8Validated descriptorpb.FeatureSet_Utf8Validation
	if utf8val := fs.Utf8Validation; utf8val != nil {
		isUTF8Validated = *utf8val
	} else {
		isUTF8Validated = *dfs.Utf8Validation
	}
	fd.L1.EditionFeatures.IsUTF8Validated = isUTF8Validated == descriptorpb.FeatureSet_VERIFY

	var messageEncoding descriptorpb.FeatureSet_MessageEncoding
	if me := fs.MessageEncoding; me != nil {
		messageEncoding = *me
	} else {
		messageEncoding = *dfs.MessageEncoding
	}
	fd.L1.EditionFeatures.IsDelimitedEncoded = messageEncoding == descriptorpb.FeatureSet_DELIMITED

	var jsonFormat descriptorpb.FeatureSet_JsonFormat
	if jf := fs.JsonFormat; jf != nil {
		jsonFormat = *jf
	} else {
		jsonFormat = *dfs.JsonFormat
	}
	fd.L1.EditionFeatures.IsJSONCompliant = jsonFormat == descriptorpb.FeatureSet_ALLOW
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
}
