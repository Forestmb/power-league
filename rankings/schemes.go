package rankings

import (
	"sort"

	"github.com/Forestmb/goff"
)

// Types contains the various types of ranking schemes offered by this package.
var Types = struct {
	// SCORE represents a rankings scheme that uses points to compare teams
	SCORE string

	// RECORD represents a rankings scheme that uses a win/loss/tie record to
	// compare teams
	RECORD string
}{
	"score",
	"record",
}

//
// Interface
//

// A Scheme is a way to rank teams based on their weekly performance
type Scheme interface {
	ID() string
	DisplayName() string
	Type() string
	CalculateWeeklyRankings(
		week int,
		teams []goff.Team,
		projected bool,
		results chan *WeeklyRanking)
}

//
// Data structures
//

// TeamRanking ranks teams based on their performance for a single week
type TeamRanking []goff.Team

func (t TeamRanking) Len() int {
	return len(t)
}

func (t TeamRanking) Less(i int, j int) bool {
	if t[i].TeamPoints.Total == t[j].TeamPoints.Total {
		return t[i].Name < t[j].Name
	}
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
	if t[i].TeamProjectedPoints.Total == t[j].TeamProjectedPoints.Total {
		return t[i].Name < t[j].Name
	}
	return t[i].TeamProjectedPoints.Total > t[j].TeamProjectedPoints.Total
}

func (t TeamProjectedRanking) Swap(i int, j int) {
	t[i], t[j] = t[j], t[i]
}

//
// Schemes
//

// GetSchemes returns the supported rankings formats
func GetSchemes() []Scheme {
	return []Scheme{
		allPlayRecord{},
		victoryPoints{},
		totalPoints{},
	}
}

type victoryPoints struct {
}

func (v victoryPoints) DisplayName() string {
	return "Victory Points"
}

func (v victoryPoints) ID() string {
	return "victory-points"
}

func (v victoryPoints) Type() string {
	return Types.SCORE
}

// CalculateWeeklyRankings for a 'Victory Points' scheme gives each team a
// win if they score more points than half of the league. All other teams get
// a loss.
func (v victoryPoints) CalculateWeeklyRankings(
	week int,
	teams []goff.Team,
	projected bool,
	results chan *WeeklyRanking) {

	// Sort teams and convert them into TeamScoreData
	if !projected {
		sort.Sort(TeamRanking(teams))
	} else {
		sort.Sort(TeamProjectedRanking(teams))
	}

	rankings := make([]*TeamScoreData, len(teams))
	for index := range teams {
		team := &teams[index]

		var score float64
		if !projected {
			score = team.TeamPoints.Total
		} else {
			score = team.TeamProjectedPoints.Total
		}

		rankings[index] = &TeamScoreData{
			Team:         team,
			FantasyScore: score,
			Record:       &goff.Record{},
			Projected:    projected,
		}
	}

	// Update ranks and assign victory points
	winCutoff := len(teams) / 2
	for i := range rankings {
		if i > 0 && rankings[i].PowerScore == rankings[i-1].PowerScore {
			rankings[i].Rank = rankings[i-1].Rank
		} else {
			rankings[i].Rank = i + 1
		}

		if rankings[i].Rank <= winCutoff {
			rankings[i].PowerScore = 1.0
		}
	}

	results <- &WeeklyRanking{
		Scheme:    v,
		Week:      week,
		Rankings:  rankings,
		Projected: projected,
	}
}

type allPlayRecord struct {
}

func (a allPlayRecord) ID() string {
	return "all-play"
}

func (a allPlayRecord) DisplayName() string {
	return "All-Play"
}

func (a allPlayRecord) Type() string {
	return Types.RECORD
}

// CalculateWeeklyRankings for a 'All-Play' ranking scheme ranks teams based
// on what their record would be if they played every other team in the
// league in a head-to-head matchup.
func (a allPlayRecord) CalculateWeeklyRankings(
	week int,
	teams []goff.Team,
	projected bool,
	results chan *WeeklyRanking) {

	// Sort teams and convert them into TeamScoreData
	if !projected {
		sort.Sort(TeamRanking(teams))
	} else {
		sort.Sort(TeamProjectedRanking(teams))
	}

	rankings := make([]*TeamScoreData, len(teams))
	for index := range teams {
		team := &teams[index]

		var score float64
		if !projected {
			score = team.TeamPoints.Total
		} else {
			score = team.TeamProjectedPoints.Total
		}

		rankings[index] = &TeamScoreData{
			Team:         team,
			FantasyScore: score,
			Record:       &goff.Record{},
			Projected:    projected,
		}
	}

	for index, team := range rankings {
		score := team.FantasyScore

		for _, otherTeam := range rankings {
			if team.Team != otherTeam.Team {
				otherScore := otherTeam.FantasyScore

				if score > otherScore {
					rankings[index].Record.Wins++
				} else if score == otherScore {
					rankings[index].Record.Ties++
				} else {
					rankings[index].Record.Losses++
				}
			}
		}
	}

	// Update ranks
	for i := range rankings {
		if i > 0 && rankings[i].Record.Wins == rankings[i-1].Record.Wins {
			rankings[i].Rank = rankings[i-1].Rank
		} else {
			rankings[i].Rank = i + 1
		}
	}

	results <- &WeeklyRanking{
		Scheme:    a,
		Week:      week,
		Rankings:  rankings,
		Projected: projected,
	}
}

type totalPoints struct {
}

func (a totalPoints) ID() string {
	return "total-points"
}

func (a totalPoints) DisplayName() string {
	return "Total Points"
}

func (a totalPoints) Type() string {
	return Types.SCORE
}

// CalculateWeeklyRankings for a 'Total Points' ranking scheme gives each team
// points based on how many fantasy points they scored.
func (a totalPoints) CalculateWeeklyRankings(
	week int,
	teams []goff.Team,
	projected bool,
	results chan *WeeklyRanking) {

	// Sort teams and convert them into TeamScoreData
	if !projected {
		sort.Sort(TeamRanking(teams))
	} else {
		sort.Sort(TeamProjectedRanking(teams))
	}

	rankings := make([]*TeamScoreData, len(teams))
	for index := range teams {
		team := &teams[index]

		var score float64
		if !projected {
			score = team.TeamPoints.Total
		} else {
			score = team.TeamProjectedPoints.Total
		}

		rankings[index] = &TeamScoreData{
			Team:         team,
			FantasyScore: score,
			PowerScore:   score,
			Record:       &goff.Record{},
			Projected:    projected,
		}
	}

	// Update ranks
	for i := range rankings {
		if i > 0 && rankings[i].PowerScore == rankings[i-1].PowerScore {
			rankings[i].Rank = rankings[i-1].Rank
		} else {
			rankings[i].Rank = i + 1
		}
	}

	results <- &WeeklyRanking{
		Scheme:    a,
		Week:      week,
		Rankings:  rankings,
		Projected: projected,
	}
}
