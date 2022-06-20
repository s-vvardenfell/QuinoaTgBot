package bot

import (
	"log"
	"strings"

	"github.com/s-vvardenfell/QuinoaTgBot/client"
	"github.com/s-vvardenfell/QuinoaTgBot/conditions"
	"github.com/s-vvardenfell/QuinoaTgBot/config"
	"github.com/sirupsen/logrus"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type QuinoaTgBot struct {
	tg     *tgbotapi.BotAPI
	client *client.Client
	cnfg   config.Config
}

func New(cnfg config.Config) *QuinoaTgBot {
	bot, err := tgbotapi.NewBotAPI(cnfg.Token)
	if err != nil {
		logrus.Fatalf("cannot create NewBotAPI, %v", err)
	}

	bot.Debug = cnfg.Debug
	return &QuinoaTgBot{
		tg:     bot,
		client: client.New(),
		cnfg:   cnfg,
	}
}

func (b *QuinoaTgBot) Work() {
	b.commandsHandling()
}

func (b *QuinoaTgBot) answerMsg() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.tg.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID

			b.tg.Send(msg)
		}
	}
}

func (b *QuinoaTgBot) commandsHandling() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.tg.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if !update.Message.IsCommand() {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		switch update.Message.Command() {
		case "help":
			msg.Text = "I understand /status, /search, /cancel (breaking search process)"
		case "status":
			msg.Text = "I'm ok."
		case "search":
			msg.Text = b.processSearchCommand(updates, update)
		default:
			msg.Text = "I don't know that command"
		}

		if _, err := b.tg.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}

func (b *QuinoaTgBot) processSearchCommand(
	updates tgbotapi.UpdatesChannel, update tgbotapi.Update) string {

	cnd := conditions.Conditions{}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Film or series?")
	msg.ParseMode = tgbotapi.ModeMarkdown
	if _, err := b.tg.Send(msg); err != nil {
		logrus.Fatal(err)
	}

	for update := range updates {

		switch update.Message.Command() {
		case "cancel":
			return "Ok, breaking search"
		}

		if cnd.Type == "" {
			cnd.Type = update.Message.Text
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.Text = "List the genres separated by a space:"
			if _, err := b.tg.Send(msg); err != nil {
				logrus.Fatal(err)
			}
			continue
		} else if cnd.Genres == nil {
			cnd.Genres = strings.Fields(update.Message.Text)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.Text = `Specify the year of issue "from":`
			if _, err := b.tg.Send(msg); err != nil {
				logrus.Fatal(err)
			}
			continue
		} else if cnd.StartYear == "" {
			cnd.StartYear = update.Message.Text
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.Text = `Specify the year of issue "before":`
			if _, err := b.tg.Send(msg); err != nil {
				logrus.Fatal(err)
			}
			continue
		} else if cnd.EndYear == "" {
			cnd.EndYear = update.Message.Text
			break
		} else {
			break
		}
	}

	return b.client.FilmsByConditions(cnd)
}
