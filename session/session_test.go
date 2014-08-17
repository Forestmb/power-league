package session

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/sessions"
	"github.com/mrjones/oauth"
)

func TestNewSessionManager(t *testing.T) {
	manager := NewSessionManager(&MockConsumer{}, mockStore())
	if manager == nil {
		t.Fatal("no manager returned")
	}
}

func TestNewSessionManagerWithCache(t *testing.T) {
	manager := NewSessionManagerWithCache(&MockConsumer{}, mockStore(), 30)
	if manager == nil {
		t.Fatal("no manager returned")
	}
}

func TestIsLoggedIn(t *testing.T) {
	store := &MockStore{
		Values: map[interface{}]interface{}{
			AccessTokenKey: &oauth.AccessToken{},
		},
	}
	manager := NewSessionManager(nil, store)
	loggedIn := manager.IsLoggedIn(&http.Request{})
	if !loggedIn {
		t.Fatal("client not logged in")
	}
}

func TestIsLoggedOut(t *testing.T) {
	store := &MockStore{
		Values: map[interface{}]interface{}{},
	}
	manager := NewSessionManager(nil, store)
	loggedIn := manager.IsLoggedIn(&http.Request{})
	if loggedIn {
		t.Fatal("client logged in")
	}
}

func TestLogout(t *testing.T) {
	store := &MockStore{
		Values: map[interface{}]interface{}{
			AccessTokenKey:  &oauth.AccessToken{},
			RequestTokenKey: &oauth.RequestToken{},
		},
	}
	manager := NewSessionManager(nil, store)
	err := manager.Logout(mockResponseWriter(), &http.Request{})
	if err != nil {
		t.Fatal("error logging out of session: %s", err)
	}

	_, ok := store.Values[AccessTokenKey].(*oauth.AccessToken)
	if ok {
		t.Fatal("session still contains a valid access token")
	}

	_, ok = store.Values[RequestTokenKey].(*oauth.RequestToken)
	if ok {
		t.Fatal("session still contains a valid request token")
	}
}

func TestLogoutSaveError(t *testing.T) {
	store := &MockStore{
		Values:    make(map[interface{}]interface{}),
		SaveError: errors.New("error"),
	}
	manager := NewSessionManager(nil, store)
	err := manager.Logout(mockResponseWriter(), &http.Request{})
	if err == nil {
		t.Fatal("error not returned on store.Save failure")
	}
}

func TestLoginCorrectURL(t *testing.T) {
	url := "http://example.com/login"
	consumer := &MockConsumer{
		RequestToken: &oauth.RequestToken{},
		LoginURL:     url,
	}
	manager := NewSessionManager(consumer, mockStore())

	loginURL, err := manager.Login(mockResponseWriter(), &http.Request{}, "http://url")
	if err != nil {
		t.Fatalf("error returned loggin in: %s", err)
	}

	if loginURL != url {
		t.Fatalf("login did not return the expected login URL\n"+
			"\texpected: %s\n\tactual: %s",
			url,
			loginURL)
	}
}

func TestLoginRequestTokenSaved(t *testing.T) {
	url := "http://example.com/login"
	consumer := &MockConsumer{
		RequestToken: &oauth.RequestToken{},
		LoginURL:     url,
	}
	store := mockStore()
	manager := NewSessionManager(consumer, store)

	_, err := manager.Login(mockResponseWriter(), &http.Request{}, "http://url")
	if err != nil {
		t.Fatalf("error returned loggin in: %s", err)
	}

	token, ok := store.Values[RequestTokenKey].(*oauth.RequestToken)
	if !ok || token == nil {
		t.Fatal("no request token saved after login")
	}
}

func TestLoginRequestTokenErr(t *testing.T) {
	consumer := &MockConsumer{
		Err: errors.New("error"),
	}
	manager := NewSessionManager(consumer, mockStore())

	_, err := manager.Login(mockResponseWriter(), &http.Request{}, "http://url")
	if err == nil {
		t.Fatal("login did not return error when consumer fails")
	}
}

func TestLoginStoreSaveError(t *testing.T) {
	consumer := &MockConsumer{
		RequestToken: &oauth.RequestToken{},
	}
	store := mockStore()
	store.SaveError = errors.New("error")
	manager := NewSessionManager(consumer, store)

	_, err := manager.Login(mockResponseWriter(), &http.Request{}, "http://url")
	if err == nil {
		t.Fatal("login did not return error when store.Save fails")
	}
}

func TestAuthenticateWithVerificationCode(t *testing.T) {
	url, _ := url.Parse(
		fmt.Sprintf("http://example.com/context?%s=%s",
			oauthVerifierKey,
			"abcd"))
	request := &http.Request{URL: url}
	consumer := &MockConsumer{
		AccessToken: &oauth.AccessToken{},
	}
	store := mockStore()
	store.Values[RequestTokenKey] = &oauth.RequestToken{}

	manager := NewSessionManager(consumer, store)
	err := manager.Authenticate(mockResponseWriter(), request)

	if err != nil {
		t.Fatalf("error when creating client with verification code")
	}
}

