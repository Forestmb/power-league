// Package rankings calculates the alternative rankings of a league based on
// data obtained from a fantasy sports provider.
package rankings

import (
	"fmt"
	"sort"

	"github.com/Forestmb/power-league/Godeps/_workspace/src/github.com/Forestmb/goff"
	"github.com/Forestmb/power-league/Godeps/_workspace/src/github.com/golang/glog"
)

//
// Configuration variables
//

// MinimizeAPICalls to the fantasy sports provider, whenever possible
var MinimizeAPICalls = true

//
// Data structures
//

// LeaguePowerData for an entire league over the course of multiple weeks
type LeaguePowerData struct {
	RankingScheme     Scheme
	OverallRankings   []*TeamPowerData
	ProjectedRankings []*TeamPowerData
	ByTeam            map[string]*TeamPowerData
	ByWeek            []*WeeklyRanking
}

// WeeklyRanking of teams based on their performance for a specific week
type WeeklyRanking struct {
	Scheme    Scheme
	Week      int
	Rankings  []*TeamScoreData
	Projected bool
}

// TeamScoreData describes the score information for a single team
type TeamScoreData struct {
	Team         *goff.Team
	FantasyScore float64
	Rank         int
	PowerScore   float64
	Record       *goff.Record
	Projected    bool
}

// TeamRankingData describes how a team was ranked in comparison to their
// peer through a particular week of the season.
type TeamRankingData struct {
	Week      int
	Rank      int
	Score     float64
	Record    *goff.Record
	Projected bool
}

// TeamPowerData describes how a team performed in the power rankings
type TeamPowerData struct {
	Team                   *goff.Team
	Rank                   int
	ProjectedRank          int
	TotalScore             float64
	ProjectedTotalScore    float64
	OverallRecord          *goff.Record
	ProjectedOverallRecord *goff.Record
	AllRankings            []*TeamRankingData
	AllScores              []*TeamScoreData
	HasProjections         bool
}

// schemeRankingWorkbook keeps track of information needed to calculate
// power rankings for a specific rankings scheme
type schemeRankingWorkbook struct {
	Scheme             Scheme
	PowerDataByTeamKey map[string]*TeamPowerData
	WeeklyRankings     []*WeeklyRanking
}

// RecordSortedTeamRankingsData allows information about how teams ranked for a
// week to be sorted by their records
type RecordSortedTeamRankingsData []*TeamRankingData

func (s RecordSortedTeamRankingsData) Len() int {
	return len(s)
}

func (s RecordSortedTeamRankingsData) Less(i, j int) bool {
	if s[i].Record.Wins == s[j].Record.Wins {
		return s[i].Record.Ties > s[j].Record.Ties
	}
	return s[i].Record.Wins > s[j].Record.Wins
}

func (s RecordSortedTeamRankingsData) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// PowerSortedTeamRankingsData allows information about how teams ranked for a
// week to be sorted by their power scores
type PowerSortedTeamRankingsData []*TeamRankingData

func (s PowerSortedTeamRankingsData) Len() int {
	return len(s)
}

func (s PowerSortedTeamRankingsData) Less(i, j int) bool {
	return s[i].Score > s[j].Score
}

func (s PowerSortedTeamRankingsData) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// RecordRankings ranks teams based on their performance over multiple weeks
type RecordRankings []*TeamPowerData

func (p RecordRankings) Len() int {
	return len(p)
}

func (p RecordRankings) Less(i, j int) bool {
	if p[i].OverallRecord.Wins == p[j].OverallRecord.Wins {
		if p[i].OverallRecord.Ties == p[j].OverallRecord.Ties {
			return p[i].Team.Name < p[j].Team.Name
		}
		return p[i].OverallRecord.Ties > p[j].OverallRecord.Ties
	}
	return p[i].OverallRecord.Wins > p[j].OverallRecord.Wins
}

func (p RecordRankings) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// PowerRankings ranks teams based on their performance over multiple weeks
type PowerRankings []*TeamPowerData

