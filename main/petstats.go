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


func petStatSheet (ctx context) {
	if  len(ctx.Args) < 1 {
		return
	}
	
	reqPet, err := getPetUser(ctx.Args[0])
	if err != nil {
		return
	}
	
	embed := *discordgo.MessageEmbed{
		Title: reqPet.Username + "'s stat sheet"
		Color: 14030101
	}
}