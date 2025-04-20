package main

import (
	"os"
	"time"

	"github.com/cyrillus31/go_rate_limiter/ratelimiter"
	"github.com/joho/godotenv"
	tele "gopkg.in/telebot.v4"
)

type BotRespWriter struct {
	bot  *tele.Bot
	user *tele.Chat
}

func (w *BotRespWriter) SetUser(user *tele.Chat) {
	w.user = user
}

func (w BotRespWriter) Write(p []byte) (n int, err error) {
	_, err = w.bot.Send(w.user, string(p))
	return len(p), err
}

func main() {
	godotenv.Load()

	settings := tele.Settings{
		Token:  os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	bot, _ := tele.NewBot(settings)

	botWriter := &BotRespWriter{bot: bot}
	rl := ratelimiter.NewRateLimiter(5, botWriter)

	bot.Handle(tele.OnText, func(ctx tele.Context) error {
		user := ctx.Chat()
		text := ctx.Text()
		botWriter.SetUser(user)
		rl.Limit(text)
		return nil
	})

	bot.Start()
}
