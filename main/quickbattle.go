package main

import(

	"github.com/bwmarrin/discordgo"
	"math/rand"
	"strconv"
)

//Entry point for the "quickbattle" or "q" command
func quickBattle(ctx context) {
	//Verify that we have enough args
	if len(ctx.Args) < 2 {
		return
	}

	//initialize pet objects to store information
	var allypet, enemypet pet

	//fill up the ally user with the information from the database
	allypet, err := getPetUser(ctx.Args[0], ctx)
	if err != nil {
		return
	}

	//fill up the enemy user
	enemypet, err = getPetUser(ctx.Args[1], ctx)
	if err != nil {
		return
	}
	//check if ally is owned by the command initiator
	if allypet.OwnerID != ctx.Msg.Author.ID {
		return
	}

	//Set the battling flag
	doingBattle := true
	
	
	//Entry point for battle loop. Ally attacks first.
	for doingBattle {
		if allypet.EffectiveHP <= 0 {
			_,_ = ctx.Session.ChannelMessageSendEmbed(ctx.Msg.ChannelID, createResultEmbed(enemypet, allypet))
			doingBattle = false
			return
		}
		
		if doesHit(&allypet, enemypet) {
			dmg := getDamage(&allypet, enemypet)
			enemypet.EffectiveHP -= dmg
			allypet.DMGCount += dmg
		}
			
		if enemypet.EffectiveHP <= 0 {
			_,_ = ctx.Session.ChannelMessageSendEmbed(ctx.Msg.ChannelID, createResultEmbed(allypet, enemypet))
			doingBattle = false
			return
		}
			
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
	chanceToHit := attacker.EffectiveACC - defender.EffectiveEVA
	
	if rand.Intn(100) < chanceToHit {
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
	var winnerName, loserName, winnerMissCount, winnerCritCount, winnerDamageCount, loserMissCount, loserCritCount, loserDamageCount discordgo.MessageEmbedField
	var resultEmbed discordgo.MessageEmbed

	winnerThumb := discordgo.MessageEmbedThumbnail{URL: "https://discordapp.com/api/v6/users/" + winner.ID + "/avatars/" + winner.Avatar + ".jpg", ProxyURL:"", Width:0, Height:0}

	embedAuthor := discordgo.MessageEmbedAuthor{URL: "", Name: "WriggleBot", IconURL: "https://discordapp.com/api/v6/users/209739190244474881/avatars/47ada5c68c51f8dc2360143c0751d656.jpg"}

	winnerName.Name = "Winner"
	winnerName.Value = winner.Username
	winnerName.Inline = true
	
	loserName.Name = "Loser"
	loserName.Value = loser.Username
	loserName.Inline = true
	
	winnerMissCount.Name = winner.Username + " misses"
	winnerMissCount.Value = strconv.FormatInt(winner.MissCount, 10)
	winnerMissCount.Inline = true
	
	loserMissCount.Name = loser.Username  + " misses"
	loserMissCount.Value = strconv.FormatInt(loser.MissCount, 10)
	loserMissCount.Inline = true
	
	winnerCritCount.Name = winner.Username + " crits"
	winnerCritCount.Value = strconv.FormatInt(winner.CritCount, 10)
	winnerCritCount.Inline = true
	
	loserCritCount.Name = loser.Username  + " crits"
	loserCritCount.Value = strconv.FormatInt(loser.CritCount, 10)
	loserCritCount.Inline = true
	
	winnerDamageCount.Name = winner.Username + " damage dealt"
	winnerDamageCount.Value = strconv.FormatFloat(winner.DMGCount, 'f', 2, 64)
	winnerDamageCount.Inline = true
	
	loserDamageCount.Name = loser.Username + " damage dealt"
	loserDamageCount.Value = strconv.FormatFloat(loser.DMGCount, 'f', 2, 64)
	loserDamageCount.Inline = true

	resultEmbed.Author = &embedAuthor
	resultEmbed.Thumbnail = &winnerThumb
	resultEmbed.Fields = []*discordgo.MessageEmbedField{&winnerName, &loserName, &winnerMissCount, &loserMissCount, &winnerCritCount, &loserCritCount, &winnerDamageCount, &loserDamageCount}
	resultEmbed.Color = 14030101

	return &resultEmbed
}
