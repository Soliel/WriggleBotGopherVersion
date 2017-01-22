package main

import (
  //"github.com/bwmarrin/discordgo"
)

//var aList map[string]*discordgo.User

func AdoptUsers(ctx Context) {
  if ctx.Args[0] == "decline" {
    _, _ = ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "You were not adopted.")
    return
  }
  
  if ctx.Args[0] == "accept" {
    //TODO: Adoption accept logic here
    return 
  }
  return

  /*if ctx.Args[0] {
    _,_ = ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "You must specify a user to adopt.")
    return
  }*/
  
}