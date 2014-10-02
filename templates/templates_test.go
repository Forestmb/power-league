package templates

import (
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

func TestWriteErrorTemplate(t *testing.T) {
	message := "This is the message"
	content := &ErrorPageContent{
		Message:    message,
		LoggedIn:   true,
		SiteConfig: mockSiteConfig(),
	}

	templates := NewTemplates()
	writer := mockWriter()
	err := templates.WriteErrorTemplate(writer, content)
	if err != nil {
		t.Fatalf("Writing error template failed with err='%s'", err.Error())
	}

	if !strings.Contains(writer.content, message) {
		t.Fatalf("Error page content did not contain expected message.\n\t"+
			"expected: %s\n\tactual: %s\n",
			message,
			writer.content)
	}
}

func TestWriteErrorTemplateError(t *testing.T) {
	content := &ErrorPageContent{
		Message:    "message",
		LoggedIn:   true,
		SiteConfig: mockSiteConfig(),
	}

	templates := NewTemplatesFromDir("dir-does-not-exist/")
	err := templates.WriteErrorTemplate(mockWriter(), content)
	if err == nil {
		t.Fatalf("Writing error template did not fail with non-existent dir")
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

func TestTemplateGetActualRankWithProjections(t *testing.T) {
	expected := 2
	teamPowerData := &rankings.TeamPowerData{
		Rank:           1,
		ProjectedRank:  expected,
		HasProjections: true,
	}
	rank := templateGetActualRank(teamPowerData)

	if rank != expected {
		t.Fatalf("Assertion error. Incorrect actual rank returned for team"+
			"power data.\n\tExpected: %d\n\tActual: %d",
			expected,
			rank)
	}
}

func TestTemplateGetActualRankWithoutProjections(t *testing.T) {
	expected := 2
	teamPowerData := &rankings.TeamPowerData{
		Rank:           expected,
		ProjectedRank:  1,
		HasProjections: false,
	}
	rank := templateGetActualRank(teamPowerData)

	if rank != expected {
		t.Fatalf("Assertion error. Incorrect actual rank returned for team"+
			"power data.\n\tExpected: %d\n\tActual: %d",
			expected,
			rank)
	}
}

func TestTemplateGetPlacingTeamsProjected(t *testing.T) {
	first := &rankings.TeamPowerData{
		Rank:           0,
		ProjectedRank:  1,
		HasProjections: true,
	}
	third := &rankings.TeamPowerData{
		Rank:           0,
		ProjectedRank:  0,
		HasProjections: false,
	}
	allTeamPowerData := []*rankings.TeamPowerData{
		first,
		first,
		third,
		third,
		third,
		&rankings.TeamPowerData{
			Rank:           4,
			ProjectedRank:  0,
			HasProjections: false,
		},
	}

	placed := templateGetPlacingTeams(allTeamPowerData)
	if len(placed) != 5 {
		t.Fatalf("Unexpected number of placed teams returned:\n\t"+
			"Expected: %d\n\tActual: %d",
			5,
			len(placed))
	}

	if placed[0] != first {
		t.Fatalf("Unexpected team returned.\n\tExpected: %+v\n\tActual: %+v",
			first,
			placed[0])
	}
	if placed[1] != first {
		t.Fatalf("Unexpected team returned.\n\tExpected: %+v\n\tActual: %+v",
			first,
			placed[1])
	}
	if placed[2] != third {
		t.Fatalf("Unexpected team returned.\n\tExpected: %+v\n\tActual: %+v",
			third,
			placed[2])
	}
	if placed[3] != third {
		t.Fatalf("Unexpected team returned.\n\tExpected: %+v\n\tActual: %+v",
			third,
			placed[3])
	}
	if placed[4] != third {
		t.Fatalf("Unexpected team returned.\n\tExpected: %+v\n\tActual: %+v",
			third,
			placed[4])
	}
}

func TestTemplateGetPlacingTeamsNotProjected(t *testing.T) {
	first := &rankings.TeamPowerData{
		Rank:           1,
		ProjectedRank:  0,
		HasProjections: false,
	}
	third := &rankings.TeamPowerData{
		Rank:           3,
		ProjectedRank:  0,
		HasProjections: false,
	}
	allTeamPowerData := []*rankings.TeamPowerData{
		first,
		first,
		third,
		third,
		third,
		&rankings.TeamPowerData{
			Rank:           4,
			ProjectedRank:  0,
			HasProjections: false,
		},
	}

	placed := templateGetPlacingTeams(allTeamPowerData)
	if len(placed) != 5 {
		t.Fatalf("Unexpected number of placed teams returned:\n\t"+
			"Expected: %d\n\tActual: %d",
			5,
			len(placed))
	}

	if placed[0] != first {
		t.Fatalf("Unexpected team returned.\n\tExpected: %+v\n\tActual: %+v",
			first,
			placed[0])
	}
	if placed[1] != first {
		t.Fatalf("Unexpected team returned.\n\tExpected: %+v\n\tActual: %+v",
			first,
			placed[1])
	}
	if placed[2] != third {
		t.Fatalf("Unexpected team returned.\n\tExpected: %+v\n\tActual: %+v",
			third,
			placed[2])
	}
	if placed[3] != third {
		t.Fatalf("Unexpected team returned.\n\tExpected: %+v\n\tActual: %+v",
			third,
			placed[3])
	}
	if placed[4] != third {
		t.Fatalf("Unexpected team returned.\n\tExpected: %+v\n\tActual: %+v",
			third,
			placed[4])
	}
}

func TestGetPlaceFromRank(t *testing.T) {
	expected := "expected"
	actual := templateGetPlaceFromRank(1, expected, "other", "another")
	if expected != actual {
		t.Fatalf("Unexpected place returned from rank.\n\tExpected: %s"+
			"\n\tActual: %s",
			expected,
			actual)
	}

	actual = templateGetPlaceFromRank(2, "other", expected, "another")
	if expected != actual {
		t.Fatalf("Unexpected place returned from rank.\n\tExpected: %s"+
			"\n\tActual: %s",
			expected,
			actual)
	}

	actual = templateGetPlaceFromRank(3, "other", "another", expected)
	if expected != actual {
		t.Fatalf("Unexpected place returned from rank.\n\tExpected: %s"+
			"\n\tActual: %s",
			expected,
			actual)
	}
}

func TestTemplateGetRankingsProjected(t *testing.T) {
	expected := []*rankings.TeamPowerData{}
	actual := templateGetRankings(
		rankings.LeaguePowerData{
			OverallRankings:   nil,
			ProjectedRankings: expected,
		},
		false)

	if actual == nil {
		t.Fatal("Projected rankings were not returned")
	}
}

func TestTemplateGetRankingsNotProjected(t *testing.T) {
	expected := []*rankings.TeamPowerData{}
	actual := templateGetRankings(
		rankings.LeaguePowerData{
			OverallRankings:   expected,
			ProjectedRankings: nil,
		},
		true)

	if actual == nil {
		t.Fatal("Actual rankings were not returned")
	}
}

func TestTemplateGetTeamPosition(t *testing.T) {
	teamID := uint64(123)
	rankings := []*rankings.TeamPowerData{
		&rankings.TeamPowerData{
			Team: &goff.Team{
				TeamID: 1,
			},
		},
		&rankings.TeamPowerData{
			Team: &goff.Team{
				TeamID: 2,
			},
		},
		&rankings.TeamPowerData{
			Team: &goff.Team{
				TeamID: teamID,
			},
		},
		&rankings.TeamPowerData{
			Team: &goff.Team{
				TeamID: 3,
			},
		},
	}

	position := templateGetTeamPosition(teamID, rankings)
	if position != 3 {
		t.Fatalf("Did not return expected position:\n\tExpected: %d\n\t"+
			"Actual: %d",
			3,
			position)
	}
}

func TestTemplateGetTeamPositionNotPresent(t *testing.T) {
	teamID := uint64(123)
	rankings := []*rankings.TeamPowerData{
		&rankings.TeamPowerData{
			Team: &goff.Team{
				TeamID: 1,
			},
		},
		&rankings.TeamPowerData{
			Team: &goff.Team{
				TeamID: 2,
			},
		},
		&rankings.TeamPowerData{
			Team: &goff.Team{
				TeamID: 3,
			},
		},
	}

	position := templateGetTeamPosition(teamID, rankings)
	if position != -1 {
		t.Fatalf("Did not return expected position:\n\tExpected: %d\n\t"+
			"Actual: %d",
			-1,
			position)
	}
}

type MockResponseWriter struct {
	content string
}

func (m *MockResponseWriter) Write(b []byte) (n int, err error) {
	m.content += string(b)
	return 0, nil
}

func mockWriter() *MockResponseWriter {
	return &MockResponseWriter{content: ""}
}

func mockLeaguePowerData() *rankings.LeaguePowerData {
	return &rankings.LeaguePowerData{
		OverallRankings: rankings.PowerRankings{
			&rankings.TeamPowerData{
				AllScores: []*rankings.TeamScoreData{
					&rankings.TeamScoreData{
						Team:       mockTeam(),
						Score:      12.0,
						Rank:       1,
						PowerScore: 36.0,
					},
				},
				Team:            mockTeam(),
				TotalPowerScore: 12.0,
				Rank:            1,
			},
			&rankings.TeamPowerData{
				AllScores: []*rankings.TeamScoreData{
					&rankings.TeamScoreData{
						Team:       mockTeam(),
						Score:      12.0,
						Rank:       1,
						PowerScore: 36.0,
					},
				},
				Team:            mockTeam(),
				TotalPowerScore: 12.0,
				Rank:            2,
			},
			&rankings.TeamPowerData{
				AllScores: []*rankings.TeamScoreData{
					&rankings.TeamScoreData{
						Team:       mockTeam(),
						Score:      12.0,
						Rank:       1,
						PowerScore: 36.0,
					},
				},
				Team:            mockTeam(),
				TotalPowerScore: 12.0,
				Rank:            3,
			},
		},
		ProjectedRankings: rankings.ProjectedPowerRankings{
			&rankings.TeamPowerData{
				AllScores: []*rankings.TeamScoreData{
					&rankings.TeamScoreData{
						Team:       mockTeam(),
						Score:      12.0,
						Rank:       1,
						PowerScore: 36.0,
					},
				},
				Team:            mockTeam(),
				TotalPowerScore: 12.0,
				Rank:            1,
			},
			&rankings.TeamPowerData{
				AllScores: []*rankings.TeamScoreData{
					&rankings.TeamScoreData{
						Team:       mockTeam(),
						Score:      12.0,
						Rank:       1,
						PowerScore: 36.0,
					},
				},
				Team:            mockTeam(),
				TotalPowerScore: 12.0,
				Rank:            2,
			},
			&rankings.TeamPowerData{
				AllScores: []*rankings.TeamScoreData{
					&rankings.TeamScoreData{
						Team:       mockTeam(),
						Score:      12.0,
						Rank:       1,
						PowerScore: 36.0,
					},
				},
				Team:            mockTeam(),
				TotalPowerScore: 12.0,
				Rank:            3,
			},
		},
	}
}

func mockTeam() *goff.Team {
	return &goff.Team{
		TeamKey: "321",
		TeamID:  321,
		TeamLogos: []goff.TeamLogo{
			goff.TeamLogo{
				Size: "medium",
				URL:  "http://example.com/image.png",
			},
		},
		Name: "TestTeam02",
	}
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
			LeagueKey:  "123",
			LeagueID:   1,
			Name:       "TestLeague",
			Teams:      mockTeams(),
			IsFinished: true,
		},
	}
}

func mockSiteConfig() *SiteConfig {
	return &SiteConfig{
		BaseContext:         "/power-rankings",
		StaticContext:       "/static",
		AnalyticsTrackingID: "",
	}
}

func mockTeams() []goff.Team {
	return []goff.Team{
		goff.Team{
			TeamKey: "123",
			TeamID:  123,
			TeamLogos: []goff.TeamLogo{
				goff.TeamLogo{
					Size: "medium",
					URL:  "http://example.com/image.png",
				},
			},
			Name: "TestTeam01",
		},
		goff.Team{
			TeamKey: "321",
			TeamID:  321,
			TeamLogos: []goff.TeamLogo{
				goff.TeamLogo{
					Size: "medium",
					URL:  "http://example.com/image.png",
				},
			},
			Name: "TestTeam02",
		},
	}
}
