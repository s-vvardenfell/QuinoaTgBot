package bot

import (
	"encoding/base64"
	"errors"
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
	sendMsgErr      = "error while sending msg to client: "
	msgSkip         = "Ок, пропустим этот вопрос"
	qType           = "Фильм или сериал? (или /skip)"
	qGenres         = "Перечислите жанры через пробел (или /skip):"
	qYearsBeg       = "Укажите год начала поиска (минимум 1900) (или /skip):"
	qKeywords       = "Укажите ключевые слова через пробел (например, название фильма или имя актера) (или /skip):"
	qCountries      = "Перечислите страны через пробел (или /skip):"
	msgWrongFormat  = "Неправильный формат или значение, попробуйте снова"
	msgWrongCommand = "Я не знаю такой команды"
)

var (
	qYearsEnd = fmt.Sprintf("Укажите год окончания поиска (максимум %d) (или /skip):", time.Now().Year()+1)
	errBreak  = errors.New("ок, прекращаю поиск")
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
		client: client.New(cnfg.ServerHost, cnfg.ServerPort, cnfg.Timeout),
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

loop:
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
			msg.Text = "Я понимаю /status, /search, /cancel (прерывание поиска)"
		case "status":
			msg.Text = "Всё в порядке"
		case "cancel":
			msg.Text = "Эта команда прерывает процесс поиска фильмов командой /search"
		case "search":
			cnd, err := b.processSearchCommand(updates, update)
			if err != nil {
				msg.Text = err.Error()
				break
			}

			msg.Text = "Поиск, ожидайте..."
			b.sendMsg(msg)

			searchResults, err := b.client.FilmsByConditions(cnd)
			if err != nil {
				msg.Text = err.Error()
				break
			}

			mg := tgbotapi.MediaGroupConfig{
				ChatID: update.Message.Chat.ID,
			}

			for i := range searchResults {
				img, err := base64.StdEncoding.DecodeString(searchResults[i].Img)
				if err != nil {
					msg.Text = err.Error()
					break
				}

				file := tgbotapi.FileBytes{
					Name:  searchResults[i].Name,
					Bytes: img,
				}

				cap := fmt.Sprintf("%s\n%s", searchResults[i].Name, searchResults[i].Ref)

				media := tgbotapi.InputMediaPhoto{
					BaseInputMedia: tgbotapi.BaseInputMedia{
						Type:    "photo",
						Media:   file,
						Caption: cap,
					},
				}
				mg.Media = append(mg.Media, media)
			}

			if _, err := b.tg.SendMediaGroup(mg); err != nil {
				logrus.Error(sendMsgErr, err)
			}

			continue loop
		default:
			msg.Text = msgWrongCommand
		}

		b.sendMsg(msg)
	}
}

func (b *QuinoaTgBot) processSearchCommand(
	updates tgbotapi.UpdatesChannel, update tgbotapi.Update) (conditions.Conditions, error) {

	cnd := conditions.Conditions{}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, qType)
	msg.ParseMode = tgbotapi.ModeMarkdown
	if _, err := b.tg.Send(msg); err != nil {
		logrus.Error(sendMsgErr, err)
	}

	for update := range updates {

		switch update.Message.Command() {
		case "cancel":
			return conditions.Conditions{}, errBreak
		}

		if !cnd.Check.Type {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			switch update.Message.Command() {
			case "skip":
				cnd.Check.Type = true
				msg.Text = msgSkip
				b.sendMsg(msg)
			default:
				cnd.Type = update.Message.Text
				cnd.Check.Type = true
			}

			msg.Text = qGenres
			b.sendMsg(msg)
			continue

		} else if !cnd.Check.Genres {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			switch update.Message.Command() {
			case "skip":
				cnd.Check.Genres = true
				msg.Text = msgSkip
				b.sendMsg(msg)
			default:
				cnd.Genres = strings.Fields(update.Message.Text)
				cnd.Check.Genres = true
			}

			msg.Text = qYearsBeg
			b.sendMsg(msg)
			continue

		} else if !cnd.Check.StartYear {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			switch update.Message.Command() {
			case "skip":
				cnd.Check.StartYear = true
				msg.Text = msgSkip
				b.sendMsg(msg)
			default:
				if !checkYear(update.Message.Text) {
					msg.Text = msgWrongFormat
					b.sendMsg(msg)
					continue
				}
				cnd.StartYear = update.Message.Text
				cnd.Check.StartYear = true
			}

			msg.Text = qYearsEnd
			b.sendMsg(msg)
			continue

		} else if !cnd.Check.EndYear {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			switch update.Message.Command() {
			case "skip":
				cnd.Check.EndYear = true
				msg.Text = msgSkip
				b.sendMsg(msg)
			default:
				if !checkYear(update.Message.Text) {
					msg.Text = msgWrongFormat
					b.sendMsg(msg)
					continue
				}

				cnd.EndYear = update.Message.Text
				cnd.Check.EndYear = true
			}

			msg.Text = qKeywords
			b.sendMsg(msg)
			continue

		} else if !cnd.Check.Keyword {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			switch update.Message.Command() {
			case "skip":
				cnd.Check.Keyword = true
				msg.Text = msgSkip
				b.sendMsg(msg)
			default:
				cnd.Keyword = update.Message.Text
				cnd.Check.Keyword = true
			}

			msg.Text = qCountries
			b.sendMsg(msg)
			continue

		} else if !cnd.Check.Countries {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			switch update.Message.Command() {
			case "skip":
				cnd.Check.Countries = true
				msg.Text = msgSkip
				b.sendMsg(msg)
			default:
				cnd.Countries = strings.Fields(update.Message.Text)
				cnd.Check.Countries = true
			}
			break

		} else {
			break
		}
	}

	return cnd, nil
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

func (b *QuinoaTgBot) sendMsg(msg tgbotapi.MessageConfig) {
	if _, err := b.tg.Send(msg); err != nil {
		logrus.Error(sendMsgErr, err)
	}
}
