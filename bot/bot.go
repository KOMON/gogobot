package bot

import (
	"fmt"
	"log"
	"os"
	"regexp"
	
	"github.com/komon/gogobot/config"
	"github.com/komon/gogobot/handler"
	"github.com/nlopes/slack"
)

func Run() int {
	f, err := os.OpenFile("jojolog", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		fmt.Printf("error opening logfile: %v", err)
		return 1
	}

	defer f.Close()

	api := slack.New(slackToken)
	logger := log.New(f, "jojobot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)
	rtm := api.NewRTM()

	router := &handler.Router{
		Handlers: populateHandlers(),
	}
	go rtm.ManageConnection()

	for {
		msg := <-rtm.IncomingEvents
		if msg.Type == "message" {
			err := handleMessage(msg.Data.(*slack.MessageEvent), router, rtm)
			if err != nil {
				logger.Printf("message handle error: %s", err.Error())
				if err.Error() == "shutdown" {
					break
				}
			}
		}
	}

	return 0
}

func handleMessage(msg *slack.MessageEvent, router *handler.Router, rtm *slack.RTM) error {
	h, err := router.Route(msg.Text)

	if err != nil {
		return err
	}

	if h != nil {

		response, err := (*h).Respond()
		if err != nil {
			return err
		}

		rtm.SendMessage(rtm.NewOutgoingMessage(response, msg.Channel))
	}

	return nil
}

func populateHandlers() []handler.Handler {
	var handlers []handler.Handler
	rs := config.PopulateResponders()

	for _, r := range rs.Responders {
		re := regexp.MustCompile(r.Regexp)

		new := &handler.Responder{
			RE:        re,
			Responses: r.Responses,
			Matches:   []string{},
		}
		handlers = append(handlers, new)
	}

	handlers = append(handlers, &handler.MtgSearchResponder{
		// [[card name|set]]
		RE: regexp.MustCompile("(?:^[^#]?\\[\\[(.*?)(\\|...)?\\]\\])+"),
	}, &handler.MtgStatsResponder{
		// #[[key:value, key:value]]
		RE: regexp.MustCompile("#\\[\\[((?:\\w+:\\s*\\w+,?\\s*)+)\\]\\]"),
	})

	return handlers
}
