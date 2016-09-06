package site

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Forestmb/goff"
	"github.com/Forestmb/power-league/rankings"
	"github.com/Forestmb/power-league/templates"
)

func TestNewSite(t *testing.T) {
	baseContext := "base-context"
	staticContext := "static-context"
	trackingID := "tracking-id"
	site := NewSite(baseContext, staticContext, "templates/", trackingID, &MockSessionManager{})

	if site == nil {
		t.Fatal("no site created")
	}
}

func TestHandleAbout(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/about", nil)
	mockTemplates := &MockTemplates{}
	site := &Site{
		config: &templates.SiteConfig{},
		sessionManager: &MockSessionManager{
			IsLoggedInRet: false,
		},
		templates: mockTemplates,
	}

	handleAbout(site, recorder, request)

	if mockTemplates.LastAboutContent.SiteConfig != site.config {
		t.Fatalf("Unexpected site config passed into templates:\n\t"+
			"Expected: %+v\n\tActual: %+v",
			*site.config,
			*(mockTemplates.LastAboutContent.SiteConfig))
	}

	if mockTemplates.LastAboutContent.LoggedIn != false {
		t.Fatal("Unexpected logged in status passed into templates:\n\t" +
			"Expected: false\n\tActual: true")
	}
}

func TestHandleAboutError(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/about", nil)
	mockTemplates := &MockTemplates{
		WriteAboutError: errors.New("error"),
	}
	site := &Site{
		config: &templates.SiteConfig{},
		sessionManager: &MockSessionManager{
			IsLoggedInRet: true,
		},
		templates: mockTemplates,
	}

	handleAbout(site, recorder, request)

	if mockTemplates.LastAboutContent.SiteConfig != site.config {
		t.Fatalf("Unexpected site config passed into templates:\n\t"+
			"Expected: %+v\n\tActual: %+v",
			*site.config,
			*(mockTemplates.LastAboutContent.SiteConfig))
	}

	if mockTemplates.LastAboutContent.LoggedIn != true {
		t.Fatal("Unexpected logged in status passed into templates:\n\t" +
			"Expected: true\n\tActual: false")
	}

	assertErrorHandledCorrectly(t, site, mockTemplates, true)
}

func TestHandleLogout(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com:8080/base/logout", nil)
	mockSessionManager := &MockSessionManager{}
	site := &Site{
		config: &templates.SiteConfig{
			BaseContext: "/base",
		},
		sessionManager: mockSessionManager,
		templates:      &MockTemplates{},
	}

	handleLogout(site, recorder, request)

	if mockSessionManager.LogoutWriter != recorder {
		t.Fatalf("Unexpected writer passed into session.Manager.Logout\n\t"+
			"Expected: %+v\n\tActual: %+v",
			*recorder,
			mockSessionManager.LogoutWriter)
	}

	if mockSessionManager.LogoutRequest != request {
		t.Fatalf("Unexpected request passed into session.Manager.Logout\n\t"+
			"Expected: %+v\n\tActual: %+v",
			*request,
			*(mockSessionManager.LogoutRequest))
	}

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Fatalf("Unexpected response code given when logging out\n\t"+
			"Expected: %d\n\tActual: %d",
			http.StatusTemporaryRedirect,
			recorder.Code)
	}

	redirectURL := recorder.HeaderMap.Get("Location")
	expected := "http://example.com:8080/base"
	if redirectURL != expected {
		t.Fatalf("Redirected to unexpected URL when logging out\n\t"+
			"Expected: %s\n\tActual: %s",
			expected,
			redirectURL)
	}
}

func TestHandleLogoutError(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com:8080/base/logout", nil)
	mockSessionManager := &MockSessionManager{
		IsLoggedInRet: true,
		LogoutError:   errors.New("error"),
	}
	mockTemplates := &MockTemplates{}
	site := &Site{
		config: &templates.SiteConfig{
			BaseContext: "/base",
		},
		sessionManager: mockSessionManager,
		templates:      mockTemplates,
	}

	handleLogout(site, recorder, request)

	if mockSessionManager.LogoutWriter != recorder {
		t.Fatalf("Unexpected writer passed into session.Manager.Logout\n\t"+
			"Expected: %+v\n\tActual: %+v",
			*recorder,
			mockSessionManager.LogoutWriter)
	}

	if mockSessionManager.LogoutRequest != request {
		t.Fatalf("Unexpected request passed into session.Manager.Logout\n\t"+
			"Expected: %+v\n\tActual: %+v",
			*request,
			*(mockSessionManager.LogoutRequest))
	}

	assertErrorHandledCorrectly(t, site, mockTemplates, true)
}

