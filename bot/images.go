package bot

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/h2non/filetype"
	"io"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
)

//go:embed images/*.png
var builtin embed.FS

type imageChoice struct {
	fs   fs.FS
	name string
}

func (i imageChoice) ReadData() ([]byte, error) {
	return fs.ReadFile(i.fs, i.name)
}
func (i imageChoice) File() (*discordgo.File, error) {
	b, err := i.ReadData()
	if err != nil {
		return nil, err
	}

	kind, _ := filetype.Match(b)

	return &discordgo.File{
		ContentType: kind.MIME.Value,
		Name:        "image." + kind.Extension,
		Reader:      bytes.NewReader(b),
	}, nil
}

var choices []imageChoice
var choiceMx sync.RWMutex

func RandomChoice() imageChoice {
	choiceMx.RLock()
	choice := choices[rand.Intn(len(choices))]
	choiceMx.RUnlock()

	return choice
}

const userImageDir = "./user-images"

var userImagesFs = os.DirFS(userImageDir)

func init() {
	entries, err := fs.ReadDir(builtin, "images")
	if err != nil {
		panic(err)
	}

	for _, e := range entries {
		choices = append(choices, imageChoice{
			fs:   builtin,
			name: path.Join("images", e.Name()),
		})
	}

	entries, err = fs.ReadDir(userImagesFs, ".")
	if err != nil {
		panic(err)
	}

	for _, e := range entries {
		fmt.Println(e.Name())
		choices = append(choices, imageChoice{
			fs:   userImagesFs,
			name: e.Name(),
		})
	}

}

func AddUserImage(ctx context.Context, url string) (*imageChoice, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Failed to parse attachment URL: %v", err)
		return nil, SafeError("mfw i couldn't download the attachment")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to make request for attachment: %v", err)
		return nil, SafeError("mfw i couldn't download the attachment")
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("Non-200 response downloading attachment: %v", err)
		return nil, SafeError("mfw the server didn't respond")
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, io.LimitReader(resp.Body, 256000))
	if err != nil {
		log.Printf("Failed to download response data: %v", err)
		return nil, SafeError("mfw i couldn't download the image")
	}

	kind, _ := filetype.Match(buf.Bytes())
	switch kind.Extension {
	case "png", "jpg", "gif":
	default:
		log.Printf("attachment must be an image (got %s)", kind.MIME.Value)
		return nil, SafeError("mfw you didn't upload an image")
	}

	imageId := uuid.New().String() + "." + kind.Extension
	w, err := os.Create(filepath.Join(userImageDir, imageId))
	if err != nil {
		log.Printf("Failed to create file: %v", err)
		return nil, SafeError("mfw i couldn't save your image")
	}

	_, err = io.Copy(w, &buf)
	closeErr := w.Close()

	if err == nil {
		err = closeErr
	}
	if err != nil {
		log.Printf("Failed to write data to file: %v", err)
		return nil, SafeError("mfw i couldn't save the image")
	}

	ic := imageChoice{
		fs:   os.DirFS(userImageDir),
		name: imageId,
	}
	choiceMx.Lock()
	choices = append(choices, ic)
	choiceMx.Unlock()

	return &ic, nil
}