func TestAuthenticateWithNoVerificationCode(t *testing.T) {
	url, _ := url.Parse("http://example.com/context")
	request := &http.Request{URL: url}
	consumer := &MockConsumer{
		AccessToken: &oauth.AccessToken{},
	}
	store := mockStore()
	store.Values[RequestTokenKey] = &oauth.RequestToken{}

	manager := NewSessionManager(consumer, store)
	err := manager.Authenticate(mockResponseWriter(), request)

	if err == nil {
		t.Fatalf("no error when creating client with no verification code")
	}
}

func TestAuthenticateWithVerificationCodeNoRequestToken(t *testing.T) {
	url, _ := url.Parse(
		fmt.Sprintf("http://example.com/context?%s=%s",
			oauthVerifierKey,
			"abcd"))
	request := &http.Request{URL: url}
	consumer := &MockConsumer{
		AccessToken: &oauth.AccessToken{},
	}

	// No RequestToken
	store := mockStore()

	manager := NewSessionManager(consumer, store)
	err := manager.Authenticate(mockResponseWriter(), request)

	if err == nil {
		t.Fatalf("no error when creating client with no request token")
	}
}

func TestAuthenticateWithVerificationCodeErrorAuthorizingToken(t *testing.T) {
	url, _ := url.Parse(
		fmt.Sprintf("http://example.com/context?%s=%s",
			oauthVerifierKey,
			"abcd"))
	request := &http.Request{URL: url}
	consumer := &MockConsumer{
		Err: errors.New("error"),
	}
	store := mockStore()
	store.Values[RequestTokenKey] = &oauth.RequestToken{}

	manager := NewSessionManager(consumer, store)
	err := manager.Authenticate(mockResponseWriter(), request)

	if err == nil {
		t.Fatalf("no error when creating client with no request token")
	}
}

func TestAuthenticateStoreGetError(t *testing.T) {
	url, _ := url.Parse(
		fmt.Sprintf("http://example.com/context?%s=%s",
			oauthVerifierKey,
			"abcd"))
	request := &http.Request{URL: url}
	consumer := &MockConsumer{
		AccessToken: &oauth.AccessToken{},
	}
	store := mockStore()
	store.GetError = errors.New("error")
	store.Values[RequestTokenKey] = &oauth.RequestToken{}

	manager := NewSessionManager(consumer, store)
	err := manager.Authenticate(mockResponseWriter(), request)

	if err != nil {
		t.Fatalf("error when creating client when store Get throws error")
	}
}

func TestAuthenticateStoreSaveError(t *testing.T) {
	url, _ := url.Parse(
		fmt.Sprintf("http://example.com/context?%s=%s",
			oauthVerifierKey,
			"abcd"))
	request := &http.Request{URL: url}
	consumer := &MockConsumer{
		AccessToken: &oauth.AccessToken{},
	}
	store := mockStore()
	store.SaveError = errors.New("error")
	store.Values[RequestTokenKey] = &oauth.RequestToken{}

	manager := NewSessionManager(consumer, store)
	err := manager.Authenticate(mockResponseWriter(), request)

	if err == nil {
		t.Fatalf("no error when creating client when store Save throws error")
	}
}

func TestGetClientNoAuthenticatedSession(t *testing.T) {
	url, _ := url.Parse("http://example.com/context")
	request := &http.Request{URL: url}
	consumer := &MockConsumer{
		AccessToken: &oauth.AccessToken{},
	}
	store := mockStore()
	store.Values[RequestTokenKey] = &oauth.RequestToken{}

	manager := NewSessionManager(consumer, store)
	_, err := manager.GetClient(mockResponseWriter(), request)

	if err == nil {
		t.Fatalf("no error when creating client with no authenticated session")
	}
}

func TestGetClientStoreGetError(t *testing.T) {
	consumer := &MockConsumer{
		AccessToken: &oauth.AccessToken{},
	}
	store := mockStore()
	store.GetError = errors.New("error")
	store.Values[AccessTokenKey] = &oauth.AccessToken{}
	store.Values[SessionIDKey] = "123"
	currentTime := time.Now()
	store.Values[LastRefreshTime] = &currentTime

	manager := NewSessionManager(consumer, store)
	client, err := manager.GetClient(mockResponseWriter(), &http.Request{})

	if err != nil {
		t.Fatalf("error creating client with existing access token: %s", err)
	}

	if client == nil {
		t.Fatalf("no client created when access token already exists")
	}
}

func TestGetClientAccessTokenExists(t *testing.T) {
	consumer := &MockConsumer{
		AccessToken: &oauth.AccessToken{},
	}
	store := mockStore()
	store.Values[AccessTokenKey] = &oauth.AccessToken{}
	store.Values[SessionIDKey] = "123"
	currentTime := time.Now()
	store.Values[LastRefreshTime] = &currentTime

	manager := NewSessionManager(consumer, store)
	client, err := manager.GetClient(mockResponseWriter(), &http.Request{})

	if err != nil {
		t.Fatalf("error creating client with existing access token: %s", err)
	}

	if client == nil {
		t.Fatalf("no client created when access token already exists")
	}
}

