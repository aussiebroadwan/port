package main

import (
	"fmt"
	"os"

	"github.com/aussiebroadwan/port/internal"
	"github.com/bwmarrin/discordgo"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	REACT_EMOJI = "port"
	THRESHOLD   = 1
)

var (
	dbConn        *gorm.DB
	portChannelId = os.Getenv("PORT_CHANNELID")
)

func main() {
	var err error
	fmt.Println("Hello, World!")

	// Open a database connection
	dbConn, err = gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Println("Error opening database: ", err)
		return
	}

	// Migrate the schema
	dbConn.AutoMigrate(&internal.Message{})

	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		fmt.Println("Error creating discord session: ", err)
		return
	}

	discord.AddHandler(messageReact)
	discord.AddHandler(messageUnreact)

	err = discord.Open()
	if err != nil {
		fmt.Println("Error opening connection: ", err)
		return
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	<-make(chan struct{})
}

func messageReact(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	if m.ChannelID == portChannelId {
		return
	}

	fmt.Printf("[%s:%s] Message React: %s\n", m.ChannelID, m.MessageID, m.Emoji.Name)
	if m.Emoji.Name == REACT_EMOJI {
		reactors, err := s.MessageReactions(m.ChannelID, m.MessageID, m.Emoji.APIName(), 100, "", "")
		if err != nil {
			fmt.Println("Error getting reactors: ", err)
			return
		}

		if len(reactors) >= THRESHOLD {
			internal.AddMessage(dbConn, m.ChannelID, m.MessageID)
		}

		message, err := internal.GetMessage(dbConn, m.ChannelID, m.MessageID)
		if err != nil {
			fmt.Println("Error getting message: ", err)
			return
		}

		message.Render(dbConn, s, m.GuildID, m.ChannelID, portChannelId, len(reactors))
	}
}

func messageUnreact(s *discordgo.Session, m *discordgo.MessageReactionRemove) {
	if m.ChannelID == portChannelId {
		return
	}

	fmt.Printf("[%s:%s] Message Unreact: %s\n", m.ChannelID, m.MessageID, m.Emoji.Name)
	// if m.Emoji.Name == REACT_EMOJI {
	// 	reactors, err := s.MessageReactions(m.ChannelID, m.MessageID, m.Emoji.APIName(), 100, "", "")
	// 	if err != nil {
	// 		fmt.Println("Error getting reactors: ", err)
	// 		return
	// 	}

	// 	if len(reactors) < THRESHOLD {
	// 		internal.RemoveMessage(dbConn, m.ChannelID, m.MessageID)
	// 	}

	// 	message, err := internal.GetMessage(dbConn, m.ChannelID, m.MessageID)
	// 	if err != nil {
	// 		fmt.Println("Error getting message: ", err)
	// 		return
	// 	}

	// 	message.Render(dbConn, s, m.GuildID, m.ChannelID, portChannelId)
	// }
}
