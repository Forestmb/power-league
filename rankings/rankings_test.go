package rankings

import (
	"errors"
	"sort"
	"testing"

	"github.com/Forestmb/power-league/Godeps/_workspace/src/github.com/Forestmb/goff"
)

func TestGetWeeklyRankingPowerScoreOrder(t *testing.T) {
	m := mockClient{
		WeekStats: map[int][]goff.Team{
			// Week 1
			1: []goff.Team{
				goff.Team{TeamKey: "a", TeamPoints: goff.Points{Total: 3.0}},
				goff.Team{TeamKey: "b", TeamPoints: goff.Points{Total: 4.0}},
				goff.Team{TeamKey: "c", TeamPoints: goff.Points{Total: 5.0}},
			},
		},
		WeekErrors: map[int]error{},
	}
	results := make(chan *WeeklyRanking)
	errorsChan := make(chan error)
	schemes := []Scheme{allPlayRecord{}}
	go GetWeeklyRanking(m, "", 1, results, errorsChan, false, schemes)
	weeklyRanking := <-results
	for i, teamData := range weeklyRanking.Rankings {
		if i > 0 && teamData.PowerScore > weeklyRanking.Rankings[i-1].PowerScore {
			t.Fatalf(
				"Assertion error. Weekly ranking not sorted in descending order."+
					"\n\t%d: %f > %d. %f\n",
				i,
				teamData.PowerScore,
				i-1,
				weeklyRanking.Rankings[i-1].PowerScore)
		}
	}
}

func TestGetWeeklyProjectedRankingPowerScoreOrder(t *testing.T) {
	m := mockClient{
		WeekStats: map[int][]goff.Team{
			// Week 1
			1: []goff.Team{
				goff.Team{
					TeamKey:             "a",
					TeamPoints:          goff.Points{Total: 3.0},
					TeamProjectedPoints: goff.Points{Total: 7.0},
				},
				goff.Team{
					TeamKey:             "b",
					TeamPoints:          goff.Points{Total: 4.0},
					TeamProjectedPoints: goff.Points{Total: 8.0},
				},
				goff.Team{
					TeamKey:             "c",
					TeamPoints:          goff.Points{Total: 5.0},
					TeamProjectedPoints: goff.Points{Total: 9.0},
				},
			},
		},
		WeekErrors: map[int]error{},
	}
	results := make(chan *WeeklyRanking)
	errorsChan := make(chan error)
	schemes := []Scheme{allPlayRecord{}}
	go GetWeeklyRanking(m, "", 1, results, errorsChan, true, schemes)
	weeklyRanking := <-results
	for i, teamData := range weeklyRanking.Rankings {
		if i > 0 && teamData.PowerScore > weeklyRanking.Rankings[i-1].PowerScore {
			t.Fatalf(
				"Assertion error. Weekly ranking not sorted in descending order."+
					"\n\t%d: %f > %d. %f\n",
				i,
				teamData.PowerScore,
				i-1,
				weeklyRanking.Rankings[i-1].PowerScore)
		}
	}
}

func TestGetWeeklyRankingCorrectPowerScores(t *testing.T) {
	m := mockClient{
		WeekStats: map[int][]goff.Team{
			// Week 1
			1: []goff.Team{
				goff.Team{TeamKey: "a", TeamPoints: goff.Points{Total: 3.0}},
				goff.Team{TeamKey: "b", TeamPoints: goff.Points{Total: 4.0}},
				goff.Team{TeamKey: "c", TeamPoints: goff.Points{Total: 5.0}},
			},
		},
		WeekErrors: map[int]error{},
	}
	results := make(chan *WeeklyRanking)
	errorsChan := make(chan error)
	schemes := []Scheme{allPlayRecord{}}
	go GetWeeklyRanking(m, "", 1, results, errorsChan, false, schemes)
	weeklyRanking := <-results
	for i, teamData := range weeklyRanking.Rankings {
		if i > 0 && teamData.PowerScore > weeklyRanking.Rankings[i-1].PowerScore {
			t.Fatalf(
				"Assertion error. Weekly ranking not sorted in descending order."+
					"\n\t%d: %f > %d. %f",
				i,
				teamData.PowerScore,
				i-1,
				weeklyRanking.Rankings[i-1].PowerScore)
		}
	}
}

