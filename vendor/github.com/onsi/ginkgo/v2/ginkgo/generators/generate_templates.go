package generators

<<<<<<< HEAD
<<<<<<< HEAD
var specText = `{{.BuildTags}}
package {{.Package}}
=======
var specText = `package {{.Package}}
>>>>>>> fe31cef4 (Update vendor github.com/MottainaiCI/lxd-compose@d928eed0eddfde18d58fe3a8ae780328c1b0d55c)
=======
var specText = `{{.BuildTags}}
package {{.Package}}
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)

import (
	{{.GinkgoImport}}
	{{.GomegaImport}}

	{{if .ImportPackage}}"{{.PackageImportPath}}"{{end}}
)

var _ = {{.GinkgoPackage}}Describe("{{.Subject}}", func() {

})
`

<<<<<<< HEAD
<<<<<<< HEAD
var agoutiSpecText = `{{.BuildTags}}
package {{.Package}}
=======
var agoutiSpecText = `package {{.Package}}
>>>>>>> fe31cef4 (Update vendor github.com/MottainaiCI/lxd-compose@d928eed0eddfde18d58fe3a8ae780328c1b0d55c)
=======
var agoutiSpecText = `{{.BuildTags}}
package {{.Package}}
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)

import (
	{{.GinkgoImport}}
	{{.GomegaImport}}
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"

	{{if .ImportPackage}}"{{.PackageImportPath}}"{{end}}
)

var _ = {{.GinkgoPackage}}Describe("{{.Subject}}", func() {
	var page *agouti.Page

	{{.GinkgoPackage}}BeforeEach(func() {
		var err error
		page, err = agoutiDriver.NewPage()
		{{.GomegaPackage}}Expect(err).NotTo({{.GomegaPackage}}HaveOccurred())
	})

	{{.GinkgoPackage}}AfterEach(func() {
		{{.GomegaPackage}}Expect(page.Destroy()).To({{.GomegaPackage}}Succeed())
	})
})
`