func TestHandleLogin(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com:8080/login", nil)
	authContext := "/auth"
	requestURL := "http://example.com:18080/request/url"
	mockSessionManager := &MockSessionManager{
		LoginURL: requestURL,
	}
	site := &Site{
		config: &templates.SiteConfig{},
		handlers: map[string]*ContextHandler{
			"auth": &ContextHandler{
				Context: authContext,
			},
		},
		sessionManager: mockSessionManager,
		templates:      &MockTemplates{},
	}

	handleLogin(site, recorder, request)

	if mockSessionManager.LoginWriter != recorder {
		t.Fatalf("Unexpected writer passed into session.Manager.Login\n\t"+
			"Expected: %+v\n\tActual: %+v",
			*recorder,
			mockSessionManager.LoginWriter)
	}

	if mockSessionManager.LoginRequest != request {
		t.Fatalf("Unexpected request passed into session.Manager.Login\n\t"+
			"Expected: %+v\n\tActual: %+v",
			*request,
			*(mockSessionManager.LoginRequest))
	}

	expectedAuthURL := "http://example.com:8080/auth"
	if mockSessionManager.LoginAuthURL != expectedAuthURL {
		t.Fatalf("Unexpected auth URL passed into session.Manager.Login\n\t"+
			"Expected: %s\n\tActual: %s",
			expectedAuthURL,
			mockSessionManager.LoginAuthURL)
	}

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Fatalf("Unexpected response code given when logging in\n\t"+
			"Expected: %d\n\tActual: %d",
			http.StatusTemporaryRedirect,
			recorder.Code)
	}

	redirectURL := recorder.HeaderMap.Get("Location")
	if redirectURL != requestURL {
		t.Fatalf("Redirected to unexpected URL when loggin in\n\t"+
			"Expected: %s\n\tActual: %s",
			requestURL,
			redirectURL)
	}
}

func TestHandleLoginError(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com:8080/login", nil)
	authContext := "/auth"
	mockSessionManager := &MockSessionManager{
		LoginError: errors.New("error"),
	}
	mockTemplates := &MockTemplates{}
	site := &Site{
		config: &templates.SiteConfig{},
		handlers: map[string]*ContextHandler{
			"auth": &ContextHandler{
				Context: authContext,
			},
		},
		sessionManager: mockSessionManager,
		templates:      mockTemplates,
	}

	handleLogin(site, recorder, request)

	if mockSessionManager.LoginWriter != recorder {
		t.Fatalf("Unexpected writer passed into session.Manager.Login\n\t"+
			"Expected: %+v\n\tActual: %+v",
			*recorder,
			mockSessionManager.LoginWriter)
	}

	if mockSessionManager.LoginRequest != request {
		t.Fatalf("Unexpected request passed into session.Manager.Login\n\t"+
			"Expected: %+v\n\tActual: %+v",
			*request,
			*(mockSessionManager.LoginRequest))
	}

	expectedAuthURL := "http://example.com:8080/auth"
	if mockSessionManager.LoginAuthURL != expectedAuthURL {
		t.Fatalf("Unexpected auth URL passed into session.Manager.Login\n\t"+
			"Expected: %s\n\tActual: %s",
			expectedAuthURL,
			mockSessionManager.LoginAuthURL)
	}

	assertErrorHandledCorrectly(t, site, mockTemplates, false)
}

func TestHandleAuthentication(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com:8080/auth", nil)
	leaguesContext := "/leagues"
	mockSessionManager := &MockSessionManager{}
	site := &Site{
		config: &templates.SiteConfig{},
		handlers: map[string]*ContextHandler{
			"showLeagues": &ContextHandler{
				Context: leaguesContext,
			},
		},
		sessionManager: mockSessionManager,
		templates:      &MockTemplates{},
	}

	handleAuthentication(site, recorder, request)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Fatalf("Unexpected response code given when authenticating\n\t"+
			"Expected: %d\n\tActual: %d",
			http.StatusTemporaryRedirect,
			recorder.Code)
	}

	redirectURL := recorder.HeaderMap.Get("Location")
	expected := "http://example.com:8080" + leaguesContext
	if redirectURL != expected {
		t.Fatalf("Redirected to unexpected URL when authenticating\n\t"+
			"Expected: %s\n\tActual: %s",
			expected,
			redirectURL)
	}
}

