package interaction

import (
	"bytes"
	_ "embed"
	"errors"
	"github.com/bwmarrin/discordgo"
)

// SafeError is a friendly error message sent in an interaction response.
// It should not contain any internal state information.
// It should take the form of `mfw ...` and be all lower-case.
type SafeError string

func (e SafeError) Error() string { return (string)(e) }

// ElseSafe checks if error is a SafeError.
// Otherwise, it returns SafeError as the contents of safeMessage.
func ElseSafe(err error, safeMessage string) SafeError {
	var safe SafeError
	if errors.As(err, &safe) {
		return safe
	}

	return SafeError(safeMessage)
}

//go:embed error.png
var errImageSrc []byte

func ErrorToInteractionResponse(err error) *discordgo.InteractionResponse {
	var se = ElseSafe(err, "mfw when something went wrong")

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: (string)(se),
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
