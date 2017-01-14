package main

import (
	"fmt"
	"flag"
	"github.com/bwmarrin/discordgo"
)

var (
	Token string
	BotID string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {

	fmt.Println("Hello World")
	dg, err := discordgo.New("Bot " + Token)

	if err != nil {
		fmt.Println("Error creating discord session, ", err)
		return
	}

	u, err := dg.User("@me")
	if err != nil {
		fmt.Println("Error obtaining account details, ", err)
	}

	BotID = u.ID

	dg.AddHandler(mCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection with Discord, ", err)
	}

	fmt.Println("Bot is now running as user: ", u.Username)

	<-make(chan struct{})
	return
}

func mCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == BotID {
		return
	}

	if m.Content == "ping" {
		_,_ = s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	if m.Content == "pong" {
		_,_ = s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}