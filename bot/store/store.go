package store

import (
	"bytes"
	"context"
	"github.com/bwmarrin/discordgo"
	"github.com/h2non/filetype/types"
)

type Store interface {
	// Next advances the Store state and returns the next available image
	Next() (*discordgo.File, error)
	// Store stores a new image in the Store
	Store(ctx context.Context, kind types.Type, data []byte) error
}

// FromContent creates a new discordgo.File attachment using the given MIME type and data
func FromContent(kind types.Type, data []byte) *discordgo.File {
	return &discordgo.File{
		ContentType: kind.MIME.Value,
		Name:        "image." + kind.Extension,
		Reader:      bytes.NewReader(data),
	}
}
