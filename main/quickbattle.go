package main

import(
	"fmt"
	"github.com/bwmarrin/discordgo"
)

type Pet struct {
	PetUser      *discordgo.User
	OwnerID      string
	EffectiveATK int
	EffectiveDEF int
	EffectiveHP  int
	EffectiveCRI int
	EffectiveEVA int
	EffectiveACC int
	Experience   int
	Level        int
}

func QuickBattle(ctx Context) {
	if len(ctx.Args) < 2 {
		return
	}

	var allypet, enemypet Pet

	allypet, err := getPetUser(ctx.Args[0], ctx)
	if err != nil {
		fmt.Println("Could not get ally pet")
		return
	}
	fmt.Println(allypet)

	enemypet, err = getPetUser(ctx.Args[1], ctx)
	if err != nil {
		fmt.Println("Could not get enemy pet")
		return
	}
	fmt.Println(enemypet)

	_,_ = ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "User lookup complete.")
}

//This helps get information from the database and constructs it into a 'Pet'
func getPetUser(userarg string, ctx Context) (Pet, error) {
	pet_user := new(Pet)
	var (
		Level, AttribATK, AttribDEF, AttribHP,
		AttribLCK, AttribEVA, AttribACC,
		TrainedATK, TrainedDEF, TrainedCRI,
		TrainedEVA, TrainedACC, Experience int
		OwnerID, PetID, PetName string
	)

	PetUser, err := ctx.Session.User(userarg)
	pet_user.PetUser = PetUser

	if err != nil {
		userReqLock.Lock()

		pet_user.PetUser, err = requestUserFromGuild(ctx.Session, ctx.Guild.ID, userarg)
		if err != nil {
			_,_ = ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "One of the Users could not be found")
			userReqLock.Unlock()
			return *pet_user, err
		}

		userReqLock.Unlock()
	}

	err = DataStore.QueryRow("SELECT * FROM pettable WHERE UserID = ?", pet_user.PetUser.ID).Scan(&PetID , &PetName, &Level, &AttribATK, &AttribDEF, &AttribHP,
													&AttribLCK, &AttribEVA, &AttribACC, &TrainedATK,
													&TrainedDEF, &TrainedCRI, &TrainedEVA, &TrainedACC,
													&Experience, &OwnerID)

	if err != nil {
		return *pet_user, err
	}

	//Stand in math for effective stats
	pet_user.EffectiveATK = AttribATK + TrainedATK
	pet_user.EffectiveDEF = AttribDEF + TrainedDEF
	pet_user.EffectiveHP  = AttribHP
	pet_user.EffectiveCRI = AttribLCK + TrainedCRI
	pet_user.EffectiveEVA = AttribEVA + TrainedEVA
	pet_user.EffectiveACC = AttribACC + TrainedACC
	pet_user.OwnerID      = OwnerID
	pet_user.Experience   = Experience
	pet_user.Level        = Level

	return *pet_user, nil
}
