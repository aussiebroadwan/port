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

func (m Message) Render(db *gorm.DB, s *discordgo.Session, guildId, channelId, postChannelId string, reacts int) {
	// TODO: Render message and/or update it

	message, err := GetMessage(db, channelId, m.MessageId)
	if err != nil {
		fmt.Println("Get Message", err)
		return
	}

	if message.PostedMessageId != "" {
		// Update message
		fmt.Println("Update message")

		if reacts == 0 {
			// Delete message
			err = s.ChannelMessageDelete(postChannelId, message.PostedMessageId)
			if err != nil {
				fmt.Println("Delete Message Err: ", err)
			}
			return
		}

		originalMessage, err := s.ChannelMessage(channelId, m.MessageId)
		if err != nil {
			fmt.Println(err)
			return
		}

		// edit.SetContent(fmt.Sprintf("%d :port:", reacts))
		_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel: channelId,
			ID:      message.PostedMessageId,
			Embed: &discordgo.MessageEmbed{
				Author: &discordgo.MessageEmbedAuthor{
					Name:    originalMessage.Author.Username,
					IconURL: originalMessage.Author.AvatarURL(""),
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: originalMessage.Timestamp.Format(time.Stamp),
				},
				Description: originalMessage.Content,
				Title:       fmt.Sprintf("%d 2 Reacts", reacts),
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
		})

		if err != nil {
			fmt.Println("Edit Message", err)
			return
		}

	} else {

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
				Title:       fmt.Sprintf("%d Reacts", reacts),
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
			fmt.Println("Send Message", err)
			return
		}

		err = UpdateMessagePost(db, channelId, m.MessageId, mess.ID)
		if err != nil {
			fmt.Println("Update Post: ", err)
			return
		}

	}
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

func UpdateMessagePost(db *gorm.DB, channelId, messageId, postedMessageId string) error {
	message, err := GetMessage(db, channelId, messageId)
	if err != nil {
		return err
	}

	message.PostedMessageId = postedMessageId
	return db.Save(&message).Error
}

// AddMessage adds a new message to the database.
// It takes a Gorm database instance, channelId, messageId, message, and owner as input.
// It returns an error, if any.
func AddMessage(db *gorm.DB, channelId string, messageId string) error {
	messageObj := Message{
		ChannelId: channelId,
		MessageId: messageId,
	}

	// Get all Messages
	messages, err := GetAllMessages(db)
	if err != nil {
		return err
	}

	// Check if message already exists
	for _, message := range messages {
		if message.ChannelId == channelId && message.MessageId == messageId {
			return nil
		}
	}

	return db.Create(&messageObj).Error
}

// RemoveMessage removes a message from the database based on message ID and channel ID.
// It takes a Gorm database instance, messageId, and channelId as input.
// It returns an error, if any.
func RemoveMessage(db *gorm.DB, channelId, messageId string) error {
	return db.Delete(&Message{}, "message_id = ? AND channel_id = ?", messageId, channelId).Error
}
