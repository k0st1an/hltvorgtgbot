package main

import (
	"bytes"
	"log"
	"text/template"

	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const limitDays = 2

type upcomingMatches struct {
	LimitDays           int
	MatchesInDateOrLive []matchesInDateOrLive
}

type matchesInDateOrLive struct {
	DateOrLive string
	IsLive     bool
	Matches    []upcomingMatch
}

type upcomingMatch struct {
	Date, Team1, Team2, Event, BestOf, Stars, URL string
}

func matches(chatID int64) {
	res, err := request(baseURL + "/matches")
	if err != nil {
		log.Println(err)
		return
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println(err)
		return
	}

	uMatches := upcomingMatches{LimitDays: limitDays}

	// Live matches
	mInDateOrLive := matchesInDateOrLive{IsLive: true}

	doc.Find(".live-match a").Each(func(i int, live *goquery.Selection) {
		uMatch := upcomingMatch{Event: live.Find(".event-name").Text()}

		live.Find(".team-name").Each(func(i int, l2 *goquery.Selection) {
			switch i {
			case 0:
				uMatch.Team1 = l2.Text()
			case 1:
				uMatch.Team2 = l2.Text()
			}
		})
		uMatch.URL = baseURL + live.AttrOr("href", "/matches")
		uMatch.BestOf = live.Find(".bestof").Text()

		if live.Find(".star").Length() > 0 {
			for i := 0; i < live.Find(".star").Length(); i++ {
				uMatch.Stars = uMatch.Stars + "+"
			}
		}
		mInDateOrLive.Matches = append(mInDateOrLive.Matches, uMatch)
	})

	uMatches.MatchesInDateOrLive = append(uMatches.MatchesInDateOrLive, mInDateOrLive)

	// Upcoming matches
	doc.Find(".match-day").Each(func(i int, s *goquery.Selection) {
		if i > limitDays-1 {
			return
		}

		mInDateOrLive = matchesInDateOrLive{DateOrLive: s.Find(".standard-headline").Text()}

		s.Find(".match").Each(func(i int, s *goquery.Selection) {
			var uMatch upcomingMatch

			if s.Find(".team").Length() == 2 {
				uMatch.Date = s.Find(".time .time").Text()
				uMatch.Event = s.Find(".event").Text()
				uMatch.BestOf = s.Find(".map-text").Text()
				uMatch.URL = baseURL + s.Find("a").AttrOr("href", "/matches")

				// Match
				s.Find(".team").Each(func(i int, s *goquery.Selection) {
					switch i {
					case 0:
						uMatch.Team1 = s.Text()
					case 1:
						uMatch.Team2 = s.Text()
					}
				})
				if s.Find(".star").Length() > 0 {
					for i := 0; i < s.Find(".star").Length(); i++ {
						uMatch.Stars = uMatch.Stars + "+"
					}
				}
			} else {
				// Event
				uMatch.Event = s.Find(".placeholder-text-cell").Text()
			}

			mInDateOrLive.Matches = append(mInDateOrLive.Matches, uMatch)
		})

		uMatches.MatchesInDateOrLive = append(uMatches.MatchesInDateOrLive, mInDateOrLive)
	})

	tpl, err := template.New("").Parse(matchesTpl)
	if err != nil {
		log.Println(err)
		return
	}

	var tplBuf bytes.Buffer

	err = tpl.Execute(&tplBuf, uMatches)
	if err != nil {
		log.Println(err)
		return
	}

	msg := tgbotapi.NewMessage(chatID, tplBuf.String())
	msg.DisableWebPagePreview = true
	msg.ParseMode = "markdown"

	_, err = bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

var matchesTpl = `
{{- range $i, $item := .MatchesInDateOrLive -}}
	{{ if $item.IsLive -}}
- Live matches:
	{{- else }}
- {{ $item.DateOrLive }}:
	{{- end }}
	{{- range $item.Matches }}
		{{- if .Team1 }}
  - {{ if not $item.IsLive }}{{ .Date }}, {{ end }}[{{ .Team1 }} vs {{ .Team2 }}]({{ .URL }}), {{ .Event }}, {{ .BestOf }}{{ if .Stars }}, {{ .Stars }}{{ end -}}
		{{ else }}
  - {{ .Event -}}
		{{ end -}}
	{{ end -}}
{{ end -}}
`
