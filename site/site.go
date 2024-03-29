// Package site publishes endpoints and handles web requests for a
// power-league site.
package site

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/Forestmb/goff"
	"github.com/Forestmb/power-league/rankings"
	"github.com/Forestmb/power-league/session"
	"github.com/Forestmb/power-league/templates"
	"github.com/golang/glog"
)

const (
	// EarliestSupportedYear that fantasy leagues will be displayed for
	EarliestSupportedYear = 2001

	// LatestSupportedYear that fantasy leagues will be displayed for
	LatestSupportedYear = 2021
)

// Site consists of the information needed to run a power rankings site
type Site struct {
	// ServeMux for this site
	ServeMux *http.ServeMux

	handlers       map[string]*ContextHandler
	sessionManager session.Manager
	config         *templates.SiteConfig
	templates      templates.Templates
}

// HandlerFunc is a handler for a given site
type HandlerFunc func(s *Site, w http.ResponseWriter, r *http.Request)

// ContextHandler describes a context handled by this site
type ContextHandler struct {
	// The context being handled
	Context string

	// The function called when the context is accessed
	Func HandlerFunc
}

// ContextHandler adds a new handler for this site
func (s *Site) ContextHandler(id string, context string, f HandlerFunc) {
	fullContext := fmt.Sprintf("%s%s", s.config.BaseContext, context)
	if fullContext != "" {
		glog.V(3).Infof("adding context handler -- id=%s, context=%s",
			id,
			fullContext)
		s.ServeMux.HandleFunc(
			fullContext,
			func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()
				f(s, w, r)
				glog.Infof("request %s took %s",
					r.URL.String(),
					time.Since(start))
			})
	}

	s.handlers[id] = &ContextHandler{
		Context: fullContext,
		Func:    f,
	}
}

// GenerateURL creates a URL to a location on the site based on an existing
// request
func (s *Site) GenerateURL(r *http.Request, context string) string {
	protocol := "https"
	if !s.config.TLS {
		protocol = "http"
	}
	return fmt.Sprintf("%s://%s%s", protocol, r.Host, context)
}

// NewSite creates a new site
func NewSite(
	tls bool,
	baseContext string,
	staticFiles string,
	templatesDir string,
	trackingID string,
	s session.Manager) *Site {

	mux := http.DefaultServeMux

	staticContext := fmt.Sprintf("%s/static/", baseContext)
	mux.Handle(staticContext,
		http.StripPrefix(staticContext,
			http.FileServer(
				http.Dir(staticFiles))))

	site := &Site{
		ServeMux:       mux,
		handlers:       make(map[string]*ContextHandler),
		sessionManager: s,
		config: &templates.SiteConfig{
			TLS:                 tls,
			BaseContext:         baseContext,
			StaticContext:       staticContext,
			AnalyticsTrackingID: trackingID,
		},
		templates: templates.NewTemplatesFromDir(templatesDir),
	}
	site.ContextHandler("base", "", handleShowLeagues)
	site.ContextHandler("showLeagues", "/", handleShowLeagues)
	site.ContextHandler("login", "/login", handleLogin)
	site.ContextHandler("logout", "/logout", handleLogout)
	site.ContextHandler("auth", "/auth", handleAuthentication)
	site.ContextHandler("league", "/league", handlePowerRankings)
	site.ContextHandler("about", "/about", handleAbout)

	return site
}

//
// Handlers
//

func handleAbout(s *Site, w http.ResponseWriter, r *http.Request) {
	glog.V(5).Infoln("in handleAbout")

	loggedIn := s.sessionManager.IsLoggedIn(r)
	aboutContent := &templates.AboutPageContent{
		LoggedIn:   loggedIn,
		SiteConfig: s.config,
	}
	err := s.templates.WriteAboutTemplate(w, aboutContent)
	if err != nil {
		glog.Warningf("error generating about page: %s", err)
		writeErrorPage(
			s,
			w,
			"There was a problem delivering you this page. "+
				"Please try again later.",
			loggedIn)
	}
}