func TestHandleAuthenticationError(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com:8080/base/auth", nil)
	baseContext := "/base"
	mockSessionManager := &MockSessionManager{
		AuthError: errors.New("error"),
	}
	site := &Site{
		config: &templates.SiteConfig{
			BaseContext: baseContext,
		},
		handlers: map[string]*ContextHandler{
			"showLeagues": &ContextHandler{
				Context: "/leagues",
			},
		},
		sessionManager: mockSessionManager,
		templates:      &MockTemplates{},
	}

	handleAuthentication(site, recorder, request)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Fatalf("Unexpected response code given when an error occurs while "+
			"authenticating\n\tExpected: %d\n\tActual: %d",
			http.StatusTemporaryRedirect,
			recorder.Code)
	}

	redirectURL := recorder.HeaderMap.Get("Location")
	expected := "http://example.com:8080" + baseContext
	if redirectURL != expected {
		t.Fatalf("Redirected to unexpected URL when an error occurs while "+
			"authenticating\n\tExpected: %s\n\tActual: %s",
			expected,
			redirectURL)
	}
}

func TestHandleShowLeagues(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com:8080/base/leagues", nil)
	baseContext := "/base"
	leagues := []goff.League{
		goff.League{
			Name: "League 1",
		},
		goff.League{
			Name: "League 2",
		},
		goff.League{
			Name: "League 3",
		},
		goff.League{
			Name: "League 4",
		},
	}
	mockTemplates := &MockTemplates{}
	mockSessionManager := &MockSessionManager{
		IsLoggedInRet: true,
		Client: &goff.Client{
			Provider: &MockedContentProvider{
				content: &goff.FantasyContent{
					Users: []goff.User{
						goff.User{
							Games: []goff.Game{
								goff.Game{
									Leagues: leagues,
								},
							},
						},
					},
				},
			},
		},
	}
	site := &Site{
		config: &templates.SiteConfig{
			BaseContext: baseContext,
		},
		sessionManager: mockSessionManager,
		templates:      mockTemplates,
	}

	handleShowLeagues(site, recorder, request)

	if mockTemplates.LastLeaguesContent.SiteConfig != site.config {
		t.Fatalf("Unexpected site config passed into templates:\n\t"+
			"Expected: %+v\n\tActual: %+v",
			*site.config,
			*(mockTemplates.LastLeaguesContent.SiteConfig))
	}

	if mockTemplates.LastLeaguesContent.LoggedIn != true {
		t.Fatal("Unexpected logged in status passed into templates:\n\t" +
			"Expected: true\n\tActual: false")
	}

	for _, yearlyLeagues := range mockTemplates.LastLeaguesContent.AllYears {
		if _, ok := goff.YearKeys[yearlyLeagues.Year]; ok {
			assertLeaguesEqual(t, yearlyLeagues.Leagues, leagues)
		}
	}
}

func TestHandleShowLeaguesNotLoggedIn(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com:8080/base/leagues", nil)
	baseContext := "/base"
	mockTemplates := &MockTemplates{}
	mockSessionManager := &MockSessionManager{
		IsLoggedInRet: false,
	}
	site := &Site{
		config: &templates.SiteConfig{
			BaseContext: baseContext,
		},
		sessionManager: mockSessionManager,
		templates:      mockTemplates,
	}

	handleShowLeagues(site, recorder, request)

	if mockTemplates.LastLeaguesContent.SiteConfig != site.config {
		t.Fatalf("Unexpected site config passed into templates:\n\t"+
			"Expected: %+v\n\tActual: %+v",
			*site.config,
			*(mockTemplates.LastLeaguesContent.SiteConfig))
	}

	if mockTemplates.LastLeaguesContent.LoggedIn != false {
		t.Fatal("Unexpected logged in status passed into templates:\n\t" +
			"Expected: false\n\tActual: true")
	}

	if mockTemplates.LastLeaguesContent.AllYears != nil {
		t.Fatalf("Leagues passed into templates when user was not logged "+
			"in: %+v",
			mockTemplates.LastLeaguesContent.AllYears)
	}
}

func TestHandleShowLeaguesGetClientError(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com:8080/base/leagues", nil)
	baseContext := "/base"
	mockTemplates := &MockTemplates{}
	mockSessionManager := &MockSessionManager{
		IsLoggedInRet: true,
		ClientError:   errors.New("error"),
	}
	site := &Site{
		config: &templates.SiteConfig{
			BaseContext: baseContext,
		},
		sessionManager: mockSessionManager,
		templates:      mockTemplates,
	}

	handleShowLeagues(site, recorder, request)

	assertErrorHandledCorrectly(t, site, mockTemplates, true)
}

