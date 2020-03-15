package main

import (
	"bytes"
	"log"
	"strings"
	"text/template"

	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func teamProfileButton(chatID int64) {
	res, err := request(baseURL + "/ranking/teams")
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var buttonsData [][]string

	doc.Find(".ranked-team").Each(func(i int, s *goquery.Selection) {
		v, _ := s.Find(".moreLink").Attr("href")
		buttonsData = append(buttonsData, []string{s.Find(".name").Text(), "teamprofile_" + baseURL + v})
	})

	msg := tgbotapi.NewMessage(chatID, "Teams:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(makeColumn(3, buttonsData)...)
	bot.Send(msg)
}

func makeColumn(nColums int, data [][]string) (rows [][]tgbotapi.InlineKeyboardButton) {
	var index int
	var buttons []tgbotapi.InlineKeyboardButton

	for _, item := range data {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(item[0], item[1]))
		index++

		if index == nColums {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(buttons...))
			index = 0
			buttons = []tgbotapi.InlineKeyboardButton{}
		}
	}
	return
}

type teamProfileData struct {
	Name, Country, WorldRanking, WeekInTop30, AveragePlayerAge string
	Players                                                    []player
	Links                                                      []string
	Trophies                                                   []string
	UpcomingEvents                                             []upcomingEvents
	UpcomingMatches                                            []upcomingMatch
}

type upcomingEvents struct {
	Name, Date string
}

type player struct {
	Name, URL string
}

type upcomingMatch struct {
	Event   string
	Matches []match
}

type match struct {
	Date, Team1, Team2, URL string
}

func teamProfile(chatID int64, u string) {
	res, err := request(u)
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatalln(err)
	}

	tp := teamProfileData{
		Name:    doc.Find(".profile-team-name").Text(),
		Country: strings.Trim(doc.Find(".team-country").Text(), " "),
	}

	// Comman info
	doc.Find(".profile-team-stat").Each(func(i int, s *goquery.Selection) {
		switch i {
		case 0:
			tp.WorldRanking = strings.Replace(s.Find("span").Text(), "#", "", 1)
		case 1:
			tp.WeekInTop30 = strings.Replace(s.Find("span").Text(), "#", "", 1)
		case 2:
			tp.AveragePlayerAge = strings.Replace(s.Find("span").Text(), "#", "", 1)
		}
	})

	// Players
	doc.Find(".bodyshot-team").Find("a").Each(func(i int, s *goquery.Selection) {
		tp.Players = append(tp.Players, player{
			Name: s.AttrOr("title", "UNKNOWN"),
			URL:  baseURL + s.AttrOr("href", u),
		})
	})

	// Links
	tp.Links = append(tp.Links, u)
	doc.Find(".profile-team-some").Find("a").Each(func(i int, s *goquery.Selection) {
		if link, ok := s.Attr("href"); ok {
			tp.Links = append(tp.Links, link)
		}
	})

	// Trophies
	if doc.Find(".trophyHolder").Length() > 0 {
		doc.Find(".trophyHolder").Each(func(i int, s *goquery.Selection) {
			if title, ok := s.Find("span").Attr("title"); ok {
				tp.Trophies = append(tp.Trophies, strings.ReplaceAll(title, "_", "\\_"))
			}
		})
	}

	// Events
	if doc.Find("#ongoingEvents").Length() > 0 {
		doc.Find("#ongoingEvents").Find(".ongoing-event").Each(func(i int, s *goquery.Selection) {
			tp.UpcomingEvents = append(tp.UpcomingEvents, upcomingEvents{
				Name: s.Find(".eventbox-eventname").Text(),
				Date: s.Find(".eventbox-date").Text()},
			)
		})
	}

	// Events and matches
	doc.Find("#matchesBox .match-table").Each(func(i int, l1 *goquery.Selection) {
		var uMatch upcomingMatch

		l1.Find(".team-row").Each(func(i int, l2 *goquery.Selection) { // Event matches
			if l2.Find(".score-cell").Text() != "-:-" { // checking what will be
				return
			}

			if len(uMatch.Event) == 0 {
				uMatch.Event = l1.Find(".text-ellipsis").Text()
			}

			// Matches
			uMatch.Matches = append(uMatch.Matches, match{
				Date:  l2.Find(".date-cell span").Text(),
				Team1: l2.Find(".team-center-cell .team-1 span").Text(),
				Team2: l2.Find(".team-center-cell .team-2 span").Text(),
				URL:   baseURL + l2.Find("a").AttrOr("href", ""),
			})
		})

		if len(uMatch.Event) > 0 {
			tp.UpcomingMatches = append(tp.UpcomingMatches, uMatch)
		}
	})

	tpl, err := template.New("").Parse(teamProfileTpl)
	if err != nil {
		log.Println(err)
		return
	}

	var tplBuf bytes.Buffer
	tpl.Execute(&tplBuf, tp)
	if err != nil {
		log.Println(err)
		return
	}

	msg := tgbotapi.NewMessage(chatID, tplBuf.String())
	msg.DisableWebPagePreview = true
	msg.ParseMode = "markdown"
	bot.Send(msg)
}

var teamProfileTpl = `
*Team name:* {{ .Name }}
*Country:* {{ .Country }}
*World ranking:* {{ .WorldRanking }}
*Weeks in top30 for core:* {{ .WeekInTop30 }}
*Average player age:* {{ .AveragePlayerAge }}
*Players:*
{{- range .Players }}
- [{{ .Name }}]({{ .URL }})
{{- end }}
*Links:*
{{- range .Links }}
- {{ . -}}
{{ end }}
*Trophies:*
{{- range .Trophies }}
- {{ . -}}
{{ end }}
*Ongoing & upcoming events and leagues:*
{{- range .UpcomingEvents }}
- {{ .Name }}, {{ .Date -}}
{{ end }}
{{ if .UpcomingMatches -}}
*Upcoming matches:*
{{- range .UpcomingMatches }}
- {{ .Event -}}:
		{{- range .Matches }}
  - {{ .Date }}, [{{ .Team1 }} vs {{ .Team2 }}]({{ .URL }})
		{{- end }}
	{{- end }}
{{- end }}
`
