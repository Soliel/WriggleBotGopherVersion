package main

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"time"
)

type (
	command struct {
		CmdFunc  func(context)
		Cooldown int	
	}
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
func (handler commandHandler) register(name string, comman func(context), cooldown int) {
	handler.Cmds[name] = command{comman, cooldown}
	if len(name) > 2 {
		handler.Cmds[name[:2]] = command{comman, cooldown}
	}
}

func (comman command) hasCooldown() bool {
	if comman.Cooldown > 0 {
		return true
	}
	return false
}

func (comman command) startCooldown(userID string) {
	cooldownMap[userID] = time.Now().Add(time.Duration(comman.Cooldown) * time.Second)
}

func (comman command) isOnCooldown(userID string) bool {
	if cooldownMap[userID].IsZero() {
		return false
	}
	if cooldownMap[userID].Sub(time.Now()) >= 0 {
		return true
		fmt.Println("command is on cooldown")
	}
	return false
}


func startCooldownTicker() {
	for tick := range tTick.C {
		for key, value := range cooldownMap {
			if tick.Sub(value) >= 0 {
				delete(cooldownMap, key)
			}
		}
	}
}