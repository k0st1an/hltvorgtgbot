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

	if doc.Find(".match-day").Length() > 0 || doc.Find(".live-match a").Length() > 0 {
		text = fmt.Sprintf("*Upcoming CS:GO matches* (%d days):\n", limitDays)
	}

	// Live matches
	if doc.Find(".live-match a").Length() > 0 {
		text = text + "- live:\n"
	}

	doc.Find(".live-match a").Each(func(i int, live *goquery.Selection) {
		text = text + fmt.Sprintf("  - %s, ", live.Find(".event-name").Text())

		text = text + "["
		live.Find(".team-name").Each(func(i int, layer2 *goquery.Selection) {
			text = text + layer2.Text()
			if i == 0 {
				text = text + " vs "
			}
		})
		text = text + fmt.Sprintf("](%s)", baseURL+live.AttrOr("href", "/matches"))
		text = text + fmt.Sprintf(", %s", live.Find(".bestof").Text())

		if live.Find(".star").Length() > 0 {
			text = text + ", "
			for i := 0; i < live.Find(".star").Length(); i++ {
				text = text + "+"
			}
		}

		text = text + "\n"
	})

	// Matches
	doc.Find(".match-day").Each(func(i int, s *goquery.Selection) {
		if i > limitDays-1 {
			return
		}

		text = text + fmt.Sprintf("- %s:\n", s.Find(".standard-headline").Text())
		s.Find(".match").Each(func(i int, s *goquery.Selection) {

			text = text + fmt.Sprintf("  - %s,", s.Find(".time .time").Text())

			if s.Find(".team").Length() == 2 {
				// Match
				text = text + "["
				s.Find(".team").Each(func(i int, s *goquery.Selection) {
					text = text + fmt.Sprintf(" %s", s.Text())
					if i == 0 {
						text = text + " vs"
					}
				})
				text = text + fmt.Sprintf("](%s)", baseURL+s.Find("a").AttrOr("href", "/matches"))

				text = text + fmt.Sprintf(", %s", s.Find(".event").Text())
				text = text + fmt.Sprintf(", %s", s.Find(".map-text").Text())

				if s.Find(".star").Length() > 0 {
					text = text + ", "
					for i := 0; i < s.Find(".star").Length(); i++ {
						text = text + "+"
					}
				}
			} else {
				// Event
				text = text + fmt.Sprintf(" %s, ", s.Find(".placeholder-text-cell").Text())
			}

			text = text + "\n"
		})
	})

	msg := tgbotapi.NewMessage(chatID, text)
	msg.DisableWebPagePreview = true
	msg.ParseMode = "markdown"

	_, err = bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}