func TestHandleShowLeaguesWriteTemplateError(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com:8080/base/leagues", nil)
	baseContext := "/base"
	mockTemplates := &MockTemplates{
		WriteLeaguesError: errors.New("error"),
	}
	mockSessionManager := &MockSessionManager{
		IsLoggedInRet: false,
	}
	site := &Site{
		config: &templates.SiteConfig{
			BaseContext: baseContext,
		},
		sessionManager: mockSessionManager,
		templates:      mockTemplates,
	}

	handleShowLeagues(site, recorder, request)

	assertErrorHandledCorrectly(t, site, mockTemplates, false)
}

func TestHandlePowerRankingsNotLoggedIn(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com:8080/league?key=3.2.1", nil)
	baseContext := "/base"
	mockSessionManager := &MockSessionManager{
		IsLoggedInRet: false,
	}
	site := &Site{
		config: &templates.SiteConfig{
			BaseContext: baseContext,
		},
		handlers:       map[string]*ContextHandler{},
		sessionManager: mockSessionManager,
		templates:      &MockTemplates{},
	}

	handlePowerRankings(site, recorder, request)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Fatalf("Unexpected response code given when attempting to access "+
			"rankings when not logged in\n\tExpected: %d\n\tActual: %d",
			http.StatusTemporaryRedirect,
			recorder.Code)
	}

	redirectURL := recorder.HeaderMap.Get("Location")
	expected := "http://example.com:8080" + baseContext
	if redirectURL != expected {
		t.Fatalf("Redirected to unexpected URL when attempting to access "+
			"rankings when not logged in\n\tExpected: %s\n\tActual: %s",
			expected,
			redirectURL)
	}
}

func TestHandlePowerRankingsNoLeagueKey(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com:8080/league", nil)
	leaguesContext := "/leagues"
	mockSessionManager := &MockSessionManager{
		IsLoggedInRet: true,
	}
	site := &Site{
		config: &templates.SiteConfig{
			BaseContext: "/base",
		},
		handlers: map[string]*ContextHandler{
			"showLeagues": &ContextHandler{
				Context: leaguesContext,
			},
		},
		sessionManager: mockSessionManager,
		templates:      &MockTemplates{},
	}

	handlePowerRankings(site, recorder, request)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Fatalf("Unexpected response code given when attempting to access "+
			"rankings when no league key is given\n\t"+
			"Expected: %d\n\tActual: %d",
			http.StatusTemporaryRedirect,
			recorder.Code)
	}

	redirectURL := recorder.HeaderMap.Get("Location")
	expected := "http://example.com:8080" + leaguesContext
	if redirectURL != expected {
		t.Fatalf("Redirected to unexpected URL when attempting to access "+
			"rankings when no league key is given\n\t"+
			"Expected: %s\n\tActual: %s",
			expected,
			redirectURL)
	}
}

func TestHandlePowerRankingsGetClientError(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com:8080/league?key=3.2.1", nil)
	baseContext := "/base"
	mockSessionManager := &MockSessionManager{
		IsLoggedInRet: true,
		ClientError:   errors.New("error"),
	}
	mockTemplates := &MockTemplates{}
	site := &Site{
		config: &templates.SiteConfig{
			BaseContext: baseContext,
		},
		handlers:       map[string]*ContextHandler{},
		sessionManager: mockSessionManager,
		templates:      mockTemplates,
	}

	handlePowerRankings(site, recorder, request)

	assertErrorHandledCorrectly(t, site, mockTemplates, true)
}

func TestHandlePowerRankingsGetLeagueErrorAccessDenied(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com:8080/league?key=3.2.1", nil)
	baseContext := "/base"
	mockSessionManager := &MockSessionManager{
		IsLoggedInRet: true,
		Client: &goff.Client{
			Provider: &MockedContentProvider{
				content: nil,
				err:     goff.ErrAccessDenied,
			},
		},
	}
	mockTemplates := &MockTemplates{}
	site := &Site{
		config: &templates.SiteConfig{
			BaseContext: baseContext,
		},
		handlers:       map[string]*ContextHandler{},
		sessionManager: mockSessionManager,
		templates:      mockTemplates,
	}

	handlePowerRankings(site, recorder, request)

	assertErrorHandledCorrectly(t, site, mockTemplates, true)
	expected := "You do not have permission to access this league."
	if mockTemplates.LastErrorContent.Message != expected {
		t.Fatalf("Unexpected error message when user does not have access "+
			"to get league information:\n\tExpected: %s\n\tActual: %s",
			expected,
			mockTemplates.LastErrorContent.Message)
	}
}

