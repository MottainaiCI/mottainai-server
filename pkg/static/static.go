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

package static

import (
	"log"
	"net/http"
	"path"
	"path/filepath"

	context "github.com/MottainaiCI/mottainai-server/pkg/context"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"

	"strings"
	"sync"

	macaron "gopkg.in/macaron.v1"
)

//TODO: Handle auth view permission
// Also add to task / namespaces the visibility : public, internal(signed in), group(in org/group/project), user (only the owner)

// FIXME: to be deleted.
type staticMap struct {
	lock sync.RWMutex
	data map[string]*http.Dir
}

func (sm *staticMap) Set(dir *http.Dir) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	sm.data[string(*dir)] = dir
}

func (sm *staticMap) Get(name string) *http.Dir {
	sm.lock.RLock()
	defer sm.lock.RUnlock()

	return sm.data[name]
}

func (sm *staticMap) Delete(name string) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	delete(sm.data, name)
}

var statics = staticMap{sync.RWMutex{}, map[string]*http.Dir{}}

// staticFileSystem implements http.FileSystem interface.
type staticFileSystem struct {
	dir *http.Dir
}

func newStaticFileSystem(directory string) staticFileSystem {
	if !filepath.IsAbs(directory) {
		directory = filepath.Join(macaron.Root, directory)
	}
	dir := http.Dir(directory)
	statics.Set(&dir)
	return staticFileSystem{&dir}
}

// Static returns a middleware handler that serves static files in the given directory.
func Static(directory string, staticOpt ...macaron.StaticOptions) macaron.Handler {
	opt := prepareStaticOptions(directory, staticOpt)

	return func(ctx *context.Context, log *log.Logger) {
		staticHandler(ctx, log, opt, func(ctx *context.Context) bool { return true })
	}
}

func AuthStatic(fn func(*context.Context) bool, directory string, staticOpt ...macaron.StaticOptions) macaron.Handler {
	opt := prepareStaticOptions(directory, staticOpt)

	return func(ctx *context.Context, log *log.Logger) {
		staticHandler(ctx, log, opt, fn)
	}
}

func prepareStaticOptions(dir string, options []macaron.StaticOptions) macaron.StaticOptions {
	var opt macaron.StaticOptions
	if len(options) > 0 {
		opt = options[0]
	}
	return prepareStaticOption(dir, opt)
}
func prepareStaticOption(dir string, opt macaron.StaticOptions) macaron.StaticOptions {
	// Defaults
	if len(opt.IndexFile) == 0 {
		opt.IndexFile = "index.html"
	}
	// Normalize the prefix if provided
	if opt.Prefix != "" {
		// Ensure we have a leading '/'
		if opt.Prefix[0] != '/' {
			opt.Prefix = "/" + opt.Prefix
		}
		// Remove any trailing '/'
		opt.Prefix = strings.TrimRight(opt.Prefix, "/")
	}
	if opt.FileSystem == nil {
		opt.FileSystem = newStaticFileSystem(dir)
	}
	return opt
}
func (fs staticFileSystem) Open(name string) (http.File, error) {
	return fs.dir.Open(name)
}

func staticHandler(ctx *context.Context, log *log.Logger, opt macaron.StaticOptions, fn func(*context.Context) bool) bool {
	if ctx.Req.Method != "GET" && ctx.Req.Method != "HEAD" {
		return false
	}

	file := ctx.Req.URL.Path
	// if we have a prefix, filter requests by stripping the prefix
	if opt.Prefix != "" {
		if !strings.HasPrefix(file, opt.Prefix) {
			return false
		}
		file = file[len(opt.Prefix):]
		if file != "" && file[0] != '/' {
			return false
		}
	}
	if !fn(ctx) {
		return false
	}
	f, err := opt.FileSystem.Open(file)
	if err != nil {
		return false
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return true // File exists but fail to open.
	}

	// Try to serve index file
	if fi.IsDir() {
		// Redirect if missing trailing slash.
		if !strings.HasSuffix(ctx.Req.URL.Path, "/") {
			http.Redirect(ctx.Resp, ctx.Req.Request, ctx.Req.URL.Path+"/", http.StatusFound)
			return true
		}

		file = path.Join(file, opt.IndexFile)
		f, err = opt.FileSystem.Open(file)
		if err != nil {
			return false // Discard error.
		}
		defer f.Close()

		fi, err = f.Stat()
		if err != nil || fi.IsDir() {
			return true
		}
	}

	if !opt.SkipLogging {
		log.Println("[Static] Serving " + file)
	}
	if len(setting.Configuration.AccessControlAllowOrigin) > 0 {
		// Set CORS headers for browser-based git clients
		ctx.Resp.Header().Set("Access-Control-Allow-Origin", setting.Configuration.AccessControlAllowOrigin)
		ctx.Resp.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		ctx.Header().Set("Access-Control-Allow-Origin", setting.Configuration.AccessControlAllowOrigin)
		ctx.Header().Set("Access-Control-Allow-Credentials", "true")
		ctx.Header().Set("Access-Control-Max-Age", "3600")
		ctx.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
	}
	// Add an Expires header to the static content
	if opt.Expires != nil {
		ctx.Resp.Header().Set("Expires", opt.Expires())
	}

	if opt.ETag {
		tag := macaron.GenerateETag(string(fi.Size()), fi.Name(), fi.ModTime().UTC().Format(http.TimeFormat))
		ctx.Resp.Header().Set("ETag", tag)
	}

	http.ServeContent(ctx.Resp, ctx.Req.Request, file, fi.ModTime(), f)
	return true
}
