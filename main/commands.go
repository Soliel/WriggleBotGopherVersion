package main

import (
	"github.com/bwmarrin/discordgo"
)

func onMessageReceived(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == BotID {
		return
	}

	if m.Content == "Fmuph" {
		_,_ = s.ChannelMessageSend(m.ChannelID, "How did you guess that?")
	}

	go insertMessageDB(m)
}

func insertMessageDB(m *discordgo.MessageCreate) {

	tx, err := DataStore.Begin()
	CheckErr(err)
	defer  tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO testtable(userID, Message) VALUES (?, ?)")
	CheckErr(err)
	defer stmt.Close()

	_, _ = stmt.Exec(m.Author.ID, m.Content)

	err = tx.Commit()
	CheckErr(err)
}

//TODO: Create a map based Command Handler that I can register functions to.