func TestHandlePowerRankingsGetLeagueError(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com:8080/league?key=3.2.1", nil)
	baseContext := "/base"
	mockSessionManager := &MockSessionManager{
		IsLoggedInRet: true,
		Client: &goff.Client{
			Provider: &MockedContentProvider{
				content: nil,
				err:     errors.New("error"),
			},
		},
	}
	mockTemplates := &MockTemplates{}
	site := &Site{
		config: &templates.SiteConfig{
			BaseContext: baseContext,
		},
		handlers:       map[string]*ContextHandler{},
		sessionManager: mockSessionManager,
		templates:      mockTemplates,
	}

	handlePowerRankings(site, recorder, request)

	assertErrorHandledCorrectly(t, site, mockTemplates, true)
}

func TestHandlePowerRankingsWriteErrorTemplateError(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://example.com:8080/league?key=3.2.1", nil)
	baseContext := "/base"
	mockSessionManager := &MockSessionManager{
		IsLoggedInRet: true,
		ClientError:   errors.New("error"),
	}
	mockTemplates := &MockTemplates{
		WriteErrorError: errors.New("another error"),
	}
	site := &Site{
		config: &templates.SiteConfig{
			BaseContext: baseContext,
		},
		handlers:       map[string]*ContextHandler{},
		sessionManager: mockSessionManager,
		templates:      mockTemplates,
	}

	handlePowerRankings(site, recorder, request)

	assertErrorHandledCorrectly(t, site, mockTemplates, true)
}

func TestChooseSchemeFromRequestURLParameter(t *testing.T) {
	unexpected := mockRecordScheme{}
	expected := mockScoreScheme{}
	request, _ := http.NewRequest(
		"GET",
		"http://example.com:8080/context?scheme="+expected.ID(),
		nil)
	actual := chooseSchemeFromRequest(request, []rankings.Scheme{unexpected, expected})

	if expected.ID() != actual.ID() {
		t.Fatalf("Unexpected scheme chosen from request using URL "+
			"parameter:\n\tExpected: %s\n\tActual: %s",
			expected.ID(),
			actual.ID())
	}

	request, _ = http.NewRequest(
		"GET",
		"http://example.com:8080/context?scheme=invalid-"+expected.ID(),
		nil)
	actual = chooseSchemeFromRequest(request, []rankings.Scheme{expected, unexpected})

	if expected.ID() != actual.ID() {
		t.Fatalf("Unexpected scheme chosen from request using URL "+
			"parameter:\n\tExpected: %s\n\tActual: %s",
			expected.ID(),
			actual.ID())
	}
}

func TestChooseSchemeFromRequestCookie(t *testing.T) {
	unexpected := mockRecordScheme{}
	expected := mockScoreScheme{}
	request, _ := http.NewRequest("GET", "http://example.com:8080/context", nil)
	request.AddCookie(&http.Cookie{
		Name:   "PowerPreference",
		Value:  expected.ID(),
		Path:   "/",
		Domain: "example.com",
	})
	actual := chooseSchemeFromRequest(request, []rankings.Scheme{unexpected, expected})

	if expected.ID() != actual.ID() {
		t.Fatalf("Unexpected scheme chosen from request using URL "+
			"parameter:\n\tExpected: %s\n\tActual: %s",
			expected.ID(),
			actual.ID())
	}

	request, _ = http.NewRequest("GET", "http://example.com:8080/context", nil)
	request.AddCookie(&http.Cookie{
		Name:   "PowerPreference",
		Value:  "invalid-id",
		Path:   "/",
		Domain: "example.com",
	})
	actual = chooseSchemeFromRequest(request, []rankings.Scheme{expected, unexpected})
	if expected.ID() != actual.ID() {
		t.Fatalf("Unexpected scheme chosen from request using URL "+
			"parameter:\n\tExpected: %s\n\tActual: %s",
			expected.ID(),
			actual.ID())
	}

	request, _ = http.NewRequest(
		"GET",
		"http://example.com:8080/context?scheme="+expected.ID(),
		nil)
	request.AddCookie(&http.Cookie{
		Name:   "PowerPreference",
		Value:  unexpected.ID(),
		Path:   "/",
		Domain: "example.com",
	})
	actual = chooseSchemeFromRequest(request, []rankings.Scheme{unexpected, expected})
	if expected.ID() != actual.ID() {
		t.Fatalf("Unexpected scheme chosen from request using URL "+
			"parameter:\n\tExpected: %s\n\tActual: %s",
			expected.ID(),
			actual.ID())
	}
}