func TestGetWeeklyProjectedRankingCorrectPowerScores(t *testing.T) {
	m := mockClient{
		WeekStats: map[int][]goff.Team{
			// Week 1
			1: []goff.Team{
				goff.Team{
					TeamKey:             "a",
					TeamPoints:          goff.Points{Total: 3.0},
					TeamProjectedPoints: goff.Points{Total: 7.0},
				},
				goff.Team{
					TeamKey:             "b",
					TeamPoints:          goff.Points{Total: 4.0},
					TeamProjectedPoints: goff.Points{Total: 8.0},
				},
				goff.Team{
					TeamKey:             "c",
					TeamPoints:          goff.Points{Total: 5.0},
					TeamProjectedPoints: goff.Points{Total: 9.0},
				},
			},
		},
		WeekErrors: map[int]error{},
	}
	results := make(chan *WeeklyRanking)
	errorsChan := make(chan error)
	schemes := []Scheme{allPlayRecord{}}
	go GetWeeklyRanking(m, "", 1, results, errorsChan, true, schemes)
	weeklyRanking := <-results
	for i, teamData := range weeklyRanking.Rankings {
		if i > 0 && teamData.PowerScore > weeklyRanking.Rankings[i-1].PowerScore {
			t.Fatalf(
				"Assertion error. Weekly ranking not sorted in descending order."+
					"\n\t%d: %f > %d. %f",
				i,
				teamData.PowerScore,
				i-1,
				weeklyRanking.Rankings[i-1].PowerScore)
		}
	}
}

func TestGetWeeklyRankingCorrectWeek(t *testing.T) {
	week := 3
	m := mockClient{
		WeekStats: map[int][]goff.Team{
			// Week 1
			week: []goff.Team{
				goff.Team{TeamKey: "a", TeamPoints: goff.Points{Total: 3.0}},
				goff.Team{TeamKey: "b", TeamPoints: goff.Points{Total: 4.0}},
				goff.Team{TeamKey: "c", TeamPoints: goff.Points{Total: 5.0}},
			},
		},
		WeekErrors: map[int]error{},
	}
	results := make(chan *WeeklyRanking)
	errorsChan := make(chan error)
	schemes := []Scheme{allPlayRecord{}}
	go GetWeeklyRanking(m, "", week, results, errorsChan, false, schemes)
	weeklyRanking := <-results
	if weeklyRanking.Week != week {
		t.Fatalf(
			"Assertion error. Incorrect week reported for weekly ranking."+
				"\n\tExpected: %d\n\tActual: %d\n",
			week,
			weeklyRanking.Week)
	}
}

func TestGetWeeklyRankingTieSamePowerScore(t *testing.T) {
	m := mockClient{
		WeekStats: map[int][]goff.Team{
			// Week 1
			1: []goff.Team{
				goff.Team{TeamKey: "a", TeamPoints: goff.Points{Total: 4.0}},
				goff.Team{TeamKey: "b", TeamPoints: goff.Points{Total: 4.0}},
				goff.Team{TeamKey: "c", TeamPoints: goff.Points{Total: 3.0}},
				goff.Team{TeamKey: "d", TeamPoints: goff.Points{Total: 5.0}},
			},
		},
		WeekErrors: map[int]error{},
	}
	results := make(chan *WeeklyRanking)
	errorsChan := make(chan error)
	schemes := []Scheme{allPlayRecord{}}
	go GetWeeklyRanking(m, "", 1, results, errorsChan, false, schemes)
	weeklyRanking := <-results

	team1 := weeklyRanking.Rankings[1]
	team2 := weeklyRanking.Rankings[2]

	if team1.Record.Wins != 1 && team2.Record.Wins != 1 {
		t.Fatalf("Assertion error. Incorrect rankings returned when tie is given")
	}

	if team1.Rank != team2.Rank {
		t.Fatalf(
			"Assertion error. Tied teams should have same rank."+
				"\n\tTeam 1: %d\n\tTeam 2: %d",
			team1.Rank,
			team2.Rank)
	}
}

