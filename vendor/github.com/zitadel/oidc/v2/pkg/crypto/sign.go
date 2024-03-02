package crypto

import (
	"encoding/json"
	"errors"

	"gopkg.in/square/go-jose.v2"
)

<<<<<<< HEAD
<<<<<<< HEAD
func Sign(object any, signer jose.Signer) (string, error) {
=======
func Sign(object interface{}, signer jose.Signer) (string, error) {
>>>>>>> fe31cef4 (Update vendor github.com/MottainaiCI/lxd-compose@d928eed0eddfde18d58fe3a8ae780328c1b0d55c)
=======
func Sign(object any, signer jose.Signer) (string, error) {
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	payload, err := json.Marshal(object)
	if err != nil {
		return "", err
	}
	return SignPayload(payload, signer)
}

func SignPayload(payload []byte, signer jose.Signer) (string, error) {
	if signer == nil {
		return "", errors.New("missing signer")
	}
	result, err := signer.Sign(payload)
	if err != nil {
		return "", err
	}
	return result.CompactSerialize()
}
