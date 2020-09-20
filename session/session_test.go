package session

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

func TestNewManager(t *testing.T) {
	manager := NewManager(&MockConsumerProvider{}, mockStore())
	if manager == nil {
		t.Fatal("no manager returned")
	}
}

func TestNewManagerWithCache(t *testing.T) {
	manager := NewManagerWithCache(&MockConsumerProvider{}, mockStore(), 30, 10)
	if manager == nil {
		t.Fatal("no manager returned")
	}
}

func TestIsLoggedIn(t *testing.T) {
	store := &MockStore{
		Values: map[interface{}]interface{}{
			AccessTokenKey: &oauth2.Token{},
		},
	}
	manager := NewManager(nil, store)
	loggedIn := manager.IsLoggedIn(&http.Request{})
	if !loggedIn {
		t.Fatal("client not logged in")
	}
}

func TestIsLoggedOut(t *testing.T) {
	store := &MockStore{
		Values: map[interface{}]interface{}{},
	}
	manager := NewManager(nil, store)
	loggedIn := manager.IsLoggedIn(&http.Request{})
	if loggedIn {
		t.Fatal("client logged in")
	}
}

func TestLogout(t *testing.T) {
	store := &MockStore{
		Values: map[interface{}]interface{}{
			AccessTokenKey: &oauth2.Token{},
		},
	}
	manager := NewManager(nil, store)
	err := manager.Logout(mockResponseWriter(), &http.Request{})
	if err != nil {
		t.Fatalf("error logging out of session: %s", err)
	}

	_, ok := store.Values[AccessTokenKey].(*oauth2.Token)
	if ok {
		t.Fatal("session still contains a valid access token")
	}
}

func TestLogoutSaveError(t *testing.T) {
	store := &MockStore{
		Values:    make(map[interface{}]interface{}),
		SaveError: errors.New("error"),
	}
	manager := NewManager(nil, store)
	err := manager.Logout(mockResponseWriter(), &http.Request{})
	if err == nil {
		t.Fatal("error not returned on store.Save failure")
	}
}

func TestLoginCorrectURL(t *testing.T) {
	url := "http://example.com/login"
	consumer := &MockConsumer{
		LoginURL: url,
	}
	manager := NewManager(mockProvider(consumer), mockStore())

	loginURL := manager.Login(mockResponseWriter(), &http.Request{})

	if loginURL != url {
		t.Fatalf("login did not return the expected login URL\n"+
			"\texpected: %s\n\tactual: %s",
			url,
			loginURL)
	}
}

func TestAuthenticateWithVerificationCode(t *testing.T) {
	consumer := &MockConsumer{
		Token: &oauth2.Token{},
	}
	store := mockStore()

	manager := NewManager(mockProvider(consumer), store)
	err := manager.Authenticate(mockResponseWriter(), defaultRequest())

	if err != nil {
		t.Fatalf("error when creating client with verification code")
	}
}

func TestAuthenticateWithWrongState(t *testing.T) {
	request := defaultRequest()
	request.Form.Set("state", oauthState+"wrong-state")

	consumer := &MockConsumer{
		Token: &oauth2.Token{},
	}
	store := mockStore()

	manager := NewManager(mockProvider(consumer), store)
	err := manager.Authenticate(mockResponseWriter(), request)

	if err == nil {
		t.Fatalf("no error when creating client with no state")
	}
}

func TestAuthenticateWithNoVerificationCode(t *testing.T) {
	request := defaultRequest()
	request.Form.Del("code")

	consumer := &MockConsumer{
		Token: &oauth2.Token{},
	}
	store := mockStore()

	manager := NewManager(mockProvider(consumer), store)
	err := manager.Authenticate(mockResponseWriter(), request)

	if err == nil {
		t.Fatalf("no error when creating client with no verification code")
	}
}

func TestAuthenticateWithVerificationCodeErrorAuthorizingToken(t *testing.T) {
	consumer := &MockConsumer{
		Err: errors.New("error"),
	}
	store := mockStore()

	manager := NewManager(mockProvider(consumer), store)
	err := manager.Authenticate(mockResponseWriter(), defaultRequest())

	if err == nil {
		t.Fatalf("no error when creating client with no request token")
	}
}

