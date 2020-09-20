// Package session defines the opertions for supporting user sessions within
// a power-league application.
package session

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Forestmb/goff"
	"github.com/golang/glog"
	"github.com/gorilla/sessions"
	"github.com/pborman/uuid"
	lru "github.com/youtube/vitess/go/cache"
	"golang.org/x/oauth2"
)

const (
	// SessionName is used to update the client session
	SessionName = "client-session"

	// AccessTokenKey updates the access token for the current session
	AccessTokenKey = "access-token"

	// SessionIDKey sets the ID for each session
	SessionIDKey = "session-id"
)

// oauthState ensures the login response matches the request
var oauthState = uuid.New()

//
// Manager interface
//

// Manager provides an interface to managing sessions for power rankings users
type Manager interface {
	Login(w http.ResponseWriter, r *http.Request) (loginURL string)
	Authenticate(w http.ResponseWriter, r *http.Request) error
	Logout(w http.ResponseWriter, r *http.Request) error
	IsLoggedIn(r *http.Request) bool
	GetClient(w http.ResponseWriter, r *http.Request) (*goff.Client, error)
}

// defaultManager is the default implementation of Manager
type defaultManager struct {
	consumerProvider         ConsumerProvider
	store                    sessions.Store
	cache                    *lru.LRUCache
	userCacheDurationSeconds int
}

// NewManager creates a new Manager that uses the given consumer for OAuth
// authentication and store to persist the sessions across requests. Each
// session client returned by `Manager.GetClient` will cache responses for up
// to 6 hours.
//
// See NewManagerWithCache
func NewManager(cp ConsumerProvider, s sessions.Store) Manager {
	return NewManagerWithCache(cp, s, 6*60*60, 10000)
}

// NewManagerWithCache creates a new Manager that uses the given consumer
// provider for OAuth authentication and store to persist the sessions across
// requests. Each session client returned by `Manager.GetClient` will cache
// responses for up to `userCacheDurationSeconds` seconds.
func NewManagerWithCache(
	cp ConsumerProvider,
	s sessions.Store,
	userCacheDurationSeconds int,
	cacheSize int64) Manager {

	gob.Register(&oauth2.Token{})
	gob.Register(&time.Time{})
	cache := lru.NewLRUCache(cacheSize)
	return &defaultManager{
		consumerProvider:         cp,
		store:                    s,
		cache:                    cache,
		userCacheDurationSeconds: userCacheDurationSeconds,
	}
}

//
// Consumer interface
//

// ConsumerProvider creates Consumers to handle authentication on behalf of a
// given request
type ConsumerProvider interface {
	Get(r *http.Request) Consumer
}

// Consumer is the interface to an OAuth2 consumer
type Consumer interface {
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	Exchange(context.Context, string, ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	Client(ctx context.Context, token *oauth2.Token) *http.Client
}

//
// Client auth
//

// Login starts a new user session within the given request and returns the URL
// that must be accessed by the user to grant authentication
func (d *defaultManager) Login(w http.ResponseWriter, r *http.Request) (loginURL string) {
	config := d.consumerProvider.Get(r)
	return config.AuthCodeURL(oauthState)
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
	_, ok := session.Values[AccessTokenKey].(*oauth2.Token)
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

	state := req.FormValue("state")
	if state != oauthState {
		return fmt.Errorf("invalid state returned for authorization, expecing '%s' got '%s'",
			oauthState,
			state)
	}

	verificationCode := req.FormValue("code")
	if verificationCode == "" {
		glog.V(2).Infoln("client not authenticated")
		return fmt.Errorf("unable to create goff client for request, "+
			"no verification code in request: %+v", req.Form)
	}
	glog.V(2).Infof("authenticating client with verification code: %s",
		verificationCode)

	consumer := d.consumerProvider.Get(req)

	accessToken, err := consumer.Exchange(req.Context(), verificationCode)
	if err != nil {
		glog.Warningf("error authorizing token: %s", err)
		return errors.New("unable to create goff client for request, " +
			"failure when authorizing request token")
	}

	session.Values = map[interface{}]interface{}{
		AccessTokenKey: accessToken,
		SessionIDKey:   uuid.New(),
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

	accessToken, ok := session.Values[AccessTokenKey].(*oauth2.Token)
	// No access token, try creating one if being verified by request
	if !ok {
		glog.V(2).Infoln("client not authenticated")
		return nil, errors.New("no access token in client session")
	}

	id, ok := session.Values[SessionIDKey].(string)
	if !ok {
		id = uuid.New()
		glog.Warningf("generating new ID, no '%s' in session -- id=%s",
			SessionIDKey,
			id)
	}

	session.Values = map[interface{}]interface{}{
		AccessTokenKey: accessToken,
		SessionIDKey:   id,
	}
	err = session.Save(req, w)
	if err != nil {
		glog.Warningf("error saving client session: %s", err)
		return nil, err
	}

	consumer := d.consumerProvider.Get(req)
	oauthClient := consumer.Client(req.Context(), accessToken)

	client := goff.NewCachedClient(
		goff.NewLRUCache(
			id,
			time.Duration(d.userCacheDurationSeconds)*time.Second,
			d.cache),
		oauthClient)
	glog.V(3).Infoln("client created successfully")
	return client, nil
}
