package main

import (
	"fmt"
	"time"
	"math"
)


//This will be used by Owners to "Train" their pets. This method should lower desire for self battles
//Command will be invoked by "wrig train <pet> <stat>, Training makes this more of a facebook game."
func trainStat(ctx context) {

	petUser, err := getPetUser(ctx.Args[0], ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	petStatMap := map[string]float64{
		"attack":   petUser.TrainedATK, "atk": petUser.TrainedATK,
		"defense":  petUser.TrainedDEF, "def": petUser.TrainedDEF,
		"evasion":  float64(petUser.TrainedEVA), "eva": float64(petUser.TrainedEVA),
		"accuracy": float64(petUser.TrainedACC), "acc": float64(petUser.TrainedACC),
		"crit":     petUser.TrainedCRI, "cri": petUser.TrainedCRI,
	}

	statMap := map[string]string {
		"attack":   "TrainedATK", "atk": "TrainedATK",
		"defense":  "TrainedDEF", "def": "TrainedDEF",
		"evasion":  "TrainedEVA", "eva": "TrainedEVA",
		"accuracy": "TrainedACC", "acc": "TrainedACC",
		"crit":     "TrainedCRI", "cri": "TrainedCRI",
	}

	if len(ctx.Args) < 2 {
		return
	}

	if statMap[ctx.Args[1]] == "" {
		return
	}

	//fmt.Println("Starting Transaction")
	tx, err := DataStore.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	//Crazy math...
	trainingDuration := time.Second * time.Duration(math.Pow(1.058473, petStatMap[ctx.Args[1]]) + 14.00)
	trainingCompletion := time.Now().Add(trainingDuration)


	_, err = tx.Exec("INSERT INTO trainingtbl VALUES(?,?,?)", petUser.ID, statMap[ctx.Args[1]], trainingCompletion.Format(time.RFC822))
	if err != nil {
		fmt.Println(err)
		return
	}

	err = tx.Commit()
	if err != nil {
		return
	}

	trainingMap[trainingCompletion] = training{TPet: petUser.ID, StatArg: statMap[ctx.Args[1]], newNum: petStatMap[ctx.Args[1]] + 1}

	_,_ = ctx.Session.ChannelMessageSend(ctx.Msg.ChannelID, "Training has begun! It will complete in: " + trainingDuration.String())
}

//This function does all the work for the training timer.
//call this in main as a goroutine.
//Anytime this Errors, gracefully exit to avoid a panic and restart it.
func startTickReceiver() {
	fmt.Println("Initializing types and maps...")

	trainingMap = make(map[time.Time]training)

	fmt.Println("Loading training tasks from the database...")
	rows, err := DataStore.Query("SELECT * FROM trainingtbl")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			//Nothing to load!
		} else {
			go startTickReceiver()
			return
		}
	}

	//Loading our database into the function
	for rows.Next() {
		var strtime, stat, petID string
		rows.Scan(&petID, &stat, &strtime)
		if err != nil {
			continue
		}
		trainingTask := training{TPet: petID, StatArg: stat}
		t, _ := time.Parse(time.RFC822, strtime)
		trainingMap[t] = trainingTask
	}

	fmt.Println("Receiving ticks!")
	for tick := range tTick.C{
		for key, value := range trainingMap {
			if tick.Sub(key) >= 0 {

				_, err = DataStore.Exec("UPDATE pettable SET " + value.StatArg + " = ? WHERE UserID = ?", value.newNum, value.TPet)
				if err != nil {
					fmt.Println(err)
					go startTickReceiver()
					return
				}

				_, err = DataStore.Exec("DELETE FROM trainingtbl WHERE petID = ?", value.TPet)
				if err != nil {
					fmt.Println(err)
					go startTickReceiver()
					return
				}

				delete(trainingMap, key)
			}
		}
	}
}