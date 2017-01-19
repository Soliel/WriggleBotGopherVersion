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
	DataStore, err = sql.Open("mysql", "root:M38a67c%%7@tcp(127.0.0.1:3306)/testschema")
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

	dg.AddHandler(onMessageReceived)

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection with Discord, ", err)
	}

	fmt.Println("Bot is now running as user: ", u.Username)

	<-make(chan struct{})
	return
}

func CheckErr(err error) {

	if err != nil {
		log.Fatal(err)
	}
}