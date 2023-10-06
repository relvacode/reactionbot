package main

import (
	"context"
	"flag"
	"github.com/relvacode/reactionbot/bot"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	flagGuildID  = flag.String("guild", os.Getenv("GUILD_ID"), "Guild ID")
	flagBotToken = flag.String("token", os.Getenv("BOT_TOKEN"), "Bot token")
)

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	err := bot.Start(ctx, *flagGuildID, *flagBotToken)
	if err != nil {
		log.Fatalln(err)
	}
}
