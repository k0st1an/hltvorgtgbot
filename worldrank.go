package main

import (
	"fmt"
	"log"

	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func worldRank(chatID int64) {
	res, err := request(baseURL + "/ranking/teams")
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatalln(err)
	}

	text := fmt.Sprintf("*%s* (team, points, change):\n", doc.Find(".regional-ranking-header").Text())

	doc.Find(".ranked-team").Each(func(i int, s *goquery.Selection) {
		text = text + fmt.Sprintf("%d. ", i+1)

		if v, ok := s.Find(".details").Attr("href"); ok {
			text = text + fmt.Sprintf(" [%s](%s)", s.Find(".name").Text(), baseURL+v)
		} else {
			text = text + fmt.Sprintf(" %s", s.Find(".name").Text())
		}

		text = text + fmt.Sprintf(", %s", s.Find(".points").Text()[1:len(s.Find(".points").Text())-1])

		if s.Find(".change").Text() != "-" {
			text = text + fmt.Sprintf(", %s\n", s.Find(".change").Text())
		} else {
			text = text + "\n"
		}
	})

	msg := tgbotapi.NewMessage(chatID, text)
	msg.DisableWebPagePreview = true
	msg.ParseMode = "markdown"
	bot.Send(msg)
}
