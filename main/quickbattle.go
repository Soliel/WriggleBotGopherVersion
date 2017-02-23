package main

import(

	"github.com/bwmarrin/discordgo"
	"math/rand"
	"strconv"
	"math"
	"errors"
)

//Entry point for the "quickbattle" or "q" command
func quickBattle(ctx context) {
	//Verify that we have enough args
	if len(ctx.Args) < 2 {
		return
	}

	//fill up the ally user with the information from the database
	allypet, err := getPetUser(ctx.Args[0], ctx)
	if err != nil {
		return
	}

	//fill up the enemy user
	enemypet, err := getPetUser(ctx.Args[1], ctx)
	if err != nil {
		return
	}
	//check if ally is owned by the command initiator

	if allypet.OwnerID != ctx.Msg.Author.ID {
		return
	}

	if allypet.Training {
		return
	}
	
	if enemypet.Training {
		return
	}

	//Set the battling flag
	doingBattle := true
	
	
	//Entry point for battle loop. Ally attacks first.
	for doingBattle {
		if allypet.EffectiveHP <= 0 {
			_,_ = ctx.Session.ChannelMessageSendEmbed(ctx.Msg.ChannelID, createResultEmbed(enemypet, allypet))
			doingBattle = false

			getLevels(&allypet, &enemypet, false, ctx)

			return
		}

		allypet.SwingCount ++
		if doesHit(&allypet, enemypet) {
			dmg := getDamage(&allypet, enemypet)
			enemypet.EffectiveHP -= dmg
			allypet.DMGCount += dmg
		}
			
		if enemypet.EffectiveHP <= 0 {
			_,_ = ctx.Session.ChannelMessageSendEmbed(ctx.Msg.ChannelID, createResultEmbed(allypet, enemypet))
			doingBattle = false

			getLevels(&allypet, &enemypet, true, ctx)
			return
		}

		enemypet.SwingCount ++
		if doesHit(&enemypet, allypet) {
			dmg := getDamage(&enemypet, allypet)
			allypet.EffectiveHP -= dmg
			enemypet.DMGCount += dmg
		}	
	}
}

//calculates the damage dealt by the attacking pet
func getDamage(attacker *pet, defender pet) (float64) {
	if doesCrit(attacker) {
		return 2*(attacker.EffectiveATK * (100.00/(defender.EffectiveDEF + 100.00)))
	}
	
	return attacker.EffectiveATK * (100.00/(defender.EffectiveDEF + 100.00))
}

//Rolls imaginary dice to see if the attack hits.
func doesHit(attacker *pet, defender pet) (bool){
	chanceToHit := float64(attacker.EffectiveACC) - defender.EffectiveEVA
	
	if float64(rand.Intn(100)) < chanceToHit {
		return true
	}
	
	//fmt.Println(attacker.PetUser.Username, " miss!")
	attacker.MissCount++
	return false
}

//Rolls imaginary dice to see if the attack crit
func doesCrit(attacker *pet) (bool) {
	critRand := float64(rand.Intn(100))
	
	if critRand < attacker.EffectiveCRI {
		//fmt.Println(attacker.PetUser.Username, " rolled a", critRand, " crit!")
		attacker.CritCount++
		return true
	}
	
	return false
	
}


//Creating the embed structure, Go throws an error when I put all the asignments for a field inline
func createResultEmbed(winner pet, loser pet) (*discordgo.MessageEmbed){

	resultEmbed := &discordgo.MessageEmbed{
		Title: "Battle Results",
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: winner.Avatar, ProxyURL:"", Width:0, Height:0},
		Author: &discordgo.MessageEmbedAuthor{URL: "", Name: "WriggleBot", IconURL: "https://discordapp.com/api/v6/users/209739190244474881/avatars/47ada5c68c51f8dc2360143c0751d656.jpg"},
		Color: 14030101,
		Fields: []*discordgo.MessageEmbedField {
			{"Winner", winner.Username, true},
			{"Loser", loser.Username, true},
			{winner.Username + " swings", strconv.FormatInt(winner.SwingCount, 10), true},
			{loser.Username + " swings", strconv.FormatInt(loser.SwingCount, 10), true},
			{winner.Username + " misses", strconv.FormatInt(winner.MissCount, 10), true},
			{loser.Username + " misses", strconv.FormatInt(loser.MissCount, 10), true},
			{winner.Username + " crits", strconv.FormatInt(winner.CritCount, 10), true},
			{loser.Username + " crits", strconv.FormatInt(loser.CritCount, 10), true},
			{winner.Username + " damage dealt", strconv.FormatFloat(winner.DMGCount, 'f', 2, 64), true},
			{loser.Username + " damage dealt", strconv.FormatFloat(loser.DMGCount, 'f', 2, 64), true},
		},
	}

	return resultEmbed
}


