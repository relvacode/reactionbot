package main

import (
	"context"
	"flag"
	"github.com/relvacode/reactionbot/bot"
	"github.com/relvacode/reactionbot/bot/store"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	flagGuildID        = flag.String("guild", os.Getenv("GUILD_ID"), "Guild ID")
	flagBotToken       = flag.String("token", os.Getenv("BOT_TOKEN"), "Bot token")
	flagUserImagesPath = flag.String("user-images-path", os.Getenv("USER_IMAGES_PATH"), "Path to store user images")
)

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	images, err := store.NewOSStore(*flagUserImagesPath)
	if err != nil {
		log.Fatalln(err)
	}

	err = bot.Start(ctx, images, *flagGuildID, *flagBotToken)
	if err != nil {
		log.Fatalln(err)
	}
}
