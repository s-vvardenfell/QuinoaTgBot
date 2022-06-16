package bot

import (
	"log"

	"github.com/s-vvardenfell/QuinoaTgBot/client"
	"github.com/s-vvardenfell/QuinoaTgBot/conditions"
	"github.com/s-vvardenfell/QuinoaTgBot/config"
	"github.com/sirupsen/logrus"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type QuinoaTgBot struct {
	tg     *tgbotapi.BotAPI
	client *client.Client
	cnfg   config.Config //нужно ли мне хранить весь конфиг? или только номер группы понадобится?
	//если общаться с ботом в личке, группа не нужна!
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
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if !update.Message.IsCommand() { // ignore any non-command Messages
			continue
		}

		// Create a new MessageConfig. We don't have text yet,
		// so we leave it empty.
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		// Extract the command from the Message.
		switch update.Message.Command() {
		case "help":
			msg.Text = "I understand /sayhi and /status."
		case "sayhi":
			msg.Text = "Hi :)"
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
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Фильм или сериал?")
	msg.ParseMode = tgbotapi.ModeMarkdown
	b.tg.Send(msg) //TODO _, err :=

	for update := range updates {
		if cnd.Type == "" {
			cnd.Type = update.Message.Text
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			// msg.ReplyToMessageID = update.Message.MessageID
			msg.Text = "Какой жанр поискать?"
			b.tg.Send(msg)
			continue
		} else if cnd.Genres == nil {
			//можно всё же отправлять клавиатуру с жанрами или галочками отметить как-то?
			//надо будет распарсить если несколько жанров и предупредить юзера
			//"перечисли жанры через пробел"
			cnd.Genres = append(cnd.Genres, update.Message.Text)
			break
		} else {
			break
		}
	}
	// return "no data"
	return b.client.FilmsByConditions(conditions.Conditions{
		Type: cnd.Type,
	})

	/*
		спросить жанр/жанры
		спросить сериал или фильм или всё равно
		или мб прислать клавиатуру? хотя это не важно для целей проекта и для client
	*/
}
