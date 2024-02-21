package store

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/h2non/filetype/types"
	"github.com/relvacode/reactionbot/bot/interaction"
	"log"
	"math/rand"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type imgHeader struct {
	path        string
	fileExtType types.Type
}

func (hdr imgHeader) open() (*discordgo.File, error) {
	b, err := os.ReadFile(hdr.path)
	if err != nil {
		return nil, err
	}

	return FromContent(hdr.fileExtType, b), nil
}

type imageListNode struct {
	next *imageListNode
	hdr  imgHeader
}

func NewOSStore(basePath string) (*OSStore, error) {
	var store = &OSStore{
		basePath: basePath,
	}

	dirEntries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	// Randomly shuffle dir entries, so we don't always start with the same image
	rand.Shuffle(len(dirEntries), func(i, j int) {
		dirEntries[i], dirEntries[j] = dirEntries[j], dirEntries[i]
	})

	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}

		// Take MIME info purely from the file extension.
		// We can assume that filenames here have already been sanitised to the correct extension.
		var (
			ext  = filepath.Ext(entry.Name())
			kind = types.NewType(strings.TrimPrefix(ext, "."), mime.TypeByExtension(ext))
		)

		switch kind.Extension {
		case "png", "jpg", "gif":
			node := &imageListNode{
				next: store.root,
				hdr: imgHeader{
					path:        filepath.Join(basePath, entry.Name()),
					fileExtType: kind,
				},
			}

			store.len++
			store.root = node
		}
	}

	store.next = store.root

	return store, nil
}

type OSStore struct {
	basePath string
	mx       sync.Mutex
	len      int
	root     *imageListNode
	next     *imageListNode
}

func (s *OSStore) shuffle() {
	if s.len == 0 {
		return
	}

	// Extract nodes into flat list
	nodes := make([]*imageListNode, 0, s.len)
	for n := s.root; n != nil; n = n.next {
		nodes = append(nodes, n)
	}

	// shuffle the list
	rand.Shuffle(len(nodes), func(i, j int) {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	})

	// Update next pointer for each node to point to the next
	for i := 0; i < len(nodes)-1; i++ {
		nodes[i].next = nodes[i+1]
	}

	// Re-pin the root node and terminate the last node
	s.root = nodes[0]
	s.next = s.root
	nodes[len(nodes)-1].next = nil
}

func (s *OSStore) advance() (*imgHeader, error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	img := s.next
	if img == nil {
		return nil, interaction.SafeError("mfw you haven't added any images yet")
	}

	// Advance to next image
	s.next = img.next
	if s.next == nil {
		// If there is no next image, then shuffle the current set and start again
		s.shuffle()
	}

	return &img.hdr, nil
}

func (s *OSStore) Next() (*discordgo.File, error) {
	hdr, err := s.advance()
	if err != nil {
		return nil, err
	}

	log.Printf("Selecting next image %s", hdr.path)

	return hdr.open()
}

func (s *OSStore) Store(_ context.Context, kind types.Type, data []byte) error {
	var (
		imageId  = uuid.New().String() + "." + kind.Extension
		fileName = filepath.Join(s.basePath, imageId)
	)

	err := os.WriteFile(fileName, data, os.FileMode(0775))
	if err != nil {
		log.Printf("Failed to create file: %v", err)
		return interaction.SafeError("mfw i couldn't save your image")
	}

	img := &imageListNode{
		hdr: imgHeader{
			path:        fileName,
			fileExtType: kind,
		},
	}

	// Put new image onto the back of the current stack
	s.mx.Lock()
	defer s.mx.Unlock()

	img.next = s.root
	s.root = img
	s.len++

	if s.next == nil {
		// This is the first node so make it the next one to advance to
		s.next = s.root
	}

	return nil
}
