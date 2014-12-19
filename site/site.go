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
)

// Site consists of the information needed to run a power rankings site
type Site struct {
	// ServeMux for this site
	ServeMux *http.ServeMux

	handlers       map[string]*ContextHandler
	sessionManager session.SessionManager
	config         *templates.SiteConfig
	templates      *templates.Templates
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
		glog.V(3).Infof("adding context handler -- id=%s, context=%s, navText=%s",
			id,
			fullContext)
		s.ServeMux.HandleFunc(
			fullContext,
			func(w http.ResponseWriter, r *http.Request) {
				f(s, w, r)
			})
	}

	s.handlers[id] = &ContextHandler{
		Context: fullContext,
		Func:    f,
	}
}

// NewSite creates a new site
func NewSite(
	baseContext string,
	staticFiles string,
	templatesDir string,
	trackingID string,
	s session.SessionManager) *Site {

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
		http.Error(w, "Error occurred when generating page content",
			http.StatusInternalServerError)
	}
}

func handleLogout(s *Site, w http.ResponseWriter, r *http.Request) {
	glog.V(5).Infoln("in handleLogout")
	s.sessionManager.Logout(w, r)
	leaguesURL := fmt.Sprintf("http://%s%s", r.Host, s.config.BaseContext)
	http.Redirect(w, r, leaguesURL, http.StatusTemporaryRedirect)
}

func handleLogin(s *Site, w http.ResponseWriter, r *http.Request) {
	glog.V(5).Infoln("in handleLogin")

	authContext := s.handlers["auth"].Context
	loginURL := fmt.Sprintf("http://%s%s", r.Host, authContext)
	requestURL, err := s.sessionManager.Login(w, r, loginURL)
	if err != nil {
		glog.Warningf("error generating login page: %s", err)
		http.Error(w, "Error occurred when authenticating user",
			http.StatusInternalServerError)
		return
	}
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

	redirectURL := fmt.Sprintf("http://%s%s", r.Host, redirectContext)
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
			http.Error(w, "Error occurred when retreiving league list",
				http.StatusInternalServerError)
			return
		}
		allYearlyLeagues = getAllYearlyLeagues(client)
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
		http.Error(w, "Error occurred when retreiving league list",
			http.StatusInternalServerError)
	}
}

func getAllYearlyLeagues(client *goff.Client) templates.AllYearlyLeagues {
	currentYear := time.Now().Year()
	numberOfYears := (currentYear - EarliestSupportedYear) + 1
	results := make(chan *templates.YearlyLeagues)
	for ; currentYear >= EarliestSupportedYear; currentYear-- {
		go getUserLeauges(client, currentYear, results)
	}

	allYearlyLeagues := make([]*templates.YearlyLeagues, numberOfYears)
	for i := 0; i < numberOfYears; i++ {
		allYearlyLeagues[i] = <-results
	}
	sort.Sort(templates.AllYearlyLeagues(allYearlyLeagues))
	return allYearlyLeagues
}

func getUserLeauges(client *goff.Client, year int, results chan *templates.YearlyLeagues) {
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

func handlePowerRankings(s *Site, w http.ResponseWriter, req *http.Request) {
	glog.V(5).Infoln("in handlePowerRankings")

	// Determine the current week
	currentWeek := -1
	leagueStarted := true
	loggedIn := s.sessionManager.IsLoggedIn(req)

	values := req.URL.Query()
	leagueKey := values.Get("key")

	var league *goff.League
	client, err := s.sessionManager.GetClient(w, req)
	if err == nil {
		glog.V(3).Infof("getting metadata -- league=%s", leagueKey)
		league, err = client.GetLeagueStandings(leagueKey)
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
			loggedIn = false
		}
	} else {
		glog.Warningf("unable to create client: %s", err)
		loggedIn = false
	}

	glog.V(3).Infof("calculating rankings -- week=%d", currentWeek)

	var rankingsContent *templates.RankingsPageContent
	if err == nil {
		var leaguePowerData *rankings.LeaguePowerData
		if leagueStarted {
			leaguePowerData, err = rankings.GetPowerData(
				&YahooClient{Client: client},
				league,
				currentWeek)

			if err != nil {
				glog.Warningf("error generating power rankings page: %s", err)
				err = s.templates.WriteErrorTemplate(
					w,
					&templates.ErrorPageContent{
						Message: "Error occurred while generating the rankings. " +
							"Try refreshing the page.",
						LoggedIn:   loggedIn,
						SiteConfig: s.config,
					})

				if err != nil {
					http.Error(
						w,
						"Error occurred when calculating rankings",
						http.StatusInternalServerError)
				}
				return
			}
		}

		rankingsContent = &templates.RankingsPageContent{
			Weeks:           currentWeek,
			League:          league,
			LeagueStarted:   leagueStarted,
			LeaguePowerData: leaguePowerData,
			LoggedIn:        loggedIn,
			SiteConfig:      s.config,
		}
	}

	err = s.templates.WriteRankingsTemplate(w, rankingsContent)
	if err != nil {
		glog.Warningf("error generating power rankings page: %s", err)
		http.Error(w, "Error occurred when calculating rankings",
			http.StatusInternalServerError)
	}

	if client != nil {
		glog.V(2).Infof("API Request Count: %d", client.RequestCount())
	}
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
	GetTeamRoster(teamKey string, week int) ([]goff.Player, error)
	GetPlayersStats(leagueKey string, week int, players []goff.Player) ([]goff.Player, error)
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
			glog.Warningf("yahoo returned team with zero points -- "+
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
				"week=%s, error=%s",
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
			"week=%s, error=%s",
			leagueKey,
			team.Name,
			week,
			err)
	}
	errors <- err
}
