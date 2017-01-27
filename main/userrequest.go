package main

import (
  "fmt"
  "github.com/bwmarrin/discordgo"
)

func onGuildMemberChunk(s *discordgo.Session, members *discordgo.GuildMembersChunk) {
	fmt.Println(members.Members[0].User.ID)
	MemChan <- members.Members[0].User
}

//When you access this function, lock it with userReqLock.
//requests a user using a string.
func requestUserFromGuild(ctx Context) (*discordgo.User) {
  _ = ctx.Session.RequestGuildMembers(ctx.Guild.ID, ctx.Args[0], 1)
	reqUser := <- MemChan
  return reqUser
}