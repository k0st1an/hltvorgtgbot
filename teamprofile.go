package main

import (
	"fmt"
	"log"
	"strings"

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

	// Comman info
	text := fmt.Sprintf("*Team name:* %s\n", doc.Find(".profile-team-name").Text())
	text = text + fmt.Sprintf("*Country:* %s\n", doc.Find(".team-country").Text())

	doc.Find(".profile-team-stat").Each(func(i int, s *goquery.Selection) {
		text = text + fmt.Sprintf("*%s*: %s\n", s.Find("b").Text(), strings.Replace(s.Find("span").Text(), "#", "", 1))
	})

	// Players
	text = text + "*Players:* "
	doc.Find(".bodyshot-team").Find("a").Each(func(i int, s *goquery.Selection) {
		text = text + fmt.Sprintf("[%s](%s%s), ", s.AttrOr("title", ""), baseURL, s.AttrOr("href", ""))
	})
	text = text[0:len(text)-2] + "\n" // delete last comma

	// Links
	text = text + "*Links:*\n"
	text = text + fmt.Sprintf("- %s\n", u)
	doc.Find(".profile-team-some").Find("a").Each(func(i int, s *goquery.Selection) {
		if link, ok := s.Attr("href"); ok {
			text = text + fmt.Sprintf("- %s\n", link)
		}
	})

	// Trophies
	if doc.Find(".trophyHolder").Length() > 0 {
		text = text + fmt.Sprintf("*Trophies (%d):*\n", doc.Find(".trophyHolder").Length())

		doc.Find(".trophyHolder").Each(func(i int, s *goquery.Selection) {
			if title, ok := s.Find("span").Attr("title"); ok {
				text = text + fmt.Sprintf("- %s\n", strings.ReplaceAll(title, "_", "\\_"))
			}
		})
	}

	// Events
	if doc.Find("#ongoingEvents").Length() > 0 {
		text = text + "*Ongoing & upcoming events and leagues:*\n"
		doc.Find("#ongoingEvents").Find(".ongoing-event").Each(func(i int, s *goquery.Selection) {
			text = text + fmt.Sprintf("- %s, %s\n", s.Find(".eventbox-eventname").Text(), s.Find(".eventbox-date").Text())
		})
	}

	// Events and matches
	matchesBool := true
	doc.Find("#matchesBox .match-table").Each(func(i int, layer1 *goquery.Selection) {
		var tmp string

		eventName := true
		layer1.Find(".team-row").Each(func(i int, layer2 *goquery.Selection) { // Event matches
			if layer2.Find(".score-cell").Text() != "-:-" { // checking what will be
				return
			}

			if matchesBool {
				matchesBool = false
				tmp = tmp + "*Upcoming matches:*\n"
			}

			if eventName {
				eventName = false
				tmp = tmp + fmt.Sprintf("- %s\n", layer1.Find(".text-ellipsis").Text()) // event name
			}

			// Matches
			tmp = tmp + fmt.Sprintf("  - %s, [", layer2.Find(".date-cell span").Text())
			tmp = tmp + fmt.Sprintf("%s vs %s](%s)\n", layer2.Find(".team-center-cell .team-1 span").Text(), layer2.Find(".team-center-cell .team-2 span").Text(), baseURL+layer2.Find("a").AttrOr("href", ""))
		})

		text = text + tmp
	})

	msg := tgbotapi.NewMessage(chatID, text)
	msg.DisableWebPagePreview = true
	msg.ParseMode = "markdown"
	bot.Send(msg)
}
