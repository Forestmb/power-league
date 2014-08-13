// Package rankings calculates the alternative rankings of a league based on
// data obtained from a fantasy sports provider.
package rankings

import (
	"sort"

	"github.com/Forestmb/goff"
	"github.com/golang/glog"
)

//
// Data structures
//

// LeaguePowerData for an entire league over the course of multiple weeks
type LeaguePowerData struct {
	OverallRankings   []*TeamPowerData
	ProjectedRankings []*TeamPowerData
	ByTeam            map[string]*TeamPowerData
	ByWeek            []*WeeklyRanking
}

// WeeklyRanking of teams based on their performance for a specific week
type WeeklyRanking struct {
	Week      int
	Rankings  []*TeamScoreData
	Projected bool
}

// TeamScoreData describes the score information for a single team
type TeamScoreData struct {
	Team       *goff.Team
	Score      float64
	Rank       int
	PowerScore float64
}

// TeamPowerData describes how a team performed in the power rankings
type TeamPowerData struct {
	AllScores           []*TeamScoreData
	Team                *goff.Team
	TotalPowerScore     float64
	ProjectedPowerScore float64
	Rank                int
	ProjectedRank       int
}

// PowerRankings ranks teams based on their performance over multiple weeks
type PowerRankings []*TeamPowerData

func (p PowerRankings) Len() int {
	return len(p)
}

func (p PowerRankings) Less(i, j int) bool {
	return p[i].TotalPowerScore > p[j].TotalPowerScore
}

func (p PowerRankings) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// ProjectedPowerRankings ranks teams based on their combined actual and
// projected performance over multiple weeks
type ProjectedPowerRankings []*TeamPowerData

func (p ProjectedPowerRankings) Len() int {
	return len(p)
}

func (p ProjectedPowerRankings) Less(i, j int) bool {
	return p[i].ProjectedPowerScore > p[j].ProjectedPowerScore
}

func (p ProjectedPowerRankings) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// TeamRanking ranks teams based on their performance for a single week
type TeamRanking []goff.Team

func (t TeamRanking) Len() int {
	return len(t)
}

func (t TeamRanking) Less(i int, j int) bool {
	return t[i].TeamPoints.Total > t[j].TeamPoints.Total
}

func (t TeamRanking) Swap(i int, j int) {
	t[i], t[j] = t[j], t[i]
}

// TeamProjectedRanking ranks teams based on their projected performance for
// a single week
type TeamProjectedRanking []goff.Team

func (t TeamProjectedRanking) Len() int {
	return len(t)
}

func (t TeamProjectedRanking) Less(i int, j int) bool {
	return t[i].TeamProjectedPoints.Total > t[j].TeamProjectedPoints.Total
}

func (t TeamProjectedRanking) Swap(i int, j int) {
	t[i], t[j] = t[j], t[i]
}

//
// Interface
//

// PowerRankingsClient is used to calculate the power rankings from a fantasy
// football statistics provider.
type PowerRankingsClient interface {
	GetAllTeamStats(leagueKey string, week int, projection bool) ([]goff.Team, error)
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
	projection bool) {

	teams, err := client.GetAllTeamStats(leagueKey, week, projection)
	if err != nil {
		glog.Warningf("couldn't retrieve team points for week %d: %s", week, err.Error())
		errors <- err
		return
	}

	// Rank the teams by sorting then assign power scores
	if !projection {
		sort.Sort(TeamRanking(teams))
	} else {
		sort.Sort(TeamProjectedRanking(teams))
	}
	rankings := make([]*TeamScoreData, len(teams))
	numTeamns := len(teams)
	for index, team := range teams {
		var score float64
		var scoreType string
		if !projection {
			score = team.TeamPoints.Total
			scoreType = "weekly rankings"
		} else {
			score = team.TeamProjectedPoints.Total
			scoreType = "weekly projections"
		}

		var powerScore float64
		var rank int
		if index > 0 && score == rankings[index-1].Score {
			powerScore = rankings[index-1].PowerScore
			rank = rankings[index-1].Rank
		} else {
			powerScore = float64(numTeamns - index)
			rank = index + 1
		}
		rankings[index] = &TeamScoreData{
			Team:       &teams[index],
			Score:      score,
			PowerScore: powerScore,
			Rank:       rank,
		}
		glog.V(4).Infof("%s -- league=%s, week=%d, rank=%d, team=%s, "+
			"fantasyScore=%f, powerScore=%f",
			scoreType,
			leagueKey,
			week,
			rank,
			team.Name,
			score,
			powerScore)
	}

	results <- &WeeklyRanking{Week: week, Rankings: rankings, Projected: projection}
}