func calcExperience(defender pet) [2]float64 {
	winner_exp := (10.00 * (math.Pow(float64(defender.Level), 1.2)))/((math.Pow(float64(defender.Level), .1162)) + 1.00)
	loser_exp  := winner_exp/2.00
	return [2]float64{winner_exp, loser_exp}
}

//The super get everything done that has to do with leveling method. 
func getLevels(attacker *pet, defender *pet, won bool, ctx context) {
	exp := calcExperience(*defender)
	aLevelReq := 10.00 * math.Pow(float64(attacker.Level), 1.2)
	dLevelReq := 10.00 * math.Pow(float64(defender.Level), 1.2)

	if won {
		attacker.Experience += exp[0]
		defender.Experience += exp[1]
	} else {
		attacker.Experience += exp[1]
		defender.Experience += exp[0]
	}

	if attacker.Experience >= aLevelReq {
		attacker.Experience -= aLevelReq
		attacker.Level += 1
		err := doPetLevelUp(*attacker)
		if err != nil {
			ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "Result storage failed, battle will not be counted.")
			return
		}
		levelPM(*attacker, ctx)
	}

	if defender.Experience >= dLevelReq {
		defender.Experience -= dLevelReq
		defender.Level += 1
		err := doPetLevelUp(*defender)
		if err != nil {
			ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "Result storage failed, battle will not be counted.")
			return
		}
		levelPM(*defender, ctx)
	}

	tx, err := DataStore.Begin()
	if err != nil {
		ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "Result storage failed, battle will not be counted.")
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE pettable SET Experience = ? WHERE UserID = ?")
	if err != nil {
		ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "Result storage failed, battle will not be counted.")
	}
	defer stmt.Close()

	_, err = stmt.Exec(attacker.Experience, attacker.ID)
	if err != nil {
		ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "Result storage failed, battle will not be counted.")
	}

	_, err = stmt.Exec(defender.Experience, defender.ID)
	if err != nil {
		ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "Result storage failed, battle will not be counted.")
	}

	tx.Commit()
	return
}

func levelPM(leveledPet pet, ctx context) {
	levelOwner, err := getOwnerFromDB(leveledPet.OwnerID)
	if err != nil {
		return
	}

	if levelOwner.OptedOut {
		return
	}

	pmChannel, err := ctx.Session.UserChannelCreate(leveledPet.OwnerID)
	if err != nil {
		return
	}

	ctx.Session.ChannelMessageSend(pmChannel.ID, "Your pet, " + leveledPet.Username + " Has leveled up to: " + strconv.FormatInt(int64(leveledPet.Level), 10))
}

//This function will apply all the necessary levelup data to a pet
func doPetLevelUp(upPet pet) (error){
  	tx, err := DataStore.Begin()
  	if err != nil {
    		//TODO: Implement a backup levelup, maybe to call this function later.
    		return errors.New("Unable to do levelup.")
  	}
  	defer tx.Rollback()

  	_, err = tx.Exec("UPDATE pettable SET Level = ?, AttribATK = ?, AttribDEF = ?, AttribHP = ?, AttribLCK = ? WHERE UserID = ?", upPet.Level, upPet.AttribATK + 2, upPet.AttribDEF + 0.25, upPet.AttribHP + 10, upPet.AttribLCK + 0.25, upPet.ID)
  	if err != nil {
    		//TODO: Implement backup levelup, maybe to call this function later.
    		return errors.New("Unable to do levelup.")
  	}

	tx.Commit()
	return nil
}

func levelOptOut(ctx context) {
	if len(ctx.Args) < 1 {
		return
	}

	if ctx.Args[0] == "true" {
		DataStore.Exec("UPDATE ownertable SET OptedOut = TRUE WHERE UserID = ?", ctx.Msg.Author.ID)
	}

	if ctx.Args[0] == "false" {
		DataStore.Exec("UPDATE ownertable SET OptedOut = FALSE WHERE UserID = ?", ctx.Msg.Author.ID)
	}

	ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "You have been opted out of levelup PMs.")
}