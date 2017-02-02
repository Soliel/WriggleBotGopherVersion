package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"database/sql"
	_"github.com/go-sql-driver/mysql"
	"strings"
	"bytes"
	"io/ioutil"
	"encoding/json"
	"sync"
)

//Setting global variables and giving DB a global scope.
const (
	PREFIX = "wrig "
)

//GLOBAL VARS
var (
	conf       *Config
	BotID      string
	DataStore  *sql.DB
	CmdHandler *CommandHandler
	MemChan    chan *discordgo.User
	AList      map[string]*discordgo.User
	userReqLock = &sync.Mutex{}
)

type Config struct {
  BotToken     string `json:"bot_token"`
  DatabaseIP   string `json:"database_ip"`
  DatabaseUser string `json:"database_user"`
  DatabasePass string `json:"database_password"`
  DatabasePort string `json:"database_port"`
  DatabaseName string `json:"database_name"`
}

func main() {
	//Create a string buffer to parse my database information from JSON.
	var buffer bytes.Buffer

	//load a json config file to make launching bot easier.
	conf = LoadConfig("config.json")

	//Concatenate database access stream inside of buffer 
	buffer.WriteString(conf.DatabaseUser)
	buffer.WriteString(":")
	buffer.WriteString(conf.DatabasePass)
	buffer.WriteString("@tcp(")
	buffer.WriteString(conf.DatabaseIP)
	buffer.WriteString(":")
	buffer.WriteString(conf.DatabasePort)
	buffer.WriteString(")/")
	buffer.WriteString(conf.DatabaseName)

	//create discord session, create a database connection, and check for errors.
	dg, err := discordgo.New("Bot " + conf.BotToken)
	DataStore, err = sql.Open("mysql", buffer.String())

	//Test for an error connecting to a database.
	err = DataStore.Ping()

	//if an error occurred creating connections log it here.
	if err != nil {
		fmt.Println("Error creating discord session, or establishing a database connection ", err)
		return
	}

	//initialize the Command Handler & Register Commands
	CmdHandler = NewCommandHandler()
	registerCommands()

	//Initialize adoption list to track current adoptions in a global scope.
	AList = make(map[string]*discordgo.User)
	MemChan = make(chan *discordgo.User)

	//close database after Main ends, should only happen when program exits.
	defer DataStore.Close()

	//select bot user.
	u, err := dg.User("@me")
	if err != nil {
		fmt.Println("Error obtaining account details, ", err)
	}

	//Save bot users ID to a global variable, used to stop command loops
	BotID = u.ID

	//Create a handler for events
	dg.AddHandler(onMessageReceived)
	dg.AddHandler(onGuildMemberChunk)


	//Start listening to discord, Events start firing.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection with Discord, ", err)
	}

	fmt.Println("Bot is now running as user: ", u.Username)

	//Lock the main thread. Keeps application running.
	<-make(chan struct{})
	return
}

func onMessageReceived(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == BotID {
		return
	}

	if len(m.Content) < len(PREFIX) {
		return
	}

	if m.Content[:len(PREFIX)] != PREFIX {
		return
	}

	content := m.Content[len(PREFIX):]
	if len(content) < 1 {
		return
	}

	content = strings.ToLower(content)
	args := strings.Fields(content)
	name := args[0]

	fmt.Println(args)
	command, found := CmdHandler.Get(name)
	if !found {
		_,_ = s.ChannelMessageSend(m.ChannelID, "This command is not valid.")
		return
	}

	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		fmt.Println("Error getting channel, ", err)
		return
	}

	guild, err := s.State.Guild(channel.GuildID)
	if err != nil {
		fmt.Println("Error getting guild, ", err)
		return
	}

	//set up my context to pass to whatever function is called.
	ctx := new(Context)
	ctx.Args = args[1:]
	ctx.Session = s
	ctx.Msg = m
	ctx.Guild = guild
	ctx.Channel = channel

	//pass command pointer and run the function
	c := *command
	go c(*ctx)
}



func LoadConfig(filename string) *Config {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error loading config, ", err)
		return nil
	}

	var confData Config
	err = json.Unmarshal(body, &confData)
	if err != nil {
		fmt.Println("Error parsing JSON data, ", err)
		return nil
	}
	return &confData
}


func registerCommands() {
	CmdHandler.Register("test", TestCommand)
	CmdHandler.Register("betatest", TestCommandTwo)
	CmdHandler.Register("adopt", AdoptUsers)
	CmdHandler.Register("quickbattle", QuickBattle)
}

func TestCommandTwo(ctx Context) {
	_,_ = ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "The second test succeded.")
}
