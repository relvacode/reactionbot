package bot

import (
	"bytes"
	"context"
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/h2non/filetype"
	"github.com/nfnt/resize"
	"github.com/relvacode/reactionbot/bot/interaction"
	"github.com/relvacode/reactionbot/bot/store"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	MaxImageBytes = 6 << 20
	MaxImageSize  = 512
)

// resizeImage img according to MaxImageSize
func resizeImage(img image.Image) (image.Image, bool) {
	if img.Bounds().Dx() <= MaxImageSize && img.Bounds().Dy() <= MaxImageSize {
		return img, false
	}

	var x uint = MaxImageSize
	var y uint

	if img.Bounds().Dy() > img.Bounds().Dx() {
		x, y = y, x
	}

	return resize.Resize(x, y, img, resize.Lanczos3), true
}

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
	n, err := io.Copy(&buf, io.LimitReader(resp.Body, MaxImageBytes+1))
	if err != nil {
		log.Printf("Failed to download response data: %v", err)
		return nil, interaction.SafeError("mfw i couldn't download the image")
	}
	if n > MaxImageBytes {
		log.Printf("Maximum image size exceeded")
		return nil, interaction.SafeError("mfw the image was too large")
	}

	kind, _ := filetype.Match(buf.Bytes())
	var img image.Image
	switch kind.Extension {
	case "png":
		img, err = png.Decode(bytes.NewReader(buf.Bytes()))
	case "jpg":
		img, err = jpeg.Decode(bytes.NewReader(buf.Bytes()))
	default:
		err = errors.New("not a supported file type")
	}

	if err != nil {
		log.Printf("Attachment doesn't look like a valid image: %v", err)
		return nil, interaction.SafeError("mfw you didn't upload an image")
	}

	// Resize image, re-encode to png
	var resized bool
	img, resized = resizeImage(img)

	// only re-encode if image was resized
	if resized {
		kind = filetype.GetType("png")

		buf.Reset()
		err = png.Encode(&buf, img)
		if err != nil {
			log.Printf("Failed to re-encode image: %v", err)
			return nil, interaction.SafeError("mfw i couldn't resize the image")
		}
	}

	err = into.Store(ctx, kind, buf.Bytes())
	if err != nil {
		log.Printf("Failed to save image into store: %v", err)
		return nil, interaction.ElseSafe(err, "mfw i couldn't save the image")
	}

	return store.FromContent(kind, buf.Bytes()), nil
}
