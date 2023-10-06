package bot

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
)

func Start(ctx context.Context, guildID string, botToken string) error {
	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		return err
	}

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "add-face",
			Description: "Add a new reaction face",

			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "image",
					Description: "The reaction face image",
					Type:        discordgo.ApplicationCommandOptionAttachment,
					Required:    true,
				},
			},
		},
		{
			Name:        "face",
			Description: "Post a random reaction image",

			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "message",
					Description: "My face when",
					Type:        discordgo.ApplicationCommandOptionString,
				},
			},
		},
	}

	commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"add-face": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var attachmentId string
			var hasAttachmentId bool
			for _, opt := range i.ApplicationCommandData().Options {
				if opt.Name == "image" {
					attachmentId = opt.Value.(string)
					hasAttachmentId = true
					break
				}
			}

			if !hasAttachmentId {
				log.Printf("Interaction did not contain an attachment id")
				_ = s.InteractionRespond(i.Interaction, ErrorToInteractionResponse(SafeError("mfw when the command was invalid")))
				return
			}

			attachmentUrl := i.ApplicationCommandData().Resolved.Attachments[attachmentId].URL
			_, err := AddUserImage(ctx, attachmentUrl)
			if err != nil {
				log.Printf("Failed to create user image: %v", err)
				_ = s.InteractionRespond(i.Interaction, ErrorToInteractionResponse(err))
				return
			}

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "mfw you add another reaction face",
					Embeds: []*discordgo.MessageEmbed{
						{
							Type: discordgo.EmbedTypeImage,
							Image: &discordgo.MessageEmbedImage{
								URL: attachmentUrl,
							},
						},
					},
				},
			})

			if err != nil {
				log.Printf("Failed to send interaction response: %v", err)
			}
		},
		"face": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var text string
			for _, opt := range i.ApplicationCommandData().Options {
				if opt.Name == "message" {
					text = opt.StringValue()
					break
				}
			}

			img, err := RandomChoice().File()
			if err != nil {
				log.Printf("Failed get image data: %v", err)
				_ = s.InteractionRespond(i.Interaction, ErrorToInteractionResponse(err))
				return
			}

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: text,
					Files:   []*discordgo.File{img},
				},
			})

			if err != nil {
				log.Printf("Failed to send interaction response: %v", err)
			}
		},
	}

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	err = dg.Open()
	if err != nil {
		return err
	}

	defer dg.Close()

	var registeredCommands = make([]*discordgo.ApplicationCommand, 0, len(commands))
	for i := range commands {
		cmd, err := dg.ApplicationCommandCreate(dg.State.User.ID, guildID, commands[i])
		if err != nil {
			return fmt.Errorf("failed to create command: %w", err)
		}

		registeredCommands = append(registeredCommands, cmd)
	}

	<-ctx.Done()
	log.Println("Shutting down")

	for _, v := range registeredCommands {
		err := dg.ApplicationCommandDelete(dg.State.User.ID, guildID, v.ID)
		if err != nil {
			log.Printf("Failed to remove command: %v", err)
		}
	}

	return nil
}
