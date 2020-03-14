package main

import (
	"bytes"
	"log"
	"text/template"

	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type worldRankingTeams struct {
	Title string
	Teams []team
}

type team struct {
	Name, Points, Change, URL string
}

func worldRanking(chatID int64) {
	res, err := request(baseURL + "/ranking/teams")
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

	ranking := worldRankingTeams{Title: doc.Find(".regional-ranking-header").Text()}

	doc.Find(".ranked-team").Each(func(i int, s *goquery.Selection) {
		teamTmp := team{
			Name:   s.Find(".name").Text(),
			Points: s.Find(".points").Text()[1 : len(s.Find(".points").Text())-1],
			URL:    baseURL + s.Find(".details").AttrOr("href", "/ranking/teams"),
		}

		if s.Find(".change").Text() != "-" {
			teamTmp.Change = s.Find(".change").Text()
		}

		ranking.Teams = append(ranking.Teams, teamTmp)
	})

	tpl, err := template.New("").Funcs(template.FuncMap{
		"add": func(n, i int) int { return n + i },
	}).Parse(worldRankingTpl)
	if err != nil {
		log.Println(err)
		return
	}

	var tplBuf bytes.Buffer

	err = tpl.Execute(&tplBuf, ranking)
	if err != nil {
		log.Println(err)
		return
	}

	msg := tgbotapi.NewMessage(chatID, tplBuf.String())
	msg.DisableWebPagePreview = true
	msg.ParseMode = "markdown"
	bot.Send(msg)
}

var worldRankingTpl = `
*{{ .Title }}*:
{{- range $i, $item := .Teams }}
{{ add $i 1 }}. [{{ $item.Name }}]({{ $item.URL }}), {{ $item.Points }}{{ if $item.Change }}, {{ $item.Change }}{{ end -}}
{{ end }}
`
