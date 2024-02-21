package bot

import (
	"bytes"
	"context"
	"github.com/bwmarrin/discordgo"
	"github.com/h2non/filetype"
	"github.com/relvacode/reactionbot/bot/interaction"
	"github.com/relvacode/reactionbot/bot/store"
	"io"
	"log"
	"net/http"
	"time"
)

// AddImage downloads the contents of url and expects it to contain an image.
// It then stores the contents of the image into the given store.Store.
// It returns the image that was just saved.
func AddImage(ctx context.Context, url string, into store.Store) (*discordgo.File, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Failed to parse attachment URL: %v", err)
		return nil, interaction.SafeError("mfw i couldn't download the attachment")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to make request for attachment: %v", err)
		return nil, interaction.SafeError("mfw i couldn't download the attachment")
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("Non-200 response downloading attachment: %v", err)
		return nil, interaction.SafeError("mfw the server didn't respond")
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, io.LimitReader(resp.Body, 256000))
	if err != nil {
		log.Printf("Failed to download response data: %v", err)
		return nil, interaction.SafeError("mfw i couldn't download the image")
	}

	kind, _ := filetype.Match(buf.Bytes())
	switch kind.Extension {
	case "png", "jpg", "gif":
	default:
		log.Printf("attachment must be an image (got %s)", kind.MIME.Value)
		return nil, interaction.SafeError("mfw you didn't upload an image")
	}

	err = into.Store(ctx, kind, buf.Bytes())
	if err != nil {
		log.Printf("Failed to save image into store: %v", err)
		return nil, interaction.ElseSafe(err, "mfw i couldn't save the image")
	}

	return store.FromContent(kind, buf.Bytes()), nil
}
