// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Indexed package import.
// See cmd/compile/internal/gc/iexport.go for the export data format.

// This file is a copy of $GOROOT/src/go/internal/gcimporter/iimport.go.

package gcimporter

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"go/constant"
	"go/token"
	"go/types"
	"io"
	"math/big"
	"sort"
	"strings"

<<<<<<< HEAD
<<<<<<< HEAD
	"golang.org/x/tools/go/types/objectpath"
	"golang.org/x/tools/internal/aliases"
	"golang.org/x/tools/internal/typesinternal"
=======
	"golang.org/x/tools/internal/typeparams"
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
	"golang.org/x/tools/go/types/objectpath"
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
)

type intReader struct {
	*bytes.Reader
	path string
}

func (r *intReader) int64() int64 {
	i, err := binary.ReadVarint(r.Reader)
	if err != nil {
		errorf("import %q: read varint error: %v", r.path, err)
	}
	return i
}

func (r *intReader) uint64() uint64 {
	i, err := binary.ReadUvarint(r.Reader)
	if err != nil {
		errorf("import %q: read varint error: %v", r.path, err)
	}
	return i
}

// Keep this in sync with constants in iexport.go.
const (
	iexportVersionGo1_11   = 0
	iexportVersionPosCol   = 1
	iexportVersionGo1_18   = 2
	iexportVersionGenerics = 2

	iexportVersionCurrent = 2
)

type ident struct {
	pkg  *types.Package
	name string
}

const predeclReserved = 32

type itag uint64

const (
	// Types
	definedType itag = iota
	pointerType
	sliceType
	arrayType
	chanType
	mapType
	signatureType
	structType
	interfaceType
	typeParamType
	instanceType
	unionType
<<<<<<< HEAD
	aliasType
)

// Object tags
const (
	varTag          = 'V'
	funcTag         = 'F'
	genericFuncTag  = 'G'
	constTag        = 'C'
	aliasTag        = 'A'
	genericAliasTag = 'B'
	typeParamTag    = 'P'
	typeTag         = 'T'
	genericTypeTag  = 'U'
=======
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
)

// IImportData imports a package from the serialized package data
// and returns 0 and a reference to the package.
// If the export data version is not recognized or the format is otherwise
// compromised, an error is returned.
func IImportData(fset *token.FileSet, imports map[string]*types.Package, data []byte, path string) (int, *types.Package, error) {
<<<<<<< HEAD
<<<<<<< HEAD
	pkgs, err := iimportCommon(fset, GetPackagesFromMap(imports), data, false, path, false, nil)
=======
	pkgs, err := iimportCommon(fset, GetPackageFromMap(imports), data, false, path, nil)
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
	pkgs, err := iimportCommon(fset, GetPackagesFromMap(imports), data, false, path, false, nil)
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	if err != nil {
		return 0, nil, err
	}
	return 0, pkgs[0], nil
}

// IImportBundle imports a set of packages from the serialized package bundle.
func IImportBundle(fset *token.FileSet, imports map[string]*types.Package, data []byte) ([]*types.Package, error) {
<<<<<<< HEAD
<<<<<<< HEAD
	return iimportCommon(fset, GetPackagesFromMap(imports), data, true, "", false, nil)
}

// A GetPackagesFunc function obtains the non-nil symbols for a set of
// packages, creating and recursively importing them as needed. An
// implementation should store each package symbol is in the Pkg
// field of the items array.
//
// Any error causes importing to fail. This can be used to quickly read
// the import manifest of an export data file without fully decoding it.
type GetPackagesFunc = func(items []GetPackagesItem) error

// A GetPackagesItem is a request from the importer for the package
// symbol of the specified name and path.
type GetPackagesItem struct {
	Name, Path string
	Pkg        *types.Package // to be filled in by GetPackagesFunc call

	// private importer state
	pathOffset uint64
	nameIndex  map[string]uint64
}

// GetPackagesFromMap returns a GetPackagesFunc that retrieves
// packages from the given map of package path to package.
//
// The returned function may mutate m: each requested package that is not
// found is created with types.NewPackage and inserted into m.
func GetPackagesFromMap(m map[string]*types.Package) GetPackagesFunc {
	return func(items []GetPackagesItem) error {
		for i, item := range items {
			pkg, ok := m[item.Path]
			if !ok {
				pkg = types.NewPackage(item.Path, item.Name)
				m[item.Path] = pkg
			}
			items[i].Pkg = pkg
		}
		return nil
	}
}

func iimportCommon(fset *token.FileSet, getPackages GetPackagesFunc, data []byte, bundle bool, path string, shallow bool, reportf ReportFunc) (pkgs []*types.Package, err error) {
=======
	return iimportCommon(fset, GetPackageFromMap(imports), data, true, "", nil)
=======
	return iimportCommon(fset, GetPackagesFromMap(imports), data, true, "", false, nil)
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
}

// A GetPackagesFunc function obtains the non-nil symbols for a set of
// packages, creating and recursively importing them as needed. An
// implementation should store each package symbol is in the Pkg
// field of the items array.
//
// Any error causes importing to fail. This can be used to quickly read
// the import manifest of an export data file without fully decoding it.
type GetPackagesFunc = func(items []GetPackagesItem) error

