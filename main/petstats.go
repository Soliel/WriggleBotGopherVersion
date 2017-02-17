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

import "errors"

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

