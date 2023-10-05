package internal

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

// Message represents a message model with Gorm fields.
type Message struct {
	gorm.Model

	ChannelId string
	MessageId string

	PostedMessageId string
}

func (m Message) Render(s *discordgo.Session, guildId, channelId, postChannelId string) {
	// TODO: Render message and/or update it

	originalMessage, err := s.ChannelMessage(channelId, m.MessageId)
	if err != nil {
		fmt.Println(err)
		return
	}

	postMessage := &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				Name:    originalMessage.Author.Username,
				IconURL: originalMessage.Author.AvatarURL(""),
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: originalMessage.Timestamp.Format(time.Stamp),
			},
			Description: originalMessage.Content,
			Color:       originalMessage.Author.AccentColor,
		},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label: "Original Message",
						Style: discordgo.LinkButton,
						URL:   fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildId, originalMessage.ChannelID, originalMessage.ID),
					},
				},
			},
		},
	}

	if len(originalMessage.Attachments) > 0 {
		postMessage.Embed.Image = &discordgo.MessageEmbedImage{
			URL: originalMessage.Attachments[0].URL,
		}
	}

	mess, err := s.ChannelMessageSendComplex(postChannelId, postMessage)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Message: %+v\n", mess)
}

// GetAllMessages retrieves all messages from the database.
// It takes a Gorm database instance as input.
// It returns a slice of Message and an error, if any.
func GetAllMessages(db *gorm.DB) ([]Message, error) {
	var messages []Message
	err := db.Model(&Message{}).Find(&messages).Error
	return messages, err
}

// GetMessage retrieves a specific message from the database based on message ID and channel ID.
// It takes a Gorm database instance, messageId, and channelId as input.
// It returns a Message and an error, if any.
func GetMessage(db *gorm.DB, channelId, messageId string) (Message, error) {
	var message Message
	err := db.Model(&Message{}).Where("message_id = ? AND channel_id = ?", messageId, channelId).First(&message).Error
	return message, err
}

// AddMessage adds a new message to the database.
// It takes a Gorm database instance, channelId, messageId, message, and owner as input.
// It returns an error, if any.
func AddMessage(db *gorm.DB, channelId string, messageId string) error {
	messageObj := Message{
		ChannelId: channelId,
		MessageId: messageId,
	}
	return db.Create(&messageObj).Error
}

// RemoveMessage removes a message from the database based on message ID and channel ID.
// It takes a Gorm database instance, messageId, and channelId as input.
// It returns an error, if any.
func RemoveMessage(db *gorm.DB, channelId, messageId string) error {
	return db.Delete(&Message{}, "message_id = ? AND channel_id = ?", messageId, channelId).Error
}