func TestGetWeeklyProjectedRankingTieSameProjectedPowerScore(t *testing.T) {
	m := mockClient{
		WeekStats: map[int][]goff.Team{
			// Week 1
			1: []goff.Team{
				goff.Team{
					TeamKey:             "a",
					TeamPoints:          goff.Points{Total: 3.0},
					TeamProjectedPoints: goff.Points{Total: 7.0},
				},
				goff.Team{
					TeamKey:             "b1",
					TeamPoints:          goff.Points{Total: 4.0},
					TeamProjectedPoints: goff.Points{Total: 8.0},
				},
				goff.Team{
					TeamKey:             "b2",
					TeamPoints:          goff.Points{Total: 5.0},
					TeamProjectedPoints: goff.Points{Total: 8.0},
				},
				goff.Team{
					TeamKey:             "c",
					TeamPoints:          goff.Points{Total: 5.0},
					TeamProjectedPoints: goff.Points{Total: 9.0},
				},
			},
		},
		WeekErrors: map[int]error{},
	}
	results := make(chan *WeeklyRanking)
	errorsChan := make(chan error)
	schemes := []Scheme{allPlayRecord{}}
	go GetWeeklyRanking(m, "", 1, results, errorsChan, true, schemes)
	weeklyRanking := <-results

	team1 := weeklyRanking.Rankings[1]
	team2 := weeklyRanking.Rankings[2]

	if team1.Record.Wins != 1 && team2.Record.Wins != 1 {
		t.Fatalf("Assertion error. Incorrect rankings returned when tie is given")
	}

	if team1.Rank != team2.Rank {
		t.Fatalf(
			"Assertion error. Tied teams should have same rank."+
				"\n\tTeam 1: %d\n\tTeam 2: %d",
			team1.Rank,
			team2.Rank)
	}
}

func TestGetWeeklyRankingError(t *testing.T) {
	m := mockFailureClient{err: errors.New("error")}
	results := make(chan *WeeklyRanking)
	errorsChan := make(chan error)
	schemes := []Scheme{allPlayRecord{}}
	go GetWeeklyRanking(m, "", 1, results, errorsChan, false, schemes)
	err := <-errorsChan
	if err == nil {
		t.Fatal("no error returned")
	}
}

func TestGetPowerDataOverallRankings(t *testing.T) {
	league := &goff.League{
		LeagueKey: "leagueID",
		EndWeek:   3,
	}
	m := mockClient{
		Matchups: map[int][]goff.Matchup{
			// Week 1
			1: []goff.Matchup{
				goff.Matchup{
					Teams: []goff.Team{
						goff.Team{TeamKey: "a", TeamPoints: goff.Points{Total: 4.0}},
						goff.Team{TeamKey: "b", TeamPoints: goff.Points{Total: 3.0}},
					},
				},
				goff.Matchup{
					Teams: []goff.Team{
						goff.Team{TeamKey: "c", TeamPoints: goff.Points{Total: 2.0}},
						goff.Team{TeamKey: "d", TeamPoints: goff.Points{Total: 1.0}},
					},
				},
			},
			2: []goff.Matchup{
				goff.Matchup{
					Teams: []goff.Team{
						goff.Team{TeamKey: "a", TeamPoints: goff.Points{Total: 4.0}},
						goff.Team{TeamKey: "b", TeamPoints: goff.Points{Total: 3.0}},
					},
				},
				goff.Matchup{
					Teams: []goff.Team{
						goff.Team{TeamKey: "c", TeamPoints: goff.Points{Total: 2.0}},
						goff.Team{TeamKey: "d", TeamPoints: goff.Points{Total: 1.0}},
					},
				},
			},
			3: []goff.Matchup{
				goff.Matchup{
					Teams: []goff.Team{
						goff.Team{TeamKey: "a", TeamPoints: goff.Points{Total: 4.0}},
						goff.Team{TeamKey: "b", TeamPoints: goff.Points{Total: 3.0}},
					},
				},
				goff.Matchup{
					Teams: []goff.Team{
						goff.Team{TeamKey: "c", TeamPoints: goff.Points{Total: 2.0}},
						goff.Team{TeamKey: "d", TeamPoints: goff.Points{Total: 1.0}},
					},
				},
			},
		},
		WeekErrors:      map[int]error{},
		StandingsLeague: league,
	}
	data, err := GetPowerData(m, league, 3)

	if err != nil {
		t.Fatalf("GetPowerData returned unexpected error: %s\n", err)
	}

	allPlayData := data[0]
	rankings := allPlayData.OverallRankings
	if rankings[0].Team.TeamKey != "a" ||
		rankings[1].Team.TeamKey != "b" ||
		rankings[2].Team.TeamKey != "c" ||
		rankings[3].Team.TeamKey != "d" {
		t.Fatalf("GetPowerData returned incorrect rankings.\n"+
			"\trankings: %+v",
			rankings)
	}

	if rankings[0].OverallRecord.Wins != 9 ||
		rankings[1].OverallRecord.Wins != 6 ||
		rankings[2].OverallRecord.Wins != 3 ||
		rankings[3].OverallRecord.Wins != 0 {
		t.Fatalf("GetPowerData returned incorrect scores.\n"+
			"\trankings: %+v",
			rankings)
	}

	if rankings[0].Rank != 1 ||
		rankings[1].Rank != 2 ||
		rankings[2].Rank != 3 ||
		rankings[3].Rank != 4 {
		t.Fatalf("GetPowerData returned incorrect ranks.\n"+
			"\trankings: %+v",
			rankings)
	}
}

