package main

import (
	"fmt"
	"flag"
	"github.com/bwmarrin/discordgo"
	"database/sql"
	_"github.com/go-sql-driver/mysql"
	"log"
)

var (
	Token     string
	BotID     string
	DataStore *sql.DB
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {

	dg, err := discordgo.New("Bot " + Token)
	DataStore, err = sql.Open("mysql", "chillout")
	err = DataStore.Ping()

	if err != nil {
		fmt.Println("Error creating discord session, or establishing a database connection ", err)
		return
	}
	defer DataStore.Close()

	u, err := dg.User("@me")
	if err != nil {
		fmt.Println("Error obtaining account details, ", err)
	}

	BotID = u.ID

	dg.AddHandler(messageCreate)
	dg.AddHandler(userUpdate)

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection with Discord, ", err)
	}

	fmt.Println("Bot is now running as user: ", u.Username)

	<-make(chan struct{})
	return
}

func userUpdate(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	fmt.Println("User Updated")
	_,_ = s.ChannelMessageSend("261665951718703114", "<@" + string(m.User.ID) + ">")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == BotID {
		return
	}

	if m.Content == "ping" {
		_,_ = s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	if m.Content == "pong" {
		_,_ = s.ChannelMessageSend(m.ChannelID, "Ping!")
	}

	go insertMessageDB(m)
}

func insertMessageDB(m *discordgo.MessageCreate) {

	tx, err := DataStore.Begin()
	checkErr(err)
	defer  tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO testtable(userID, Message) VALUES (?, ?)")
	checkErr(err)
	defer stmt.Close()

	_, _ = stmt.Exec(m.Author.ID, m.Content)

	err = tx.Commit()
	checkErr(err)
}

func checkErr(err error) {

	if err != nil {
		log.Fatal(err)
	}
}