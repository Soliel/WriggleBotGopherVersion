package main

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
)

type (
	Command func(Context)
	CmdMap map[string]Command
	CommandHandler struct {
		Cmds CmdMap
	}
	
	Context struct {
		Msg     *discordgo.MessageCreate
		Session *discordgo.Session
		Guild   *discordgo.Guild
		Channel *discordgo.Channel
		Args    []string
	}
)



//create a command handler and associated map and pass the memory address back.
func NewCommandHandler() *CommandHandler {
	return &CommandHandler{make(CmdMap)}
}

//return a map of registered commands
func (handler CommandHandler) GetCmds() CmdMap {
	return handler.Cmds
}

//Pull a specific command from the map and return the memory address of the function as well as a true or false for error checking
func (handler CommandHandler) Get(name string) (*Command, bool) {
	cmd, found := handler.Cmds[name]
	fmt.Println(cmd, found)
	return &cmd, found
}

//Adds a command to the global command map and assigns it's name to be accessed by.
func (handler CommandHandler) Register(name string, command Command) {
	handler.Cmds[name] = command
	if len(name) > 1 {
		handler.Cmds[name[:1]] = command
	}
}

func TestCommand(ctx Context) {
	_,_ = ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "The Test has succeded.")
}