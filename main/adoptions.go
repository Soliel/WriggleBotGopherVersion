/*
TODO: Implement a timeout for adoptions, and clear the user out of the map
TODO: During user Selection check for adoption/adopting status
 */
package main

import (
	"fmt"
	"strings"
	"time"
)

func AdoptUsers(ctx Context) {

	if len(ctx.Args) < 1 {
		_, _ = ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "You must specify a user to adopt.")
		return
	}


	if ctx.Args[0] == "decline" {
		_, _ = ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "You were not adopted.")
		delete(AList, ctx.Msg.Author.ID)
 		return
	}
  
	if ctx.Args[0] == "accept" {
		if AList[ctx.Msg.Author.ID] == nil {
			return
		}

		tx, err := DataStore.Begin()
		if err != nil {
			ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "Transaction could not be started, Adoption Aborted.")
			delete(AList, ctx.Msg.Author.ID)
			fmt.Println(err)
			return
		}
		defer tx.Rollback()

		stmt, err := tx.Prepare("INSERT INTO pettable VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
		if err != nil {
			ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "Could not prepare SQL statement, adoption aborted.")
			delete(AList, ctx.Msg.Author.ID)
			fmt.Println(err)
			return
		}
		defer stmt.Close()
			
		_, err = stmt.Exec(ctx.Msg.Author.ID, ctx.Msg.Author.Username, 1, 10, 10, 20, 1, 5, 80, 0, 0, 0, 0, 0, 0, AList[ctx.Msg.Author.ID].ID)
		if err != nil {
			ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "Could not execute SQL statement with the database, adoption aborted.")
			delete(AList, ctx.Msg.Author.ID)
			fmt.Println(err)
			return
		}
			
		dupOwn := checkDuplicateOwners(AList[ctx.Msg.Author.ID].ID)
		if !dupOwn {
			stmt, _ = tx.Prepare("INSERT INTO ownertable VALUES(?,?,?,?)")
			stmt.Exec(AList[ctx.Msg.Author.ID].ID, AList[ctx.Msg.Author.ID].Username, 1, 1)
		}
			
		if dupOwn {
			
			var petamnt int
			
			err = DataStore.QueryRow("SELECT PetAmount FROM ownertable WHERE UserID = ?", AList[ctx.Msg.Author.ID].ID).Scan(&petamnt)
			if err != nil {
				ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "SQL Query failed to find Owner, adoption aborting.")
				return
			}
			
			stmt, err = tx.Prepare("UPDATE ownertable SET PetAmount = ? WHERE UserID = ?")
			if err != nil {
				ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "Could not prepare SQL statement, adoption aborted.")
				return
			}
			fmt.Println(stmt)
			
			res, err := stmt.Exec(petamnt + 1, AList[ctx.Msg.Author.ID].ID)
			if err != nil {
				fmt.Println(err)
				ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "Could not execute SQL statement with the database, adoption aborted.")
				return
			}
			fmt.Println(res)
		}
		
		err = tx.Commit()
		if err != nil {
			ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "Could not commit changes to database, adoption aborted.")
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
	
	//Search for user by ID in case they entered one.
	pet_user, err := ctx.Session.User(ctx.Args[0])
	if err == nil {
		dupErr := checkDuplicatePets(pet_user.ID)
		if dupErr != nil {
			ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, dupErr.Error())
			return
		}
		AList[pet_user.ID] = ctx.Msg.Author
		go timeoutAdoption(pet_user.ID)
		ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, pet_user.Username + " Do you accept the adoption? if so type ``wrig adopt accept``, ``wrig adopt decline`` otherwise.")
	}
	
	//Start a query for the user, Causes Guild Member Chunk event to fire. 
	if err != nil {
		fmt.Println("User could not be parsed with ID, showing error, ", err)
		fmt.Println("Attempting to locate user by name")
		
		//We lock the function with UserReqLock to ensure it is the only one being requested.
		userReqLock.Lock()
		reqUser, err := requestUserFromGuild(ctx)
		if err != nil {
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
	return
}

//This should always be called in a goroutine. Creates an intentional "race" so that after 15 seconds the adoption times out.
func timeoutAdoption(key string) {
	time.Sleep(time.Second*15)
	fmt.Println("No response in 15 seconds, adoption aborting.")
	delete(AList, key)
	return
}
