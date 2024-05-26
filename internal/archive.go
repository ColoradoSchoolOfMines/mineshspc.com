package internal

import (
	"github.com/ColoradoSchoolOfMines/mineshspc.com/internal/templates"
)

var archiveInfo = []templates.YearInfo{
	{
		Year: 2024,
		RecapParagraphs: []string{
			"The 2024 competition returned to an in-person only competition, but we also had an open division. We gave a separate set of prizes for teams consisting of only first-time competitors. We did not award prizes for the open division.",
			"The in-person competition had 27 teams while the open division had 31 teams.",
		},
		Links: []templates.Link{
			{URL: "/static/2024-solutions.pdf", Title: "Solution Sketch Slides"},
			{URL: "https://sumnerevans.com/posts/school/2024-hspc/", Title: "Competition Recap and Solution Sketches"},
			{URL: "https://mines-hspc.kattis.com/contests/mines-hspc24/problems", Title: "Problems"},
		},
		Results: []templates.CompetitionResult{
			{
				Name:      "Overall Winners",
				Shortname: "Overall",
				Teams: []templates.WinningTeam{
					{Place: "1st", Name: "Innovation Center 1", School: "Innovation Center SVVSD", Location: "Longmont"},
					{Place: "2nd", Name: "Sigma Scripters", School: "Arapahoe High School", Location: "Centennial"},
					{Place: "3nd", Name: "CyberRebels2", School: "Columbine High School", Location: "Littleton"},
				},
			},
			{
				Name:      "First-Time Team Winners",
				Shortname: "FirstTime",
				Teams: []templates.WinningTeam{
					{Place: "1st", Name: "Loopy Groupies", School: "Chatfield Senior High School", Location: "Littleton"},
					{Place: "2nd", Name: "Lorem Ipsum", School: "Warren Tech", Location: "Lakewood"},
					{Place: "3nd", Name: "the cows(mooooooooooooo)", School: "Cherry Creek High School", Location: "Greenwood Village"},
				},
			},
		},
	},
	{
		Year: 2023,
		RecapParagraphs: []string{
			"The 2023 competition again featured two divisions: beginner and advanced. As with 2022, it was a hybrid competition, but we awarded prizes for both in-person and remote winners in both divisions.",
			"The advanced division featured 31 teams, while the beginner division had 34 teams.",
		},
		Links: []templates.Link{
			{URL: "/static/2023-solutions.pdf", Title: "Solution Sketch Slides"},
			{URL: "https://sumnerevans.com/posts/school/2023-hspc/", Title: "Competition Recap and Solution Sketches"},
			{URL: "https://mines23advanced.kattis.com/problems", Title: "Advanced Problems"},
			{URL: "https://mines23beginner.kattis.com/problems", Title: "Beginner Problems"},
		},
		Results: []templates.CompetitionResult{
			{
				Name:      "Advanced In-Person",
				Shortname: "AdvancedInPerson",
				Teams: []templates.WinningTeam{
					{Place: "1st", Name: "Code Rats", School: "Futures Lab", Location: "Fort Collins, Colorado"},
					{Place: "2nd", Name: "The Spanish Inquisition", School: "Regis Jesuit High School", Location: "Aurora, Colorado"},
					{Place: "3nd", Name: "CA is 202", School: "Colorado Academy", Location: "Denver, Colorado"},
				},
			},
			{
				Name:      "Beginner In-Person",
				Shortname: "BeginnerInPerson",
				Teams: []templates.WinningTeam{
					{Place: "1st", Name: "Spaghetti Code and Meatballs", School: "Warren Tech", Location: "Lakewood, Colorado"},
					{Place: "2nd", Name: "Innovation Center 1", School: "Innovation Center SVVSD", Location: "Longmont, Colorado"},
					{Place: "3nd", Name: "Team LuLo", School: "Colorado Academy", Location: "Denver, Colorado"},
				},
			},
			{
				Name:      "Advanced Remote",
				Shortname: "AdvancedRemote",
				Teams: []templates.WinningTeam{
					{Place: "1st", Name: "River Hill Team #1", School: "River Hill High School", Location: "Clarksville, Maryland"},
					{Place: "2nd", Name: "CreekCyberBruins", School: "Cherry Creek High School", Location: "Greenwood Village, Colorado"},
					{Place: "3nd", Name: "JMS", School: "Bergen County Academies", Location: "Bergen County, New Jersey"},
				},
			},
			{
				Name:      "Beginner Remote",
				Shortname: "BeginnerRemote",
				Teams: []templates.WinningTeam{
					{Place: "1st", Name: "Wormhole", School: "Voice of Calling NPO", Location: "Northridge, California"},
					{Place: "2nd", Name: "Lineup", School: "Voice of Calling NPO", Location: "Northridge, California"},
					{Place: "3nd", Name: "River Hill Team #2", School: "River Hill High School", Location: "Clarksville, Maryland"},
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
		Links: []templates.Link{
			{URL: "https://sumnerevans.com/posts/school/2022-hspc/", Title: "Competition Recap and Solution Sketches"},
			{URL: "https://mines22advanced.kattis.com/problems", Title: "Advanced Problems"},
			{URL: "https://mines22beginner.kattis.com/problems", Title: "Beginner Problems"},
		},
		Results: []templates.CompetitionResult{
			{
				Name: "Advanced",
				Teams: []templates.WinningTeam{
					{Place: "1st", Name: "Pen A Team", School: "PEN Academy", Location: "Cresskill, New Jersey"},
					{Place: "2nd", Name: "Cherry Creek Cobras", School: "Cherry Creek High School", Location: "Greenwood Village, Colorado"},
					{Place: "3nd", Name: "River Hill Team 1", School: "River Hill High School", Location: "Clarksville, Maryland"},
					{Place: "4th", Name: "The Spanish Inquisition", School: "Regis Jesuit High School", Location: "Aurora, Colorado"},
				},
			},
			{
				Name: "Beginner",
				Teams: []templates.WinningTeam{
					{Place: "1st", Name: "LLL", School: "Future Forward at Bollman", Location: "Thornton, Colorado"},
					{Place: "2nd", Name: "Error 404: Name not found", School: "Colorado Academy", Location: "Denver, Colorado"},
					{Place: "3nd", Name: "Liberty 1", School: "Liberty Common School", Location: "Fort Collins, Colorado"},
					{Place: "4th", Name: "Cool Cats", School: "Arvada West High School", Location: "Arvada, Colorado"},
				},
			},
		},
	},
	{
		Year: 2021,
		RecapParagraphs: []string{
			"The 2021 competition was an all-remote competition featuring 55 teams from across the nation.",
		},
		Links: []templates.Link{
			{URL: "https://sumnerevans.com/posts/school/2021-hspc/", Title: "Competition Recap and Solution Sketches"},
			{URL: "https://mines21.kattis.com/problems", Title: "Problems"},
		},
		Results: []templates.CompetitionResult{
			{
				Teams: []templates.WinningTeam{
					{Place: "1st", Name: "River Hill HS Team 1", School: "River Hill High School", Location: "Clarksville, Maryland"},
					{Place: "2nd", Name: "PEN A Team", School: "PEN Academy", Location: "Cresskill, New Jersey"},
					{Place: "3nd", Name: "River Hill HS Team 2", School: "River Hill High School", Location: "Clarksville, Maryland"},
				},
			},
		},
	},
	{
		Year: 2020,
		RecapParagraphs: []string{
			"Due to COVID, the 2020 competition was the first all-remote HSPC competition. The competition featured 30 teams.",
		},
		Links: []templates.Link{
			{URL: "https://sumnerevans.com/posts/school/2020-hspc/", Title: "Competition Recap and Solution Sketches"},
			{URL: "https://mines20.kattis.com/problems", Title: "Problems"},
		},
		Results: []templates.CompetitionResult{
			{
				Teams: []templates.WinningTeam{
					{Place: "1st", Name: "Installation Wizards", School: "STEM School Highlands Ranch", Location: "Highlands Ranch, Colorado"},
					{Place: "2nd", Name: "i", School: "STEM School Highlands Ranch", Location: "Highlands Ranch, Colorado"},
					{Place: "3nd", Name: "Sun Devils", School: "Kent Denver", Location: "Denver, Colorado"},
				},
			},
		},
	},
	{
		Year: 2019,
		RecapParagraphs: []string{
			"The second ever CS@Mines High School Programming Competition featured 22 teams from all around Colorado and from as far as Steamboat Springs.",
		},
		Links: []templates.Link{
			{URL: "https://sumnerevans.com/posts/school/2019-hspc/", Title: "Competition Recap and Solution Sketches"},
			{URL: "https://mines19.kattis.com/problems", Title: "Problems"},
		},
		Results: []templates.CompetitionResult{
			{
				Teams: []templates.WinningTeam{
					{Place: "1st", Name: "STEM Team 1", School: "STEM School Highlands Ranch", Location: "Highlands Ranch, Colorado"},
					{Place: "2nd", Name: "IntrospectionExceptions", School: "Colorado Academy", Location: "Lakewood, Colorado"},
					{Place: "3nd", Name: "Team 2", School: "?", Location: "?"},
				},
			},
		},
	},
	{
		Year: 2018,
		RecapParagraphs: []string{
			"The first ever CS@Mines High School Programming Competition featured 22 teams.",
		},
		Links: []templates.Link{
			{URL: "https://mines18.kattis.com/problems", Title: "Problems"},
		},
		Results: []templates.CompetitionResult{
			{
				Teams: []templates.WinningTeam{
					{Place: "1st", Name: "The Crummies", School: "Warren Tech", Location: "Arvada, Colorado"},
					{Place: "2nd", Name: "The Bean Beans", School: "Colorado Academy", Location: "Lakewood, Colorado"},
					{Place: "3nd", Name: "Warriors", School: "Arapahoe High School", Location: "Centennial, Colorado"},
				},
			},
		},
	},
}