func TestGetUserLeagues(t *testing.T) {
	year := "2012"
	client := &MockUserLeaguesClient{
		Leagues: map[string][]goff.League{
			year: []goff.League{
				goff.League{
					Name: "League 1",
				},
				goff.League{
					Name: "League 2",
				},
				goff.League{
					Name: "League 3",
				},
			},
		},
	}
	results := make(chan *templates.YearlyLeagues)
	go getUserLeauges(client, 2012, results)

	yearlyLeagues := <-results
	if yearlyLeagues.Year != year {
		t.Fatalf("Leagues returned for the wrong year:\n\t"+
			"Expected: %s\n\tActual: %s",
			year,
			yearlyLeagues.Year)
	}

	assertLeaguesEqual(t, yearlyLeagues.Leagues, client.Leagues[year])
}

func TestGetUserLeaguesError(t *testing.T) {
	year := "2012"
	client := &MockUserLeaguesClient{
		Error: errors.New("error"),
	}
	results := make(chan *templates.YearlyLeagues)
	go getUserLeauges(client, 2012, results)

	yearlyLeagues := <-results
	if yearlyLeagues.Year != year {
		t.Fatalf("Leagues returned for the wrong year:\n\t"+
			"Expected: %s\n\tActual: %s",
			year,
			yearlyLeagues.Year)
	}

	if len(yearlyLeagues.Leagues) != 0 {
		t.Fatalf("Leagues present after client returned error: %+v",
			yearlyLeagues.Leagues)
	}
}

func TestGetAllYearlyLeagues(t *testing.T) {
	client := &MockUserLeaguesClient{
		Leagues: map[string][]goff.League{
			"2012": []goff.League{
				goff.League{
					Name: "League 1",
				},
				goff.League{
					Name: "League 2",
				},
				goff.League{
					Name: "League 3",
				},
			},
			"2010": []goff.League{
				goff.League{
					Name: "League 1",
				},
				goff.League{
					Name: "League 2",
				},
				goff.League{
					Name: "League 3",
				},
			},
			"2007": []goff.League{
				goff.League{
					Name: "League 1",
				},
				goff.League{
					Name: "League 2",
				},
				goff.League{
					Name: "League 3",
				},
			},
			"2006": []goff.League{
				goff.League{
					Name: "League 1",
				},
				goff.League{
					Name: "League 2",
				},
				goff.League{
					Name: "League 3",
				},
			},
		},
	}
	allYearlyLeagues := getAllYearlyLeagues(client)

	for _, yearlyLeagues := range allYearlyLeagues {
		assertLeaguesEqual(t, yearlyLeagues.Leagues, client.Leagues[yearlyLeagues.Year])
	}
}

func TestYahooClientNoZeroPoints(t *testing.T) {
	goffClient := &MockGoffClient{
		AllTeamStats: []goff.Team{
			goff.Team{
				TeamPoints: goff.Points{
					Total: 10.0,
				},
			},
			goff.Team{
				TeamPoints: goff.Points{
					Total: 11.0,
				},
			},
			goff.Team{
				TeamPoints: goff.Points{
					Total: 12.0,
				},
			},
		},
	}
	client := &YahooClient{Client: goffClient}

	_, err := client.GetAllTeamStats("123", 12, false)
	if err != nil {
		t.Fatalf("error returned getting team stats: %s", err)
	}
}

func TestYahooClientAllTeamStatsError(t *testing.T) {
	goffClient := &MockGoffClient{
		AllTeamStatsError: errors.New("error"),
	}
	client := &YahooClient{Client: goffClient}

	_, err := client.GetAllTeamStats("123", 12, false)
	if err == nil {
		t.Fatalf("error was not returned after getting all team stats failed")
	}
}

func TestYahooClientZeroPointsRosterError(t *testing.T) {
	goffClient := &MockGoffClient{
		AllTeamStats: []goff.Team{
			goff.Team{
				TeamPoints: goff.Points{
					Total: 0.0,
				},
			},
		},
		TeamRosterError: errors.New("error"),
	}
	client := &YahooClient{Client: goffClient}

	_, err := client.GetAllTeamStats("123", 12, false)
	if err == nil {
		t.Fatalf("no error getting all team stats when team had 0 points " +
			"and getting the roster failed")
	}
}

