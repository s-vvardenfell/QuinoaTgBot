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
	client *client.QuinoaTgBotClient
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
		client: client.New(cnfg.ServerHost, cnfg.ServerPort),
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
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			switch update.Message.Command() {
			case "skip":
				cnd.Type = "no"
				msg.Text = "Ok, skipping this question"
				if _, err := b.tg.Send(msg); err != nil {
					logrus.Error(sendMsgErr, err)
				}
			default:
				cnd.Type = update.Message.Text
			}

			msg.Text = "List the genres separated by a space:"
			if _, err := b.tg.Send(msg); err != nil {
				logrus.Error(sendMsgErr, err)
			}

			continue

		} else if cnd.Genres == nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			switch update.Message.Command() {
			case "skip":
				cnd.Genres = []string{"no"}
				msg.Text = "Ok, skipping this question"
				if _, err := b.tg.Send(msg); err != nil {
					logrus.Error(sendMsgErr, err)
				}
			default:
				cnd.Genres = strings.Fields(update.Message.Text)
			}

			msg.Text = `Specify the year of issue "from" (since 1900):`
			if _, err := b.tg.Send(msg); err != nil {
				logrus.Error(sendMsgErr, err)
			}
			continue

		} else if cnd.StartYear == "" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			switch update.Message.Command() {
			case "skip":
				cnd.StartYear = "no"
				msg.Text = "Ok, skipping this question"
				if _, err := b.tg.Send(msg); err != nil {
					logrus.Error(sendMsgErr, err)
				}
			default:
				if !checkYear(update.Message.Text) {
					msg.Text = `Wrong year format or value, please, try again`
					if _, err := b.tg.Send(msg); err != nil {
						logrus.Error(sendMsgErr, err)
					}
					continue
				}
				cnd.StartYear = update.Message.Text
			}

			msg.Text = fmt.Sprintf(`Specify the year of issue "before" (up to %d):`, time.Now().Year())
			if _, err := b.tg.Send(msg); err != nil {
				logrus.Error(sendMsgErr, err)
			}
			continue

		} else if cnd.EndYear == "" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			switch update.Message.Command() {
			case "skip":
				cnd.EndYear = "no"
				msg.Text = "Ok, skipping this question"
				if _, err := b.tg.Send(msg); err != nil {
					logrus.Error(sendMsgErr, err)
				}
			default:
				if !checkYear(update.Message.Text) {
					msg.Text = `Wrong year format or value, please, try again`
					if _, err := b.tg.Send(msg); err != nil {
						logrus.Error(sendMsgErr, err)
					}
					continue
				}

				cnd.EndYear = update.Message.Text
			}

			msg.Text = "Specify keywords (like film name etc, separated by space):"
			if _, err := b.tg.Send(msg); err != nil {
				logrus.Error(sendMsgErr, err)
			}
			continue

		} else if cnd.Keyword == "" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			switch update.Message.Command() {
			case "skip":
				cnd.Keyword = "no"
				msg.Text = "Ok, skipping this question"
				if _, err := b.tg.Send(msg); err != nil {
					logrus.Error(sendMsgErr, err)
				}
			default:
				cnd.Keyword = update.Message.Text
			}

			msg.Text = "List the countries separated by a space:"
			if _, err := b.tg.Send(msg); err != nil {
				logrus.Error(sendMsgErr, err)
			}
			continue

		} else if cnd.Countries == nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			switch update.Message.Command() {
			case "skip":
				cnd.Countries = []string{"no"}
				msg.Text = "Ok, skipping this question"
				if _, err := b.tg.Send(msg); err != nil {
					logrus.Error(sendMsgErr, err)
				}
			default:
				cnd.Countries = strings.Fields(update.Message.Text)
			}
			break

		} else {
			break
		}
	}

	return b.client.FilmsByConditions(cnd)
}

func checkYear(y string) bool {
	d, err := strconv.Atoi(y)
	if err != nil {
		return false
	}

	if d >= 1900 && d <= time.Now().Year()+1 {
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
