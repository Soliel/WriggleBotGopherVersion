package main

import (
	"fmt"
	"strings"
)

func AdoptUsers(ctx Context) {

	if len(ctx.Args) < 1 {
		_,_ = ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "You must specify a user to adopt.")
		return
	}

	if ctx.Args[0] == "decline" {
		_, _ = ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "You were not adopted.")
		delete(AList, ctx.Msg.Author.ID)
 		return
	}
  
	if ctx.Args[0] == "accept" {
		fmt.Println(AList[ctx.Msg.Author.ID])
		return
  }

	//Get the userID from a mention.
	if strings.HasPrefix(ctx.Args[0], "<@") {
		ctx.Args[0] = strings.Trim(ctx.Args[0], "<@>")
	}
	
	//Search for user by ID in case they entered one.
	pet_user, err := ctx.Session.User(ctx.Args[0])
	if err == nil {
		AList[pet_user.ID] = ctx.Msg.Author
	}
	
	//Start a query for the user, Causes Guild Member Chunk event to fire. 
	if err != nil {
		fmt.Println("User could not be parsed with ID, showing error, ", err)
		fmt.Println("Attempting to locate user by name")
		
		//We lock the function with UserReqLock to ensure it is the only one being requested.
		userReqLock.Lock()
		reqUser := requestUserFromGuild(ctx)
		AList[reqUser.ID] = ctx.Msg.Author
		userReqLock.Unlock()
		
		return
	}
	return

}