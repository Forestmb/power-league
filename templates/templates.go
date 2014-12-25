// Package templates centralizes access to template functions that render the
// power-league content as HTML.
package templates

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"strconv"
	"strings"

	"github.com/Forestmb/goff"
	"github.com/Forestmb/power-league/rankings"
	"github.com/golang/glog"
)

const (
	// DefaultBaseDir is the default location of the directory containing the
	// template files.
	DefaultBaseDir = "html/"

	baseTemplate     = "base.html"
	aboutTemplate    = "about.html"
	errorTemplate    = "error.html"
	leaguesTemplate  = "leagues.html"
	rankingsTemplate = "rankings.html"
)

// Templates provides programmtic access to power rankings templates
type Templates struct {
	// Directory containing the HTML template files
	BaseDir string
}

// NewTemplates creates new templates using the `DefaultBaseDir`
func NewTemplates() *Templates {
	return NewTemplatesFromDir(DefaultBaseDir)
}

// NewTemplatesFromDir creates new templates using the given directory as the base
func NewTemplatesFromDir(dir string) *Templates {
	glog.V(2).Infof("creating new templates -- dir=%s", dir)
	return &Templates{
		BaseDir: dir,
	}
}

//
// Template Data Structures
//

// RankingsPageContent is used to show an given league's power rankings up
// to a certain week in the season.
type RankingsPageContent struct {
	Weeks           int
	League          *goff.League
	LeagueStarted   bool
	LeaguePowerData *rankings.LeaguePowerData
	LoggedIn        bool
	SiteConfig      *SiteConfig
}

// YearlyLeagues describes the leagues for a user for a given year.
type YearlyLeagues struct {
	Year    string
	Leagues []goff.League
}

// AllYearlyLeagues contains leagues for multiple years.
type AllYearlyLeagues []*YearlyLeagues

// LeaguesPageContent describes the leagues a user has participated in across
// multiple years.
type LeaguesPageContent struct {
	AllYears   AllYearlyLeagues
	LoggedIn   bool
	SiteConfig *SiteConfig
}

// AboutPageContent describes the power rankings web site.
type AboutPageContent struct {
	LoggedIn   bool
	SiteConfig *SiteConfig
}

// ErrorPageContent describes an error that has occurred in the application.
type ErrorPageContent struct {
	Message    string
	LoggedIn   bool
	SiteConfig *SiteConfig
}

// SiteConfig provides configuration info about the site that can be used on
// all pages.
type SiteConfig struct {
	BaseContext         string
	StaticContext       string
	AnalyticsTrackingID string
}

// SiteLink represents a link to a page in the site's navigation.
type SiteLink struct {
	Link string
	Name string
}

// WriteRankingsTemplate writes the raknings template to the given writer
func (t *Templates) WriteRankingsTemplate(w io.Writer, content *RankingsPageContent) error {
	funcMap := template.FuncMap{
		"getPowerScore":     templateGetPowerScore,
		"getRecord":         templateGetRecord,
		"getRankings":       templateGetRankings,
		"getActualRank":     templateGetActualRank,
		"getPlacingTeams":   templateGetPlacingTeams,
		"getPlaceFromRank":  templateGetPlaceFromRank,
		"getTeamPosition":   templateGetTeamPosition,
		"getRankOffset":     templateGetRankOffset,
		"getCSVContent":     templateGetCSVContent,
		"getExportFilename": templateGetExportFilename,
	}
	template, err := template.New(rankingsTemplate).Funcs(funcMap).ParseFiles(
		t.BaseDir+baseTemplate,
		t.BaseDir+rankingsTemplate)
	if err != nil {
		return err
	}
	return template.Execute(w, content)
}

// WriteAboutTemplate writes the about page template to the given writer
func (t *Templates) WriteAboutTemplate(w io.Writer, content *AboutPageContent) error {
	template, err := template.New(aboutTemplate).ParseFiles(
		t.BaseDir+baseTemplate,
		t.BaseDir+aboutTemplate)
	if err != nil {
		return err
	}
	return template.Execute(w, content)
}

// WriteLeaguesTemplate writes the leagues page template to the given writer
func (t *Templates) WriteLeaguesTemplate(w io.Writer, content *LeaguesPageContent) error {

	funcMap := template.FuncMap{
		"getTitleFromYear": templateGetTitleFromYear,
	}
	template, err := template.New(leaguesTemplate).Funcs(funcMap).ParseFiles(
		t.BaseDir+baseTemplate,
		t.BaseDir+leaguesTemplate)
	if err != nil {
		return err
	}

	return template.Execute(w, content)
}

// WriteErrorTemplate writes the error page template to the given writer
func (t *Templates) WriteErrorTemplate(w io.Writer, content *ErrorPageContent) error {
	template, err := template.New(errorTemplate).ParseFiles(
		t.BaseDir+baseTemplate,
		t.BaseDir+errorTemplate)
	if err != nil {
		return err
	}
	return template.Execute(w, content)
}

//
// Template functions
//

func templateGetTitleFromYear(year string) string {
	if year == "nfl" || year == "" {
		return "Current Leagues"
	}
	return year
}

func templateGetActualRank(teamData *rankings.TeamPowerData) int {
	if teamData.HasProjections {
		return teamData.ProjectedRank
	}
	return teamData.Rank
}

func templateGetPlaceFromRank(rank int, places ...string) string {
	return places[rank-1]
}

func templateGetPlacingTeams(powerData []*rankings.TeamPowerData) []*rankings.TeamPowerData {
	var placingTeams []*rankings.TeamPowerData
	for _, teamData := range powerData {
		if templateGetActualRank(teamData) <= 3 {
			placingTeams = append(placingTeams, teamData)
		}
	}
	return placingTeams
}

