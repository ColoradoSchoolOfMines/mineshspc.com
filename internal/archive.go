package internal

import (
	"net/http"
)

type Link struct {
	URL   string
	Title string
}

type WinningTeam struct {
	Place    string
	Name     string
	School   string
	Location string
}

type CompetitionResult struct {
	Name      string
	Shortname string
	Teams     []WinningTeam
}

type YearInfo struct {
	Year            int
	RecapParagraphs []string
	Links           []Link
	Results         []CompetitionResult
}

func (a *Application) GetArchiveTemplate(*http.Request) map[string]any {
	return map[string]any{
		"YearInfo": []YearInfo{
			{
				Year: 2023,
				RecapParagraphs: []string{
					"The 2023 competition again featured two divisions: beginner and advanced. As with 2022, it was a hybrid competition, but we awarded prizes for both in-person and remote winners in both divisions.",
					"The advanced division featured 31 teams, while the beginner division had 34 teams.",
				},
				Links: []Link{
					{"/static/2023-solutions.pdf", "Solution Sketch Slides"},
					{"https://sumnerevans.com/posts/school/2023-hspc/", "Competition Recap and Solution Sketches"},
					{"https://mines23advanced.kattis.com/problems", "Advanced Problems"},
					{"https://mines23beginner.kattis.com/problems", "Beginner Problems"},
				},
				Results: []CompetitionResult{
					{
						Name:      "Advanced In-Person",
						Shortname: "AdvancedInPerson",
						Teams: []WinningTeam{
							{"1st", "Code Rats", "Futures Lab", "Fort Collins, Colorado"},
							{"2nd", "The Spanish Inquisition", "Regis Jesuit High School", "Aurora, Colorado"},
							{"3nd", "CA is 202", "Colorado Academy", "Denver, Colorado"},
						},
					},
					{
						Name:      "Beginner In-Person",
						Shortname: "BeginnerInPerson",
						Teams: []WinningTeam{
							{"1st", "Spaghetti Code and Meatballs", "Warren Tech", "Lakewood, Colorado"},
							{"2nd", "Innovation Center 1", "Innovation Center SVVSD", "Longmont, Colorado"},
							{"3nd", "Team LuLo", "Colorado Academy", "Denver, Colorado"},
						},
					},
					{
						Name:      "Advanced Remote",
						Shortname: "AdvancedRemote",
						Teams: []WinningTeam{
							{"1st", "River Hill Team #1", "River Hill High School", "Clarksville, Maryland"},
							{"2nd", "CreekCyberBruins", "Cherry Creek High School", "Greenwood Village, Colorado"},
							{"3nd", "JMS", "Bergen County Academies", "Bergen County, New Jersey"},
						},
					},
					{
						Name:      "Beginner Remote",
						Shortname: "BeginnerRemote",
						Teams: []WinningTeam{
							{"1st", "Wormhole", "Voice of Calling NPO", "Northridge, California"},
							{"2nd", "Lineup", "Voice of Calling NPO", "Northridge, California"},
							{"3nd", "River Hill Team #2", "River Hill High School", "Clarksville, Maryland"},
						},
					},
				},
			},
			{
				Year: 2022,
				RecapParagraphs: []string{
					"The 2022 competition was the first to feature two divisions: a beginner division and an advanced division. It was also the first hybrid competition with both remote and in-person contestants.",
					"The advanced division had 26 teams, while the beginner division had 39 teams. Due to the number of teams, we decided to give awards to first place through fourth place.",
				},
				Links: []Link{
					{"https://sumnerevans.com/posts/school/2022-hspc/", "Competition Recap and Solution Sketches"},
					{"https://mines22advanced.kattis.com/problems", "Advanced Problems"},
					{"https://mines22beginner.kattis.com/problems", "Beginner Problems"},
				},
				Results: []CompetitionResult{
					{
						Name: "Advanced",
						Teams: []WinningTeam{
							{"1st", "Pen A Team", "PEN Academy", "Cresskill, New Jersey"},
							{"2nd", "Cherry Creek Cobras", "Cherry Creek High School", "Greenwood Village, Colorado"},
							{"3nd", "River Hill Team 1", "River Hill High School", "Clarksville, Maryland"},
							{"4th", "The Spanish Inquisition", "Regis Jesuit High School", "Aurora, Colorado"},
						},
					},
					{
						Name: "Beginner",
						Teams: []WinningTeam{
							{"1st", "LLL", "Future Forward at Bollman", "Thornton, Colorado"},
							{"2nd", "Error 404: Name not found", "Colorado Academy", "Denver, Colorado"},
							{"3nd", "Liberty 1", "Liberty Common School", "Fort Collins, Colorado"},
							{"4th", "Cool Cats", "Arvada West High School", "Arvada, Colorado"},
						},
					},
				},
			},
			{
				Year: 2021,
				RecapParagraphs: []string{
					"The 2021 competition was an all-remote competition featuring 55 teams from across the nation.",
				},
				Links: []Link{
					{"https://sumnerevans.com/posts/school/2021-hspc/", "Competition Recap and Solution Sketches"},
					{"https://mines21.kattis.com/problems", "Problems"},
				},
				Results: []CompetitionResult{
					{
						Teams: []WinningTeam{
							{"1st", "River Hill HS Team 1", "River Hill High School", "Clarksville, Maryland"},
							{"2nd", "PEN A Team", "PEN Academy", "Cresskill, New Jersey"},
							{"3nd", "River Hill HS Team 2", "River Hill High School", "Clarksville, Maryland"},
						},
					},
				},
			},
			{
				Year: 2020,
				RecapParagraphs: []string{
					"Due to COVID, the 2020 competition was the first all-remote HSPC competition. The competition featured 30 teams.",
				},
				Links: []Link{
					{"https://sumnerevans.com/posts/school/2020-hspc/", "Competition Recap and Solution Sketches"},
					{"https://mines20.kattis.com/problems", "Problems"},
				},
				Results: []CompetitionResult{
					{
						Teams: []WinningTeam{
							{"1st", "Installation Wizards", "STEM School Highlands Ranch", "Highlands Ranch, Colorado"},
							{"2nd", "i", "STEM School Highlands Ranch", "Highlands Ranch, Colorado"},
							{"3nd", "Sun Devils", "Kent Denver", "Denver, Colorado"},
						},
					},
				},
			},
			{
				Year: 2019,
				RecapParagraphs: []string{
					"The second ever CS@Mines High School Programming Competition featured 22 teams from all around Colorado and from as far as Steamboat Springs.",
				},
				Links: []Link{
					{"https://sumnerevans.com/posts/school/2019-hspc/", "Competition Recap and Solution Sketches"},
					{"https://mines19.kattis.com/problems", "Problems"},
				},
				Results: []CompetitionResult{
					{
						Teams: []WinningTeam{
							{"1st", "STEM Team 1", "STEM School Highlands Ranch", "Highlands Ranch, Colorado"},
							{"2nd", "IntrospectionExceptions", "Colorado Academy", "Lakewood, Colorado"},
							{"3nd", "Team 2", "?", "?"},
						},
					},
				},
			},
			{
				Year: 2018,
				RecapParagraphs: []string{
					"The first ever CS@Mines High School Programming Competition featured 22 teams.",
				},
				Links: []Link{
					{"https://mines18.kattis.com/problems", "Problems"},
				},
				Results: []CompetitionResult{
					{
						Teams: []WinningTeam{
							{"1st", "The Crummies", "Warren Tech", "Arvada, Colorado"},
							{"2nd", "The Bean Beans", "Colorado Academy", "Lakewood, Colorado"},
							{"3nd", "Warriors", "Arapahoe High School", "Centennial, Colorado"},
						},
					},
				},
			},
		},
	}
}
