package main

import (
	"fmt"
	"log"

	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const limitDays = 2

func matches(chatID int64) {
	res, err := request(baseURL + "/matches")
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var text string
	matchesBool := true
	doc.Find(".match-day").Each(func(i int, s *goquery.Selection) {
		if matchesBool {
			text = fmt.Sprintf("*Upcoming CS:GO matches* (%d days):\n", limitDays)
			matchesBool = false
		}

		if i > limitDays-1 {
			return
		}

		text = text + fmt.Sprintf("- %s:\n", s.Find(".standard-headline").Text())
		s.Find(".match").Each(func(i int, s *goquery.Selection) {

			text = text + fmt.Sprintf("  - %s,", s.Find(".time .time").Text())

			if s.Find(".team").Length() == 2 {
				// Match
				s.Find(".team").Each(func(i int, s *goquery.Selection) {
					text = text + fmt.Sprintf(" %s", s.Text())
					if i == 0 {
						text = text + " vs"
					}
				})

				text = text + fmt.Sprintf(", %s", s.Find(".event").Text())
				text = text + fmt.Sprintf(", %s, ", s.Find(".map-text").Text())
				if s.Find(".star").Length() > 0 {
					for i := 0; i < s.Find(".star").Length(); i++ {
						text = text + "+"
					}
					text = text + ", "
				}
			} else {
				// Event
				text = text + fmt.Sprintf(" %s, ", s.Find(".placeholder-text-cell").Text())
			}

			text = text + fmt.Sprintf("[match page](%s)\n", baseURL+s.Find("a").AttrOr("href", "http://none"))
		})
	})

	msg := tgbotapi.NewMessage(chatID, text)
	msg.DisableWebPagePreview = true
	msg.ParseMode = "markdown"
	_, err = bot.Send(msg)
	if err != nil {
		log.Fatalln(err)
	}
}
