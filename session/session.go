// Package session defines the opertions for supporting user sessions within
// a power-league application.
package session

import (
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"

	"github.com/Forestmb/goff"
	"github.com/golang/glog"
	"github.com/gorilla/sessions"
	"github.com/mrjones/oauth"
)

const (
	// SessionName is used to update the client session
	SessionName = "client-session"

	// AccessTokenKey updates the access token for the current session
	AccessTokenKey = "access-token"

	// RequestTokenKey updates the request token for the current session
	RequestTokenKey = "request-token"

	// oauthVerifierKey acceses  the verification code after oauth
	// authentication
	oauthVerifierKey = "oauth_verifier"
)

//
// SessionManager interface
//

// SessionManager provides an interface to managing sessions for power
// rankings users
type SessionManager interface {
	Login(w http.ResponseWriter, r *http.Request, redirectURL string) (loginURL string, err error)
	Authenticate(w http.ResponseWriter, r *http.Request) error
	Logout(w http.ResponseWriter, r *http.Request) error
	IsLoggedIn(r *http.Request) bool
	GetClient(w http.ResponseWriter, r *http.Request) (*goff.Client, error)
}

// defaultManager is the default implementation of SessionManager
type defaultManager struct {
	consumer Consumer
	store    sessions.Store
}

// NewSessionManager creates a new SessionManager that uses the given
// consumer for OAuth authentication and store to persist the sessions across
// requests
func NewSessionManager(c Consumer, s sessions.Store) SessionManager {
	gob.Register(&oauth.RequestToken{})
	gob.Register(&oauth.AccessToken{})
	return &defaultManager{
		consumer: c,
		store:    s,
	}
}

//
// Consumer interface
//

// Consumer is the interface to an OAuth consumer
type Consumer interface {
	GetRequestTokenAndUrl(url string) (r *oauth.RequestToken, requestURL string, err error)
	AuthorizeToken(r *oauth.RequestToken, verificationCode string) (*oauth.AccessToken, error)
	RefreshToken(accessToken *oauth.AccessToken) (*oauth.AccessToken, error)
	Get(url string, data map[string]string, token *oauth.AccessToken) (*http.Response, error)
}

//
// Client auth
//

// Login starts a new user session within the given request and returns the URL
// that must be access by the user to grant authentication
func (d *defaultManager) Login(
	w http.ResponseWriter,
	r *http.Request,
	redirectURL string) (loginURL string, err error) {

	token, loginURL, err := d.consumer.GetRequestTokenAndUrl(redirectURL)
	if err != nil {
		glog.Warningf("error getting request token: %s", err)
		return "", err
	}

	session, _ := d.store.Get(r, SessionName)
	session.Values = map[interface{}]interface{}{
		RequestTokenKey: token,
	}
	err = session.Save(r, w)

	if err != nil {
		glog.Warningf("error saving client login in session: %s", err)
		return "", err
	}

	glog.V(3).Infoln("client login saved in session")
	return loginURL, nil
}

// Logout ends a user session
func (d *defaultManager) Logout(w http.ResponseWriter, r *http.Request) error {
	session, _ := d.store.Get(r, SessionName)

	session.Values = make(map[interface{}]interface{})
	err := session.Save(r, w)
	if err != nil {
		glog.Warningf("error saving client logout in session: %s", err)
		return err
	}

	glog.V(3).Infoln("client logout saved in session")
	return nil
}

// IsLoggedIn returns whether or not the user represented by the given request
// is logged in.
func (d *defaultManager) IsLoggedIn(req *http.Request) bool {
	session, _ := d.store.Get(req, SessionName)
	_, ok := session.Values[AccessTokenKey].(*oauth.AccessToken)
	return ok
}

// Authenticate uses the verification code in the request and a request token to
// authenticate the user and create an access token.
func (d *defaultManager) Authenticate(w http.ResponseWriter, req *http.Request) error {
	session, err := d.store.Get(req, SessionName)
	if err != nil {
		glog.Warningf("error getting session: %s", err)
		// continue since a new one should have been created
	}

	values := req.URL.Query()
	verificationCode := values.Get(oauthVerifierKey)
	if verificationCode == "" {
		glog.V(2).Infoln("client not authenticated")
		return fmt.Errorf("unable to create goff client for request, "+
			"no verification code in URL: %s", req.URL.String())
	}
	glog.V(2).Infoln("authenticating client with verification code: %s",
		verificationCode)

	rtoken, ok := session.Values[RequestTokenKey].(*oauth.RequestToken)
	if !ok {
		glog.Warningf("error authenticating user, "+
			"no request token in session: %s",
			err)
		return errors.New("unable to create goff client for request, " +
			"no request token in session")
	}

	accessToken, err := d.consumer.AuthorizeToken(rtoken, verificationCode)
	if err != nil {
		glog.Warningf("error authorizing token: %s", err)
		return errors.New("unable to create goff client for request, " +
			"failure when authorizing request token")
	}

	// Only save SESSION_HANDLE_PARAM to reduce cookie size
	sessionParam := accessToken.AdditionalData[oauth.SESSION_HANDLE_PARAM]
	accessToken.AdditionalData = map[string]string{
		oauth.SESSION_HANDLE_PARAM: sessionParam,
	}

	session.Values = map[interface{}]interface{}{
		AccessTokenKey: accessToken,
	}
	err = session.Save(req, w)
	if err != nil {
		glog.Warningf("error saving client session: %s", err)
		return err
	}

	glog.Infoln("client authenticated")
	return nil
}

// GetClient returns the goff.Client for the user represented by the given
// request. The return value can be used to make fantasy API requests
func (d *defaultManager) GetClient(w http.ResponseWriter, req *http.Request) (*goff.Client, error) {
	session, err := d.store.Get(req, SessionName)
	if err != nil {
		glog.Warningf("error getting session: %s", err)
		// continue since a new one should have been created
	}

	accessToken, ok := session.Values[AccessTokenKey].(*oauth.AccessToken)
	// No access token, try creating one if being verified by request
	if !ok {
		glog.V(2).Infoln("client not authenticated")
		return nil, errors.New("no access token in client session")
	}

	accessToken, err = d.consumer.RefreshToken(accessToken)
	if err != nil {
		glog.Warningf("error refreshing token: %s", err)
		return nil, errors.New("unable to create goff client for request, " +
			"failure when refreshing acces token")
	}

	glog.V(2).Infoln("client token refreshed")

	// Only save SESSION_HANDLE_PARAM to reduce cookie size
	sessionParam := accessToken.AdditionalData[oauth.SESSION_HANDLE_PARAM]
	accessToken.AdditionalData = map[string]string{
		oauth.SESSION_HANDLE_PARAM: sessionParam,
	}

	session.Values = map[interface{}]interface{}{
		AccessTokenKey: accessToken,
	}
	err = session.Save(req, w)
	if err != nil {
		glog.Warningf("error saving client session: %s", err)
		return nil, err
	}

	client := goff.NewOAuthClient(d.consumer, accessToken)
	glog.V(3).Infoln("client created successfully")
	return client, nil
}
