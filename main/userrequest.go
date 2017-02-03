package main

import (
	"github.com/bwmarrin/discordgo"
	//"fmt"
	"errors"
	"time"
)

func onGuildMemberChunk(s *discordgo.Session, members *discordgo.GuildMembersChunk) {
	MemChan <- members.Members[0].User
}

//When you access this function, lock it with userReqLock.
//requests a user using a string.
func requestUserFromGuild(s *discordgo.Session, guild string, user string) (*discordgo.User, error){
	_ = s.RequestGuildMembers(guild, user, 1)
	select{
		case reqUser := <- MemChan:
			return reqUser, nil
		case <- time.After(time.Second):
			return nil, errors.New("Discord request timed out.")
	}
}

func checkDuplicatePets(ID string) (e error){
	var OID string
	err := DataStore.QueryRow("SELECT OwnerID FROM pettable WHERE UserID = ?", ID).Scan(&OID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {return nil}
		return err
	}

	if OID != "" {
		return errors.New("Pet already has an owner")
	}

	return  nil
}

func checkDuplicateOwners(ID string) (exists bool) {
	var OID string
	err := DataStore.QueryRow("SELECT UserID FROM ownertable WHERE UserID = ?", ID).Scan(&OID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return false
		}
	}
	return true
}
