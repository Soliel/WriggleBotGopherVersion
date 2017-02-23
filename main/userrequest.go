package main

import (
	"github.com/bwmarrin/discordgo"
	//"fmt"
	"errors"
	"time"
	"math"
	"strings"
)

//Defines a pet, lots of information
type pet struct {
	Username     string;  OwnerID      string
	Avatar       string;  ID           string
	EffectiveATK float64; AttribATK    float64
	EffectiveDEF float64; AttribDEF    float64
	EffectiveHP  float64; AttribHP     float64
	EffectiveCRI float64; AttribLCK    float64
	DMGCount     float64; AttribEVA    float64
	EffectiveEVA float64; AttribACC    float64
	EffectiveACC float64; TrainedCRI   float64
	TrainedATK   float64; TrainedEVA   float64
	TrainedDEF   float64; TrainedACC   float64
	CritCount    int64;   MissCount    int64
	SwingCount   int64;   Training     bool
	Experience   float64; Level        int
}

type owner struct {
	ID        string
	Username  string
	PetAmount int
	OptedOut  bool
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

func checkDuplicatePets(ID string) (isowned bool, e error){
	var OID string
	err := DataStore.QueryRow("SELECT OwnerID FROM pettable WHERE UserID = ?", ID).Scan(&OID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {return false, errors.New("Pet does not exist")}
		return false, err
	}

	if OID != "" {
		return true, nil
	}

	return  false, nil
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

	var PetUser *discordgo.User
	var reqPet  pet

	userReqLock.Lock()
	PetUser, err := requestUserFromGuild(ctx.Session, ctx.Guild.ID, userid)
	if err != nil {
		var modifiedID string

		if strings.HasPrefix(userid, "<@") {
			modifiedID = strings.Trim(userid, "<@!>")
		}

		PetUser, err = ctx.Session.User(modifiedID)
		if err != nil {
			userReqLock.Unlock()
			return reqPet, err
		}
	}

	userReqLock.Unlock()
	reqPet, err = getPetFromDB(PetUser.ID)

	if PetUser.Avatar == "" {
		reqPet.Avatar = "https://discordapp.com/assets/dd4dbc0016779df1378e7812eabaa04d.png" //Sets the avatar data to discords default if we do not have any data.
	} else {
		reqPet.Avatar   = "https://discordapp.com/api/v6/users/" + PetUser.ID + "/avatars/" + PetUser.Avatar + ".jpg"
	}

	reqPet.Username = PetUser.Username
	reqPet.ID       = PetUser.ID

	return reqPet, nil
}

func getPetFromDB(petID string) (pet, error) {
	reqPet := new(pet)

	var (
		AttribATK, AttribDEF, AttribHP, AttribLCK, AttribACC,
		TrainedATK, TrainedDEF, TrainedCRI, Experience, AttribEVA, TrainedEVA, TrainedACC float64
		OwnerID, PetID, PetName string
		Level int
		Training bool
	)

	//Grab the pet from the database based on userID
	err := DataStore.QueryRow("SELECT * FROM pettable WHERE UserID = ?", petID).Scan(&PetID , &PetName, &Level, &AttribATK, &AttribDEF, &AttribHP,
													&AttribLCK, &AttribEVA, &AttribACC, &TrainedATK,
													&TrainedDEF, &TrainedCRI, &TrainedEVA, &TrainedACC,
													&Experience, &OwnerID, &Training)

	if err != nil {
		return *reqPet, err
	}

	//Stand in math for effective stats
	reqPet.EffectiveATK = AttribATK*(1+(0.002 * TrainedATK))
	reqPet.EffectiveDEF = AttribDEF + TrainedDEF
	reqPet.EffectiveHP  = AttribHP * (1 + (.0045 * TrainedDEF))
	reqPet.EffectiveCRI = (0.04*AttribLCK) + (0.08 * TrainedCRI)
	reqPet.EffectiveEVA = AttribEVA + (400/( 1 + math.Pow(math.E, -(.013 * (TrainedEVA-200))))) - 27.655
	reqPet.EffectiveACC = AttribACC + (400/( 1 + math.Pow(math.E, -(.013 * (TrainedACC-200))))) - 27.655
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
	reqPet.Training     = Training
	
	return *reqPet, nil
}

//Function will be useful for future purposes

func getOwnerFromDB(OwnerID string) (owner, error) {
	var(
		ID, Username     string
		PetAmount int
		OptedOut bool
	)

	var reqOwner owner
	
	err := DataStore.QueryRow("SELECT * FROM ownertable WHERE UserID = ?", OwnerID).Scan(&ID, &Username, &PetAmount, &OptedOut)

	if err != nil{
		return reqOwner, err
	}

	return owner{ID: ID, Username: Username, PetAmount: PetAmount, OptedOut: OptedOut}, nil
}