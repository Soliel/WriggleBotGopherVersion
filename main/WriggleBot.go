package main

import (
	"fmt"
	"flag"
	"github.com/bwmarrin/discordgo"
	"database/sql"
	_"github.com/go-sql-driver/mysql"
	"log"
	"strings"
)

//Setting global variables and giving DB a global scope.
const (
	PREFIX = "wrig "
)

var (
	Token      string
	BotID      string
	DataStore  *sql.DB
	CmdHandler *CommandHandler
)

//getting command line token flags
//TODO: Update to config file instead.
func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {

	//create discord session, create a database connection, and check for errors. 
	dg, err := discordgo.New("Bot " + Token)
	DataStore, err = sql.Open("mysql", "root:M38a67c%%7@tcp(127.0.0.1:3306)/wriggletest")
	//Test for an error connecting to a database.
	err = DataStore.Ping()

	//if an error occured creating connections log it here.
	if err != nil {
		fmt.Println("Error creating discord session, or establishing a database connection ", err)
		return
	}
	
	//initialize the Command Handler & Register Commands.
	CmdHandler = NewCommandHandler()
	registerCommands()
	
	//close database after Main ends, should only happen when program exits.
	defer DataStore.Close()

	//select bot user.
	u, err := dg.User("@me")
	if err != nil {
		fmt.Println("Error obtaining account details, ", err)
	}

	//Save bot users ID to a global variable, used to stop command loops
	BotID = u.ID

	//Create a message handler
	dg.AddHandler(onMessageReceived)

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
	
	fmt.Println("Checking against length of prefix.")
	if len(m.Content) < len(PREFIX) {
		return
	}
	
	fmt.Println("Checking prefix")
	if m.Content[:len(PREFIX)] != PREFIX {
		return
	}
	
	fmt.Println("Parsing arguements")
	content := m.Content[len(PREFIX):]
	if len(content) < 1 {
		return
	}
	
	fmt.Println("Searching for commands.")
	args := strings.Fields(content)
	name := strings.ToLower(args[0])
	fmt.Println(args)
	command, found := CmdHandler.Get(name)
	if !found {
		_,_ = s.ChannelMessageSend(m.ChannelID, "This command is not valid.")
		return
	}
	
	//set up my context to pass to whatever function is called.
	ctx := new(Context)
	ctx.Args = args[1:]
	ctx.Session = s
	ctx.Msg = m
	
	//pass command pointer and run the function
	c := *command
	c(*ctx)

}

//Make fatal errors easier.
func CheckErr(err error) {

	if err != nil {
		log.Fatal(err)
	}
}

func registerCommands() {
	CmdHandler.Register("test", TestCommand)
	CmdHandler.Register("betatest", TestCommandTwo)
}

func TestCommandTwo(ctx Context) {
	_,_ = ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "The second test succeded.")
}