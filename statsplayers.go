package main

import (
	"bytes"
	"log"
	"text/template"

	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const statsPlayersTop = 30

type statsPlayersData struct {
	RatingVersion string
	Players       []ratePlayer
}

type ratePlayer struct {
	Name, Team, Maps, KDDiff, KD, Rate, URL string
}

func statsPlayers(chatID int64) {
	res, err := request(baseURL + "/stats/players")
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var rate statsPlayersData

	rate.RatingVersion = doc.Find(".ratingDesc").Text()

	doc.Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		if s.Index() < statsPlayersTop {
			rPlayer := ratePlayer{
				Name:   s.Find("td a").Text(),
				Team:   s.Find("td a img").AttrOr("title", "UNKNOWN"),
				KDDiff: s.Find(".kdDiffCol").Text(),
				Rate:   s.Find(".ratingCol").Text(),
				URL:    baseURL + s.Find("td a").AttrOr("href", baseURL),
			}

			s.Find(".statsDetail").Each(func(i int, s *goquery.Selection) {
				switch i {
				case 0:
					rPlayer.Maps = s.Text()
				case 1:
					rPlayer.KD = s.Text()
				}
			})

			rate.Players = append(rate.Players, rPlayer)
		}
	})

	tpl, err := template.New("").Funcs(template.FuncMap{
		"add": func(n, i int) int { return n + i },
	}).Parse(statsPlayerTpl)
	if err != nil {
		log.Println(err)
		return
	}

	var tplBuf bytes.Buffer

	err = tpl.Execute(&tplBuf, rate)
	if err != nil {
		log.Println(err)
		return
	}

	msg := tgbotapi.NewMessage(chatID, tplBuf.String())
	msg.DisableWebPagePreview = true
	msg.ParseMode = "markdown"
	bot.Send(msg)
}

var statsPlayerTpl = `
*Stats players* (name, team, maps, K/D Diff, K/D, rating {{ .RatingVersion }}):
{{- range $i, $item := .Players }}
{{ add $i 1 }}. [{{ $item.Name }}]({{ $item.URL }}), {{ $item.Team }}, {{ $item.Maps }}, {{ $item.KDDiff }}, {{ $item.KD }}, {{ $item.Rate -}}
{{ end -}}
`
