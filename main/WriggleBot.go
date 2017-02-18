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
	"time"
)

//Setting global variables and giving DB a global scope.
const (
	PREFIX = "wrig "
)

//GLOBAL VARS
var (
	conf         *config
	BotID        string
	DataStore    *sql.DB
	CmdHandler   *commandHandler
	MemChan      chan *discordgo.User
	AList        map[string]*discordgo.User
	userReqLock  = &sync.Mutex{}
	tTick        *time.Ticker
	trainingMap  map[time.Time]training
	cooldownMap  map[string]time.Time
)

type config struct {
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
	conf = loadConfig("config.json")

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
	CmdHandler = newCommandHandler()
	registerCommands()

	//Initialize adoption list to track current adoptions in a global scope.
	AList = make(map[string]*discordgo.User)
	cooldownMap = make(map[string]time.Time)
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

	tTick = time.NewTicker(time.Second)
	go startTickReceiver()
	go startCooldownTicker()
	defer tTick.Stop()

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

	comman, found := CmdHandler.get(name)
	if !found {
		return
	}
	
	if comman.hasCooldown() {
		if comman.isOnCooldown(m.Author.ID) {
			return
		}
		comman.startCooldown(m.Author.ID)
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
	ctx := new(context)
	ctx.Args = args[1:]
	ctx.Session = s
	ctx.Msg = m
	ctx.Guild = guild
	ctx.Channel = channel

	//pass command pointer and run the function
	c := comman.CmdFunc
	go c(*ctx)
}



func loadConfig(filename string) *config {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error loading config, ", err)
		return nil
	}

	var confData config
	err = json.Unmarshal(body, &confData)
	if err != nil {
		fmt.Println("Error parsing JSON data, ", err)
		return nil
	}
	return &confData
}


func registerCommands() {
	CmdHandler.register("adopt",       adoptUsers,   0)
	CmdHandler.register("quickbattle", quickBattle,  15)
	CmdHandler.register("train",       trainStat,    0)
	CmdHandler.register("pets",        listPets,     0)
	CmdHandler.register("statsheet",   petStatSheet, 0)
	CmdHandler.register("abandon",     abandon,      0)
	CmdHandler.register("flee",        flee,         0)
	CmdHandler.register("update",      sendBotWideNotice, 0)
	//CmdHandler.register("showalist", showAdoptions)
}

func sendBotWideNotice(ctx context) {
	if ctx.Msg.Author.ID != "96013796681736192" {
		return
	}
	
	embed := &discordgo.MessageEmbed {
		Title: "WriggleBot Announcement",
		Color: 14030101,
		Author: &discordgo.MessageEmbedAuthor{URL: "", Name: "WriggleBot", IconURL: "https://discordapp.com/api/v6/users/209739190244474881/avatars/47ada5c68c51f8dc2360143c0751d656.jpg"},
		Fields: []*discordgo.MessageEmbedField{
			{"", `Dear WriggleBot users:
WriggleBot has been completely rewritten! Most of the stuff has stayed the same except for a few major changes:
1.) The commands addstat and remstat have been removed. These commands have been replaced with the command train, which looks like this:
wrig train <petname> <petstats>
This command will wait an amount of time based on the stats level and then "Levelup" the stat.
2.) the command battle has been replaced with quickbattle, which still looks the same. Some foreshadowing of things to come.
wrig quickbattle <pet1> <pet2>
3.) WriggleBot will no longer respond to invalid commands. This includes commands where users are not found. or if you don't own a certain pet. This change will hopefully cut down on the spam Wriggle causes.
4.) Commands can now be used by a 2 letter prefix. quickbattle is really long right? well, now you can also use:
wrig qu <pet1> <pet2> 
to call the command. this will work for every command.
5.) Formatted output has been replaced with rich embeds. This message is one of those. WriggleBot said she deserved to be beautiful.
6.) Unfortunately as a result of some massive under-the-hood changes and the removal of the older stat system in lieu of training. The old database is no longer compatible, as a result Pets will be completely reset. Everyone is now back to square one. again, I'm sorry. This shouldn't happen ever again.

I'm sorry for how long it has been since there has been an update to our friend Wriggle. but they should now start coming more regularly. Expect future updates to include:
- Further progression introducing an item and currency system.
- More Challenge with Raids and party battles coming
- More interaction with the interactive battles.
- More communication from me. 

Thank you for using WriggleBot, and if you have any questions feel free to contact me through discord at the username Soliel#0897
		`, true},
		},
	}
	
	_, err := ctx.Session.ChannelMessageSendEmbed(ctx.Msg.ChannelID, embed)
	if err != nil {
		fmt.Println(err)
	}
}
