package templates

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http/httptest"
	"sort"
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
		SchemeToShow:    mockScoreScheme{},
		Schemes:         []rankings.Scheme{mockScoreScheme{}, mockRecordScheme{}},
		League:          &(mockLeagues()[0]),
		LeaguePowerData: []*rankings.LeaguePowerData{mockLeaguePowerData()},
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
		SchemeToShow:    mockScoreScheme{},
		Schemes:         []rankings.Scheme{mockScoreScheme{}, mockRecordScheme{}},
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
		SchemeToShow:    mockScoreScheme{},
		Schemes:         []rankings.Scheme{mockScoreScheme{}, mockRecordScheme{}},
		LeaguePowerData: []*rankings.LeaguePowerData{mockLeaguePowerData()},
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

func TestWriteTemplateSafeWithCodeResponseWriter(t *testing.T) {
	message := "This is the message"
	content := &ErrorPageContent{
		Message:    message,
		LoggedIn:   true,
		SiteConfig: mockSiteConfig(),
	}

	templates := NewTemplates().(*defaultTemplates)
	template, err := template.New(errorTemplate).ParseFiles(
		templates.baseDir+baseTemplate,
		templates.baseDir+errorTemplate)

	recorder := httptest.NewRecorder()
	code := 123
	err = writeTemplateSafeWithCode(recorder, template, content, code)
	if err != nil {
		t.Fatalf("Writing error template safely failed with err='%s'",
			err.Error())
	}

	if recorder.Code != code {
		t.Fatalf("Did not write expected error code:\n\tExpected: %d"+
			"\n\tActual: %d",
			code,
			recorder.Code)
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
	teamPowerData := &rankings.TeamPowerData{
		AllRankings: []*rankings.TeamRankingData{
			&rankings.TeamRankingData{
				Score: 3.0,
			},
			&rankings.TeamRankingData{
				Score: 7.0,
			},
			&rankings.TeamRankingData{
				Score: 17.0,
			},
		},
	}
	for i, rankingData := range teamPowerData.AllRankings {
		scoreStr := templateGetPowerScore(i+1, teamPowerData)
		expectedScoreStr := fmt.Sprintf("%.2f", rankingData.Score)
		if scoreStr != expectedScoreStr {
			t.Fatalf("Assertion error. Incorrect power score calculated for given week."+
				"\n\tExpected: %s\n\tActual: %s",
				expectedScoreStr,
				scoreStr)
		}
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

	actual = templateGetPlaceFromRank(4, "other", "another", expected)
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

func TestTemplateGetRankOffset(t *testing.T) {
	offset := templateGetRankOffset(2, 3)
	expected := "-1"
	if offset != expected {
		t.Fatalf("Unexpected offset returned:\n\tExpected: %s\n\tActual: %s",
			expected,
			offset)
	}

	offset = templateGetRankOffset(2, 2)
	expected = "+0"
	if offset != expected {
		t.Fatalf("Unexpected offset returned:\n\tExpected: %s\n\tActual: %s",
			expected,
			offset)
	}

	offset = templateGetRankOffset(4, 2)
	expected = "+2"
	if offset != expected {
		t.Fatalf("Unexpected offset returned:\n\tExpected: %s\n\tActual: %s",
			expected,
			offset)
	}
}

func TestTemplateGetRankForPreviousWeek(t *testing.T) {
	teamData := mockLeaguePowerData().OverallRankings[0]
	previousWeek := templateGetRankForPreviousWeek(teamData, 1)
	if previousWeek != nil {
		t.Fatalf("Rank for previous returned for Week 1 when no rank should"+
			"have been returned: %+v",
			previousWeek)
	}

	previousWeek = templateGetRankForPreviousWeek(teamData, 2)
	expectedRank := 1
	if previousWeek.Rank != expectedRank {
		t.Fatalf("Unexpected rank returned for previous week:\n\t"+
			"Expected: %d\n\tActual: %d",
			expectedRank,
			previousWeek.Rank)
	}
	expectedOffset := -2
	if previousWeek.Rank != expectedRank {
		t.Fatalf("Unexpected rank offset returned for previous week:\n\t"+
			"Expected: %d\n\tActual: %d",
			expectedOffset,
			previousWeek.Offset)
	}
}

func TestTemplateGetRankForScheme(t *testing.T) {
	scheme := mockScoreScheme{}
	powerData := mockLeaguePowerData()
	powerData.RankingScheme = scheme

	otherScheme := mockRecordScheme{}
	otherPowerData := mockLeaguePowerData()
	otherPowerData.RankingScheme = otherScheme

	actual, err := templateGetRankForScheme(
		scheme.ID(),
		[]*rankings.LeaguePowerData{otherPowerData, powerData})

	expected := -1
	for _, teamData := range powerData.OverallRankings {
		if teamData.Team.IsOwnedByCurrentLogin {
			expected = teamData.Rank
		}
	}

	if expected == -1 {
		t.Fatalf("Invalid test input, no team owned by current login: %+v",
			powerData)
	} else if err != nil {
		t.Fatalf("Unexpected error getting rank for scheme: %s", err)
	} else if expected != actual {
		t.Fatalf("Unexpected rank returned for scheme:\n\t"+
			"Expected: %d\n\tActual: %d",
			expected,
			actual)
	}

	_, err = templateGetRankForScheme(
		scheme.ID(),
		[]*rankings.LeaguePowerData{otherPowerData})
	if err == nil {
		t.Fatal("No error returned when getting rank for a scheme that " +
			"does not exist")
	}
}

func TestTemplateGetAbsoluteValue(t *testing.T) {
	for input, expected := range map[int]int{
		0:   0,
		10:  10,
		-10: 10,
		2:   2,
		-2:  2,
	} {
		actual := templateGetAbsoluteValue(input)
		if actual != expected {
			t.Fatalf("Unexpected result when calculating absolute value\n\t"+
				"Expected: %d\n\tActual: %d",
				expected,
				actual)
		}
	}
}

func TestTemplateGetExportFilename(t *testing.T) {
	league := mockLeagues()[0]
	filename := templateGetExportFilename(&league)
	if len(filename) == 0 {
		t.Fatal("Returned empty export filename for league")
	}
}

func TestTemplateGetCSVContentRecordScheme(t *testing.T) {
	leagueData := mockLeaguePowerData()
	csvBase64 := templateGetCSVContent(leagueData)
	csv, err := base64.StdEncoding.DecodeString(csvBase64)
	if err != nil {
		t.Fatalf("Error decoding content from Base64: %s", err)
	}
	csvScanner := bufio.NewScanner(bytes.NewReader(csv))

	csvScanner.Scan()
	titleLine := csvScanner.Text()
	expectedTitle :=
		"Rank," +
			"Projected Rank," +
			"Team," +
			"Manager," +
			"Mock Scheme Record," +
			"Projected Mock Scheme Record," +
			"League Rank," +
			"League Rank Offset," +
			"League Record"

	for _, weeklyRanking := range leagueData.ByWeek {
		var weekStr string
		if weeklyRanking.Projected {
			weekStr = fmt.Sprintf(",[Projected] Week %d ", weeklyRanking.Week)
		} else {
			weekStr = fmt.Sprintf(",Week %d ", weeklyRanking.Week)
		}
		expectedTitle += weekStr + "Fantasy Points"
		expectedTitle += weekStr + "Weekly Rank"
		expectedTitle += weekStr + "Overall Rank"
		expectedTitle += weekStr + "Overall Mock Scheme Record"
	}

	if titleLine != expectedTitle {
		t.Fatalf("Unexpected title line in CSV content:\n\tExpected: %s"+
			"\n\tActual: %s",
			expectedTitle,
			titleLine)
	}

	for _, teamData := range leagueData.OverallRankings {
		csvScanner.Scan()
		teamContent := csvScanner.Text()
		expectedContent :=
			fmt.Sprintf(
				"%d,"+
					"%d,"+
					"%s,"+
					"%s,"+
					"%d-%d-%d,"+
					"%d-%d-%d,"+
					"%d,"+
					"%s,"+
					"%d-%d-%d",
				teamData.Rank,
				teamData.ProjectedRank,
				teamData.Team.Name,
				teamData.Team.Managers[0].Nickname,
				teamData.OverallRecord.Wins,
				teamData.OverallRecord.Losses,
				teamData.OverallRecord.Ties,
				teamData.ProjectedOverallRecord.Wins,
				teamData.ProjectedOverallRecord.Losses,
				teamData.ProjectedOverallRecord.Ties,
				teamData.Team.TeamStandings.Rank,
				templateGetRankOffset(teamData.Rank, teamData.Team.TeamStandings.Rank),
				teamData.Team.TeamStandings.Record.Wins,
				teamData.Team.TeamStandings.Record.Losses,
				teamData.Team.TeamStandings.Record.Ties)
		for i, teamScoreData := range teamData.AllScores {
			ranking := teamData.AllRankings[i]
			expectedContent +=
				fmt.Sprintf(
					",%.2f"+
						",%d"+
						",%d"+
						",%d-%d-%d",
					teamScoreData.FantasyScore,
					teamScoreData.Rank,
					ranking.Rank,
					ranking.Record.Wins,
					ranking.Record.Losses,
					ranking.Record.Ties)
		}

		if teamContent != expectedContent {
			t.Fatalf("Unexpected content returned for team:\n\tExpected: %s\n\tActual: %s\n",
				expectedContent,
				teamContent)
		}
	}
}

func TestTemplateGetCSVContentScoreScheme(t *testing.T) {
	leagueData := mockLeaguePowerData()
	leagueData.RankingScheme = mockScoreScheme{}
	csvBase64 := templateGetCSVContent(leagueData)
	csv, err := base64.StdEncoding.DecodeString(csvBase64)
	if err != nil {
		t.Fatalf("Error decoding content from Base64: %s", err)
	}
	csvScanner := bufio.NewScanner(bytes.NewReader(csv))

	csvScanner.Scan()
	titleLine := csvScanner.Text()
	expectedTitle :=
		"Rank," +
			"Projected Rank," +
			"Team," +
			"Manager," +
			"Mock Scheme Points," +
			"Projected Mock Scheme Points," +
			"League Rank," +
			"League Rank Offset," +
			"League Record"

	for _, weeklyRanking := range leagueData.ByWeek {
		var weekStr string
		if weeklyRanking.Projected {
			weekStr = fmt.Sprintf(",[Projected] Week %d ", weeklyRanking.Week)
		} else {
			weekStr = fmt.Sprintf(",Week %d ", weeklyRanking.Week)
		}
		expectedTitle += weekStr + "Fantasy Points"
		expectedTitle += weekStr + "Weekly Rank"
		expectedTitle += weekStr + "Overall Rank"
		expectedTitle += weekStr + "Overall Mock Scheme Points"
	}

	if titleLine != expectedTitle {
		t.Fatalf("Unexpected title line in CSV content:\n\tExpected: %s"+
			"\n\tActual: %s",
			expectedTitle,
			titleLine)
	}

	for _, teamData := range leagueData.OverallRankings {
		csvScanner.Scan()
		teamContent := csvScanner.Text()
		expectedContent :=
			fmt.Sprintf(
				"%d,"+
					"%d,"+
					"%s,"+
					"%s,"+
					"%.2f,"+
					"%.2f,"+
					"%d,"+
					"%s,"+
					"%d-%d-%d",
				teamData.Rank,
				teamData.ProjectedRank,
				teamData.Team.Name,
				teamData.Team.Managers[0].Nickname,
				teamData.TotalScore,
				teamData.ProjectedTotalScore,
				teamData.Team.TeamStandings.Rank,
				templateGetRankOffset(teamData.Rank, teamData.Team.TeamStandings.Rank),
				teamData.Team.TeamStandings.Record.Wins,
				teamData.Team.TeamStandings.Record.Losses,
				teamData.Team.TeamStandings.Record.Ties)
		for i, teamScoreData := range teamData.AllScores {
			ranking := teamData.AllRankings[i]
			expectedContent +=
				fmt.Sprintf(
					",%.2f"+
						",%d"+
						",%d"+
						",%.2f",
					teamScoreData.FantasyScore,
					teamScoreData.Rank,
					ranking.Rank,
					ranking.Score)
		}

		if teamContent != expectedContent {
			t.Fatalf("Unexpected content returned for team:\n\tExpected: %s\n\tActual: %s\n",
				expectedContent,
				teamContent)
		}
	}
}

func TestTemplateGetRecord(t *testing.T) {
	teamPowerData := &rankings.TeamPowerData{
		AllRankings: []*rankings.TeamRankingData{
			&rankings.TeamRankingData{
				Record: &goff.Record{
					Wins:   1,
					Losses: 2,
					Ties:   3,
				},
			},
			&rankings.TeamRankingData{
				Record: &goff.Record{
					Wins:   3,
					Losses: 4,
					Ties:   5,
				},
			},
			&rankings.TeamRankingData{
				Record: &goff.Record{
					Wins:   6,
					Losses: 10,
					Ties:   14,
				},
			},
		},
	}

	record := templateGetRecord(1, teamPowerData)
	if record.Wins != 1 ||
		record.Losses != 2 ||
		record.Ties != 3 {
		t.Fatalf("Incorrect cumulative record returned:\n\tExpected: "+
			"%d - %d - %d\n\tActual: %d - %d - %d",
			1, 2, 3, record.Wins, record.Losses, record.Ties)
	}

	record = templateGetRecord(3, teamPowerData)
	if record.Wins != 6 ||
		record.Losses != 10 ||
		record.Ties != 14 {
		t.Fatalf("Incorrect cumulative record returned:\n\tExpected: "+
			"%d - %d - %d\n\tActual: %d - %d - %d",
			6, 10, 14, record.Wins, record.Losses, record.Ties)
	}
}

func TestSortAllYearlyLeagues(t *testing.T) {
	leagues := []*YearlyLeagues{
		&YearlyLeagues{
			Year:    "2013",
			Leagues: mockLeagues(),
		},
		&YearlyLeagues{
			Year:    "2014",
			Leagues: mockLeagues(),
		},
		&YearlyLeagues{
			Year:    "2001",
			Leagues: mockLeagues(),
		},
		&YearlyLeagues{
			Year:    "2005",
			Leagues: mockLeagues(),
		},
	}
	sort.Sort(AllYearlyLeagues(leagues))
	if leagues[0].Year != "2014" ||
		leagues[1].Year != "2013" ||
		leagues[2].Year != "2005" ||
		leagues[3].Year != "2001" {
		t.Fatalf("Unexpected order after sorting yearly leagues:\n\t"+
			"Expected: 2014, 2013, 2005, 2001\n\tActual: %s, %s, %s, %s\n",
			leagues[0].Year,
			leagues[1].Year,
			leagues[2].Year,
			leagues[3].Year)
	}
}

type MockResponseWriter struct {
	content string
}

func (m *MockResponseWriter) Write(b []byte) (n int, err error) {
	m.content += string(b)
	return len(b), nil
}

func mockWriter() *MockResponseWriter {
	return &MockResponseWriter{content: ""}
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

func mockLeaguePowerData() *rankings.LeaguePowerData {
	ownerTeam := mockTeam()
	ownerTeam.IsOwnedByCurrentLogin = true
	return &rankings.LeaguePowerData{
		RankingScheme: mockRecordScheme{},
		OverallRankings: rankings.PowerRankings{
			&rankings.TeamPowerData{
				AllScores: []*rankings.TeamScoreData{
					&rankings.TeamScoreData{
						Team:         mockTeam(),
						FantasyScore: 12.0,
						Rank:         1,
						PowerScore:   36.0,
					},
					&rankings.TeamScoreData{
						Team:         mockTeam(),
						FantasyScore: 22.0,
						Rank:         3,
						PowerScore:   36.0,
					},
					&rankings.TeamScoreData{
						Team:         mockTeam(),
						FantasyScore: 32.0,
						Rank:         2,
						PowerScore:   36.0,
					},
				},
				Team:                mockTeam(),
				TotalScore:          12.0,
				ProjectedTotalScore: 13.0,
				OverallRecord: &goff.Record{
					Wins:   1,
					Losses: 2,
					Ties:   3,
				},
				ProjectedOverallRecord: &goff.Record{
					Wins:   1,
					Losses: 2,
					Ties:   3,
				},
				Rank:          1,
				ProjectedRank: 3,
				AllRankings: []*rankings.TeamRankingData{
					&rankings.TeamRankingData{
						Week:  1,
						Rank:  1,
						Score: 12.0,
						Record: &goff.Record{
							Wins:   3,
							Losses: 3,
							Ties:   0,
						},
						Projected: false,
					},
					&rankings.TeamRankingData{
						Week:  2,
						Rank:  1,
						Score: 22.0,
						Record: &goff.Record{
							Wins:   3,
							Losses: 7,
							Ties:   1,
						},
						Projected: false,
					},
					&rankings.TeamRankingData{
						Week:  3,
						Rank:  3,
						Score: 32.0,
						Record: &goff.Record{
							Wins:   7,
							Losses: 9,
							Ties:   2,
						},
						Projected: true,
					},
				},
			},
			&rankings.TeamPowerData{
				AllScores: []*rankings.TeamScoreData{
					&rankings.TeamScoreData{
						Team:         ownerTeam,
						FantasyScore: 12.0,
						Rank:         1,
						PowerScore:   36.0,
					},
					&rankings.TeamScoreData{
						Team:         ownerTeam,
						FantasyScore: 22.0,
						Rank:         3,
						PowerScore:   36.0,
					},
					&rankings.TeamScoreData{
						Team:         ownerTeam,
						FantasyScore: 32.0,
						Rank:         2,
						PowerScore:   36.0,
					},
				},
				Team:                ownerTeam,
				TotalScore:          12.0,
				ProjectedTotalScore: 14.0,
				OverallRecord: &goff.Record{
					Wins:   1,
					Losses: 2,
					Ties:   3,
				},
				ProjectedOverallRecord: &goff.Record{
					Wins:   1,
					Losses: 2,
					Ties:   3,
				},
				Rank:          2,
				ProjectedRank: 2,
				AllRankings: []*rankings.TeamRankingData{
					&rankings.TeamRankingData{
						Week:  1,
						Rank:  2,
						Score: 12.0,
						Record: &goff.Record{
							Wins:   3,
							Losses: 3,
							Ties:   0,
						},
						Projected: false,
					},
					&rankings.TeamRankingData{
						Week:  2,
						Rank:  3,
						Score: 22.0,
						Record: &goff.Record{
							Wins:   3,
							Losses: 7,
							Ties:   1,
						},
						Projected: false,
					},
					&rankings.TeamRankingData{
						Week:  3,
						Rank:  2,
						Score: 32.0,
						Record: &goff.Record{
							Wins:   7,
							Losses: 9,
							Ties:   2,
						},
						Projected: true,
					},
				},
			},
			&rankings.TeamPowerData{
				AllScores: []*rankings.TeamScoreData{
					&rankings.TeamScoreData{
						Team:         mockTeam(),
						FantasyScore: 12.0,
						Rank:         1,
						PowerScore:   36.0,
					},
					&rankings.TeamScoreData{
						Team:         mockTeam(),
						FantasyScore: 22.0,
						Rank:         3,
						PowerScore:   36.0,
					},
					&rankings.TeamScoreData{
						Team:         mockTeam(),
						FantasyScore: 32.0,
						Rank:         2,
						PowerScore:   36.0,
					},
				},
				Team:                mockTeam(),
				TotalScore:          12.0,
				ProjectedTotalScore: 15.0,
				OverallRecord: &goff.Record{
					Wins:   1,
					Losses: 2,
					Ties:   3,
				},
				ProjectedOverallRecord: &goff.Record{
					Wins:   1,
					Losses: 2,
					Ties:   3,
				},
				Rank:          3,
				ProjectedRank: 1,
				AllRankings: []*rankings.TeamRankingData{
					&rankings.TeamRankingData{
						Week:  1,
						Rank:  3,
						Score: 12.0,
						Record: &goff.Record{
							Wins:   3,
							Losses: 3,
							Ties:   0,
						},
						Projected: false,
					},
					&rankings.TeamRankingData{
						Week:  2,
						Rank:  1,
						Score: 22.0,
						Record: &goff.Record{
							Wins:   3,
							Losses: 7,
							Ties:   1,
						},
						Projected: false,
					},
					&rankings.TeamRankingData{
						Week:  3,
						Rank:  1,
						Score: 32.0,
						Record: &goff.Record{
							Wins:   7,
							Losses: 9,
							Ties:   2,
						},
						Projected: true,
					},
				},
			},
		},
		ProjectedRankings: rankings.ProjectedPowerRankings{
			&rankings.TeamPowerData{
				AllScores: []*rankings.TeamScoreData{
					&rankings.TeamScoreData{
						Team:         mockTeam(),
						FantasyScore: 12.0,
						Rank:         1,
						PowerScore:   36.0,
					},
				},
				Team:                mockTeam(),
				TotalScore:          12.0,
				ProjectedTotalScore: 15.0,
				OverallRecord: &goff.Record{
					Wins:   1,
					Losses: 2,
					Ties:   3,
				},
				ProjectedOverallRecord: &goff.Record{
					Wins:   1,
					Losses: 2,
					Ties:   3,
				},
				Rank:          1,
				ProjectedRank: 1,
				AllRankings: []*rankings.TeamRankingData{
					&rankings.TeamRankingData{
						Week:  1,
						Rank:  1,
						Score: 12.0,
						Record: &goff.Record{
							Wins:   3,
							Losses: 3,
							Ties:   0,
						},
						Projected: false,
					},
					&rankings.TeamRankingData{
						Week:  2,
						Rank:  2,
						Score: 22.0,
						Record: &goff.Record{
							Wins:   3,
							Losses: 7,
							Ties:   1,
						},
						Projected: false,
					},
					&rankings.TeamRankingData{
						Week:  3,
						Rank:  3,
						Score: 32.0,
						Record: &goff.Record{
							Wins:   7,
							Losses: 9,
							Ties:   2,
						},
						Projected: true,
					},
				},
			},
			&rankings.TeamPowerData{
				AllScores: []*rankings.TeamScoreData{
					&rankings.TeamScoreData{
						Team:         mockTeam(),
						FantasyScore: 12.0,
						Rank:         1,
						PowerScore:   36.0,
					},
				},
				Team:                mockTeam(),
				TotalScore:          12.0,
				ProjectedTotalScore: 14.0,
				OverallRecord: &goff.Record{
					Wins:   1,
					Losses: 2,
					Ties:   3,
				},
				ProjectedOverallRecord: &goff.Record{
					Wins:   1,
					Losses: 2,
					Ties:   3,
				},
				Rank:          2,
				ProjectedRank: 2,
				AllRankings: []*rankings.TeamRankingData{
					&rankings.TeamRankingData{
						Week:  1,
						Rank:  2,
						Score: 12.0,
						Record: &goff.Record{
							Wins:   3,
							Losses: 3,
							Ties:   0,
						},
						Projected: false,
					},
					&rankings.TeamRankingData{
						Week:  2,
						Rank:  3,
						Score: 22.0,
						Record: &goff.Record{
							Wins:   3,
							Losses: 7,
							Ties:   1,
						},
						Projected: false,
					},
					&rankings.TeamRankingData{
						Week:  3,
						Rank:  2,
						Score: 32.0,
						Record: &goff.Record{
							Wins:   7,
							Losses: 9,
							Ties:   2,
						},
						Projected: true,
					},
				},
			},
			&rankings.TeamPowerData{
				AllScores: []*rankings.TeamScoreData{
					&rankings.TeamScoreData{
						Team:         mockTeam(),
						FantasyScore: 12.0,
						Rank:         1,
						PowerScore:   36.0,
					},
				},
				Team:                mockTeam(),
				TotalScore:          12.0,
				ProjectedTotalScore: 13.0,
				OverallRecord: &goff.Record{
					Wins:   1,
					Losses: 2,
					Ties:   3,
				},
				ProjectedOverallRecord: &goff.Record{
					Wins:   1,
					Losses: 2,
					Ties:   3,
				},
				Rank:          3,
				ProjectedRank: 3,
				AllRankings: []*rankings.TeamRankingData{
					&rankings.TeamRankingData{
						Week:  1,
						Rank:  3,
						Score: 12.0,
						Record: &goff.Record{
							Wins:   3,
							Losses: 3,
							Ties:   0,
						},
						Projected: false,
					},
					&rankings.TeamRankingData{
						Week:  2,
						Rank:  1,
						Score: 22.0,
						Record: &goff.Record{
							Wins:   3,
							Losses: 7,
							Ties:   1,
						},
						Projected: false,
					},
					&rankings.TeamRankingData{
						Week:  3,
						Rank:  1,
						Score: 32.0,
						Record: &goff.Record{
							Wins:   7,
							Losses: 9,
							Ties:   2,
						},
						Projected: true,
					},
				},
			},
		},
		ByWeek: []*rankings.WeeklyRanking{
			&rankings.WeeklyRanking{
				Week:      1,
				Projected: false,
				Rankings: []*rankings.TeamScoreData{
					&rankings.TeamScoreData{
						Team:         mockTeam(),
						FantasyScore: 12.0,
						Rank:         1,
						PowerScore:   36.0,
					},
					&rankings.TeamScoreData{
						Team:         mockTeam(),
						FantasyScore: 34.0,
						Rank:         2,
						PowerScore:   23.0,
					},
					&rankings.TeamScoreData{
						Team:         mockTeam(),
						FantasyScore: 38.0,
						Rank:         3,
						PowerScore:   21.0,
					},
				},
			},
			&rankings.WeeklyRanking{
				Week:      2,
				Projected: false,
				Rankings: []*rankings.TeamScoreData{
					&rankings.TeamScoreData{
						Team:         mockTeam(),
						FantasyScore: 12.0,
						Rank:         1,
						PowerScore:   36.0,
					},
					&rankings.TeamScoreData{
						Team:         mockTeam(),
						FantasyScore: 34.0,
						Rank:         2,
						PowerScore:   23.0,
					},
					&rankings.TeamScoreData{
						Team:         mockTeam(),
						FantasyScore: 38.0,
						Rank:         3,
						PowerScore:   21.0,
					},
				},
			},
			&rankings.WeeklyRanking{
				Week:      3,
				Projected: true,
				Rankings: []*rankings.TeamScoreData{
					&rankings.TeamScoreData{
						Team:         mockTeam(),
						FantasyScore: 12.0,
						Rank:         1,
						PowerScore:   36.0,
					},
					&rankings.TeamScoreData{
						Team:         mockTeam(),
						FantasyScore: 34.0,
						Rank:         2,
						PowerScore:   23.0,
					},
					&rankings.TeamScoreData{
						Team:         mockTeam(),
						FantasyScore: 38.0,
						Rank:         3,
						PowerScore:   21.0,
					},
				},
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
		Managers: []goff.Manager{
			goff.Manager{
				Nickname: "Manager",
			},
		},
		Name: "TestTeam02",
		TeamStandings: goff.TeamStandings{
			Rank: 1,
			Record: goff.Record{
				Wins:   4,
				Losses: 2,
				Ties:   2,
			},
		},
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
			Managers: []goff.Manager{
				goff.Manager{
					Nickname: "Manager 1",
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
			Managers: []goff.Manager{
				goff.Manager{
					Nickname: "Manager 2",
				},
			},
			Name: "TestTeam02",
		},
	}
}
