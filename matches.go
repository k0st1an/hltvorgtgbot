package main

import (
	"fmt"
	"log"

	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const limitDays = 3

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

			text = text + fmt.Sprintf("[match page](%s)\n", baseURL+s.Find("a").AttrOr("href", "https://www.hltv.org/matches"))
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

func liveMatches(chatID int64) {
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

	if doc.Find(".live-match a").Length() == 0 {
		text = "No live matches"
	}

	doc.Find(".live-match a").Each(func(i int, layer1 *goquery.Selection) {
		text = fmt.Sprintf("- %s, ", layer1.Find(".event-name").Text())

		layer1.Find(".team-name").Each(func(i int, layer2 *goquery.Selection) {
			text = text + layer2.Text()
			if i == 0 {
				text = text + " vs "
			}
		})

		text = text + fmt.Sprintf(", %s", layer1.Find(".bestof").Text())

		if layer1.Find(".star").Length() > 0 {
			text = text + ", "
			for i := 0; i < layer1.Find(".star").Length(); i++ {
				text = text + "+"
			}
		}

		text = text + fmt.Sprintf(", [match page](%s)\n", baseURL+layer1.Find("a").AttrOr("href", "https://www.hltv.org/matches"))
		text = text + "  - "

		// fmt.Println(">>", layer1.Find(".map").Length())

		var jump1, jump2, cycle int

		switch layer1.Find(".bestof").Text() {
		case "Best of 1":
			jump1 = 1
			jump2 = 2
			cycle = 0
		case "Best of 3":
			jump1 = 3
			jump2 = 6
			cycle = 2
		}

		layer1.Find(".map").Each(func(i int, layer2 *goquery.Selection) {
			if i > cycle {
				return
			}

			// fmt.Println(">>>", layer1.Find(".map").Text())

			t1 := goquery.NewDocumentFromNode(layer1.Find(".map").Get(i + jump1))
			t2 := goquery.NewDocumentFromNode(layer1.Find(".map").Get(i + jump2))

			text = text + fmt.Sprintf("%s: %s / %s, ", layer2.Text(), t1.Find("span").Text(), t2.Find("span").Text())
		})

		layer1.Find(".total").Each(func(i int, layer2 *goquery.Selection) {
			if i == 0 { // skip first total, thead
				return
			}
			if i == 1 {
				text = text + " maps: "
			}

			text = text + fmt.Sprintf("%s / ", layer2.Find("span").Text())
		})

		if layer1.Find(".total").Length() > 0 {
			text = text[:len(text)-3] + "\n"
		} else {
			text = text[:len(text)-2] + "\n"
		}
	})

	msg := tgbotapi.NewMessage(chatID, text)
	msg.DisableWebPagePreview = true
	msg.ParseMode = "markdown"

	_, err = bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}