func TestGetProjectedPowerDataOverallRankings(t *testing.T) {
	league := &goff.League{
		LeagueKey: "leagueID",
		EndWeek:   3,
		Standings: []goff.Team{
			goff.Team{
				TeamKey: "a",
				TeamStandings: goff.TeamStandings{
					Rank: 1,
					Record: goff.Record{
						Wins:   5,
						Losses: 2,
						Ties:   0,
					},
					PointsFor:     12345.0,
					PointsAgainst: 54321.0,
				},
			},
			goff.Team{
				TeamKey: "b",
				TeamStandings: goff.TeamStandings{
					Rank: 2,
					Record: goff.Record{
						Wins:   3,
						Losses: 3,
						Ties:   1,
					},
					PointsFor:     1234.0,
					PointsAgainst: 4321.0,
				},
			},
			goff.Team{
				TeamKey: "c",
				TeamStandings: goff.TeamStandings{
					Rank: 3,
					Record: goff.Record{
						Wins:   0,
						Losses: 6,
						Ties:   1,
					},
					PointsFor:     234.0,
					PointsAgainst: 321.0,
				},
			},
		},
	}
	m := mockClient{
		Matchups: map[int][]goff.Matchup{
			1: []goff.Matchup{
				goff.Matchup{
					Teams: []goff.Team{
						goff.Team{
							TeamKey:             "a",
							TeamPoints:          goff.Points{Total: 4.0},
							TeamProjectedPoints: goff.Points{Total: 7.0},
						},
						goff.Team{
							TeamKey:             "b",
							TeamPoints:          goff.Points{Total: 3.0},
							TeamProjectedPoints: goff.Points{Total: 8.0},
						},
					},
				},
				goff.Matchup{
					Teams: []goff.Team{
						goff.Team{
							TeamKey:             "c",
							TeamPoints:          goff.Points{Total: 2.0},
							TeamProjectedPoints: goff.Points{Total: 9.0},
						},
						goff.Team{
							TeamKey:             "d",
							TeamPoints:          goff.Points{Total: 1.0},
							TeamProjectedPoints: goff.Points{Total: 10.0},
						},
					},
				},
			},
		},
		WeekStats: map[int][]goff.Team{
			2: []goff.Team{
				goff.Team{
					TeamKey:             "a",
					TeamPoints:          goff.Points{Total: 4.0},
					TeamProjectedPoints: goff.Points{Total: 7.0},
				},
				goff.Team{
					TeamKey:             "b",
					TeamPoints:          goff.Points{Total: 3.0},
					TeamProjectedPoints: goff.Points{Total: 8.0},
				},
				goff.Team{
					TeamKey:             "c",
					TeamPoints:          goff.Points{Total: 2.0},
					TeamProjectedPoints: goff.Points{Total: 9.0},
				},
				goff.Team{
					TeamKey:             "d",
					TeamPoints:          goff.Points{Total: 1.0},
					TeamProjectedPoints: goff.Points{Total: 10.0},
				},
			},
			3: []goff.Team{
				goff.Team{
					TeamKey:             "a",
					TeamPoints:          goff.Points{Total: 4.0},
					TeamProjectedPoints: goff.Points{Total: 7.0},
				},
				goff.Team{
					TeamKey:             "b",
					TeamPoints:          goff.Points{Total: 3.0},
					TeamProjectedPoints: goff.Points{Total: 8.0},
				},
				goff.Team{
					TeamKey:             "c",
					TeamPoints:          goff.Points{Total: 2.0},
					TeamProjectedPoints: goff.Points{Total: 9.0},
				},
				goff.Team{
					TeamKey:             "d",
					TeamPoints:          goff.Points{Total: 1.0},
					TeamProjectedPoints: goff.Points{Total: 10.0},
				},
			},
		},
		WeekErrors:      map[int]error{},
		StandingsLeague: league,
	}
	data, err := GetPowerData(m, league, 1)

	if err != nil {
		t.Fatalf("GetPowerData returned unexpected error: %s\n", err)
	}

	allPlayData := data[0]
	rankings := allPlayData.ProjectedRankings
	if rankings[0].Team.TeamKey != "d" ||
		rankings[1].Team.TeamKey != "c" ||
		rankings[2].Team.TeamKey != "b" ||
		rankings[3].Team.TeamKey != "a" {
		t.Fatalf("GetPowerData returned incorrect rankings.\n"+
			"\trankings: %+v\n",
			rankings)
	}

	if rankings[0].ProjectedOverallRecord.Wins != 6 ||
		rankings[1].ProjectedOverallRecord.Wins != 5 ||
		rankings[2].ProjectedOverallRecord.Wins != 4 ||
		rankings[3].ProjectedOverallRecord.Wins != 3 {
		t.Fatalf("GetPowerData returned incorrect scores.\n"+
			"\trankings: %+v\n",
			rankings)
	}

	if rankings[0].ProjectedRank != 1 ||
		rankings[1].ProjectedRank != 2 ||
		rankings[2].ProjectedRank != 3 ||
		rankings[3].ProjectedRank != 4 {
		t.Fatalf("GetPowerData returned incorrect ranks.\n"+
			"\trankings: %d, %d, %d, %d\n",
			rankings[0].ProjectedRank,
			rankings[1].ProjectedRank,
			rankings[2].ProjectedRank,
			rankings[3].ProjectedRank)
	}
}