func (p PowerRankings) Len() int {
	return len(p)
}

func (p PowerRankings) Less(i, j int) bool {
	if p[i].TotalScore == p[j].TotalScore {
		if p[i].ProjectedTotalScore == p[j].ProjectedTotalScore {
			return p[i].Team.Name < p[j].Team.Name
		}
		return p[i].ProjectedTotalScore > p[j].ProjectedTotalScore
	}
	return p[i].TotalScore > p[j].TotalScore
}

func (p PowerRankings) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// ProjectedRecordRankings ranks teams based on their combined actual and
// projected performance over multiple weeks
type ProjectedRecordRankings []*TeamPowerData

func (p ProjectedRecordRankings) Len() int {
	return len(p)
}

func (p ProjectedRecordRankings) Less(i, j int) bool {
	if p[i].ProjectedOverallRecord.Wins == p[j].ProjectedOverallRecord.Wins {
		return p[i].ProjectedOverallRecord.Ties > p[j].ProjectedOverallRecord.Ties
	}
	return p[i].ProjectedOverallRecord.Wins > p[j].ProjectedOverallRecord.Wins
}

func (p ProjectedRecordRankings) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// ProjectedPowerRankings ranks teams based on their combined actual and
// projected performance over multiple weeks
type ProjectedPowerRankings []*TeamPowerData

func (p ProjectedPowerRankings) Len() int {
	return len(p)
}

func (p ProjectedPowerRankings) Less(i, j int) bool {
	return p[i].ProjectedTotalScore > p[j].ProjectedTotalScore
}

func (p ProjectedPowerRankings) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

//
// Interface
//

// PowerRankingsClient calculates the power rankings from a fantasy football
// statistics provider.
type PowerRankingsClient interface {
	GetAllTeamStats(leagueKey string, week int, projection bool) ([]goff.Team, error)
	GetLeagueStandings(leagueKey string) (*goff.League, error)
	GetMatchupsForWeekRange(leagueKey string, startWeek, endWeek int) (map[int][]goff.Matchup, error)
}

//
// Functions
//

// GetWeeklyRanking returns a league's rankings for a specific week
func GetWeeklyRanking(
	client PowerRankingsClient,
	leagueKey string,
	week int,
	results chan *WeeklyRanking,
	errors chan error,
	projection bool,
	schemes []Scheme) {

	teams, err := client.GetAllTeamStats(leagueKey, week, projection)
	if err != nil {
		glog.Warningf("couldn't retrieve team points for week %d: %s", week, err.Error())
		errors <- err
		return
	}

	for _, scheme := range schemes {
		teamsForSchemes := make([]goff.Team, len(teams))
		copy(teamsForSchemes, teams)
		go scheme.CalculateWeeklyRankings(week, teamsForSchemes, projection, results)
	}
}

// GetWeeklyRankingFromMatchups rankings teams for a given week using
// matchups
func GetWeeklyRankingFromMatchups(
	week int,
	matchups []goff.Matchup,
	results chan *WeeklyRanking,
	schemes []Scheme) {

	var teams []goff.Team
	for _, matchup := range matchups {
		teams = append(teams, matchup.Teams[0])
		teams = append(teams, matchup.Teams[1])
	}

	for _, scheme := range schemes {
		teamsForSchemes := make([]goff.Team, len(teams))
		copy(teamsForSchemes, teams)
		go scheme.CalculateWeeklyRankings(week, teamsForSchemes, false, results)
	}
}

