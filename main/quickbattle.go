package main

import(
  "fmt"
)

type pet {
  petUser *discordgo.User
  string OwnerID 
  string OwnerName
  int EffectiveATK
  int EffectiveDEF
  int EffectiveHP
  int EffectiveCRI
  int EffectiveEVA
  int EffectiveACC
}

func quickbattle(ctx Context) {
  if len(ctx.Args) < 2 {
    return
  }
}

//Use this to get users from the database for battle, search for designates who to Search discord for
//0 - Don't search; 1 - Search for Friendly; 2 - Search for enemy; 3 - Search for  both.
//Search in any case that a snowflake user is not provided.
func getbattleusers(string friendly, string enemy, int searchfor) (friendly pet, enemy pet, err error){
  
}