func TestGetPowerDataClientError(t *testing.T) {
	league := &goff.League{
		LeagueKey: "leagueID",
		EndWeek:   3,
	}
	m := mockClient{
		MatchupsError: errors.New("error"),
		WeekStats: map[int][]goff.Team{
			// Week 1
			1: []goff.Team{
				goff.Team{TeamKey: "a", TeamPoints: goff.Points{Total: 4.0}},
				goff.Team{TeamKey: "b", TeamPoints: goff.Points{Total: 3.0}},
				goff.Team{TeamKey: "c", TeamPoints: goff.Points{Total: 2.0}},
				goff.Team{TeamKey: "d", TeamPoints: goff.Points{Total: 1.0}},
			},
			3: []goff.Team{
				goff.Team{TeamKey: "a", TeamPoints: goff.Points{Total: 4.0}},
				goff.Team{TeamKey: "b", TeamPoints: goff.Points{Total: 3.0}},
				goff.Team{TeamKey: "c", TeamPoints: goff.Points{Total: 2.0}},
				goff.Team{TeamKey: "d", TeamPoints: goff.Points{Total: 1.0}},
			},
		},
		WeekErrors: map[int]error{
			2: errors.New("error"),
		},
		StandingsLeague: league,
	}
	data, err := GetPowerData(m, league, 3)
	if err == nil {
		t.Fatalf("GetPowerData did not return error\n\tdata: %+v\n", data)
	}
}

func TestGetPowerDataTies(t *testing.T) {
	league := &goff.League{
		LeagueKey: "leagueID",
		EndWeek:   4,
		Settings: goff.Settings{
			UsesPlayoff:      true,
			PlayoffStartWeek: 1,
		},
	}
	m := mockClient{
		// Teams a/b will be tied after 4 weeks
		WeekStats: map[int][]goff.Team{
			// Week 1
			1: []goff.Team{
				goff.Team{TeamKey: "a", TeamPoints: goff.Points{Total: 4.0}},
				goff.Team{TeamKey: "b", TeamPoints: goff.Points{Total: 5.0}},
				goff.Team{TeamKey: "c", TeamPoints: goff.Points{Total: 2.0}},
				goff.Team{TeamKey: "d", TeamPoints: goff.Points{Total: 1.0}},
			},
			2: []goff.Team{
				goff.Team{TeamKey: "a", TeamPoints: goff.Points{Total: 5.0}},
				goff.Team{TeamKey: "b", TeamPoints: goff.Points{Total: 4.0}},
				goff.Team{TeamKey: "c", TeamPoints: goff.Points{Total: 2.0}},
				goff.Team{TeamKey: "d", TeamPoints: goff.Points{Total: 1.0}},
			},
			3: []goff.Team{
				goff.Team{TeamKey: "a", TeamPoints: goff.Points{Total: 4.0}},
				goff.Team{TeamKey: "b", TeamPoints: goff.Points{Total: 5.0}},
				goff.Team{TeamKey: "c", TeamPoints: goff.Points{Total: 2.0}},
				goff.Team{TeamKey: "d", TeamPoints: goff.Points{Total: 1.0}},
			},
			4: []goff.Team{
				goff.Team{TeamKey: "a", TeamPoints: goff.Points{Total: 5.0}},
				goff.Team{TeamKey: "b", TeamPoints: goff.Points{Total: 4.0}},
				goff.Team{TeamKey: "c", TeamPoints: goff.Points{Total: 2.0}},
				goff.Team{TeamKey: "d", TeamPoints: goff.Points{Total: 1.0}},
			},
		},
		WeekErrors:      map[int]error{},
		StandingsLeague: league,
	}
	data, err := GetPowerData(m, league, 4)

	if err != nil {
		t.Fatalf("GetPowerData returned unexpected error: %s\n", err)
	}

	allPlayData := data[0]
	rankings := allPlayData.OverallRankings
	if rankings[0].Rank != 1 ||
		rankings[1].Rank != 1 ||
		rankings[2].Rank != 3 ||
		rankings[3].Rank != 4 {
		t.Fatalf("GetPowerData did not correctly rank teams after a tie\n"+
			"\trankings: %+v\n",
			rankings)
	}
}

