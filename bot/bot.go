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
	sendMsgErr = "error while sending msg to client: "
	msgSkip    = "Ок, пропустим этот вопрос"
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
			msg.Text = "Я понимаю /status, /search, /cancel (прерываение процедуры поиска)"
		case "status":
			msg.Text = "Всё в порядке"
		case "cancel":
			msg.Text = "Эта команда прерывает процесс выбора фильма команды /search"
		case "search":
			cnd, err := b.processSearchCommand(updates, update)
			if err != nil {
				msg.Text = err.Error()
				break
			}

			msg.Text = "Поиск, ожидайте..."
			if _, err := b.tg.Send(msg); err != nil {
				logrus.Error(sendMsgErr, err)
			}

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
					break //точно ли?
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
			msg.Text = "Я не знаю такой команды"
		}

		if _, err := b.tg.Send(msg); err != nil {
			logrus.Error(sendMsgErr, err)
		}
	}
}

//TODO обработать возврат ошибок
func (b *QuinoaTgBot) processSearchCommand(
	updates tgbotapi.UpdatesChannel, update tgbotapi.Update) (conditions.Conditions, error) {

	cnd := conditions.Conditions{}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Фильм или сериал? (или /skip)")
	msg.ParseMode = tgbotapi.ModeMarkdown
	if _, err := b.tg.Send(msg); err != nil {
		logrus.Error(sendMsgErr, err)
	}

	for update := range updates {

		switch update.Message.Command() {
		case "cancel":
			return conditions.Conditions{}, errors.New("ок, прекращаю поиск")
		}

		if cnd.Type == "" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			switch update.Message.Command() {
			case "skip":
				cnd.Type = "no"
				msg.Text = msgSkip
				if _, err := b.tg.Send(msg); err != nil {
					logrus.Error(sendMsgErr, err)
				}
			default:
				cnd.Type = update.Message.Text
			}

			msg.Text = "Перечислите жанры через пробел (или /skip):"
			if _, err := b.tg.Send(msg); err != nil {
				logrus.Error(sendMsgErr, err)
			}

			continue

		} else if cnd.Genres == nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			switch update.Message.Command() {
			case "skip":
				cnd.Genres = []string{"no"}
				msg.Text = msgSkip
				if _, err := b.tg.Send(msg); err != nil {
					logrus.Error(sendMsgErr, err)
				}
			default:
				cnd.Genres = strings.Fields(update.Message.Text)
			}

			msg.Text = `Укажите год начала поиска (минимум 1900) (или /skip):`
			if _, err := b.tg.Send(msg); err != nil {
				logrus.Error(sendMsgErr, err)
			}
			continue

		} else if cnd.StartYear == "" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			switch update.Message.Command() {
			case "skip":
				cnd.StartYear = "no"
				msg.Text = msgSkip
				if _, err := b.tg.Send(msg); err != nil {
					logrus.Error(sendMsgErr, err)
				}
			default:
				if !checkYear(update.Message.Text) {
					msg.Text = "Неправильный формат или значение, попробуйте снова"
					if _, err := b.tg.Send(msg); err != nil {
						logrus.Error(sendMsgErr, err)
					}
					continue
				}
				cnd.StartYear = update.Message.Text
			}

			msg.Text = fmt.Sprintf(
				"Укажите год окончания поиска (максимум %d) (или /skip):", time.Now().Year()+1)
			if _, err := b.tg.Send(msg); err != nil {
				logrus.Error(sendMsgErr, err)
			}
			continue

		} else if cnd.EndYear == "" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			switch update.Message.Command() {
			case "skip":
				cnd.EndYear = "no"
				msg.Text = msgSkip
				if _, err := b.tg.Send(msg); err != nil {
					logrus.Error(sendMsgErr, err)
				}
			default:
				if !checkYear(update.Message.Text) {
					msg.Text = "Неправильный формат или значение, попробуйте снова"
					if _, err := b.tg.Send(msg); err != nil {
						logrus.Error(sendMsgErr, err)
					}
					continue
				}

				cnd.EndYear = update.Message.Text
			}

			msg.Text = "Укажите ключевые слова через пробел (например, название фильма или имя актера) (или /skip):"
			if _, err := b.tg.Send(msg); err != nil {
				logrus.Error(sendMsgErr, err)
			}
			continue

		} else if cnd.Keyword == "" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			switch update.Message.Command() {
			case "skip":
				cnd.Keyword = "no"
				msg.Text = msgSkip
				if _, err := b.tg.Send(msg); err != nil {
					logrus.Error(sendMsgErr, err)
				}
			default:
				cnd.Keyword = update.Message.Text
			}

			msg.Text = "Перечислите страны через пробел:"
			if _, err := b.tg.Send(msg); err != nil {
				logrus.Error(sendMsgErr, err)
			}
			continue

		} else if cnd.Countries == nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

			switch update.Message.Command() {
			case "skip":
				cnd.Countries = []string{"no"}
				msg.Text = msgSkip
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
