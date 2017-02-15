package main

import (
	"github.com/bwmarrin/discordgo"
	//"fmt"
	"errors"
	"time"
)

//Defines a pet, lots of information
type pet struct {
	PetUser      *discordgo.User
	OwnerID      string
	EffectiveATK float64; AttribATK    float64
	EffectiveDEF float64; AttribDEF    float64
	EffectiveHP  float64; AttribHP     float64
	EffectiveCRI float64; AttribLCK    float64
	DMGCount     float64; AttribEVA    int
	EffectiveEVA int;     AttribACC    int
	EffectiveACC int;     TrainedCRI   float64
	TrainedATK   float64; TrainedEVA   int
	TrainedDEF   float64; TrainedACC   int
	CritCount    int64;   MissCount    int64
	Experience   int;     Level        int
}

//Called when requestUserFromGuild is called, it returns a guildMemberChunk asynchronously, uses a channel to return it's value back to it's requester.
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

//This helps get information from the database and constructs it into a 'Pet'
func getPetUser(userid string, ctx context) (pet, error) {

	PetUser, err := ctx.Session.User(userid)
	reqPet, err := getPetFromDB(userid)

	//if asking for a user by snowflake ID fails, request a user from the guild list.
	if err != nil {
		userReqLock.Lock()	
		petUser, err = requestUserFromGuild(ctx.Session, ctx.Guild.ID, userid)
		if err != nil {
			_,_ = ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "One of the Users could not be found")
			userReqLock.Unlock()
			return *reqPet, err
		}
		reqPet, err = getPetFromDB(PetUser.ID)
		userReqLock.Unlock()
	}
	
	reqPet.PetUser = PetUser
	return *reqPet, nil
}

func getPetFromDB(petID string) (pet, error) {
	reqPet := new(pet)
	
	var (
		AttribATK, AttribDEF, AttribHP, AttribLCK, 
		TrainedATK, TrainedDEF, TrainedCRI float64
		OwnerID, PetID, PetName string
		Level, Experience, AttribEVA, AttribACC, TrainedEVA, TrainedACC int
	)
	
	//Grab the pet from the database based on userID
	err = DataStore.QueryRow("SELECT * FROM pettable WHERE UserID = ?", petID).Scan(&PetID , &PetName, &Level, &AttribATK, &AttribDEF, &AttribHP,
													&AttribLCK, &AttribEVA, &AttribACC, &TrainedATK,
													&TrainedDEF, &TrainedCRI, &TrainedEVA, &TrainedACC,
													&Experience, &OwnerID)

	if err != nil {
		return *reqPet, err
	}

	//Stand in math for effective stats
	reqPet.EffectiveATK = AttribATK*(1+(0.2 * TrainedATK))
	reqPet.EffectiveDEF = AttribDEF + TrainedDEF
	reqPet.EffectiveHP  = AttribHP * (1 + (.0013 * TrainedDEF))
	reqPet.EffectiveCRI = (0.04*AttribLCK) + (0.08 * TrainedCRI)
	reqPet.EffectiveEVA = AttribEVA + TrainedEVA
	reqPet.EffectiveACC = AttribACC + TrainedACC
	reqPet.AttribATK    = AttribATK
	reqPet.AttribDEF    = AttribDEF
	reqPet.AttribHP     = AttribHP
	reqPet.AttribLCK    = AttribLCK
	reqPet.AttribACC    = AttribACC
	reqPet.AttribEVA    = AttribEVA
	reqPet.TrainedATK   = TrainedATK
	reqPet.TrainedDEF   = TrainedDEF
	reqPet.TrainedCRI   = TrainedCRI
	reqPet.TrainedACC   = TrainedACC
	reqPet.TrainedEVA   = TrainedEVA
	reqPet.OwnerID      = OwnerID
	reqPet.Experience   = Experience
	reqPet.Level        = Level
	
	return *reqPet, nil
}