func TestAuthenticateStoreGetError(t *testing.T) {
	consumer := &MockConsumer{
		Token: &oauth2.Token{},
	}
	store := mockStore()
	store.GetError = errors.New("error")

	manager := NewManager(mockProvider(consumer), store)
	err := manager.Authenticate(mockResponseWriter(), defaultRequest())

	if err != nil {
		t.Fatalf("error when creating client when store Get throws error")
	}
}

func TestAuthenticateStoreSaveError(t *testing.T) {
	consumer := &MockConsumer{
		Token: &oauth2.Token{},
	}
	store := mockStore()
	store.SaveError = errors.New("error")

	manager := NewManager(mockProvider(consumer), store)
	err := manager.Authenticate(mockResponseWriter(), defaultRequest())

	if err == nil {
		t.Fatalf("no error when creating client when store Save throws error")
	}
}

func TestGetClientNoAuthenticatedSession(t *testing.T) {
	consumer := &MockConsumer{
		Token: &oauth2.Token{},
	}
	store := mockStore()

	manager := NewManager(mockProvider(consumer), store)
	_, err := manager.GetClient(mockResponseWriter(), defaultRequest())

	if err == nil {
		t.Fatalf("no error when creating client with no authenticated session")
	}
}

func TestGetClientStoreGetError(t *testing.T) {
	consumer := &MockConsumer{
		Token: &oauth2.Token{},
	}
	store := mockStore()
	store.GetError = errors.New("error")
	store.Values[AccessTokenKey] = &oauth2.Token{}
	store.Values[SessionIDKey] = "123"

	manager := NewManager(mockProvider(consumer), store)
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
		Token: &oauth2.Token{},
	}
	store := mockStore()
	store.Values[AccessTokenKey] = &oauth2.Token{}
	store.Values[SessionIDKey] = "123"

	manager := NewManager(mockProvider(consumer), store)
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
		Token: &oauth2.Token{},
	}
	store := mockStore()
	store.Values[AccessTokenKey] = &oauth2.Token{}
	store.Values[SessionIDKey] = "123"

	manager := NewManager(mockProvider(consumer), store)
	_, err := manager.GetClient(mockResponseWriter(), &http.Request{})

	if err != nil {
		t.Fatalf("error creating client with existing access token: %s", err)
	}

	token, ok := store.Values[AccessTokenKey].(*oauth2.Token)
	if !ok || token == nil {
		t.Fatal("no token saved in session")
	}
}

func TestGetClientAccessStoreSaveError(t *testing.T) {
	consumer := &MockConsumer{
		Token: &oauth2.Token{},
	}
	store := mockStore()
	store.Values[AccessTokenKey] = &oauth2.Token{}
	store.Values[SessionIDKey] = "123"
	store.SaveError = errors.New("error")

	manager := NewManager(mockProvider(consumer), store)
	_, err := manager.GetClient(mockResponseWriter(), &http.Request{})

	if err == nil {
		t.Fatalf("no error returned, saving token should have failed")
	}
}

type MockConsumerProvider struct {
	Consumer *MockConsumer
}

func (p *MockConsumerProvider) Get(r *http.Request) Consumer {
	return p.Consumer
}

func mockProvider(c *MockConsumer) *MockConsumerProvider {
	return &MockConsumerProvider{Consumer: c}
}

type MockConsumer struct {
	Token    *oauth2.Token
	LoginURL string
	Err      error
}

func (m *MockConsumer) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return m.LoginURL
}

func (m *MockConsumer) Exchange(ctx context.Context, verificationCode string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return m.Token, m.Err
}

func (m *MockConsumer) Client(ctx context.Context, token *oauth2.Token) *http.Client {
	return &http.Client{}
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

func defaultRequest() *http.Request {
	values := url.Values{}
	values.Add("state", oauthState)
	values.Add("code", "abcd")
	return &http.Request{
		Form: values,
	}
}
