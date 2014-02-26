package templates

import (
	"io"
	"strings"
	"testing"

	"github.com/Forestmb/goff"
	"github.com/Forestmb/power-league/rankings"
)

func TestWriteLeaguesTemplate(t *testing.T) {
	content := &LeaguesPageContent{
		AllYears:   mockAllLeagues(),
		LoggedIn:   true,
		SiteConfig: mockSiteConfig(),
	}

	templates := NewTemplates()
	err := templates.WriteLeaguesTemplate(mockWriter(), content)
	if err != nil {
		t.Fatalf("Writing league list template failed with err='%s'", err.Error())
	}
}

func TestWriteLeaguesTemplateError(t *testing.T) {
	content := &LeaguesPageContent{
		AllYears:   mockAllLeagues(),
		LoggedIn:   true,
		SiteConfig: mockSiteConfig(),
	}

	templates := NewTemplatesFromDir("dir-does-not-exist/")
	err := templates.WriteLeaguesTemplate(mockWriter(), content)
	if err == nil {
		t.Fatalf("Writing leagues template did not fail with non-existent dir")
	}
}

func TestWriteRankingsTemplate(t *testing.T) {
	content := &RankingsPageContent{
		Weeks:           12,
		League:          &(mockLeagues()[0]),
		LeaguePowerData: mockLeaguePowerData(),
		SiteConfig:      mockSiteConfig(),
	}

	templates := NewTemplates()
	err := templates.WriteRankingsTemplate(mockWriter(), content)
	if err != nil {
		t.Fatalf("Writing rankingstemplate failed with err='%s'", err.Error())
	}
}

func TestWriteRankingsTemplateNilLeaguePowerData(t *testing.T) {
	content := &RankingsPageContent{
		Weeks:           12,
		League:          &(mockLeagues()[0]),
		LeaguePowerData: nil,
		SiteConfig:      mockSiteConfig(),
	}

	templates := NewTemplates()
	err := templates.WriteRankingsTemplate(mockWriter(), content)
	if err != nil {
		t.Fatalf("Writing rankings template failed with err='%s'", err.Error())
	}
}

func TestWriteRankingsTemplateError(t *testing.T) {
	content := &RankingsPageContent{
		Weeks:           12,
		League:          &(mockLeagues()[0]),
		LeaguePowerData: mockLeaguePowerData(),
		SiteConfig:      mockSiteConfig(),
	}

	templates := NewTemplatesFromDir("dir-does-not-exist/")
	err := templates.WriteRankingsTemplate(mockWriter(), content)
	if err == nil {
		t.Fatalf("Writing rankings template did not fail with non-existent dir")
	}
}

func TestWriteAboutTemplate(t *testing.T) {
	content := &AboutPageContent{
		LoggedIn:   true,
		SiteConfig: mockSiteConfig(),
	}

	templates := NewTemplates()
	err := templates.WriteAboutTemplate(mockWriter(), content)
	if err != nil {
		t.Fatalf("Writing about template failed with err='%s'", err.Error())
	}
}

func TestWriteAboutTemplateError(t *testing.T) {
	content := &AboutPageContent{
		LoggedIn:   true,
		SiteConfig: mockSiteConfig(),
	}

	templates := NewTemplatesFromDir("dir-does-not-exist/")
	err := templates.WriteAboutTemplate(mockWriter(), content)
	if err == nil {
		t.Fatalf("Writing about template did not fail with non-existent dir")
	}
}

func TestTemplateGetTitleFromYear(t *testing.T) {
	year := "2011"
	out := templateGetTitleFromYear(year)
	if !strings.Contains(out, year) {
		t.Fatalf("Assertion error. Title does not contain year."+
			"\n\tTitle: %s\n\tYear: %s",
			out,
			year)
	}
}

func TestTemplateGetTitleFromYearDefault(t *testing.T) {
	out := templateGetTitleFromYear("nfl")
	if len(out) < 1 {
		t.Fatal("Assertion error. Title was not generated for default year code")
	}
}

func TestTemplateGetTitleFromYearEmpty(t *testing.T) {
	out := templateGetTitleFromYear("")
	if len(out) < 1 {
		t.Fatal("Assertion error. Title was not generated for empty year")
	}
}

func TestTemplateGetPowerScore(t *testing.T) {
	teamScores := []*rankings.TeamScoreData{
		&rankings.TeamScoreData{PowerScore: 3.0},
		&rankings.TeamScoreData{PowerScore: 4.0},
		&rankings.TeamScoreData{PowerScore: 5.0},
		&rankings.TeamScoreData{PowerScore: 6.0},
	}
	expectedScoreStr := "18"
	scoreStr := templateGetPowerScore(4, teamScores)
	if scoreStr != expectedScoreStr {
		t.Fatalf("Assertion error. Incorrect power score calculated for given week."+
			"\n\tExpected: %s\n\tActual: %s",
			expectedScoreStr,
			scoreStr)
	}
}

type MockResponseWriter struct {
	content string
}

func (m *MockResponseWriter) Write(b []byte) (n int, err error) {
	m.content = string(b)
	return 0, nil
}

func mockWriter() io.Writer {
	return &MockResponseWriter{}
}

func mockLeaguePowerData() *rankings.LeaguePowerData {
	return &rankings.LeaguePowerData{}
}

func mockAllLeagues() []*YearlyLeagues {
	return []*YearlyLeagues{
		&YearlyLeagues{
			Year:    "2012",
			Leagues: mockLeagues(),
		},
		&YearlyLeagues{
			Year:    "2011",
			Leagues: mockLeagues(),
		},
	}
}

func mockLeagues() []goff.League {
	return []goff.League{
		goff.League{
			LeagueKey: "123",
			LeagueID:  1,
			Name:      "TestLeague",
			Teams:     mockTeams(),
		},
	}
}

func mockSiteConfig() *SiteConfig {
	return &SiteConfig{
		BaseContext: "/power-rankings",
		NavLinks: []SiteLink{
			SiteLink{
				Link: "/example",
				Name: "example",
			},
		},
		StaticContext: "/static",
	}
}

func mockTeams() goff.Teams {
	return goff.Teams{
		List: []goff.Team{
			goff.Team{
				TeamKey: "123",
				TeamID:  123,
				Name:    "TestTeam01",
			},
			goff.Team{
				TeamKey: "321",
				TeamID:  321,
				Name:    "TestTeam02",
			},
		},
	}
}
