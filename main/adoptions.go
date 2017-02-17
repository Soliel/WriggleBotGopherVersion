/*
TODO: Switch all debugging messages into logrus or zap.
 */
package main

import (
	"fmt"
	"strings"
	"time"
	"bytes"
	"github.com/bwmarrin/discordgo"
)

func adoptUsers(ctx context) {

	if len(ctx.Args) < 1 {
		//ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "You must specify a user to adopt.")
		return
	}


	if ctx.Args[0] == "decline" {
		ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "You were not adopted.")
		delete(AList, ctx.Msg.Author.ID)
 		return
	}
  
	if ctx.Args[0] == "accept" {
		if AList[ctx.Msg.Author.ID] == nil {
			return
		}

		tx, err := DataStore.Begin()
		if err != nil {
			//ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "Transaction could not be started, Adoption Aborted.")
			delete(AList, ctx.Msg.Author.ID)
			fmt.Println(err)
			return
		}
		defer tx.Rollback()

		/*stmt, err := tx.Prepare()
		if err != nil {
			//ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "Could not prepare SQL statement, adoption aborted.")
			delete(AList, ctx.Msg.Author.ID)
			fmt.Println(err)
			return
		}
		defer stmt.Close()*/
			
		_, err = tx.Exec("INSERT INTO pettable VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?, ?)", ctx.Msg.Author.ID, ctx.Msg.Author.Username, 1, 10, 10, 20, 1, 5, 80, 0, 0, 0, 0, 0, 0, AList[ctx.Msg.Author.ID].ID, false)
		if err != nil {
			//ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "Could not execute SQL statement with the database, adoption aborted.")
			delete(AList, ctx.Msg.Author.ID)
			fmt.Println(err)
			return
		}
			
		dupOwn := checkDuplicateOwners(AList[ctx.Msg.Author.ID].ID)
		if !dupOwn {
			//stmt, _ = tx.Prepare()
			tx.Exec("INSERT INTO ownertable VALUES(?,?,?)", AList[ctx.Msg.Author.ID].ID, AList[ctx.Msg.Author.ID].Username, 1)
		}
			
		if dupOwn {
			
			var petamnt int
			
			err = DataStore.QueryRow("SELECT PetAmount FROM ownertable WHERE UserID = ?", AList[ctx.Msg.Author.ID].ID).Scan(&petamnt)
			if err != nil {
				//ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "SQL Query failed to find Owner, adoption aborting.")
				return
			}
			
			/*stmt, err = tx.Prepare()
			if err != nil {
				//ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "Could not prepare SQL statement, adoption aborted.")
				return
			}*/
			
			_, err := tx.Exec("UPDATE ownertable SET PetAmount = ? WHERE UserID = ?", petamnt + 1, AList[ctx.Msg.Author.ID].ID)
			if err != nil {
				fmt.Println(err)
				//ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "Could not execute SQL statement with the database, adoption aborted.")
				return
			}
		}
		
		err = tx.Commit()
		if err != nil {
			//ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "Could not commit changes to database, adoption aborted.")
			delete(AList, ctx.Msg.Author.ID)
			fmt.Println(err)
			return
		}

		ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "You have been adopted!")
		return
	}

	//Get the Snowflake ID from a mention.
	if strings.HasPrefix(ctx.Args[0], "<@") {
		ctx.Args[0] = strings.Trim(ctx.Args[0], "<@>")
	}


	//Start a query for the user, Causes Guild Member Chunk event to fire
	//We lock the function with UserReqLock to ensure it is the only one being requested.
	userReqLock.Lock()
	reqUser, err := requestUserFromGuild(ctx.Session, ctx.Guild.ID, ctx.Args[0])
	if err != nil {
		reqUser, err := ctx.Session.User(ctx.Args[0])
		if err == nil {
			userReqLock.Unlock()
			dupErr := checkDuplicatePets(reqUser.ID)
			if dupErr != nil {
				ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, dupErr.Error())
				return
			}
			AList[reqUser.ID] = ctx.Msg.Author
			go timeoutAdoption(reqUser.ID)
			ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, reqUser.Username + " Do you accept the adoption? if so type ``wrig adopt accept``, ``wrig adopt decline`` otherwise.")
			return 
		}
		userReqLock.Unlock()
		return
	}

	dupErr := checkDuplicatePets(reqUser.ID)
	if dupErr != nil {
		ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, dupErr.Error())
		userReqLock.Unlock()
		return
	}

	AList[reqUser.ID] = ctx.Msg.Author
	go timeoutAdoption(reqUser.ID)
	userReqLock.Unlock()

	ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, reqUser.Username + " Do you accept the adoption? if so type ``wrig adopt accept``, ``wrig adopt decline`` otherwise.")
	return
}

//This should always be called in a goroutine. Creates an intentional "race" so that after 15 seconds the adoption times out.
func timeoutAdoption(key string) {
	time.Sleep(time.Second*15)
	fmt.Println("No response in 15 seconds, adoption aborting.")
	delete(AList, key)
	return
}

func listPets(ctx context) {
	rows, err := DataStore.Query("SELECT FriendlyName FROM pettable WHERE OwnerID = ?", ctx.Msg.Author.ID)
	if err != nil {
		return
	}

	var buffer bytes.Buffer

	for rows.Next() {
		var petName string
		rows.Scan(&petName)
		if err != nil {
			continue
		}
		buffer.WriteString("\n" + petName)
	}

	var listEmbed discordgo.MessageEmbed

	embedAuthor := discordgo.MessageEmbedAuthor{URL: "", Name: "WriggleBot", IconURL: "https://discordapp.com/api/v6/users/209739190244474881/avatars/47ada5c68c51f8dc2360143c0751d656.jpg"}
	petList     := discordgo.MessageEmbedField{Name: ctx.Msg.Author.Username + "'s pets", Value: buffer.String(), Inline: true}

	listEmbed.Author = &embedAuthor
	listEmbed.Color  = 14030101
	listEmbed.Fields = []*discordgo.MessageEmbedField{&petList}

	ctx.Session.ChannelMessageSendEmbed(ctx.Msg.ChannelID, &listEmbed)
}

func showAdoptions(ctx context) {
	var buffer bytes.Buffer
	
	for key, value := range AList {
		buffer.WriteString("\n" + "AList[" + key + "]" + value.Username)
	}
	
	ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, buffer.String())
}