// A GetPackagesItem is a request from the importer for the package
// symbol of the specified name and path.
type GetPackagesItem struct {
	Name, Path string
	Pkg        *types.Package // to be filled in by GetPackagesFunc call

	// private importer state
	pathOffset uint64
	nameIndex  map[string]uint64
}

// GetPackagesFromMap returns a GetPackagesFunc that retrieves
// packages from the given map of package path to package.
//
// The returned function may mutate m: each requested package that is not
// found is created with types.NewPackage and inserted into m.
func GetPackagesFromMap(m map[string]*types.Package) GetPackagesFunc {
	return func(items []GetPackagesItem) error {
		for i, item := range items {
			pkg, ok := m[item.Path]
			if !ok {
				pkg = types.NewPackage(item.Path, item.Name)
				m[item.Path] = pkg
			}
			items[i].Pkg = pkg
		}
		return nil
	}
}

<<<<<<< HEAD
func iimportCommon(fset *token.FileSet, getPackage GetPackageFunc, data []byte, bundle bool, path string, insert InsertType) (pkgs []*types.Package, err error) {
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
func iimportCommon(fset *token.FileSet, getPackages GetPackagesFunc, data []byte, bundle bool, path string, shallow bool, reportf ReportFunc) (pkgs []*types.Package, err error) {
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	const currentVersion = iexportVersionCurrent
	version := int64(-1)
	if !debug {
		defer func() {
			if e := recover(); e != nil {
				if bundle {
					err = fmt.Errorf("%v", e)
				} else if version > currentVersion {
					err = fmt.Errorf("cannot import %q (%v), export data is newer version - update tool", path, e)
				} else {
					err = fmt.Errorf("internal error while importing %q (%v); please report an issue", path, e)
				}
			}
		}()
	}

	r := &intReader{bytes.NewReader(data), path}

	if bundle {
		if v := r.uint64(); v != bundleVersion {
			errorf("unknown bundle format version %d", v)
		}
	}

	version = int64(r.uint64())
	switch version {
	case iexportVersionGo1_18, iexportVersionPosCol, iexportVersionGo1_11:
	default:
		if version > iexportVersionGo1_18 {
			errorf("unstable iexport format version %d, just rebuild compiler and std library", version)
		} else {
			errorf("unknown iexport format version %d", version)
		}
	}

	sLen := int64(r.uint64())
	var fLen int64
	var fileOffset []uint64
<<<<<<< HEAD
<<<<<<< HEAD
	if shallow {
=======
	if insert != nil {
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
	if shallow {
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
		// Shallow mode uses a different position encoding.
		fLen = int64(r.uint64())
		fileOffset = make([]uint64, r.uint64())
		for i := range fileOffset {
			fileOffset[i] = r.uint64()
		}
	}
	dLen := int64(r.uint64())

	whence, _ := r.Seek(0, io.SeekCurrent)
	stringData := data[whence : whence+sLen]
	fileData := data[whence+sLen : whence+sLen+fLen]
	declData := data[whence+sLen+fLen : whence+sLen+fLen+dLen]
	r.Seek(sLen+fLen+dLen, io.SeekCurrent)

	p := iimporter{
		version: int(version),
		ipath:   path,
<<<<<<< HEAD
<<<<<<< HEAD
		aliases: aliases.Enabled(),
		shallow: shallow,
		reportf: reportf,
=======
		insert:  insert,
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
		shallow: shallow,
		reportf: reportf,
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)

		stringData:  stringData,
		stringCache: make(map[uint64]string),
		fileOffset:  fileOffset,
		fileData:    fileData,
		fileCache:   make([]*token.File, len(fileOffset)),
		pkgCache:    make(map[uint64]*types.Package),

		declData: declData,
		pkgIndex: make(map[*types.Package]map[string]uint64),
		typCache: make(map[uint64]types.Type),
		// Separate map for typeparams, keyed by their package and unique
		// name.
		tparamIndex: make(map[ident]types.Type),

		fake: fakeFileSet{
			fset:  fset,
			files: make(map[string]*fileInfo),
		},
	}
	defer p.fake.setLines() // set lines for files in fset

	for i, pt := range predeclared() {
		p.typCache[uint64(i)] = pt
	}

<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	// Gather the relevant packages from the manifest.
	items := make([]GetPackagesItem, r.uint64())
	uniquePkgPaths := make(map[string]bool)
	for i := range items {
<<<<<<< HEAD
=======
	pkgList := make([]*types.Package, r.uint64())
	for i := range pkgList {
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
		pkgPathOff := r.uint64()
		pkgPath := p.stringAt(pkgPathOff)
		pkgName := p.stringAt(r.uint64())
		_ = r.uint64() // package height; unused by go/types

		if pkgPath == "" {
			pkgPath = path
		}
<<<<<<< HEAD
<<<<<<< HEAD
		items[i].Name = pkgName
		items[i].Path = pkgPath
		items[i].pathOffset = pkgPathOff
=======
		pkg := getPackage(pkgPath, pkgName)
		if pkg == nil {
			errorf("internal error: getPackage returned nil package for %s", pkgPath)
		} else if pkg.Name() != pkgName {
			errorf("conflicting names %s and %s for package %q", pkg.Name(), pkgName, path)
		}
		if i == 0 && !bundle {
			p.localpkg = pkg
		}

		p.pkgCache[pkgPathOff] = pkg
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
		items[i].Name = pkgName
		items[i].Path = pkgPath
		items[i].pathOffset = pkgPathOff
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)

		// Read index for package.
		nameIndex := make(map[string]uint64)
		nSyms := r.uint64()
<<<<<<< HEAD
<<<<<<< HEAD
		// In shallow mode, only the current package (i=0) has an index.
		assert(!(shallow && i > 0 && nSyms != 0))
=======
		// In shallow mode we don't expect an index for other packages.
		assert(nSyms == 0 || p.localpkg == pkg || p.insert == nil)
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
		// In shallow mode, only the current package (i=0) has an index.
		assert(!(shallow && i > 0 && nSyms != 0))
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
		for ; nSyms > 0; nSyms-- {
			name := p.stringAt(r.uint64())
			nameIndex[name] = r.uint64()
		}

<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
		items[i].nameIndex = nameIndex

		uniquePkgPaths[pkgPath] = true
	}
	// Debugging #63822; hypothesis: there are duplicate PkgPaths.
	if len(uniquePkgPaths) != len(items) {
		reportf("found duplicate PkgPaths while reading export data manifest: %v", items)
	}

	// Request packages all at once from the client,
	// enabling a parallel implementation.
	if err := getPackages(items); err != nil {
		return nil, err // don't wrap this error
	}

	// Check the results and complete the index.
	pkgList := make([]*types.Package, len(items))
	for i, item := range items {
		pkg := item.Pkg
		if pkg == nil {
			errorf("internal error: getPackages returned nil package for %q", item.Path)
		} else if pkg.Path() != item.Path {
			errorf("internal error: getPackages returned wrong path %q, want %q", pkg.Path(), item.Path)
		} else if pkg.Name() != item.Name {
			errorf("internal error: getPackages returned wrong name %s for package %q, want %s", pkg.Name(), item.Path, item.Name)
		}
		p.pkgCache[item.pathOffset] = pkg
		p.pkgIndex[pkg] = item.nameIndex
<<<<<<< HEAD
=======
		p.pkgIndex[pkg] = nameIndex
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
		pkgList[i] = pkg
	}

	if bundle {
		pkgs = make([]*types.Package, r.uint64())
		for i := range pkgs {
			pkg := p.pkgAt(r.uint64())
			imps := make([]*types.Package, r.uint64())
			for j := range imps {
				imps[j] = p.pkgAt(r.uint64())
			}
			pkg.SetImports(imps)
			pkgs[i] = pkg
		}
	} else {
		if len(pkgList) == 0 {
			errorf("no packages found for %s", path)
			panic("unreachable")
		}
		pkgs = pkgList[:1]

		// record all referenced packages as imports
		list := append(([]*types.Package)(nil), pkgList[1:]...)
		sort.Sort(byPath(list))
		pkgs[0].SetImports(list)
	}

	for _, pkg := range pkgs {
		if pkg.Complete() {
			continue
		}

		names := make([]string, 0, len(p.pkgIndex[pkg]))
		for name := range p.pkgIndex[pkg] {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			p.doDecl(pkg, name)
		}

		// package was imported completely and without errors
		pkg.MarkComplete()
	}

	// SetConstraint can't be called if the constraint type is not yet complete.
<<<<<<< HEAD
	// When type params are created in the typeParamTag case of (*importReader).obj(),
=======
	// When type params are created in the 'P' case of (*importReader).obj(),
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
	// the associated constraint type may not be complete due to recursion.
	// Therefore, we defer calling SetConstraint there, and call it here instead
	// after all types are complete.
	for _, d := range p.later {
<<<<<<< HEAD
<<<<<<< HEAD
		d.t.SetConstraint(d.constraint)
=======
		typeparams.SetTypeParamConstraint(d.t, d.constraint)
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
		d.t.SetConstraint(d.constraint)
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	}

	for _, typ := range p.interfaceList {
		typ.Complete()
	}

<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	// Workaround for golang/go#61561. See the doc for instanceList for details.
	for _, typ := range p.instanceList {
		if iface, _ := typ.Underlying().(*types.Interface); iface != nil {
			iface.Complete()
		}
	}

<<<<<<< HEAD
=======
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	return pkgs, nil
}

type setConstraintArgs struct {
<<<<<<< HEAD
<<<<<<< HEAD
	t          *types.TypeParam
=======
	t          *typeparams.TypeParam
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
	t          *types.TypeParam
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	constraint types.Type
}

type iimporter struct {
	version int
	ipath   string

<<<<<<< HEAD
<<<<<<< HEAD
	aliases bool
	shallow bool
	reportf ReportFunc // if non-nil, used to report bugs
=======
	localpkg *types.Package
	insert   func(pkg *types.Package, name string) // "shallow" mode only
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
	shallow bool
	reportf ReportFunc // if non-nil, used to report bugs
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)

	stringData  []byte
	stringCache map[uint64]string
	fileOffset  []uint64 // fileOffset[i] is offset in fileData for info about file encoded as i
	fileData    []byte
	fileCache   []*token.File // memoized decoding of file encoded as i
	pkgCache    map[uint64]*types.Package

	declData    []byte
	pkgIndex    map[*types.Package]map[string]uint64
	typCache    map[uint64]types.Type
	tparamIndex map[ident]types.Type

	fake          fakeFileSet
	interfaceList []*types.Interface

<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	// Workaround for the go/types bug golang/go#61561: instances produced during
	// instantiation may contain incomplete interfaces. Here we only complete the
	// underlying type of the instance, which is the most common case but doesn't
	// handle parameterized interface literals defined deeper in the type.
	instanceList []types.Type // instances for later completion (see golang/go#61561)

<<<<<<< HEAD
=======
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	// Arguments for calls to SetConstraint that are deferred due to recursive types
	later []setConstraintArgs

	indent int // for tracing support
}

func (p *iimporter) trace(format string, args ...interface{}) {
	if !trace {
		// Call sites should also be guarded, but having this check here allows
		// easily enabling/disabling debug trace statements.
		return
	}
	fmt.Printf(strings.Repeat("..", p.indent)+format+"\n", args...)
}

func (p *iimporter) doDecl(pkg *types.Package, name string) {
	if debug {
		p.trace("import decl %s", name)
		p.indent++
		defer func() {
			p.indent--
			p.trace("=> %s", name)
		}()
	}
	// See if we've already imported this declaration.
	if obj := pkg.Scope().Lookup(name); obj != nil {
		return
	}

	off, ok := p.pkgIndex[pkg][name]
	if !ok {
<<<<<<< HEAD
<<<<<<< HEAD
		// In deep mode, the index should be complete. In shallow
		// mode, we should have already recursively loaded necessary
		// dependencies so the above Lookup succeeds.
=======
		// In "shallow" mode, call back to the application to
		// find the object and insert it into the package scope.
		if p.insert != nil {
			assert(pkg != p.localpkg)
			p.insert(pkg, name) // "can't fail"
			return
		}
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
		// In deep mode, the index should be complete. In shallow
		// mode, we should have already recursively loaded necessary
		// dependencies so the above Lookup succeeds.
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
		errorf("%v.%v not in index", pkg, name)
	}

	r := &importReader{p: p, currPkg: pkg}
	r.declReader.Reset(p.declData[off:])

	r.obj(name)
}

func (p *iimporter) stringAt(off uint64) string {
	if s, ok := p.stringCache[off]; ok {
		return s
	}

	slen, n := binary.Uvarint(p.stringData[off:])
	if n <= 0 {
		errorf("varint failed")
	}
	spos := off + uint64(n)
	s := string(p.stringData[spos : spos+slen])
	p.stringCache[off] = s
	return s
}

func (p *iimporter) fileAt(index uint64) *token.File {
	file := p.fileCache[index]
	if file == nil {
		off := p.fileOffset[index]
		file = p.decodeFile(intReader{bytes.NewReader(p.fileData[off:]), p.ipath})
		p.fileCache[index] = file
	}
	return file
}

func (p *iimporter) decodeFile(rd intReader) *token.File {
	filename := p.stringAt(rd.uint64())
	size := int(rd.uint64())
	file := p.fake.fset.AddFile(filename, -1, size)

	// SetLines requires a nondecreasing sequence.
	// Because it is common for clients to derive the interval
	// [start, start+len(name)] from a start position, and we
	// want to ensure that the end offset is on the same line,
	// we fill in the gaps of the sparse encoding with values
	// that strictly increase by the largest possible amount.
	// This allows us to avoid having to record the actual end
	// offset of each needed line.

	lines := make([]int, int(rd.uint64()))
	var index, offset int
	for i, n := 0, int(rd.uint64()); i < n; i++ {
		index += int(rd.uint64())
		offset += int(rd.uint64())
		lines[index] = offset

		// Ensure monotonicity between points.
		for j := index - 1; j > 0 && lines[j] == 0; j-- {
			lines[j] = lines[j+1] - 1
		}
	}

	// Ensure monotonicity after last point.
	for j := len(lines) - 1; j > 0 && lines[j] == 0; j-- {
		size--
		lines[j] = size
	}

	if !file.SetLines(lines) {
		errorf("SetLines failed: %d", lines) // can't happen
	}
	return file
}

func (p *iimporter) pkgAt(off uint64) *types.Package {
	if pkg, ok := p.pkgCache[off]; ok {
		return pkg
	}
	path := p.stringAt(off)
	errorf("missing package %q in %q", path, p.ipath)
	return nil
}

func (p *iimporter) typAt(off uint64, base *types.Named) types.Type {
	if t, ok := p.typCache[off]; ok && canReuse(base, t) {
		return t
	}

	if off < predeclReserved {
		errorf("predeclared type missing from cache: %v", off)
	}

	r := &importReader{p: p}
	r.declReader.Reset(p.declData[off-predeclReserved:])
	t := r.doType(base)

	if canReuse(base, t) {
		p.typCache[off] = t
	}
	return t
}

// canReuse reports whether the type rhs on the RHS of the declaration for def
// may be re-used.
//
// Specifically, if def is non-nil and rhs is an interface type with methods, it
// may not be re-used because we have a convention of setting the receiver type
// for interface methods to def.
func canReuse(def *types.Named, rhs types.Type) bool {
	if def == nil {
		return true
	}
<<<<<<< HEAD
	iface, _ := aliases.Unalias(rhs).(*types.Interface)
=======
	iface, _ := rhs.(*types.Interface)
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
	if iface == nil {
		return true
	}
	// Don't use iface.Empty() here as iface may not be complete.
	return iface.NumEmbeddeds() == 0 && iface.NumExplicitMethods() == 0
}

type importReader struct {
	p          *iimporter
	declReader bytes.Reader
	currPkg    *types.Package
	prevFile   string
	prevLine   int64
	prevColumn int64
}

func (r *importReader) obj(name string) {
	tag := r.byte()
	pos := r.pos()

	switch tag {
<<<<<<< HEAD
	case aliasTag:
		typ := r.typ()
		// TODO(adonovan): support generic aliases:
		// if tag == genericAliasTag {
		// 	tparams := r.tparamList()
		// 	alias.SetTypeParams(tparams)
		// }
		r.declare(aliases.NewAlias(r.p.aliases, pos, r.currPkg, name, typ))

	case constTag:
=======
	case 'A':
		typ := r.typ()

		r.declare(types.NewTypeName(pos, r.currPkg, name, typ))

	case 'C':
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
		typ, val := r.value()

		r.declare(types.NewConst(pos, r.currPkg, name, typ, val))

<<<<<<< HEAD
	case funcTag, genericFuncTag:
		var tparams []*types.TypeParam
		if tag == genericFuncTag {
=======
	case 'F', 'G':
		var tparams []*types.TypeParam
		if tag == 'G' {
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
			tparams = r.tparamList()
		}
		sig := r.signature(nil, nil, tparams)
		r.declare(types.NewFunc(pos, r.currPkg, name, sig))

<<<<<<< HEAD
	case typeTag, genericTypeTag:
=======
	case 'T', 'U':
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
		// Types can be recursive. We need to setup a stub
		// declaration before recursing.
		obj := types.NewTypeName(pos, r.currPkg, name, nil)
		named := types.NewNamed(obj, nil, nil)
		// Declare obj before calling r.tparamList, so the new type name is recognized
		// if used in the constraint of one of its own typeparams (see #48280).
		r.declare(obj)
<<<<<<< HEAD
		if tag == genericTypeTag {
			tparams := r.tparamList()
			named.SetTypeParams(tparams)
=======
		if tag == 'U' {
			tparams := r.tparamList()
<<<<<<< HEAD
			typeparams.SetForNamed(named, tparams)
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
			named.SetTypeParams(tparams)
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
		}

		underlying := r.p.typAt(r.uint64(), named).Underlying()
		named.SetUnderlying(underlying)

		if !isInterface(underlying) {
			for n := r.uint64(); n > 0; n-- {
				mpos := r.pos()
				mname := r.ident()
				recv := r.param()

				// If the receiver has any targs, set those as the
				// rparams of the method (since those are the
				// typeparams being used in the method sig/body).
<<<<<<< HEAD
				_, recvNamed := typesinternal.ReceiverNamed(recv)
				targs := recvNamed.TypeArgs()
				var rparams []*types.TypeParam
				if targs.Len() > 0 {
					rparams = make([]*types.TypeParam, targs.Len())
					for i := range rparams {
						rparams[i] = aliases.Unalias(targs.At(i)).(*types.TypeParam)
=======
				base := baseType(recv.Type())
				assert(base != nil)
				targs := base.TypeArgs()
				var rparams []*types.TypeParam
				if targs.Len() > 0 {
					rparams = make([]*types.TypeParam, targs.Len())
					for i := range rparams {
<<<<<<< HEAD
						rparams[i] = targs.At(i).(*typeparams.TypeParam)
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
						rparams[i] = targs.At(i).(*types.TypeParam)
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
					}
				}
				msig := r.signature(recv, rparams, nil)

				named.AddMethod(types.NewFunc(mpos, r.currPkg, mname, msig))
			}
		}

<<<<<<< HEAD
	case typeParamTag:
=======
	case 'P':
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
		// We need to "declare" a typeparam in order to have a name that
		// can be referenced recursively (if needed) in the type param's
		// bound.
		if r.p.version < iexportVersionGenerics {
			errorf("unexpected type param type")
		}
		name0 := tparamName(name)
		tn := types.NewTypeName(pos, r.currPkg, name0, nil)
<<<<<<< HEAD
<<<<<<< HEAD
		t := types.NewTypeParam(tn, nil)
=======
		t := typeparams.NewTypeParam(tn, nil)
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
		t := types.NewTypeParam(tn, nil)
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)

		// To handle recursive references to the typeparam within its
		// bound, save the partial type in tparamIndex before reading the bounds.
		id := ident{r.currPkg, name}
		r.p.tparamIndex[id] = t
		var implicit bool
		if r.p.version >= iexportVersionGo1_18 {
			implicit = r.bool()
		}
		constraint := r.typ()
		if implicit {
<<<<<<< HEAD
			iface, _ := aliases.Unalias(constraint).(*types.Interface)
			if iface == nil {
				errorf("non-interface constraint marked implicit")
			}
			iface.MarkImplicit()
=======
			iface, _ := constraint.(*types.Interface)
			if iface == nil {
				errorf("non-interface constraint marked implicit")
			}
<<<<<<< HEAD
			typeparams.MarkImplicit(iface)
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
			iface.MarkImplicit()
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
		}
		// The constraint type may not be complete, if we
		// are in the middle of a type recursion involving type
		// constraints. So, we defer SetConstraint until we have
		// completely set up all types in ImportData.
		r.p.later = append(r.p.later, setConstraintArgs{t: t, constraint: constraint})

<<<<<<< HEAD
	case varTag:
=======
	case 'V':
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
		typ := r.typ()

		r.declare(types.NewVar(pos, r.currPkg, name, typ))

	default:
		errorf("unexpected tag: %v", tag)
	}
}

func (r *importReader) declare(obj types.Object) {
	obj.Pkg().Scope().Insert(obj)
}

func (r *importReader) value() (typ types.Type, val constant.Value) {
	typ = r.typ()
	if r.p.version >= iexportVersionGo1_18 {
		// TODO: add support for using the kind.
		_ = constant.Kind(r.int64())
	}

	switch b := typ.Underlying().(*types.Basic); b.Info() & types.IsConstType {
	case types.IsBoolean:
		val = constant.MakeBool(r.bool())

	case types.IsString:
		val = constant.MakeString(r.string())

	case types.IsInteger:
		var x big.Int
		r.mpint(&x, b)
		val = constant.Make(&x)

	case types.IsFloat:
		val = r.mpfloat(b)

	case types.IsComplex:
		re := r.mpfloat(b)
		im := r.mpfloat(b)
		val = constant.BinaryOp(re, token.ADD, constant.MakeImag(im))

	default:
		if b.Kind() == types.Invalid {
			val = constant.MakeUnknown()
			return
		}
		errorf("unexpected type %v", typ) // panics
		panic("unreachable")
	}

	return
}

func intSize(b *types.Basic) (signed bool, maxBytes uint) {
	if (b.Info() & types.IsUntyped) != 0 {
		return true, 64
	}

	switch b.Kind() {
	case types.Float32, types.Complex64:
		return true, 3
	case types.Float64, types.Complex128:
		return true, 7
	}

	signed = (b.Info() & types.IsUnsigned) == 0
	switch b.Kind() {
	case types.Int8, types.Uint8:
		maxBytes = 1
	case types.Int16, types.Uint16:
		maxBytes = 2
	case types.Int32, types.Uint32:
		maxBytes = 4
	default:
		maxBytes = 8
	}

	return
}

func (r *importReader) mpint(x *big.Int, typ *types.Basic) {
	signed, maxBytes := intSize(typ)

	maxSmall := 256 - maxBytes
	if signed {
		maxSmall = 256 - 2*maxBytes
	}
	if maxBytes == 1 {
		maxSmall = 256
	}

	n, _ := r.declReader.ReadByte()
	if uint(n) < maxSmall {
		v := int64(n)
		if signed {
			v >>= 1
			if n&1 != 0 {
				v = ^v
			}
		}
		x.SetInt64(v)
		return
	}

	v := -n
	if signed {
		v = -(n &^ 1) >> 1
	}
	if v < 1 || uint(v) > maxBytes {
		errorf("weird decoding: %v, %v => %v", n, signed, v)
	}
	b := make([]byte, v)
	io.ReadFull(&r.declReader, b)
	x.SetBytes(b)
	if signed && n&1 != 0 {
		x.Neg(x)
	}
}

func (r *importReader) mpfloat(typ *types.Basic) constant.Value {
	var mant big.Int
	r.mpint(&mant, typ)
	var f big.Float
	f.SetInt(&mant)
	if f.Sign() != 0 {
		f.SetMantExp(&f, int(r.int64()))
	}
	return constant.Make(&f)
}

func (r *importReader) ident() string {
	return r.string()
}

func (r *importReader) qualifiedIdent() (*types.Package, string) {
	name := r.string()
	pkg := r.pkg()
	return pkg, name
}

func (r *importReader) pos() token.Pos {
<<<<<<< HEAD
<<<<<<< HEAD
	if r.p.shallow {
		// precise offsets are encoded only in shallow mode
=======
	if r.p.insert != nil { // shallow mode
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
	if r.p.shallow {
		// precise offsets are encoded only in shallow mode
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
		return r.posv2()
	}
	if r.p.version >= iexportVersionPosCol {
		r.posv1()
	} else {
		r.posv0()
	}

	if r.prevFile == "" && r.prevLine == 0 && r.prevColumn == 0 {
		return token.NoPos
	}
	return r.p.fake.pos(r.prevFile, int(r.prevLine), int(r.prevColumn))
}

func (r *importReader) posv0() {
	delta := r.int64()
	if delta != deltaNewFile {
		r.prevLine += delta
	} else if l := r.int64(); l == -1 {
		r.prevLine += deltaNewFile
	} else {
		r.prevFile = r.string()
		r.prevLine = l
	}
}

func (r *importReader) posv1() {
	delta := r.int64()
	r.prevColumn += delta >> 1
	if delta&1 != 0 {
		delta = r.int64()
		r.prevLine += delta >> 1
		if delta&1 != 0 {
			r.prevFile = r.string()
		}
	}
}

func (r *importReader) posv2() token.Pos {
	file := r.uint64()
	if file == 0 {
		return token.NoPos
	}
	tf := r.p.fileAt(file - 1)
	return tf.Pos(int(r.uint64()))
}

func (r *importReader) typ() types.Type {
	return r.p.typAt(r.uint64(), nil)
}

func isInterface(t types.Type) bool {
<<<<<<< HEAD
	_, ok := aliases.Unalias(t).(*types.Interface)
=======
	_, ok := t.(*types.Interface)
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
	return ok
}

func (r *importReader) pkg() *types.Package { return r.p.pkgAt(r.uint64()) }
func (r *importReader) string() string      { return r.p.stringAt(r.uint64()) }

func (r *importReader) doType(base *types.Named) (res types.Type) {
	k := r.kind()
	if debug {
		r.p.trace("importing type %d (base: %s)", k, base)
		r.p.indent++
		defer func() {
			r.p.indent--
			r.p.trace("=> %s", res)
		}()
	}
	switch k {
	default:
		errorf("unexpected kind tag in %q: %v", r.p.ipath, k)
		return nil

<<<<<<< HEAD
	case aliasType, definedType:
=======
	case definedType:
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
		pkg, name := r.qualifiedIdent()
		r.p.doDecl(pkg, name)
		return pkg.Scope().Lookup(name).(*types.TypeName).Type()
	case pointerType:
		return types.NewPointer(r.typ())
	case sliceType:
		return types.NewSlice(r.typ())
	case arrayType:
		n := r.uint64()
		return types.NewArray(r.typ(), int64(n))
	case chanType:
		dir := chanDir(int(r.uint64()))
		return types.NewChan(dir, r.typ())
	case mapType:
		return types.NewMap(r.typ(), r.typ())
	case signatureType:
		r.currPkg = r.pkg()
		return r.signature(nil, nil, nil)

	case structType:
		r.currPkg = r.pkg()

		fields := make([]*types.Var, r.uint64())
		tags := make([]string, len(fields))
		for i := range fields {
<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
			var field *types.Var
			if r.p.shallow {
				field, _ = r.objectPathObject().(*types.Var)
			}

<<<<<<< HEAD
=======
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
			fpos := r.pos()
			fname := r.ident()
			ftyp := r.typ()
			emb := r.bool()
			tag := r.string()

<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
			// Either this is not a shallow import, the field is local, or the
			// encoded objectPath failed to produce an object (a bug).
			//
			// Even in this last, buggy case, fall back on creating a new field. As
			// discussed in iexport.go, this is not correct, but mostly works and is
			// preferable to failing (for now at least).
			if field == nil {
				field = types.NewField(fpos, r.currPkg, fname, ftyp, emb)
			}

			fields[i] = field
<<<<<<< HEAD
=======
			fields[i] = types.NewField(fpos, r.currPkg, fname, ftyp, emb)
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
			tags[i] = tag
		}
		return types.NewStruct(fields, tags)

	case interfaceType:
		r.currPkg = r.pkg()

		embeddeds := make([]types.Type, r.uint64())
		for i := range embeddeds {
			_ = r.pos()
			embeddeds[i] = r.typ()
		}

		methods := make([]*types.Func, r.uint64())
		for i := range methods {
<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
			var method *types.Func
			if r.p.shallow {
				method, _ = r.objectPathObject().(*types.Func)
			}

<<<<<<< HEAD
=======
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
			mpos := r.pos()
			mname := r.ident()

			// TODO(mdempsky): Matches bimport.go, but I
			// don't agree with this.
			var recv *types.Var
			if base != nil {
				recv = types.NewVar(token.NoPos, r.currPkg, "", base)
			}
<<<<<<< HEAD
<<<<<<< HEAD
			msig := r.signature(recv, nil, nil)

			if method == nil {
				method = types.NewFunc(mpos, r.currPkg, mname, msig)
			}
			methods[i] = method
=======

			msig := r.signature(recv, nil, nil)
			methods[i] = types.NewFunc(mpos, r.currPkg, mname, msig)
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
			msig := r.signature(recv, nil, nil)

			if method == nil {
				method = types.NewFunc(mpos, r.currPkg, mname, msig)
			}
			methods[i] = method
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
		}

		typ := newInterface(methods, embeddeds)
		r.p.interfaceList = append(r.p.interfaceList, typ)
		return typ

	case typeParamType:
		if r.p.version < iexportVersionGenerics {
			errorf("unexpected type param type")
		}
		pkg, name := r.qualifiedIdent()
		id := ident{pkg, name}
		if t, ok := r.p.tparamIndex[id]; ok {
			// We're already in the process of importing this typeparam.
			return t
		}
		// Otherwise, import the definition of the typeparam now.
		r.p.doDecl(pkg, name)
		return r.p.tparamIndex[id]

	case instanceType:
		if r.p.version < iexportVersionGenerics {
			errorf("unexpected instantiation type")
		}
		// pos does not matter for instances: they are positioned on the original
		// type.
		_ = r.pos()
		len := r.uint64()
		targs := make([]types.Type, len)
		for i := range targs {
			targs[i] = r.typ()
		}
		baseType := r.typ()
		// The imported instantiated type doesn't include any methods, so
		// we must always use the methods of the base (orig) type.
		// TODO provide a non-nil *Environment
<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
		t, _ := types.Instantiate(nil, baseType, targs, false)

		// Workaround for golang/go#61561. See the doc for instanceList for details.
		r.p.instanceList = append(r.p.instanceList, t)
<<<<<<< HEAD
=======
		t, _ := typeparams.Instantiate(nil, baseType, targs, false)
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
		return t

	case unionType:
		if r.p.version < iexportVersionGenerics {
			errorf("unexpected instantiation type")
		}
<<<<<<< HEAD
<<<<<<< HEAD
		terms := make([]*types.Term, r.uint64())
		for i := range terms {
			terms[i] = types.NewTerm(r.bool(), r.typ())
		}
		return types.NewUnion(terms)
=======
		terms := make([]*typeparams.Term, r.uint64())
=======
		terms := make([]*types.Term, r.uint64())
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
		for i := range terms {
			terms[i] = types.NewTerm(r.bool(), r.typ())
		}
<<<<<<< HEAD
		return typeparams.NewUnion(terms)
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
		return types.NewUnion(terms)
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	}
}

func (r *importReader) kind() itag {
	return itag(r.uint64())
}

<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
// objectPathObject is the inverse of exportWriter.objectPath.
//
// In shallow mode, certain fields and methods may need to be looked up in an
// imported package. See the doc for exportWriter.objectPath for a full
// explanation.
func (r *importReader) objectPathObject() types.Object {
	objPath := objectpath.Path(r.string())
	if objPath == "" {
		return nil
	}
	pkg := r.pkg()
	obj, err := objectpath.Object(pkg, objPath)
	if err != nil {
		if r.p.reportf != nil {
			r.p.reportf("failed to find object for objectPath %q: %v", objPath, err)
		}
	}
	return obj
}

func (r *importReader) signature(recv *types.Var, rparams []*types.TypeParam, tparams []*types.TypeParam) *types.Signature {
<<<<<<< HEAD
	params := r.paramList()
	results := r.paramList()
	variadic := params.Len() > 0 && r.bool()
	return types.NewSignatureType(recv, rparams, tparams, params, results, variadic)
}

func (r *importReader) tparamList() []*types.TypeParam {
=======
func (r *importReader) signature(recv *types.Var, rparams []*typeparams.TypeParam, tparams []*typeparams.TypeParam) *types.Signature {
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	params := r.paramList()
	results := r.paramList()
	variadic := params.Len() > 0 && r.bool()
	return types.NewSignatureType(recv, rparams, tparams, params, results, variadic)
}

<<<<<<< HEAD
func (r *importReader) tparamList() []*typeparams.TypeParam {
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
func (r *importReader) tparamList() []*types.TypeParam {
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	n := r.uint64()
	if n == 0 {
		return nil
	}
<<<<<<< HEAD
<<<<<<< HEAD
	xs := make([]*types.TypeParam, n)
	for i := range xs {
		// Note: the standard library importer is tolerant of nil types here,
		// though would panic in SetTypeParams.
		xs[i] = aliases.Unalias(r.typ()).(*types.TypeParam)
=======
	xs := make([]*typeparams.TypeParam, n)
	for i := range xs {
		// Note: the standard library importer is tolerant of nil types here,
		// though would panic in SetTypeParams.
		xs[i] = r.typ().(*typeparams.TypeParam)
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
=======
	xs := make([]*types.TypeParam, n)
	for i := range xs {
		// Note: the standard library importer is tolerant of nil types here,
		// though would panic in SetTypeParams.
		xs[i] = r.typ().(*types.TypeParam)
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	}
	return xs
}

func (r *importReader) paramList() *types.Tuple {
	xs := make([]*types.Var, r.uint64())
	for i := range xs {
		xs[i] = r.param()
	}
	return types.NewTuple(xs...)
}

func (r *importReader) param() *types.Var {
	pos := r.pos()
	name := r.ident()
	typ := r.typ()
	return types.NewParam(pos, r.currPkg, name, typ)
}

func (r *importReader) bool() bool {
	return r.uint64() != 0
}

func (r *importReader) int64() int64 {
	n, err := binary.ReadVarint(&r.declReader)
	if err != nil {
		errorf("readVarint: %v", err)
	}
	return n
}

func (r *importReader) uint64() uint64 {
	n, err := binary.ReadUvarint(&r.declReader)
	if err != nil {
		errorf("readUvarint: %v", err)
	}
	return n
}

func (r *importReader) byte() byte {
	x, err := r.declReader.ReadByte()
	if err != nil {
		errorf("declReader.ReadByte: %v", err)
	}
	return x
}
<<<<<<< HEAD
=======

func baseType(typ types.Type) *types.Named {
	// pointer receivers are never types.Named types
	if p, _ := typ.(*types.Pointer); p != nil {
		typ = p.Elem()
	}
	// receiver base types are always (possibly generic) types.Named types
	n, _ := typ.(*types.Named)
	return n
}
>>>>>>> 59b7cc43 (Update vendor github.com/docker/docker@v23.0.2+incompatible, k8s.io/api@v0.26.2)