func templateGetRecord(week int, teamPowerData *rankings.TeamPowerData) *goff.Record {
	return teamPowerData.AllRankings[week-1].OverallRecord
}

func templateGetPowerScore(week int, teamPowerData *rankings.TeamPowerData) string {
	return fmt.Sprintf("%.0f", teamPowerData.AllRankings[week-1].PowerScore)
}

func templateGetRankings(powerData rankings.LeaguePowerData, finished bool) []*rankings.TeamPowerData {
	if finished {
		return powerData.OverallRankings
	}
	return powerData.ProjectedRankings
}

func templateGetTeamPosition(teamID uint64, rankings []*rankings.TeamPowerData) int {
	for index, scoreData := range rankings {
		if scoreData.Team.TeamID == teamID {
			return index + 1
		}
	}
	return -1
}

func templateGetRankOffset(powerRank, leagueRank int) string {
	offset := powerRank - leagueRank
	return fmt.Sprintf("%+d", offset)
}

// templateGetExportFilename creates a filename for a file containing the
// power rankings data for that league
func templateGetExportFilename(l *goff.League) string {
	return fmt.Sprintf("power-rankings-%s",
		strings.Replace(
			strings.ToLower(l.Name),
			" ",
			"-",
			-1))
}

func templateGetCSVContent(leagueData *rankings.LeaguePowerData) string {
	var buffer bytes.Buffer
	separator := ","
	buffer.WriteString("Rank,")
	buffer.WriteString("Projected Rank,")
	buffer.WriteString("Team,")
	buffer.WriteString("Manager,")
	buffer.WriteString("Power Points,")
	buffer.WriteString("Projected Power Points,")
	buffer.WriteString("All-Play Record,")
	buffer.WriteString("Projected All-Play Record,")
	buffer.WriteString("League Rank,")
	buffer.WriteString("League Rank Offset,")
	buffer.WriteString("League Record")
	for index, weeklyRanking := range leagueData.ByWeek {
		var weekStr string
		if weeklyRanking.Projected {
			weekStr = fmt.Sprintf(",[Projected] Week %d ", index+1)
		} else {
			weekStr = fmt.Sprintf(",Week %d ", index+1)
		}
		buffer.WriteString(weekStr)
		buffer.WriteString("Fantasy Points")
		buffer.WriteString(weekStr)
		buffer.WriteString("Weekly Rank")
		buffer.WriteString(weekStr)
		buffer.WriteString("Overall Power Points")
		buffer.WriteString(weekStr)
		buffer.WriteString("Overall All-Play Record")
	}
	buffer.WriteString("\n")
	for _, teamData := range leagueData.OverallRankings {
		buffer.WriteString(strconv.Itoa(teamData.Rank))
		buffer.WriteString(separator)
		buffer.WriteString(strconv.Itoa(teamData.ProjectedRank))
		buffer.WriteString(separator)
		buffer.WriteString(teamData.Team.Name)
		buffer.WriteString(separator)
		buffer.WriteString(teamData.Team.Managers[0].Nickname)
		buffer.WriteString(separator)
		buffer.WriteString(
			strconv.FormatFloat(teamData.TotalPowerScore, 'f', 2, 64))
		buffer.WriteString(separator)
		buffer.WriteString(
			strconv.FormatFloat(teamData.ProjectedPowerScore, 'f', 2, 64))
		buffer.WriteString(separator)
		writeRecordToBuffer(&buffer, teamData.OverallRecord)
		buffer.WriteString(separator)
		writeRecordToBuffer(&buffer, teamData.OverallProjectedRecord)
		buffer.WriteString(separator)
		buffer.WriteString(strconv.Itoa(teamData.Team.TeamStandings.Rank))
		buffer.WriteString(separator)
		buffer.WriteString(
			templateGetRankOffset(
				teamData.Rank,
				teamData.Team.TeamStandings.Rank))
		buffer.WriteString(separator)
		writeRecordToBuffer(&buffer, &teamData.Team.TeamStandings.Record)
		for index, weeklyScore := range teamData.AllScores {
			buffer.WriteString(separator)
			buffer.WriteString(strconv.FormatFloat(weeklyScore.Score, 'f', 2, 64))
			buffer.WriteString(separator)
			buffer.WriteString(strconv.Itoa(weeklyScore.Rank))
			buffer.WriteString(separator)
			buffer.WriteString(
				strconv.FormatFloat(
					teamData.AllRankings[index].PowerScore, 'f', 2, 64))
			buffer.WriteString(separator)
			writeRecordToBuffer(
				&buffer, teamData.AllRankings[index].OverallRecord)
		}
		buffer.WriteString("\n")
	}

	return base64.StdEncoding.EncodeToString(buffer.Bytes())
}

func writeRecordToBuffer(buffer *bytes.Buffer, r *goff.Record) {
	buffer.WriteString(strconv.Itoa(r.Wins))
	buffer.WriteString("-")
	buffer.WriteString(strconv.Itoa(r.Losses))
	buffer.WriteString("-")
	buffer.WriteString(strconv.Itoa(r.Ties))
}

//
// Sorting
//

func (y AllYearlyLeagues) Len() int {
	return len(y)
}

func (y AllYearlyLeagues) Less(i, j int) bool {
	firstYear, _ := strconv.Atoi(y[i].Year)
	secondYear, _ := strconv.Atoi(y[j].Year)
	return firstYear > secondYear
}

func (y AllYearlyLeagues) Swap(i, j int) {
	y[i], y[j] = y[j], y[i]
}
