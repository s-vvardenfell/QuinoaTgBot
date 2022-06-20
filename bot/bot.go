package bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/s-vvardenfell/QuinoaTgBot/client"
	"github.com/s-vvardenfell/QuinoaTgBot/conditions"
	"github.com/s-vvardenfell/QuinoaTgBot/config"
	"github.com/sirupsen/logrus"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	sendMsgErr = "error while sending msg to client: "
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
			logrus.Error(sendMsgErr, err)
		}
	}
}

func (b *QuinoaTgBot) processSearchCommand(
	updates tgbotapi.UpdatesChannel, update tgbotapi.Update) string {

	cnd := conditions.Conditions{}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Film or series?")
	msg.ParseMode = tgbotapi.ModeMarkdown
	if _, err := b.tg.Send(msg); err != nil {
		logrus.Error(sendMsgErr, err)
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
				logrus.Error(sendMsgErr, err)
			}
			continue
		} else if cnd.Genres == nil {
			cnd.Genres = strings.Fields(update.Message.Text)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.Text = `Specify the year of issue "from" (since 1900):`
			if _, err := b.tg.Send(msg); err != nil {
				logrus.Error(sendMsgErr, err)
			}
			continue
		} else if cnd.StartYear == "" {
			if !CheckYear(update.Message.Text) {
				msg.Text = `Wrong year format or value, please, try again`
				if _, err := b.tg.Send(msg); err != nil {
					logrus.Error(sendMsgErr, err)
				}
				continue
			}

			cnd.StartYear = update.Message.Text
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.Text = fmt.Sprintf(`Specify the year of issue "before" (up to %d):`, time.Now().Year())
			if _, err := b.tg.Send(msg); err != nil {
				logrus.Error(sendMsgErr, err)
			}
			continue
		} else if cnd.EndYear == "" {
			if !CheckYear(update.Message.Text) {
				msg.Text = `Wrong year format or value, please, try again`
				if _, err := b.tg.Send(msg); err != nil {
					logrus.Error(sendMsgErr, err)
				}
				continue
			}

			cnd.EndYear = update.Message.Text
			break
		} else {
			break
		}
	}

	return b.client.FilmsByConditions(cnd)
}

func CheckYear(y string) bool {
	d, err := strconv.Atoi(y)
	if err != nil {
		return false
	}

	if d >= 1900 && d <= time.Now().Year() {
		for _, d := range y {
			if !unicode.IsDigit(d) {
				return false
			}
		}
	} else {
		return false
	}
	return true
}
