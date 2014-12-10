package rankings

import (
	"errors"
	"sort"
	"testing"

	"github.com/Forestmb/goff"
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
	go GetWeeklyRanking(m, "", 1, results, errorsChan, false)
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
	go GetWeeklyRanking(m, "", 1, results, errorsChan, true)
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
	go GetWeeklyRanking(m, "", 1, results, errorsChan, false)
	weeklyRanking := <-results
	for i, teamData := range weeklyRanking.Rankings {
		if i > 0 && teamData.Score > weeklyRanking.Rankings[i-1].Score {
			t.Fatalf(
				"Assertion error. Weekly ranking not sorted in descending order."+
					"\n\t%d: %f > %d. %f",
				i,
				teamData.Score,
				i-1,
				weeklyRanking.Rankings[i-1].Score)
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
	go GetWeeklyRanking(m, "", 1, results, errorsChan, true)
	weeklyRanking := <-results
	for i, teamData := range weeklyRanking.Rankings {
		if i > 0 && teamData.Score > weeklyRanking.Rankings[i-1].Score {
			t.Fatalf(
				"Assertion error. Weekly ranking not sorted in descending order."+
					"\n\t%d: %f > %d. %f",
				i,
				teamData.Score,
				i-1,
				weeklyRanking.Rankings[i-1].Score)
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
	go GetWeeklyRanking(m, "", week, results, errorsChan, false)
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
	go GetWeeklyRanking(m, "", 1, results, errorsChan, false)
	weeklyRanking := <-results

	team1 := weeklyRanking.Rankings[1]
	team2 := weeklyRanking.Rankings[2]

	if team1.Score != 4.0 && team2.Score != 4.0 {
		t.Fatalf("Assertion error. Incorrect rankings returned when tie is given")
	}

	if team1.PowerScore != team2.PowerScore {
		t.Fatalf(
			"Assertion error. Tied teams should have same power score."+
				"\n\tTeam 1: %f\n\tTeam 2: %f",
			team1.PowerScore,
			team2.PowerScore)
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
	go GetWeeklyRanking(m, "", 1, results, errorsChan, true)
	weeklyRanking := <-results

	team1 := weeklyRanking.Rankings[1]
	team2 := weeklyRanking.Rankings[2]

	if team1.Score != 8.0 && team2.Score != 8.0 {
		t.Fatalf("Assertion error. Incorrect rankings returned when tie is given")
	}

	if team1.PowerScore != team2.PowerScore {
		t.Fatalf(
			"Assertion error. Tied teams should have same projected power score."+
				"\n\tTeam 1: %f\n\tTeam 2: %f",
			team1.PowerScore,
			team2.PowerScore)
	}
}

func TestGetWeeklyRankingError(t *testing.T) {
	m := mockFailureClient{err: errors.New("error")}
	results := make(chan *WeeklyRanking)
	errorsChan := make(chan error)
	go GetWeeklyRanking(m, "", 1, results, errorsChan, false)
	err := <-errorsChan
	if err == nil {
		t.Fatal("no error returned")
	}
}

func TestGetPowerDataOverallRankings(t *testing.T) {
	m := mockClient{
		WeekStats: map[int][]goff.Team{
			// Week 1
			1: []goff.Team{
				goff.Team{TeamKey: "a", TeamPoints: goff.Points{Total: 4.0}},
				goff.Team{TeamKey: "b", TeamPoints: goff.Points{Total: 3.0}},
				goff.Team{TeamKey: "c", TeamPoints: goff.Points{Total: 2.0}},
				goff.Team{TeamKey: "d", TeamPoints: goff.Points{Total: 1.0}},
			},
			2: []goff.Team{
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
		WeekErrors: map[int]error{},
	}
	data, err := GetPowerData(m, "leagueID", 3, 3)

	if err != nil {
		t.Fatalf("GetPowerData returned unexpected error: %s\n", err)
	}

	rankings := data.OverallRankings
	if rankings[0].Team.TeamKey != "a" ||
		rankings[1].Team.TeamKey != "b" ||
		rankings[2].Team.TeamKey != "c" ||
		rankings[3].Team.TeamKey != "d" {
		t.Fatalf("GetPowerData returned incorrect rankings.\n"+
			"\trankings: %+v",
			rankings)
	}

	if rankings[0].TotalPowerScore != 9.0 ||
		rankings[1].TotalPowerScore != 6.0 ||
		rankings[2].TotalPowerScore != 3.0 ||
		rankings[3].TotalPowerScore != 0.0 {
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
	m := mockClient{
		WeekStats: map[int][]goff.Team{
			1: []goff.Team{
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
		WeekErrors: map[int]error{},
	}
	data, err := GetPowerData(m, "leagueID", 1, 3)

	if err != nil {
		t.Fatalf("GetPowerData returned unexpected error: %s\n", err)
	}

	rankings := data.ProjectedRankings
	if rankings[0].Team.TeamKey != "d" ||
		rankings[1].Team.TeamKey != "c" ||
		rankings[2].Team.TeamKey != "b" ||
		rankings[3].Team.TeamKey != "a" {
		t.Fatalf("GetPowerData returned incorrect rankings.\n"+
			"\trankings: %+v\n",
			rankings)
	}

	if rankings[0].ProjectedPowerScore != 6.0 ||
		rankings[1].ProjectedPowerScore != 5.0 ||
		rankings[2].ProjectedPowerScore != 4.0 ||
		rankings[3].ProjectedPowerScore != 3.0 {
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
	m := mockClient{
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
	}
	data, err := GetPowerData(m, "leagueID", 3, 3)
	if err == nil {
		t.Fatalf("GetPowerData did not return error\n\tdata: %+v\n", data)
	}
}

func TestGetPowerDataTies(t *testing.T) {
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
		WeekErrors: map[int]error{},
	}
	data, err := GetPowerData(m, "leagueID", 4, 4)

	if err != nil {
		t.Fatalf("GetPowerData returned unexpected error: %s\n", err)
	}

	rankings := data.OverallRankings
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
		&TeamPowerData{TotalPowerScore: 3.0},
		&TeamPowerData{TotalPowerScore: 2.0},
		&TeamPowerData{TotalPowerScore: 4.0},
		&TeamPowerData{TotalPowerScore: 4.0},
		&TeamPowerData{TotalPowerScore: 1.0},
		&TeamPowerData{TotalPowerScore: 5.0},
	}
	sort.Sort(PowerRankings(teamData))
	for i, team := range teamData {
		if i > 0 && team.TotalPowerScore > teamData[i-1].TotalPowerScore {
			t.Fatalf(
				"Assertion error. Team points not sorted in descending order."+
					"\n\t%d: %f > %d. %f\n",
				i,
				team.TotalPowerScore,
				i-1,
				teamData[i-1].TotalPowerScore)
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

type mockClient struct {
	WeekStats  map[int][]goff.Team
	WeekErrors map[int]error
}

func (m mockClient) GetAllTeamStats(leagueKey string, week int, projected bool) ([]goff.Team, error) {
	return m.WeekStats[week], m.WeekErrors[week]
}