// GetPowerData returns a league's power rankings up to the given week and
// projections until the end of the season.
func GetPowerData(client PowerRankingsClient, leagueKey string, currentWeek int, numWeeks int) (*LeaguePowerData, error) {
	resultsChan := make(chan *WeeklyRanking)
	errorsChan := make(chan error)
	for week := 1; week <= currentWeek; week++ {
		go GetWeeklyRanking(client, leagueKey, week, resultsChan, errorsChan, false)
	}
	for week := currentWeek + 1; week <= numWeeks; week++ {
		go GetWeeklyRanking(client, leagueKey, week, resultsChan, errorsChan, true)
	}

	// Calculate power score for each team
	powerDataByTeamKey := make(map[string]*TeamPowerData)
	weeklyRankings := make([]*WeeklyRanking, numWeeks)
	for week := 1; week <= numWeeks; week++ {
		select {
		case err := <-errorsChan:
			glog.Warningf("error calculating weekly ranking -- "+
				"league=%s, error=%s",
				leagueKey,
				err)
			return nil, err
		case weeklyRanking := <-resultsChan:
			glog.V(4).Infof(
				"received weekly ranking -- week=%d",
				weeklyRanking.Week)
			weekIndex := weeklyRanking.Week - 1
			weeklyRankings[weekIndex] = weeklyRanking
			for _, teamScoreData := range weeklyRanking.Rankings {
				powerData, ok := powerDataByTeamKey[teamScoreData.Team.TeamKey]
				if !ok {
					powerData = &TeamPowerData{
						AllScores:           make([]*TeamScoreData, numWeeks),
						Team:                teamScoreData.Team,
						TotalPowerScore:     0.0,
						ProjectedPowerScore: 0.0,
					}
					powerDataByTeamKey[teamScoreData.Team.TeamKey] = powerData
				}
				powerData.AllScores[weekIndex] = teamScoreData
				if !weeklyRanking.Projected {
					powerData.TotalPowerScore += teamScoreData.PowerScore
				}
				powerData.ProjectedPowerScore += teamScoreData.PowerScore
				glog.V(4).Infof(
					"adding team score data -- team=%s, projection=%t, "+
						"powerScore=%f, totalPowerScore=%f, projectedPowerScore=%f",
					teamScoreData.Team.Name,
					weeklyRanking.Projected,
					teamScoreData.PowerScore,
					powerData.TotalPowerScore,
					powerData.ProjectedPowerScore)
			}
		}
	}

	glog.V(2).Infof("ranking teams -- league=%s", leagueKey)
	// Calculate the actual power rankings
	sortedPowerData := make([]*TeamPowerData, len(powerDataByTeamKey))
	index := 0
	for _, powerData := range powerDataByTeamKey {
		sortedPowerData[index] = powerData
		index++
	}
	sort.Sort(PowerRankings(sortedPowerData))
	for i, powerData := range sortedPowerData {
		// Handle ties
		if i > 0 && powerData.TotalPowerScore == sortedPowerData[i-1].TotalPowerScore {
			powerData.Rank = sortedPowerData[i-1].Rank
		} else {
			powerData.Rank = i + 1
		}
		glog.V(4).Infof("overall rankings -- league=%s, rank=%d, team=%s, "+
			"total=%f",
			leagueKey,
			powerData.Rank,
			powerData.Team.Name,
			powerData.TotalPowerScore)
	}

	glog.V(2).Infof("projecting rankings -- league=%s", leagueKey)
	sortedProjectionData := make([]*TeamPowerData, len(powerDataByTeamKey))
	index = 0
	for _, powerData := range powerDataByTeamKey {
		sortedProjectionData[index] = powerData
		index++
	}
	sort.Sort(ProjectedPowerRankings(sortedProjectionData))
	for i, powerData := range sortedProjectionData {
		// Handle ties
		if i > 0 &&
			powerData.ProjectedPowerScore ==
				sortedProjectionData[i-1].ProjectedPowerScore {
			powerData.ProjectedRank = sortedProjectionData[i-1].ProjectedRank
		} else {
			powerData.ProjectedRank = i + 1
		}
		glog.V(4).Infof("projected rankings -- league=%s, rank=%d, team=%s, "+
			"projected=%f",
			leagueKey,
			powerData.ProjectedRank,
			powerData.Team.Name,
			powerData.ProjectedPowerScore)
	}

	return &LeaguePowerData{
		OverallRankings:   sortedPowerData,
		ProjectedRankings: sortedProjectionData,
		ByTeam:            powerDataByTeamKey,
		ByWeek:            weeklyRankings,
	}, nil
}
