package lxd

import (
	"github.com/go-macaroon-bakery/macaroon-bakery/v3/httpbakery"
)

<<<<<<< HEAD:vendor/github.com/canonical/lxd/client/lxd_candid.go
// setupBakeryClient initializes the bakeryClient with a new client, sets its http field,
// and adds any existing interactors.
=======
>>>>>>> fe31cef4 (Update vendor github.com/MottainaiCI/lxd-compose@d928eed0eddfde18d58fe3a8ae780328c1b0d55c):vendor/github.com/lxc/lxd/client/lxd_candid.go
func (r *ProtocolLXD) setupBakeryClient() {
	r.bakeryClient = httpbakery.NewClient()
	r.bakeryClient.Client = r.http
	if r.bakeryInteractor != nil {
		for _, interactor := range r.bakeryInteractor {
			r.bakeryClient.AddInteractor(interactor)
		}
	}
}