func handleLogout(s *Site, w http.ResponseWriter, r *http.Request) {
	glog.V(5).Infoln("in handleLogout")
	err := s.sessionManager.Logout(w, r)
	if err != nil {
		writeErrorPage(
			s,
			w,
			"There was a problem logging you out.",
			s.sessionManager.IsLoggedIn(r))
		return
	}
	leaguesURL := s.GenerateURL(r, s.config.BaseContext)
	http.Redirect(w, r, leaguesURL, http.StatusTemporaryRedirect)
}

func handleLogin(s *Site, w http.ResponseWriter, r *http.Request) {
	glog.V(5).Infoln("in handleLogin")

	requestURL := s.sessionManager.Login(w, r)

	http.Redirect(w, r, requestURL, http.StatusTemporaryRedirect)
}

func handleAuthentication(s *Site, w http.ResponseWriter, r *http.Request) {
	glog.V(5).Infoln("in handleAuthentication")

	redirectContext := s.handlers["showLeagues"].Context
	err := s.sessionManager.Authenticate(w, r)
	if err != nil {
		glog.Warningf("authentication failed: %s", err)
		redirectContext = s.config.BaseContext
	}

	redirectURL := s.GenerateURL(r, redirectContext)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func handleShowLeagues(s *Site, w http.ResponseWriter, req *http.Request) {
	glog.V(5).Infoln("in handleShowLeagues")
	loggedIn := s.sessionManager.IsLoggedIn(req)
	var allYearlyLeagues []*templates.YearlyLeagues
	if loggedIn {
		client, err := s.sessionManager.GetClient(w, req)
		if err != nil {
			glog.Warningf("error getting client to retreive league list: %s",
				err)
			writeErrorPage(
				s,
				w,
				"There was a problem getting your leagues from Yahoo. "+
					"Please try again later.",
				loggedIn)
			return
		}

		allYearlyLeagues, err = getAllYearlyLeagues(client)
		if err != nil {
			glog.Warningf("error getting all yearly leagues: %s", err)
			writeErrorPage(
				s,
				w,
				"There was a problem delivering you your power rankings. "+
					"Please try again later.",
				loggedIn)
			return
		}
		glog.V(2).Infof("API Request Count: %d", client.RequestCount())
	} else {
		glog.V(2).Infoln("user not logged in, can't show leagues")
	}

	leaguesContent := &templates.LeaguesPageContent{
		AllYears:   allYearlyLeagues,
		LoggedIn:   loggedIn,
		SiteConfig: s.config,
	}
	err := s.templates.WriteLeaguesTemplate(w, leaguesContent)
	if err != nil {
		glog.Warningf("error generating league overview page: %s", err)
		writeErrorPage(
			s,
			w,
			"There was a problem generating your list of leagues. "+
				"Please try again later.",
			loggedIn)
	}
}

func handlePowerRankings(s *Site, w http.ResponseWriter, req *http.Request) {
	glog.V(5).Infoln("in handlePowerRankings")

	loggedIn := s.sessionManager.IsLoggedIn(req)
	if !loggedIn {
		homePage := s.GenerateURL(req, s.config.BaseContext)
		http.Redirect(w, req, homePage, http.StatusTemporaryRedirect)
		return
	}

	// Determine the current week
	currentWeek := -1
	leagueStarted := true
	values := req.URL.Query()
	leagueKey := values.Get("key")
	if leagueKey == "" {
		leaguesContext := s.handlers["showLeagues"].Context
		leaguesURL := s.GenerateURL(req, leaguesContext)
		http.Redirect(w, req, leaguesURL, http.StatusTemporaryRedirect)
		return
	}

	var league *goff.League
	client, err := s.sessionManager.GetClient(w, req)
	if err == nil {
		glog.V(3).Infof("getting metadata -- league=%s", leagueKey)
		league, err = client.GetLeagueMetadata(leagueKey)
		if err == nil {
			if league.IsFinished {
				glog.V(3).Infoln("league is finished")
				currentWeek = league.CurrentWeek
			} else {
				currentWeek = league.CurrentWeek - 1
			}

			leagueStarted = league.DraftStatus == "postdraft"
		} else {
			glog.Warningf("unable to get current week from league metadata: %s", err)
		}
	} else {
		glog.Warningf("unable to create client: %s", err)
	}

	var rankingsContent *templates.RankingsPageContent
	if err == nil {
		glog.V(3).Infof("calculating rankings -- week=%d", currentWeek)
		var leaguePowerData []*rankings.LeaguePowerData
		var schemes []rankings.Scheme
		var chosenScheme rankings.Scheme
		if leagueStarted {
			leaguePowerData, err = rankings.GetPowerData(
				&YahooClient{Client: client},
				league,
				currentWeek)
			if err == nil {
				for _, powerData := range leaguePowerData {
					schemes = append(schemes, powerData.RankingScheme)
				}
				chosenScheme = chooseSchemeFromRequest(req, schemes)
			}
		}

		if err == nil {
			rankingsContent = &templates.RankingsPageContent{
				Weeks:           currentWeek,
				League:          league,
				LeagueStarted:   leagueStarted,
				SchemeToShow:    chosenScheme,
				Schemes:         schemes,
				LeaguePowerData: leaguePowerData,
				LoggedIn:        loggedIn,
				SiteConfig:      s.config,
			}

			err = s.templates.WriteRankingsTemplate(w, rankingsContent)
		}
	}

	if err != nil {
		glog.Warningf("error generating power rankings page: %s", err)
		if err == goff.ErrAccessDenied {
			writeErrorPage(
				s,
				w,
				"You do not have permission to access this league.",
				loggedIn)
		} else {
			writeErrorPage(
				s,
				w,
				"There was a problem delivering you your power rankings. "+
					"Please try again later.",
				loggedIn)
		}
	}

	if client != nil {
		glog.V(2).Infof("API Request Count: %d", client.RequestCount())
	}
}

// Respond to an HTTP request with an error page
func writeErrorPage(
	s *Site,
	w http.ResponseWriter,
	message string,
	loggedIn bool) {

	err := s.templates.WriteErrorTemplate(
		w,
		&templates.ErrorPageContent{
			Message:    message,
			LoggedIn:   loggedIn,
			SiteConfig: s.config,
		})

	if err != nil {
		http.Error(w, message, http.StatusInternalServerError)
	}
}

//
// userLeaguesClient
//

// userLeaguesClient returns the leagues a user is a participant of for the
// given year
type userLeaguesClient interface {
	GetUserLeagues(year string) ([]goff.League, error)
}

func getAllYearlyLeagues(client userLeaguesClient) (templates.AllYearlyLeagues, error) {
	currentYear := LatestSupportedYear
	numberOfYears := (currentYear - EarliestSupportedYear) + 1
	results := make(chan *templates.YearlyLeagues)
	for ; currentYear >= EarliestSupportedYear; currentYear-- {
		go getUserLeauges(client, currentYear, results)
	}

	allYearlyLeagues := make([]*templates.YearlyLeagues, numberOfYears)
	for i := 0; i < numberOfYears; i++ {
		result := <-results
		if result.Leagues == nil {
			return nil, fmt.Errorf("Error occurred while obtaining user leagues for year %s",
				result.Year)
		}
		allYearlyLeagues[i] = result
	}
	sort.Sort(templates.AllYearlyLeagues(allYearlyLeagues))
	return allYearlyLeagues, nil
}

func getUserLeauges(client userLeaguesClient, year int, results chan *templates.YearlyLeagues) {
	yearStr := fmt.Sprintf("%d", year)
	glog.V(2).Infof("getting user leagues -- year=%d", year)
	leagues, err := client.GetUserLeagues(yearStr)
	if err != nil {
		glog.Warningf("unable to get leagues for year '%d': %s", year, err)
		leagues = nil
	}
	glog.V(2).Infof("got user leagues -- year=%d, leagues=%v", year, leagues)
	yearlyLeagues := &templates.YearlyLeagues{
		Year:    yearStr,
		Leagues: leagues,
	}
	results <- yearlyLeagues
}

//
// PowerRankingsClient
//

// YahooClient implements rankings.PowerRankingsClient
type YahooClient struct {
	Client yahooGoffClient
}

// Methods needed to circumvent Yahoo's handling of playoff teams on bye weeks
type yahooGoffClient interface {
	GetAllTeamStats(leagueKey string, week int) ([]goff.Team, error)
	GetLeagueStandings(leagueKey string) (*goff.League, error)
	GetTeamRoster(teamKey string, week int) ([]goff.Player, error)
	GetPlayersStats(leagueKey string, week int, players []goff.Player) ([]goff.Player, error)
	GetMatchupsForWeekRange(leagueKey string, startWeek, endWeek int) (map[int][]goff.Matchup, error)
}

// GetLeagueStandings gets a league containing the current standings.
func (y *YahooClient) GetLeagueStandings(leagueKey string) (*goff.League, error) {
	return y.Client.GetLeagueStandings(leagueKey)
}

// GetMatchupsForWeekRange returns a list of matchups for each week in the
// requested range.
func (y *YahooClient) GetMatchupsForWeekRange(leagueKey string, startWeek, endWeek int) (map[int][]goff.Matchup, error) {
	return y.Client.GetMatchupsForWeekRange(leagueKey, startWeek, endWeek)
}

// GetAllTeamStats gets teams stats for a given week from the Yahoo fantasy
// football API
func (y *YahooClient) GetAllTeamStats(leagueKey string, week int, projection bool) ([]goff.Team, error) {
	teams, err := y.Client.GetAllTeamStats(leagueKey, week)
	if err != nil {
		return nil, err
	}

	// Yahoo returns 0 points for teams that aren't participating in a matchup in the
	// playoffs. To handle this, get the stats for each starting player and calculate
	// the total score manually.
	numScoresToCalculate := 0
	calculateErrors := make(chan error)
	for index, team := range teams {
		score := team.TeamPoints.Total
		if score == 0.0 && !projection {
			glog.V(2).Infof("yahoo returned team with zero points -- "+
				"league=%s, week=%d, team=%s",
				leagueKey,
				week,
				team.Name)
			numScoresToCalculate++
			go calculateTeamScore(y, leagueKey, week, &teams[index], calculateErrors)
		}
	}

	for i := 0; i < numScoresToCalculate; i++ {
		err := <-calculateErrors
		if err != nil {
			return nil, err
		}
	}

	return teams, nil
}

func calculateTeamScore(y *YahooClient, leagueKey string, week int, team *goff.Team, errors chan error) {
	allPlayers, err := y.Client.GetTeamRoster(team.TeamKey, week)
	if err == nil {
		score := 0.0
		var players []goff.Player
		for _, player := range allPlayers {
			if player.SelectedPosition.Position != "BN" {
				players = append(players, player)
			}
		}
		players, err = y.Client.GetPlayersStats(leagueKey, week, players)
		if err == nil {
			for _, player := range players {
				score += player.PlayerPoints.Total
			}
		} else {
			glog.Warningf("error getting players stats -- league=%s, team=%s, "+
				"week=%d, error=%s",
				leagueKey,
				team.Name,
				week,
				err)
		}
		team.TeamPoints.Total = score
		glog.V(2).Infof("manually calculated team score -- "+
			"league=%s, week=%d, team=%s, score=%f",
			leagueKey,
			week,
			team.Name,
			score)
	} else {
		glog.Warningf("error getting team roster -- league=%s, team=%s, "+
			"week=%d, error=%s",
			leagueKey,
			team.Name,
			week,
			err)
	}
	errors <- err
}

func chooseSchemeFromRequest(
	req *http.Request,
	schemes []rankings.Scheme) rankings.Scheme {

	// Always use URL parameter if given
	values := req.URL.Query()
	schemeID := values.Get("scheme")
	if schemeID != "" {
		for _, scheme := range schemes {
			if schemeID == scheme.ID() {
				return scheme
			}
		}
	}

	// Fallback to user preference (saved in cookie)
	powerPreferenceCookie, err := req.Cookie("PowerPreference")
	if err == nil {
		for _, scheme := range schemes {
			if powerPreferenceCookie.Value == scheme.ID() {
				return scheme
			}
		}
	}

	// Default to the first scheme
	return schemes[0]
}
