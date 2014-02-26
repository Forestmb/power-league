package site

import (
	"errors"
	"net/http"
	"testing"

	"github.com/Forestmb/goff"
)

func TestNewSite(t *testing.T) {
	baseContext := "base-context"
	staticContext := "static-context"
	site := NewSite(baseContext, staticContext, "templates/", &MockSessionManager{})

	if site == nil {
		t.Fatal("no site created")
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

	_, err := client.GetAllTeamStats("123", 12)
	if err != nil {
		t.Fatalf("error returned getting team stats: %s", err)
	}
}

func TestYahooClientAllTeamStatsError(t *testing.T) {
	goffClient := &MockGoffClient{
		AllTeamStatsError: errors.New("error"),
	}
	client := &YahooClient{Client: goffClient}

	_, err := client.GetAllTeamStats("123", 12)
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

	_, err := client.GetAllTeamStats("123", 12)
	if err == nil {
		t.Fatalf("no error getting all team stats when team had 0 points " +
			"and getting the roster failed: %s")
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

	_, err := client.GetAllTeamStats("123", 12)
	if err == nil {
		t.Fatalf("no error getting all team stats when team had 0 points " +
			"and getting player stats failed: %s")
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

	teams, err := client.GetAllTeamStats("123", 12)
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

type MockGoffClient struct {
	AllTeamStats      []goff.Team
	AllTeamStatsError error
	TeamRoster        []goff.Player
	TeamRosterError   error
	PlayersStats      []goff.Player
	PlayersStatsError error
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

type MockSessionManager struct {
	LoginURL      string
	LoginError    error
	LogoutError   error
	AuthError     error
	IsLoggedInRet bool
	Client        *goff.Client
	ClientError   error
}

func (m *MockSessionManager) Login(w http.ResponseWriter, r *http.Request, redirectURL string) (loginURL string, err error) {
	return m.LoginURL, m.LoginError
}

func (m *MockSessionManager) Authenticate(w http.ResponseWriter, r *http.Request) error {
	return m.AuthError
}

func (m *MockSessionManager) Logout(w http.ResponseWriter, r *http.Request) error {
	return m.LogoutError
}

func (m *MockSessionManager) IsLoggedIn(r *http.Request) bool {
	return m.IsLoggedInRet
}

func (m *MockSessionManager) GetClient(w http.ResponseWriter, r *http.Request) (*goff.Client, error) {
	return m.Client, m.ClientError
}
