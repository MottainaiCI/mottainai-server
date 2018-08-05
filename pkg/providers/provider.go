/*

This is a Goth (https://github.com/markbates/goth) provider re-adaptation for Go-Macaron.

Copyright (C) 2017-2018  Ettore Di Giacinto <mudler@gentoo.org>

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
package gothic

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	//	"os"

	"github.com/MottainaiCI/mottainai-server/pkg/context"

	"github.com/markbates/goth"
)

/*
BeginAuthHandler is a convenience handler for starting the authentication process.
It expects to be able to get the name of the provider from the query parameters
as either "provider" or ":provider".

BeginAuthHandler will redirect the user to the appropriate authentication end-point
for the requested provider.

See https://github.com/markbates/goth/examples/main.go to see this in action.
*/
func BeginAuthHandler(ctx *context.Context) {

	res := ctx.Resp
	req := ctx.Req.Request
	url, err := GetAuthURL(ctx)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(res, err)
		return
	}

	http.Redirect(res, req, url, http.StatusTemporaryRedirect)
}

// SetState sets the state string associated with the given request.
// If no state string is associated with the request, one will be generated.
// This state is sent to the provider and can be retrieved during the
// callback.
var SetState = func(req *http.Request) string {
	state := req.URL.Query().Get("state")
	if len(state) > 0 {
		return state
	}

	// If a state query param is not passed in, generate a random
	// base64-encoded nonce so that the state on the auth URL
	// is unguessable, preventing CSRF attacks, as described in
	//
	// https://auth0.com/docs/protocols/oauth2/oauth-state#keep-reading
	nonceBytes := make([]byte, 64)
	_, err := io.ReadFull(rand.Reader, nonceBytes)
	if err != nil {
		panic("gothic: source of randomness unavailable: " + err.Error())
	}
	return base64.URLEncoding.EncodeToString(nonceBytes)
}

// GetState gets the state returned by the provider during the callback.
// This is used to prevent CSRF attacks, see
// http://tools.ietf.org/html/rfc6749#section-10.12
var GetState = func(req *http.Request) string {
	return req.URL.Query().Get("state")
}

/*
GetAuthURL starts the authentication process with the requested provided.
It will return a URL that should be used to send users to.

It expects to be able to get the name of the provider from the query parameters
as either "provider" or ":provider".

I would recommend using the BeginAuthHandler instead of doing all of these steps
yourself, but that's entirely up to you.
*/
func GetAuthURL(ctx *context.Context) (string, error) {

	res := ctx.Resp
	req := ctx.Req.Request
	providerName, err := GetProviderName(req)
	if err != nil {
		return "", err
	}

	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return "", err
	}
	sess, err := provider.BeginAuth(SetState(req))
	if err != nil {
		return "", err
	}

	url, err := sess.GetAuthURL()
	if err != nil {
		return "", err
	}
	err = StoreInSession(ctx, providerName, sess.Marshal(), req, res)

	if err != nil {
		return "", err
	}

	return url, err
}

/*
CompleteUserAuth does what it says on the tin. It completes the authentication
process and fetches all of the basic information about the user from the provider.

It expects to be able to get the name of the provider from the query parameters
as either "provider" or ":provider".

See https://github.com/markbates/goth/examples/main.go to see this in action.
*/
var CompleteUserAuth = func(ctx *context.Context) (goth.User, error) {

	res := ctx.Resp
	req := ctx.Req.Request
	defer Logout(res, req)

	providerName, err := GetProviderName(req)
	if err != nil {
		return goth.User{}, err
	}

	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return goth.User{}, err
	}

	value, err := GetFromSession(ctx, providerName, req)
	if err != nil {
		return goth.User{}, err
	}

	sess, err := provider.UnmarshalSession(value)
	if err != nil {
		return goth.User{}, err
	}

	err = validateState(req, sess)
	if err != nil {
		return goth.User{}, err
	}

	user, err := provider.FetchUser(sess)
	if err == nil {
		// user can be found with existing session data
		return user, err
	}

	// get new token and retry fetch
	_, err = sess.Authorize(provider, req.URL.Query())
	if err != nil {
		return goth.User{}, err
	}

	err = StoreInSession(ctx, providerName, sess.Marshal(), req, res)

	if err != nil {
		return goth.User{}, err
	}

	gu, err := provider.FetchUser(sess)
	return gu, err
}

// validateState ensures that the state token param from the original
// AuthURL matches the one included in the current (callback) request.
func validateState(req *http.Request, sess goth.Session) error {
	rawAuthURL, err := sess.GetAuthURL()
	if err != nil {
		return err
	}

	authURL, err := url.Parse(rawAuthURL)
	if err != nil {
		return err
	}

	originalState := authURL.Query().Get("state")
	if originalState != "" && (originalState != req.URL.Query().Get("state")) {
		return errors.New("state token mismatch")
	}
	return nil
}

// Logout invalidates a user session.
func Logout(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// GetProviderName is a function used to get the name of a provider
// for a given request. By default, this provider is fetched from
// the URL query string. If you provide it in a different way,
// assign your own function to this variable that returns the provider
// name for your request.
var GetProviderName = getProviderName

func getProviderName(req *http.Request) (string, error) {

	// get all the used providers

	// try to get it from the url param "provider"
	if p := req.URL.Query().Get("provider"); p != "" {
		return p, nil
	}

	// try to get it from the url param ":provider"
	if p := req.URL.Query().Get(":provider"); p != "" {
		return p, nil
	}

	//  try to get it from the go-context's value of "provider" key
	if p, ok := req.Context().Value("provider").(string); ok {
		return p, nil
	}

	// if not found then return an empty string with the corresponding error
	return "", errors.New("you must select a provider")
}

// StoreInSession stores a specified key/value pair in the session.
func StoreInSession(ctx *context.Context, key string, value string, req *http.Request, res http.ResponseWriter) error {
	return ctx.Session.Set(key, value)
}

// GetFromSession retrieves a previously-stored value from the session.
// If no value has previously been stored at the specified key, it will return an error.
func GetFromSession(ctx *context.Context, key string, req *http.Request) (string, error) {
	value := ctx.Session.Get(key)

	if _, ok := value.(string); !ok {
		return "", errors.New("No value")
	}
	return value.(string), nil
}
