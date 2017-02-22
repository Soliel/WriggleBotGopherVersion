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

	_, err = DataStore.Exec("SET NAMES utf8")
	if err != nil {
		fmt.Println("Failure to set databse connection to UTF 8")
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

	var args []string

	//If someones name has spaces this allows them to fix it.
	if strings.Contains(content, "\"") {
		tempArgs := strings.Split(content, "\"")
		for s := range tempArgs {
			tempArgs[s] = strings.TrimSpace(tempArgs[s])
			if tempArgs[s] != "" {
				args = append(args, tempArgs[s])
			}
		}

		args[0] = strings.TrimPrefix(args[0], " ")
		args[0] = strings.TrimSuffix(args[0], " ")
		if strings.Contains(args[0], " ") {
			firstArgs := strings.Fields(args[0])
			args = append(firstArgs[:], args[1:]...)
		}
	} else {
		args = strings.Fields(content)
	}
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

	//set up my context to pass to whatever function is called.
	ctx := new(context)
	ctx.Args = args[1:]
	ctx.Session = s
	ctx.Msg = m
	ctx.Channel = channel

	guild, err := s.State.Guild(channel.GuildID)
	if err == nil {
		ctx.Guild = guild
	}


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
	CmdHandler.register("help",        help,         0)
	CmdHandler.register("optout",      levelOptOut,  0)
	//CmdHandler.register("showalist", showAdoptions)
}

//TODO: Move help dialog to a configuration file.
func help(ctx context) {
	var embed discordgo.MessageEmbed
	var embedField discordgo.MessageEmbedField
	var textOnly bool
	if len(ctx.Args) == 0 {
		message  := "Thank you for using WriggleBot\n\n" +
		"The commands are:\n\n" +
		"Adopt <User to Adopt|accept|decline> \n\n" +
		"Train <pet to train> <stat to train> \n\n" +
		"Quickbattle <your pet> <pet to battle> \n\n" +
		"Pets \n\n" +
		"Statsheet <pet to view> \n\n" +
		"Abandon <pet to abandon> \n\n" +
		"Flee \n\n" +
		"OptOut <true|false>\n\n" +
		"To find out what a specific command does, type:\n\n" +
		"wrig help <command>"

		embedField.Value = message
		embedField.Name = "Main Help"
		embedField.Inline = false
	}

	if len(ctx.Args) == 1 {
		if ctx.Args[0] == "textonly" {
			textOnly = true
		}
	}

	if len(ctx.Args) > 0 {
		switch ctx.Args[0] {
		case "adopt":
			message := "This command adopts a user.\n\n" +
				"Once run this command will start an adoption of another user, turning them into your pet. " +
				"adoptions will last for 15 seconds before they time out, " +
				"Adopt a user using:\n" +
				"``wrig adopt <user>``\n\n" +
				"within 15 seconds user must type:\n\n" +
				"``wrig adopt accept\n\n``" +
				"In order to be adopted. Otherwise the command will do nothing.\n\n" +
				"You may either type the users username, nick, or mention them in order to adopt them.\n\n" +
				"If you do not know what a pet is, please type wrig help pets"

			embedField.Value = message
			embedField.Name = "Adopt Help"

		case "quickbattle":
			message := "This command allows you to battle your pet against someone elses.\n\n" +
				"``wrig quickbattle <your pet> <enemy pet>``\n\n" +
				"This command will not work with mentions.\n\n" +
				"If you do not know what a pet is, please type wrig help pets"

			embedField.Value = message
			embedField.Name = "Quick Battle Help"

		case "train":
			message := "This command allows you to train your pet, increasing it's effectiveness in battles.\n\n" +
				"Once run, this command will trigger a timer, after the timers duration is up your pet will gain a level in a certain stat\n\n" +
				"``wrig train <your pet> <stat to train>``\n\n" +
				"The training duration increases as you level up a stat but will not exceed 24 hours.\n\n" +
				"Valid stats are:\n\n" +
				"Attack\n" +
				"Defense\n" +
				"Crit\n" +
				"Evasion\n" +
				"Accuracy\n\n" +
				"This command will not work with mentions."

			embedField.Value = message
			embedField.Name = "Training Help"

		case "pets":
			message := "This command will bring up the list of pets you own.\n\n" +
				"``wrig pets``\n\n" +
				"Pets are other users that have been adopted by someone. " +
				"They can be used to battle against one another and, eventually, to challenge harder fights.\n\n" +
				"A pet will grow as it gains experience and will \"Level up.\" A notification is sent to the owner upon level up unless they have opted out.\n\n" +
				"A pet can also get stronger by training using the train command"

			embedField.Value = message
			embedField.Name = "Pets Help"

		case "statsheet":
			message := "This command will bring up a formatted list of the designated pet's combat stats and level." +
				"``wrig statsheet <pet>``\n\n" +
				"This command may be used regardless of whether you own the pet or not.\n\n" +
				"This command will not work with mentions."

			embedField.Value = message
			embedField.Name = "Stat Sheet Help"

		case "abandon":
			message := "This command will allow you to abandon a pet if you do not like them or want to give them to someone else.\n\n" +
				"``wrig abandon <your pet>``\n\n" +
				"This command will not work with mentions."

			embedField.Value = message
			embedField.Name = "Abandonment Help"

		case "flee":
			message := "This command is only usable by pets and allows them to leave an owner they do not like. " +
				"Intended for instances where you are not recieving training or battle experience\n\n" +
				"``wrig flee``\n\n" +
				"This will cause you to flee from the owner and cannot be undone unless you get adopted by them again."

			embedField.Value = message
			embedField.Name = "Fleeing Help"

		case "optout":
			message := "This command will allow you to opt in or out of level up PMs.\n\n" +
				"By default this feature is opted out.\n\n" +
				"To opt in:\n" +
				"``wrig optout false``\n\n" +
				"To opt out again:\n" +
				"``wrig optout true``"

			embedField.Value = message
			embedField.Name = "Leveling Message Opt Out Help"
		}

		if len(ctx.Args) == 2 {
			if ctx.Args[1] == "textonly" {
				textOnly = true
			}
		}
	}
	embed.Color = 14030101
	embed.Fields = []*discordgo.MessageEmbedField{&embedField}

	pmChannel, err := ctx.Session.UserChannelCreate(ctx.Msg.Author.ID)
	if err != nil {
		return
	}

	if textOnly {
		ctx.Session.ChannelMessageSend(pmChannel.ID, embedField.Value)
		return
	}

	_, err  = ctx.Session.ChannelMessageSendEmbed(pmChannel.ID, &embed)
	if err != nil {
		fmt.Println(err)
	}
}