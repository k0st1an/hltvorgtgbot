package main

import (
	"flag"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	baseURL = "https://www.hltv.org"
)

var (
	bot *tgbotapi.BotAPI
)

func main() {
	pathToConf := flag.String("conf", "./hltvorgtgbot.yml", "path to config")
	flag.Parse()
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)

	readConf(*pathToConf)

	var err error

	bot, err = tgbotapi.NewBotAPI(conf.Token)
	if err != nil {
		log.Fatalln(err)
	}

	bot.Debug = conf.Debug
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil {
			switch strings.Split(update.CallbackQuery.Data, "_")[0] {
			case "teamprofile":
				go teamProfile(update.CallbackQuery.Message.Chat.ID, strings.Split(update.CallbackQuery.Data, "_")[1])
			}
		}

		if update.Message != nil { // ignore any non-Message updates
			switch msg := update.Message.Command(); msg {
			case "worldrank":
				go worldRanking(update.Message.Chat.ID)
			case "teamprofile":
				go teamProfileButton(update.Message.Chat.ID)
			case "statsplayers":
				go statsPlayers(update.Message.Chat.ID)
			case "matches":
				go matches(update.Message.Chat.ID)
			}
		}
	}
}
