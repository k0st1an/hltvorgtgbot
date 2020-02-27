package main

import (
	"fmt"
	"log"

	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const statsPlayersTop = 30

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

	text := "*Stats players* (name, team, maps, K/D Diff, K/D, rating 1.0):\n"

	doc.Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		if s.Index() < statsPlayersTop {
			if link, ok := s.Find("td a").Attr("href"); ok {
				text = text + fmt.Sprintf("%d. [%s](%s): ", i+1, s.Find("td a").Text(), baseURL+link)
			} else {
				text = text + fmt.Sprintf("%d. %s: ", i+1, s.Find("td a").Text())
			}

			text = text + fmt.Sprintf("%s, ", s.Find("td a img").AttrOr("title", "UNKNOWN")) // team name

			s.Find(".statsDetail").Each(func(i int, s *goquery.Selection) {
				if i == 0 {
					text = text + fmt.Sprintf("%s, ", s.Text())
				}
			})

			text = text + fmt.Sprintf("%s, ", s.Find(".kdDiffCol").Text())

			s.Find(".statsDetail").Each(func(i int, s *goquery.Selection) {
				if i == 1 {
					text = text + fmt.Sprintf("%s, ", s.Text())
				}
			})

			text = text + fmt.Sprintf("%s\n", s.Find(".ratingCol").Text())
		}
	})

	msg := tgbotapi.NewMessage(chatID, text)
	msg.DisableWebPagePreview = true
	msg.ParseMode = "markdown"
	bot.Send(msg)
}