// GetPowerData returns a league's power rankings up to the given week and
// projections until the end of the season.
func GetPowerData(client PowerRankingsClient, l *goff.League, currentWeek int) ([]*LeaguePowerData, error) {
	endWeek := l.EndWeek
	leagueKey := l.LeagueKey

	league, err := client.GetLeagueStandings(leagueKey)
	if err != nil {
		return nil, err
	}

	resultsChan := make(chan *WeeklyRanking)
	errorsChan := make(chan error)
	schemes := GetSchemes()

	// Getting matchups for a span of multiple weeks results in less API calls
	// to the fantasy sports provider, and thus a lower risk of being
	// throttled. However it should be noted that this particular request
	// actually takes longer than requesting data for each week individually.
	matchupsEnd := 0
	if MinimizeAPICalls {
		matchupsEnd = lastWeekMatchupsAreAvailable(currentWeek, league)
		glog.V(2).Infof("getting weekly matchups -- weekStart=%d, weekEnd=%d",
			1,
			matchupsEnd)
		allMatchups, err := client.GetMatchupsForWeekRange(leagueKey, 1, matchupsEnd)
		if err != nil {
			return nil, err
		}

		for week, matchups := range allMatchups {
			go GetWeeklyRankingFromMatchups(week, matchups, resultsChan, schemes)
		}
	}

	for week := matchupsEnd + 1; week <= currentWeek; week++ {
		go GetWeeklyRanking(client, leagueKey, week, resultsChan, errorsChan, false, schemes)
	}

	// Get projections
	for week := currentWeek + 1; week <= endWeek; week++ {
		go GetWeeklyRanking(client, leagueKey, week, resultsChan, errorsChan, true, schemes)
	}

	teamDataByTeamKey := make(map[string]goff.Team)
	for _, team := range league.Standings {
		teamDataByTeamKey[team.TeamKey] = team
	}

	schemeWorkbooks := make(map[string]schemeRankingWorkbook)
	for i, s := range schemes {
		schemeWorkbooks[s.ID()] = schemeRankingWorkbook{
			Scheme:             schemes[i],
			PowerDataByTeamKey: make(map[string]*TeamPowerData),
			WeeklyRankings:     make([]*WeeklyRanking, endWeek),
		}
	}

	// Calculate power score for each team
	for week := 1; week <= endWeek*len(schemes); week++ {
		select {
		case err := <-errorsChan:
			glog.Warningf("error calculating weekly ranking -- "+
				"league=%s, error=%s",
				leagueKey,
				err)
			return nil, err
		case weeklyRanking := <-resultsChan:
			scheme := weeklyRanking.Scheme
			powerDataByTeamKey :=
				schemeWorkbooks[scheme.ID()].PowerDataByTeamKey
			weeklyRankings :=
				schemeWorkbooks[scheme.ID()].WeeklyRankings

			glog.V(2).Infof(
				"received weekly ranking -- week=%d, scheme=%s",
				weeklyRanking.Week,
				scheme.DisplayName())

			weekIndex := weeklyRanking.Week - 1
			weeklyRankings[weekIndex] = weeklyRanking
			for _, teamScoreData := range weeklyRanking.Rankings {
				powerData, ok := powerDataByTeamKey[teamScoreData.Team.TeamKey]
				if !ok {
					teamData, ok := teamDataByTeamKey[teamScoreData.Team.TeamKey]
					if ok {
						teamScoreData.Team.TeamStandings = teamData.TeamStandings
					}
					powerData = &TeamPowerData{
						AllScores:              make([]*TeamScoreData, endWeek),
						Team:                   teamScoreData.Team,
						TotalScore:             0.0,
						ProjectedTotalScore:    0.0,
						OverallRecord:          &goff.Record{},
						ProjectedOverallRecord: &goff.Record{},
						AllRankings:            make([]*TeamRankingData, endWeek),
						HasProjections:         weeklyRanking.Projected,
					}
					powerDataByTeamKey[teamScoreData.Team.TeamKey] = powerData
				}
				powerData.AllScores[weekIndex] = teamScoreData
				if scheme.Type() == Types.RECORD {
					addRecord(powerData.ProjectedOverallRecord, teamScoreData.Record)
				} else {
					powerData.ProjectedTotalScore += teamScoreData.PowerScore
				}
				if weeklyRanking.Projected {
					powerData.HasProjections = true
				} else {
					if scheme.Type() == Types.RECORD {
						addRecord(powerData.OverallRecord, teamScoreData.Record)
					} else {
						powerData.TotalScore += teamScoreData.PowerScore
					}
				}
			}
		}
	}

	var leaguePowerData []*LeaguePowerData
	for _, scheme := range schemes {
		workbook := schemeWorkbooks[scheme.ID()]
		powerDataByTeamKey := workbook.PowerDataByTeamKey
		weeklyRankings := workbook.WeeklyRankings
		createWeeklyTeamRankings(scheme, powerDataByTeamKey, endWeek)
		glog.V(2).Infof("ranking teams -- league=%s, numTeams=%d",
			leagueKey,
			len(powerDataByTeamKey))
		sortedPowerData := make([]*TeamPowerData, len(powerDataByTeamKey))
		index := 0
		for _, powerData := range powerDataByTeamKey {
			sortedPowerData[index] = powerData
			index++
		}

		if scheme.Type() == Types.RECORD {
			sort.Sort(RecordRankings(sortedPowerData))
			for i, powerData := range sortedPowerData {
				// Handle ties
				if i > 0 &&
					powerData.OverallRecord.Wins ==
						sortedPowerData[i-1].OverallRecord.Wins &&
					powerData.OverallRecord.Ties ==
						sortedPowerData[i-1].OverallRecord.Ties {
					powerData.Rank = sortedPowerData[i-1].Rank
				} else {
					powerData.Rank = i + 1
				}
				glog.V(4).Infof("overall rankings -- league=%s, rank=%d, team=%s, "+
					"total=%s",
					leagueKey,
					powerData.Rank,
					powerData.Team.Name,
					recordString(powerData.OverallRecord))
			}
		} else {
			sort.Sort(PowerRankings(sortedPowerData))
			for i, powerData := range sortedPowerData {
				// Handle ties
				if i > 0 &&
					powerData.TotalScore ==
						sortedPowerData[i-1].TotalScore {
					powerData.Rank = sortedPowerData[i-1].Rank
				} else {
					powerData.Rank = i + 1
				}
				glog.V(4).Infof("overall rankings -- league=%s, rank=%d, team=%s, "+
					"total=%f",
					leagueKey,
					powerData.Rank,
					powerData.Team.Name,
					powerData.TotalScore)
			}
		}

		glog.V(2).Infof("projecting rankings -- league=%s", leagueKey)
		sortedProjectionData := make([]*TeamPowerData, len(powerDataByTeamKey))
		index = 0
		for _, powerData := range powerDataByTeamKey {
			sortedProjectionData[index] = powerData
			index++
		}
		if scheme.Type() == Types.RECORD {
			sort.Sort(ProjectedRecordRankings(sortedProjectionData))
			for i, powerData := range sortedProjectionData {
				// Handle ties
				if i > 0 &&
					powerData.ProjectedOverallRecord.Wins ==
						sortedProjectionData[i-1].ProjectedOverallRecord.Wins &&
					powerData.ProjectedOverallRecord.Ties ==
						sortedProjectionData[i-1].ProjectedOverallRecord.Ties {
					powerData.ProjectedRank = sortedProjectionData[i-1].ProjectedRank
				} else {
					powerData.ProjectedRank = i + 1
				}
				glog.V(4).Infof("projected rankings -- league=%s, rank=%d, team=%s, "+
					"projected=%s",
					leagueKey,
					powerData.ProjectedRank,
					powerData.Team.Name,
					recordString(powerData.ProjectedOverallRecord))
			}
		} else {
			sort.Sort(ProjectedPowerRankings(sortedProjectionData))
			for i, powerData := range sortedProjectionData {
				// Handle ties
				if i > 0 &&
					powerData.ProjectedTotalScore ==
						sortedProjectionData[i-1].ProjectedTotalScore {
					powerData.ProjectedRank = sortedProjectionData[i-1].ProjectedRank
				} else {
					powerData.ProjectedRank = i + 1
				}
				glog.V(4).Infof("projected rankings -- league=%s, rank=%d, team=%s, "+
					"projected=%f",
					leagueKey,
					powerData.ProjectedRank,
					powerData.Team.Name,
					powerData.ProjectedTotalScore)
			}
		}

		leaguePowerData = append(leaguePowerData, &LeaguePowerData{
			RankingScheme:     scheme,
			OverallRankings:   sortedPowerData,
			ProjectedRankings: sortedProjectionData,
			ByTeam:            powerDataByTeamKey,
			ByWeek:            weeklyRankings,
		})
	}

	return leaguePowerData, nil
}

