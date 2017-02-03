package main

import(
	//"fmt"
	"github.com/bwmarrin/discordgo"
)

type pet struct {
	PetUser      *discordgo.User
	OwnerID      string
	EffectiveATK float64
	EffectiveDEF float64
	EffectiveHP  float64
	EffectiveCRI float64
	EffectiveEVA float64
	EffectiveACC float64
	Experience   int
	Level        int
}

func quickBattle(ctx context) {
	if len(ctx.Args) < 2 {
		return
	}

	var allypet, enemypet pet

	allypet, err := getPetUser(ctx.Args[0], ctx)
	if err != nil {
		return
	}

	enemypet, err = getPetUser(ctx.Args[1], ctx)
	if err != nil {
		return
	}


	doingBattle := true
	
	
		for doingBattle {
		if allypet.EffectiveHP <= 0 {
			_,_ = ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "You lost")
			doingBattle = false
			return
		}
		
		enemypet.EffectiveHP -= getDamage(allypet, enemypet)
		
		if enemypet.EffectiveHP <= 0 {
			_,_ = ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "You won!")
			doingBattle = false
			return
		}
		
		allypet.EffectiveHP -= getDamage(enemypet, allypet)
	}
}

//TODO: add Crit 
func getDamage(attacker pet, defender pet) (float64) {
	return attacker.EffectiveATK * (100.00/(defender.EffectiveDEF + 100.00))
}

//This helps get information from the database and constructs it into a 'Pet'
func getPetUser(userarg string, ctx context) (pet, error) {
	reqPet := new(pet)
	var (
		AttribATK, AttribDEF, AttribHP,
		AttribLCK, AttribEVA, AttribACC,
		TrainedATK, TrainedDEF, TrainedCRI,
		TrainedEVA, TrainedACC float64
		OwnerID, PetID, PetName string
		Level, Experience int
	)

	PetUser, err := ctx.Session.User(userarg)
	reqPet.PetUser = PetUser

	if err != nil {
		userReqLock.Lock()

		reqPet.PetUser, err = requestUserFromGuild(ctx.Session, ctx.Guild.ID, userarg)
		if err != nil {
			_,_ = ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "One of the Users could not be found")
			userReqLock.Unlock()
			return *reqPet, err
		}

		userReqLock.Unlock()
	}

	err = DataStore.QueryRow("SELECT * FROM pettable WHERE UserID = ?", reqPet.PetUser.ID).Scan(&PetID , &PetName, &Level, &AttribATK, &AttribDEF, &AttribHP,
													&AttribLCK, &AttribEVA, &AttribACC, &TrainedATK,
													&TrainedDEF, &TrainedCRI, &TrainedEVA, &TrainedACC,
													&Experience, &OwnerID)

	if err != nil {
		return *reqPet, err
	}

	//Stand in math for effective stats
	reqPet.EffectiveATK = AttribATK + TrainedATK
	reqPet.EffectiveDEF = AttribDEF + TrainedDEF
	reqPet.EffectiveHP  = AttribHP
	reqPet.EffectiveCRI = AttribLCK + TrainedCRI
	reqPet.EffectiveEVA = AttribEVA + TrainedEVA
	reqPet.EffectiveACC = AttribACC + TrainedACC
	reqPet.OwnerID      = OwnerID
	reqPet.Experience   = Experience
	reqPet.Level        = Level

	return *reqPet, nil
}