func TestYahooClientZeroPointsPlayerStatsError(t *testing.T) {
	goffClient := &MockGoffClient{
		AllTeamStats: []goff.Team{
			goff.Team{
				TeamPoints: goff.Points{
					Total: 0.0,
				},
			},
		},
		TeamRoster: []goff.Player{
			goff.Player{
				SelectedPosition: goff.SelectedPosition{
					Position: "BN",
				},
			},
			goff.Player{
				SelectedPosition: goff.SelectedPosition{
					Position: "BN",
				},
			},
			goff.Player{
				SelectedPosition: goff.SelectedPosition{
					Position: "QB",
				},
			},
			goff.Player{
				SelectedPosition: goff.SelectedPosition{
					Position: "RB",
				},
			},
		},
		PlayersStatsError: errors.New("error"),
	}
	client := &YahooClient{Client: goffClient}

	_, err := client.GetAllTeamStats("123", 12, false)
	if err == nil {
		t.Fatalf("no error getting all team stats when team had 0 points " +
			"and getting player stats failed")
	}
}

func TestYahooClientAllTeamStatsZeroPoints(t *testing.T) {
	goffClient := &MockGoffClient{
		AllTeamStats: []goff.Team{
			goff.Team{
				TeamPoints: goff.Points{
					Total: 10.0,
				},
			},
			goff.Team{
				TeamPoints: goff.Points{
					Total: 11.0,
				},
			},
			goff.Team{
				TeamPoints: goff.Points{
					Total: 12.0,
				},
			},
			goff.Team{
				TeamPoints: goff.Points{
					Total: 0.0,
				},
			},
		},
		TeamRoster: []goff.Player{
			goff.Player{
				SelectedPosition: goff.SelectedPosition{
					Position: "BN",
				},
			},
			goff.Player{
				SelectedPosition: goff.SelectedPosition{
					Position: "BN",
				},
			},
			goff.Player{
				SelectedPosition: goff.SelectedPosition{
					Position: "QB",
				},
			},
			goff.Player{
				SelectedPosition: goff.SelectedPosition{
					Position: "RB",
				},
			},
		},
		PlayersStats: []goff.Player{
			goff.Player{
				PlayerPoints: goff.Points{
					Total: 12.0,
				},
			},
			goff.Player{
				PlayerPoints: goff.Points{
					Total: 11.0,
				},
			},
		},
	}
	client := &YahooClient{Client: goffClient}

	teams, err := client.GetAllTeamStats("123", 12, false)
	if err != nil {
		t.Fatalf("unexpected error getting all team stats: %s", err)
	}

	if teams[3].TeamPoints.Total != 23.0 {
		t.Fatalf("incorrect total calculated for a zeroed team\n"+
			"\texpected: %f\n\tactual: %f",
			23.0,
			teams[3].TeamPoints.Total)
	}
}

func assertErrorHandledCorrectly(
	t *testing.T,
	s *Site,
	m *MockTemplates,
	loggedIn bool) {

	if m.LastErrorContent == nil {
		t.Fatal("No error content passed into templates")
	}

	if m.LastErrorContent.Message == "" {
		t.Fatal("No message passed into error page")
	}

	if m.LastErrorContent.SiteConfig != s.config {
		t.Fatalf("Wrong site config passed into error page:\n\t"+
			"Expected: %+v\n\tActual: %+v",
			*(s.config),
			*(m.LastErrorContent.SiteConfig))
	}

	if m.LastErrorContent.LoggedIn != loggedIn {
		if loggedIn {
			t.Fatal("User should be logged in when displaying error page")
		} else {
			t.Fatal("User should not be logged in when displaying error page")
		}
	}
}

type MockGoffClient struct {
	AllTeamStats      []goff.Team
	AllTeamStatsError error
	TeamRoster        []goff.Player
	TeamRosterError   error
	PlayersStats      []goff.Player
	PlayersStatsError error
	Matchups          map[int][]goff.Matchup
	MatchupsError     error
	StandingsLeague   *goff.League
	StandingsError    error
}

func (m *MockGoffClient) GetAllTeamStats(leagueKey string, week int) ([]goff.Team, error) {
	return m.AllTeamStats, m.AllTeamStatsError
}

func (m *MockGoffClient) GetTeamRoster(teamKey string, week int) ([]goff.Player, error) {
	return m.TeamRoster, m.TeamRosterError
}

func (m *MockGoffClient) GetPlayersStats(leagueKey string, week int, players []goff.Player) ([]goff.Player, error) {
	return m.PlayersStats, m.PlayersStatsError
}

func (m *MockGoffClient) GetMatchupsForWeekRange(leagueKey string, startWeek, endWeek int) (map[int][]goff.Matchup, error) {
	return m.Matchups, m.MatchupsError
}

func (m *MockGoffClient) GetLeagueStandings(leagueKey string) (*goff.League, error) {
	return m.StandingsLeague, m.StandingsError
}

