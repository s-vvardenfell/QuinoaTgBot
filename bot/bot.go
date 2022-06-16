package bot

import (
	"log"

	"github.com/s-vvardenfell/QuinoaTgBot/config"
	"github.com/sirupsen/logrus"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type QuinoaTgBot struct {
	tg   *tgbotapi.BotAPI
	cnfg config.Config //нужно ли мне хранить весь конфиг? или только номер группы понадобится?
}

func New(cnfg config.Config) *QuinoaTgBot {
	bot, err := tgbotapi.NewBotAPI(cnfg.Token)
	if err != nil {
		logrus.Fatalf("cannot create NewBotAPI, %v", err)
	}

	bot.Debug = cnfg.Debug
	return &QuinoaTgBot{
		tg:   bot,
		cnfg: cnfg,
	}
}

func (b *QuinoaTgBot) Work() {
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
