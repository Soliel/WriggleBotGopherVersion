/* This file holds helper functions for playing with pet stats and commands
to allow pets and users to view, set, and reset their stat points
STATS formulae 
EffectiveATK = AttribATK * (1 + (0.02* TrainedATK) | AttribATK + 2 /lvl
EffectiveDEF = AttribDEF + TrainedDEF | Damage % = 100/(100 + EffectiveDEF) | AttribDEF = .25 /lvl
EffectiveHP  = AttribHP * (0.0013 * TrainedDEF) | AttribHP + 10 /lvl
EffectiveCRI = (AttribLCK * 0.04) + (TrainedCRI * 0.08) | Crit % = EffectiveCRI/100 | .25 /lvl
EffectiveEVA = AttribEVA + TrainedEVA | Dodge % = EffectiveEVA / 100 | 0 /lvl
EffectiveACC = AttribACC + TrainedACC | Hit %   = EffectiveACC / 100 | 0 /lvl

Experience Formulae
Winner Experience = (10 * (enemyLevel ^ 1.2))/(1 + (enemyLevel ^.1162))
Loser  Experience = Winner Experience/2

Level Requirement = 10 * (level ^ 1.2)
*/
package main

import (
	"github.com/bwmarrin/discordgo"
	"strconv"
)

func petStatSheet (ctx context) {
	if  len(ctx.Args) < 1 {
		return
	}
	
	reqPet, err := getPetUser(ctx.Args[0], ctx)
	if err != nil {
		return
	}
	
	embed := &discordgo.MessageEmbed{
		Title: reqPet.Username + "'s stat sheet",
		Color: 14030101,
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: "https://discordapp.com/api/v6/users/" + reqPet.ID + "/avatars/" + reqPet.Avatar + ".jpg", ProxyURL:"", Width:0, Height:0},
		Author: &discordgo.MessageEmbedAuthor{URL: "", Name: "WriggleBot", IconURL: "https://discordapp.com/api/v6/users/209739190244474881/avatars/47ada5c68c51f8dc2360143c0751d656.jpg"},
		Fields: []*discordgo.MessageEmbedField{
			{"Level", strconv.FormatInt(int64(reqPet.Level), 10), true},
			{"Experience", strconv.FormatFloat(reqPet.Experience, 'f', 2, 64), true},
			{"Attack Attribute", strconv.FormatFloat(reqPet.EffectiveATK, 'f', 0, 64), true},
			{"Attack Level", strconv.FormatFloat(reqPet.TrainedATK, 'f', 0, 64), true},
			{"Defense Attribute", strconv.FormatFloat(reqPet.EffectiveDEF, 'f', 0, 64), true},
			{"Defense Level", strconv.FormatFloat(reqPet.TrainedDEF, 'f', 0, 64), true},
			{"Health Attribute", strconv.FormatFloat(reqPet.AttribHP, 'f', 0, 64), true},
			{"Health Points", strconv.FormatFloat(reqPet.EffectiveHP, 'f', 0, 64), true},
			{"Luck Attribute", strconv.FormatFloat(reqPet.EffectiveCRI, 'f', 0, 64), true},
			{"Crit Level", strconv.FormatFloat(reqPet.TrainedCRI, 'f',0, 64), true},
			{"Evasion Attribute", strconv.FormatInt(int64(reqPet.EffectiveEVA), 10), true},
			{"Evasion Level", strconv.FormatInt(int64(reqPet.TrainedEVA), 10), true},
			{"Accuracy Attribute", strconv.FormatInt(int64(reqPet.EffectiveACC), 10), true},
			{"Accuracy Level", strconv.FormatInt(int64(reqPet.TrainedACC), 10), true},
		},
	}

	ctx.Session.ChannelMessageSendEmbed(ctx.Msg.ChannelID, embed)
}