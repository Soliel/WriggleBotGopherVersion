package main

import (
	"github.com/bwmarrin/discordgo"
	//"fmt"
)

type (
	command func(context)
	cmdMap map[string]command
	commandHandler struct {
		Cmds cmdMap
	}
	
	context struct {
		Msg     *discordgo.MessageCreate
		Session *discordgo.Session
		Guild   *discordgo.Guild
		Channel *discordgo.Channel
		Args    []string
	}
)



//create a command handler and associated map and pass the memory address back.
func newCommandHandler() *commandHandler {
	return &commandHandler{make(cmdMap)}
}

//return a map of registered commands
func (handler commandHandler) getCmds() cmdMap {
	return handler.Cmds
}

//Pull a specific command from the map and return the memory address of the function as well as a true or false for error checking
func (handler commandHandler) get(name string) (*command, bool) {
	cmd, found := handler.Cmds[name]
	return &cmd, found
}

//Adds a command to the global command map and assigns it's name to be accessed by.
func (handler commandHandler) register(name string, comman command) {
	handler.Cmds[name] = comman
	if len(name) > 1 {
		handler.Cmds[name[:1]] = comman
	}
}

func testCommand(ctx context) {
	_,_ = ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "The Test has succeded.")
}