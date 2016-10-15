package main

import (
	//	"fmt"
	"log"
	"os"

	"github.com/nlopes/slack"
)

func main() {
	handlers := populateResponders()
	handlers = populateComplexResponders(handlers)
	api := slack.New(TOKEN)
	//api.SetDebug(true)
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for {
		msg := <-rtm.IncomingEvents
		if msg.Type == "message" {
			handleMessage(msg.Data.(*slack.MessageEvent), handlers, rtm)
		}
	}
}

func handleMessage(msg *slack.MessageEvent, handlers []Handler, rtm *slack.RTM) {
	for _, handler := range handlers {
		if handler.Match(msg.Text) {
			go handler.Respond(rtm, msg.Channel)
		}
	}
}

func populateComplexResponders(handlers []Handler) []Handler {
	handlers = append(handlers, &MtgSearchResponder{
		re: MTGSEARCH_REGEXP,
	}, &MtgStatsResponder{
		re: MTGSTATS_REGEXP,
	})
	return handlers
}