// Find the last week in a season that matchups can be used to gather data for
// the power rankings. (Matchups cannot be used for playoff games or
// projections)
func lastWeekMatchupsAreAvailable(currentWeek int, l *goff.League) int {
	if l.Settings.UsesPlayoff && currentWeek >= l.Settings.PlayoffStartWeek {
		return l.Settings.PlayoffStartWeek - 1
	}
	return currentWeek
}

// Update each team in the power data map to have their overall ranking
// in the power league for each week in a season
func createWeeklyTeamRankings(
	scheme Scheme,
	powerDataByTeamKey map[string]*TeamPowerData,
	endWeek int) {

	// Calculate the overall rankings for each week
	weeklyTeamRankings := make([][]*TeamRankingData, endWeek)
	for i := 0; i < endWeek; i++ {
		weeklyTeamRankings[i] = make([]*TeamRankingData, len(powerDataByTeamKey))
		j := 0
		for _, powerData := range powerDataByTeamKey {
			weeklyScore := powerData.AllScores[i]
			powerData.AllRankings[i] = &TeamRankingData{
				Week:  i + 1,
				Score: weeklyScore.PowerScore,
				Record: &goff.Record{
					Wins:   weeklyScore.Record.Wins,
					Losses: weeklyScore.Record.Losses,
					Ties:   weeklyScore.Record.Ties,
				},
				Projected: weeklyScore.Projected,
			}
			if i > 0 {
				previousRanking := powerData.AllRankings[i-1]
				powerData.AllRankings[i].Score +=
					previousRanking.Score
				addRecord(
					powerData.AllRankings[i].Record,
					previousRanking.Record)
			}
			weeklyTeamRankings[i][j] = powerData.AllRankings[i]
			j++
		}

		// Sort by the cumulative power scores and assign ranks for this week
		if scheme.Type() == Types.RECORD {
			sort.Sort(RecordSortedTeamRankingsData(weeklyTeamRankings[i]))
			for j, rankingsData := range weeklyTeamRankings[i] {
				if j > 0 &&
					rankingsData.Record.Wins ==
						weeklyTeamRankings[i][j-1].Record.Wins &&
					rankingsData.Record.Ties ==
						weeklyTeamRankings[i][j-1].Record.Ties {
					rankingsData.Rank = weeklyTeamRankings[i][j-1].Rank
				} else {
					rankingsData.Rank = j + 1
				}
			}
		} else {
			sort.Sort(PowerSortedTeamRankingsData(weeklyTeamRankings[i]))
			for j, rankingsData := range weeklyTeamRankings[i] {
				if j > 0 &&
					rankingsData.Score ==
						weeklyTeamRankings[i][j-1].Score {
					rankingsData.Rank = weeklyTeamRankings[i][j-1].Rank
				} else {
					rankingsData.Rank = j + 1
				}
			}
		}
	}
}

// addRecord adds to the first record the wins/losses/ties of the second
func addRecord(r, toAdd *goff.Record) {
	r.Wins += toAdd.Wins
	r.Losses += toAdd.Losses
	r.Ties += toAdd.Ties
}

// recordString returns the string representation of a record
func recordString(r *goff.Record) string {
	return fmt.Sprintf("%d-%d-%d", r.Wins, r.Losses, r.Ties)
}
