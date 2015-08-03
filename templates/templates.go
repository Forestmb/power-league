// Package templates centralizes access to template functions that render the
// power-league content as HTML.
package templates

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Forestmb/power-league/Godeps/_workspace/src/github.com/Forestmb/goff"
	"github.com/Forestmb/power-league/Godeps/_workspace/src/github.com/golang/glog"
	"github.com/Forestmb/power-league/rankings"
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
type Templates interface {
	WriteAboutTemplate(w io.Writer, content *AboutPageContent) error
	WriteErrorTemplate(w io.Writer, content *ErrorPageContent) error
	WriteLeaguesTemplate(w io.Writer, content *LeaguesPageContent) error
	WriteRankingsTemplate(w io.Writer, content *RankingsPageContent) error
}

// defaultTemplates provides programmtic access to power rankings templates
// created from files within a single base directory
type defaultTemplates struct {
	// Directory containing the HTML template files
	baseDir string
}

// NewTemplates creates new templates using the `DefaultBaseDir`
func NewTemplates() Templates {
	return NewTemplatesFromDir(DefaultBaseDir)
}

// NewTemplatesFromDir creates new templates using the given directory as the base
func NewTemplatesFromDir(dir string) Templates {
	glog.V(2).Infof("creating new templates -- dir=%s", dir)
	return &defaultTemplates{
		baseDir: dir,
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
	SchemeIDToShow  string
	Schemes         []rankings.Scheme
	LeaguePowerData []*rankings.LeaguePowerData
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

// PreviousRank defines the rank a team had for a previous week and the
// offset between that rank and the current rank.
type PreviousRank struct {
	Rank   int
	Offset int
}

// WriteRankingsTemplate writes the raknings template to the given writer
func (t *defaultTemplates) WriteRankingsTemplate(w io.Writer, content *RankingsPageContent) error {
	funcMap := template.FuncMap{
		"getPowerScore":          templateGetPowerScore,
		"getRecord":              templateGetRecord,
		"getRankings":            templateGetRankings,
		"getActualRank":          templateGetActualRank,
		"getPlacingTeams":        templateGetPlacingTeams,
		"getPlaceFromRank":       templateGetPlaceFromRank,
		"getTeamPosition":        templateGetTeamPosition,
		"getRankOffset":          templateGetRankOffset,
		"getRankForPreviousWeek": templateGetRankForPreviousWeek,
		"getAbsoluteValue":       templateGetAbsoluteValue,
		"getCSVContent":          templateGetCSVContent,
		"getExportFilename":      templateGetExportFilename,
	}
	template, err := template.New(rankingsTemplate).Funcs(funcMap).ParseFiles(
		t.baseDir+baseTemplate,
		t.baseDir+rankingsTemplate)
	if err != nil {
		return err
	}
	return writeTemplateSafe(w, template, content)
}

// WriteAboutTemplate writes the about page template to the given writer
func (t *defaultTemplates) WriteAboutTemplate(w io.Writer, content *AboutPageContent) error {
	template, err := template.New(aboutTemplate).ParseFiles(
		t.baseDir+baseTemplate,
		t.baseDir+aboutTemplate)
	if err != nil {
		return err
	}
	return writeTemplateSafe(w, template, content)
}

// WriteLeaguesTemplate writes the leagues page template to the given writer
func (t *defaultTemplates) WriteLeaguesTemplate(w io.Writer, content *LeaguesPageContent) error {

	funcMap := template.FuncMap{
		"getTitleFromYear": templateGetTitleFromYear,
	}
	template, err := template.New(leaguesTemplate).Funcs(funcMap).ParseFiles(
		t.baseDir+baseTemplate,
		t.baseDir+leaguesTemplate)
	if err != nil {
		return err
	}
	return writeTemplateSafe(w, template, content)
}

// WriteErrorTemplate writes the error page template to the given writer
//
// If the io.Writer is an http.ResponseWriter, this function will write an
// error code of http.StatusInternalServerError (500) before writing the
// actual content.
func (t *defaultTemplates) WriteErrorTemplate(w io.Writer, content *ErrorPageContent) error {
	template, err := template.New(errorTemplate).ParseFiles(
		t.baseDir+baseTemplate,
		t.baseDir+errorTemplate)
	if err != nil {
		return err
	}
	return writeTemplateSafeWithCode(
		w, template, content, http.StatusInternalServerError)
}

// Write to a writer if and only if the template can be executed successfully
// with the given content as input.
func writeTemplateSafe(w io.Writer, t *template.Template, content interface{}) error {
	return writeTemplateSafeWithCode(w, t, content, http.StatusOK)
}

// Write to a writer if and only if the template can be executed successfully
// with the given content as input.
//
// If the given writer is an http.ResponseWriter, the httpResponseCode will be
// written to the header before any content is written. If the writer is not
// an http.ResponseWriter this field is ignored.
func writeTemplateSafeWithCode(
	w io.Writer,
	t *template.Template,
	content interface{},
	httpResponseCode int) error {

	buffer := &bytes.Buffer{}
	err := t.Execute(buffer, content)
	if err == nil {
		if rw, ok := w.(http.ResponseWriter); ok {
			rw.WriteHeader(httpResponseCode)
		}
		_, err = io.Copy(w, buffer)
	}
	return err
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
	return teamPowerData.AllRankings[week-1].Record
}

func templateGetPowerScore(week int, teamPowerData *rankings.TeamPowerData) string {
	return fmt.Sprintf("%.2f", teamPowerData.AllRankings[week-1].Score)
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

func templateGetRankOffset(rank, otherRank int) string {
	offset := rank - otherRank
	return fmt.Sprintf("%+d", offset)
}

func templateGetAbsoluteValue(value int) int {
	if value < 0 {
		return value * -1
	}
	return value
}

func templateGetRankForPreviousWeek(teamPowerData *rankings.TeamPowerData, currentWeek int) *PreviousRank {
	if currentWeek > 1 {
		previousWeekIndex := (currentWeek - 1) - 1
		currentWeekRank := teamPowerData.AllRankings[currentWeek-1].Rank
		previousWeekRank := teamPowerData.AllRankings[previousWeekIndex].Rank
		return &PreviousRank{
			Rank:   previousWeekRank,
			Offset: previousWeekRank - currentWeekRank,
		}
	}
	return nil
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
	scheme := leagueData.RankingScheme
	var buffer bytes.Buffer
	separator := ","
	buffer.WriteString("Rank,")
	buffer.WriteString("Projected Rank,")
	buffer.WriteString("Team,")
	buffer.WriteString("Manager,")
	if scheme.Type() == rankings.RECORD {
		buffer.WriteString(scheme.DisplayName())
		buffer.WriteString(" Record,")
		buffer.WriteString("Projected ")
		buffer.WriteString(scheme.DisplayName())
		buffer.WriteString(" Record,")
	} else {
		buffer.WriteString(scheme.DisplayName())
		buffer.WriteString(",")
		buffer.WriteString("Projected ")
		buffer.WriteString(scheme.DisplayName())
		buffer.WriteString(",")
	}
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
		buffer.WriteString("Overall Rank")
		buffer.WriteString(weekStr)
		buffer.WriteString("Overall ")
		buffer.WriteString(scheme.DisplayName())
		if scheme.Type() == rankings.RECORD {
			buffer.WriteString(" Record")
		}
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
		if scheme.Type() == rankings.RECORD {
			writeRecordToBuffer(&buffer, teamData.OverallRecord)
			buffer.WriteString(separator)
			writeRecordToBuffer(&buffer, teamData.ProjectedOverallRecord)
		} else {
			buffer.WriteString(
				strconv.FormatFloat(teamData.TotalScore, 'f', 2, 64))
			buffer.WriteString(separator)
			buffer.WriteString(
				strconv.FormatFloat(teamData.ProjectedTotalScore, 'f', 2, 64))
		}
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
			buffer.WriteString(strconv.FormatFloat(weeklyScore.FantasyScore, 'f', 2, 64))
			buffer.WriteString(separator)
			buffer.WriteString(strconv.Itoa(weeklyScore.Rank))
			buffer.WriteString(separator)
			buffer.WriteString(strconv.Itoa(teamData.AllRankings[index].Rank))
			buffer.WriteString(separator)
			if scheme.Type() == rankings.RECORD {
				writeRecordToBuffer(
					&buffer, teamData.AllRankings[index].Record)
			} else {
				buffer.WriteString(
					strconv.FormatFloat(
						teamData.AllRankings[index].Score, 'f', 2, 64))
			}
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
