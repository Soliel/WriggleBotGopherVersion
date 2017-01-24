package main

import (
	"fmt"
	"strings"
	"github.com/bwmarrin/discordgo"
)

var aList map[string]*discordgo.User

func AdoptUsers(ctx Context) {

	if len(ctx.Args) < 1 {
		_,_ = ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "You must specify a user to adopt.")
		return
	}

	if ctx.Args[0] == "decline" {
   	 	_, _ = ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "You were not adopted.")
   	 	return
  	}
  
  	if ctx.Args[0] == "accept" {
 	   	//TODO: Adoption accept logic here
  	  	return
  	}

	if strings.HasPrefix(ctx.Args[0], "<@") {
		ctx.Args[0] = strings.Trim(ctx.Args[0], "<@>")
	}

	pet_user, err := ctx.Session.User(ctx.Args[0])
	if err != nil {
		fmt.Println("User could not be parsed with ID, showing error, ", err)
		fmt.Println("Attempting to locate user by name")

		aList[ctx.Args[0]] = ctx.Msg.Author

		_ = ctx.Session.RequestGuildMembers(ctx.Guild.ID, ctx.Args[0], 1)
		return
	}

	fmt.Println(pet_user.Username)
	return

}