func TestPowerRankingsSort(t *testing.T) {
	var teamData = []*TeamPowerData{
		&TeamPowerData{TotalScore: 3.0},
		&TeamPowerData{TotalScore: 2.0},
		&TeamPowerData{TotalScore: 4.0, Team: &goff.Team{Name: "Name"}},
		&TeamPowerData{TotalScore: 4.0, Team: &goff.Team{Name: "Name2"}},
		&TeamPowerData{TotalScore: 1.0},
		&TeamPowerData{TotalScore: 5.0},
	}
	sort.Sort(PowerRankings(teamData))
	for i, team := range teamData {
		if i > 0 && team.TotalScore > teamData[i-1].TotalScore {
			t.Fatalf(
				"Assertion error. Team points not sorted in descending order."+
					"\n\t%d: %f > %d. %f\n",
				i,
				team.TotalScore,
				i-1,
				teamData[i-1].TotalScore)
		}
	}
}

func TestTeamRankingSort(t *testing.T) {
	var teams = []goff.Team{
		goff.Team{TeamPoints: goff.Points{Total: 5.0}},
		goff.Team{TeamPoints: goff.Points{Total: 2.0}},
		goff.Team{TeamPoints: goff.Points{Total: 4.0}},
		goff.Team{TeamPoints: goff.Points{Total: 3.0}},
		goff.Team{TeamPoints: goff.Points{Total: 1.0}},
	}
	sort.Sort(TeamRanking(teams))
	for i, team := range teams {
		if i > 0 && team.TeamPoints.Total > teams[i-1].TeamPoints.Total {
			t.Fatalf(
				"Assertion error. Team points not sorted in descending order."+
					"\n\t%d: %f > %d. %f\n",
				i,
				team.TeamPoints.Total,
				i-1,
				teams[i-1].TeamPoints.Total)
		}
	}
}

type mockFailureClient struct {
	err error
}

func (m mockFailureClient) GetAllTeamStats(leagueKey string, week int, projected bool) ([]goff.Team, error) {
	return nil, m.err
}

func (m mockFailureClient) GetMatchupsForWeekRange(leagueKey string, startWeek, endWeek int) (map[int][]goff.Matchup, error) {
	return nil, m.err
}

func (m mockFailureClient) GetLeagueStandings(leagueKey string) (*goff.League, error) {
	return nil, m.err
}

type mockClient struct {
	WeekStats  map[int][]goff.Team
	WeekErrors map[int]error

	Matchups      map[int][]goff.Matchup
	MatchupsError error

	StandingsLeague *goff.League
	StandingsError  error
}

func (m mockClient) GetAllTeamStats(leagueKey string, week int, projected bool) ([]goff.Team, error) {
	return m.WeekStats[week], m.WeekErrors[week]
}

func (m mockClient) GetMatchupsForWeekRange(leagueKey string, startWeek, endWeek int) (map[int][]goff.Matchup, error) {
	return m.Matchups, m.MatchupsError
}

func (m mockClient) GetLeagueStandings(leagueKey string) (*goff.League, error) {
	return m.StandingsLeague, m.StandingsError
}