func TestGetClientAccessTokenExistsTokenSaved(t *testing.T) {
	consumer := &MockConsumer{
		AccessToken: &oauth.AccessToken{},
	}
	store := mockStore()
	store.Values[AccessTokenKey] = &oauth.AccessToken{}
	store.Values[SessionIDKey] = "123"
	currentTime := time.Now()
	store.Values[LastRefreshTime] = &currentTime

	manager := NewSessionManager(consumer, store)
	_, err := manager.GetClient(mockResponseWriter(), &http.Request{})

	if err != nil {
		t.Fatalf("error creating client with existing access token: %s", err)
	}

	token, ok := store.Values[AccessTokenKey].(*oauth.AccessToken)
	if !ok || token == nil {
		t.Fatal("no token saved in session")
	}
}

func TestGetClientLastRefreshTimeOldRefreshesToken(t *testing.T) {
	consumer := &MockConsumer{
		AccessToken: &oauth.AccessToken{},
	}

	store := mockStore()
	store.Values[AccessTokenKey] = &oauth.AccessToken{}
	store.Values[SessionIDKey] = "123"
	refreshTime := time.Now().Add(-61 * time.Minute)
	store.Values[LastRefreshTime] = &refreshTime

	manager := NewSessionManager(consumer, store)
	client, err := manager.GetClient(mockResponseWriter(), &http.Request{})

	if err != nil {
		t.Fatalf("error creating client when access token needed to be"+
			"refreshed: %s", err)
	}

	if client == nil {
		t.Fatalf("no client created when access token needed to be refresehd")
	}
}

func TestGetClientAccessTokenExistsRefreshFails(t *testing.T) {
	consumer := &MockConsumer{
		Err: errors.New("error"),
	}
	store := mockStore()
	store.Values[AccessTokenKey] = &oauth.AccessToken{}

	manager := NewSessionManager(consumer, store)
	_, err := manager.GetClient(mockResponseWriter(), &http.Request{})

	if err == nil {
		t.Fatalf("no error returned, refreshing token should have failed")
	}
}

func TestGetClientAccessStoreSaveError(t *testing.T) {
	consumer := &MockConsumer{
		AccessToken: &oauth.AccessToken{},
	}
	store := mockStore()
	store.Values[AccessTokenKey] = &oauth.AccessToken{}
	store.Values[SessionIDKey] = "123"
	currentTime := time.Now()
	store.Values[LastRefreshTime] = &currentTime
	store.SaveError = errors.New("error")

	manager := NewSessionManager(consumer, store)
	_, err := manager.GetClient(mockResponseWriter(), &http.Request{})

	if err == nil {
		t.Fatalf("no error returned, saving token should have failed")
	}
}

type MockConsumer struct {
	AccessToken  *oauth.AccessToken
	RequestToken *oauth.RequestToken
	LoginURL     string
	Err          error
}

func (m *MockConsumer) GetRequestTokenAndUrl(url string) (*oauth.RequestToken, string, error) {
	return m.RequestToken, m.LoginURL, m.Err
}

func (m *MockConsumer) AuthorizeToken(r *oauth.RequestToken, code string) (*oauth.AccessToken, error) {
	return m.AccessToken, m.Err
}

func (m *MockConsumer) RefreshToken(a *oauth.AccessToken) (*oauth.AccessToken, error) {
	return m.AccessToken, m.Err
}

func (m *MockConsumer) Get(url string, data map[string]string, token *oauth.AccessToken) (*http.Response, error) {
	return &http.Response{}, nil
}

type MockResponseWriter struct {
	content string
}

func (m *MockResponseWriter) Header() http.Header {
	return nil
}

func (m *MockResponseWriter) Write(b []byte) (n int, err error) {
	m.content = string(b)
	return 0, nil
}

func (m *MockResponseWriter) WriteHeader(i int) {
}

func mockResponseWriter() http.ResponseWriter {
	return &MockResponseWriter{}
}

func mockStore() *MockStore {
	return &MockStore{
		Values: make(map[interface{}]interface{}),
	}
}

type MockStore struct {
	Values    map[interface{}]interface{}
	GetError  error
	NewError  error
	SaveError error
}

func (m *MockStore) Get(req *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(m, name)
	session.Values = m.Values
	return session, m.GetError
}

func (m *MockStore) New(r *http.Request, name string) (*sessions.Session, error) {
	if m.NewError != nil {
		return nil, m.NewError
	}

	session := sessions.NewSession(m, name)
	session.Values = m.Values
	return session, nil
}

func (m *MockStore) Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error {
	m.Values = s.Values
	return m.SaveError
}
