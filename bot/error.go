package bot

import (
	"bytes"
	_ "embed"
	"errors"
	"github.com/bwmarrin/discordgo"
)

type SafeError string

func (e SafeError) Error() string { return (string)(e) }

//go:embed error.png
var errImageSrc []byte

func ErrorToInteractionResponse(err error) *discordgo.InteractionResponse {
	var text = "mfw when something went wrong"
	var se SafeError
	if errors.As(err, &se) {
		text = (string)(se)
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: text,
			Files: []*discordgo.File{
				{
					Name:        "error.png",
					ContentType: "image/png",
					Reader:      bytes.NewReader(errImageSrc),
				},
			},
		},
	}
}