type MockSessionManager struct {
	LoginWriter   http.ResponseWriter
	LoginRequest  *http.Request
	LoginAuthURL  string
	LogoutWriter  http.ResponseWriter
	LogoutRequest *http.Request

	LoginURL      string
	LoginError    error
	LogoutError   error
	AuthError     error
	IsLoggedInRet bool
	Client        *goff.Client
	ClientError   error
}

func (m *MockSessionManager) Login(w http.ResponseWriter, r *http.Request, redirectURL string) (loginURL string, err error) {
	m.LoginWriter = w
	m.LoginRequest = r
	m.LoginAuthURL = redirectURL

	return m.LoginURL, m.LoginError
}

func (m *MockSessionManager) Authenticate(w http.ResponseWriter, r *http.Request) error {
	return m.AuthError
}

func (m *MockSessionManager) Logout(w http.ResponseWriter, r *http.Request) error {
	m.LogoutWriter = w
	m.LogoutRequest = r
	return m.LogoutError
}

func (m *MockSessionManager) IsLoggedIn(r *http.Request) bool {
	return m.IsLoggedInRet
}

func (m *MockSessionManager) GetClient(w http.ResponseWriter, r *http.Request) (*goff.Client, error) {
	return m.Client, m.ClientError
}

type MockUserLeaguesClient struct {
	Leagues map[string][]goff.League
	Error   error
}

func (m *MockUserLeaguesClient) GetUserLeagues(year string) ([]goff.League, error) {
	return m.Leagues[year], m.Error
}

type MockTemplates struct {
	WriteAboutError    error
	WriteErrorError    error
	WriteLeaguesError  error
	WriteRankingsError error

	LastAboutContent    *templates.AboutPageContent
	LastErrorContent    *templates.ErrorPageContent
	LastLeaguesContent  *templates.LeaguesPageContent
	LastRankingsContent *templates.RankingsPageContent
}

func (m *MockTemplates) WriteRankingsTemplate(w io.Writer, content *templates.RankingsPageContent) error {
	m.LastRankingsContent = content
	return m.WriteRankingsError
}

func (m *MockTemplates) WriteAboutTemplate(w io.Writer, content *templates.AboutPageContent) error {
	m.LastAboutContent = content
	return m.WriteAboutError
}

func (m *MockTemplates) WriteLeaguesTemplate(w io.Writer, content *templates.LeaguesPageContent) error {
	m.LastLeaguesContent = content
	return m.WriteLeaguesError
}

func (m *MockTemplates) WriteErrorTemplate(w io.Writer, content *templates.ErrorPageContent) error {
	m.LastErrorContent = content
	return m.WriteErrorError
}

// MockedContentProvider creates a goff.ContentProvider that returns the
// given content and error whenever provider.Get is called.
type MockedContentProvider struct {
	lastGetURL string
	content    *goff.FantasyContent
	err        error
	count      int
}

func (m *MockedContentProvider) Get(url string) (*goff.FantasyContent, error) {
	m.lastGetURL = url
	m.count++
	return m.content, m.err
}

func (m *MockedContentProvider) RequestCount() int {
	return m.count
}

type mockScoreScheme struct{}

func (m mockScoreScheme) ID() string {
	return "score-id"
}

func (m mockScoreScheme) DisplayName() string {
	return "Mock Scheme Points"
}

func (m mockScoreScheme) Type() string {
	return rankings.Types.SCORE
}

func (m mockScoreScheme) CalculateWeeklyRankings(
	week int,
	teams []goff.Team,
	projected bool,
	results chan *rankings.WeeklyRanking) {
}

type mockRecordScheme struct{}

func (m mockRecordScheme) ID() string {
	return "record-id"
}

func (m mockRecordScheme) DisplayName() string {
	return "Mock Scheme"
}

func (m mockRecordScheme) Type() string {
	return rankings.Types.RECORD
}

func (m mockRecordScheme) CalculateWeeklyRankings(
	week int,
	teams []goff.Team,
	projected bool,
	results chan *rankings.WeeklyRanking) {
}

func assertLeaguesEqual(t *testing.T, actualLeagues, expectedLeagues []goff.League) {
	if len(actualLeagues) != len(expectedLeagues) {
		t.Fatalf("Unexpected number of leagues returned:\n\t"+
			"Expected: %d\n\tActual: %d",
			len(expectedLeagues),
			len(actualLeagues))
	}

	for i, actual := range actualLeagues {
		expected := expectedLeagues[i]
		if actual.Name != expected.Name {
			t.Fatalf("Unexpected league returned:\n\t"+
				"Expected: %+v\n\tActual: %+v",
				expected,
				actual)
		}
	}
}
