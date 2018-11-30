package main

import "math"

type address uintptr

type gameVariable struct {
	parentOffset     address
	offsets          []address
	previousVariable interface{}
	variable         interface{}
}

func (gv *gameVariable) Get() {
	gv.previousVariable = gv.variable
	var pointer address
	getValue(&pointer, exeBaseAddress+gv.parentOffset)
	for i := 0; i < len(gv.offsets)-1; i++ {
		getValue(&pointer, pointer+gv.offsets[i])
	}
	getValue(&gv.variable, pointer+gv.offsets[len(gv.offsets)-1])
	switch gv.variable.(type) {
	case int:
		gv.variable = int(gv.variable.(int))
	case float32:
		gv.variable = float32(gv.variable.(float32))
	case float64:
		gv.variable = float64(gv.variable.(float64))
	case bool:
		gv.variable = bool(gv.variable.(bool))
	case string:
		gv.variable = string(gv.variable.(string))
	case address:
		gv.variable = address(gv.variable.(address))
	case uintptr:
		gv.variable = uintptr(gv.variable.(uintptr))
	}
}

func (gv *gameVariable) Reset(i interface{}) {
	gv.variable = i
	gv.previousVariable = i
}

func (gv *gameVariable) GetPreviousVariable() interface{} {
	return gv.previousVariable
}

func (gv *gameVariable) GetVariable() interface{} {
	return gv.variable
}

type gameStringVariable struct {
	stringVariable gameVariable
	variable       string
}

// maxSize is used to check the maximum size of the string array in dd.
// if maxSize is 15, the pointer points directly to the char array
// if maxSize is 31 the char array holds the address of where the new
// string is stored. the offset of the maxSize is always 0x14.
func (gsv *gameStringVariable) Get() {
	lengthOffset := gsv.stringVariable.offsets[0] + 0x10
	lengthVariable := gameVariable{parentOffset: gameStatsAddress, offsets: []address{lengthOffset}, variable: 0}
	lengthVariable.Get()
	length := lengthVariable.variable

	maxSizeOffset := gsv.stringVariable.offsets[0] + 0x14
	maxSizeVariable := gameVariable{parentOffset: gameStatsAddress, offsets: []address{maxSizeOffset}, variable: 0}
	maxSizeVariable.Get()
	maxSize := maxSizeVariable.variable.(int)

	iterations := int(math.Log2(float64(maxSize)+1)) - 4

	for i := 0; i < iterations; i++ {
		gsv.stringVariable.offsets = append(gsv.stringVariable.offsets, 0x0)
	}
	gsv.stringVariable.variable = string(make([]byte, length.(int)))
	gsv.stringVariable.Get()

	gsv.variable = gsv.stringVariable.variable.(string)[:length.(int)]
	gsv.stringVariable.variable = ""
}

func (gsv *gameStringVariable) GetVariable() interface{} {
	return string(gsv.variable)
}

type gameReplayIDVariable struct {
	replayIDVariable gameVariable
	variable         int
}

func (gridv *gameReplayIDVariable) GetVariable() interface{} {
	return int(gridv.variable)
}

func (gridv *gameReplayIDVariable) Get() {
	gridv.replayIDVariable.Get()
	gridv.variable = gridv.replayIDVariable.variable.(int)
	gridv.replayIDVariable.variable = "XXXXXX"
}

type gameCapture struct {
	status              status
	isAlive             bool
	isReplay            bool
	playerJustDied      bool
	playerID            int
	playerName          string
	deathType           int
	replayPlayerID      int
	replayPlayerName    string
	timer               float32
	gems                int
	totalGems           int
	totalGemsAtDeath    int
	level2time          float32
	level3time          float32
	level4time          float32
	homing              int
	homingAtDeath       int
	homingMax           int
	homingMaxTime       float32
	enemiesAlive        int
	enemiesAliveMax     int
	enemiesAliveMaxTime float32
	enemiesKilled       int
	daggersFired        int
	daggersHit          int
	accuracy            float64
}

func (gc *gameCapture) ResetGameVariables() {
	gc.deathType = 0
	deathType.Reset(0)
	gc.timer = 0.0
	timer.Reset(0.0)
	gc.gems = 0
	gems.Reset(0)
	gc.totalGems = 0
	gc.totalGemsAtDeath = 0
	totalGems.Reset(0)
	gc.level2time = 0.0
	gc.level3time = 0.0
	gc.level4time = 0.0
	gc.homing = 0
	gc.homingAtDeath = 0
	homing.Reset(0)
	gc.homingMax = 0
	gc.homingMaxTime = 0.0
	gc.enemiesAlive = 0
	enemiesAlive.Reset(0)
	gc.enemiesAliveMax = 0
	gc.enemiesAliveMaxTime = 0.0
	gc.enemiesKilled = 0
	enemiesKilled.Reset(0)
	gc.daggersFired = 0
	daggersFired.Reset(0)
	gc.daggersHit = 0
	daggersHit.Reset(0)
	gc.accuracy = 0.0
}

func (gc *gameCapture) GetPlayerVariables() {
	if handle != 0 {
		playerID.Get()
		playerName.Get()

		gc.playerID = playerID.GetVariable().(int)
		gc.playerName = playerName.GetVariable().(string)
	}
}

func (gc *gameCapture) GetReplayPlayerVariables() {
	if handle != 0 {
		replayPlayerID.Get()
		replayPlayerName.Get()

		gc.replayPlayerID = replayPlayerID.GetVariable().(int)
		gc.replayPlayerName = replayPlayerName.GetVariable().(string)
	}
}

func (gc *gameCapture) ResetReplayPlayerVariables() {
	gc.replayPlayerID = 0
	gc.replayPlayerName = ""
}

func (gc *gameCapture) GetStatus() status {
	return gc.status
}

// This function is heavily commented because of the weird nature
// of reading how another program's memory is working. The logic
// is a bit squirrely.
func (gc *gameCapture) GetGameVariables() {

	// Get the handle before retrieving any variables. This helps ensure that the
	// user hasn't closed dd in the interim.
	getHandle()

	// If the game is not found, reset the playerName to ""
	// This is to make sure that 'connecting' is appropriately set
	// on reopening dd.exe multiple times.
	if handle == 0 {

		gc.status = statusNotConnected
		gc.playerName = ""

	} else {

		isAlive.Get()
		isReplay.Get()
		timer.Get()
		gems.Get()
		totalGems.Get()
		homing.Get()
		enemiesAlive.Get()
		enemiesKilled.Get()
		daggersFired.Get()
		daggersHit.Get()

		gc.isAlive = isAlive.GetVariable().(bool)

		if !isAlive.GetVariable().(bool) && isAlive.GetPreviousVariable().(bool) {
			gc.playerJustDied = true
			deathType.Get()

			gc.timer = timer.GetVariable().(float32)
			gc.deathType = deathType.GetVariable().(int)
			gc.totalGemsAtDeath = totalGems.GetPreviousVariable().(int)
			gc.homingAtDeath = homing.GetPreviousVariable().(int)

			gc.status = statusIsDead
		}

		if gc.isAlive {

			if gc.playerName == "" {
				gc.GetPlayerVariables()
			}

			gc.isReplay = isReplay.GetVariable().(bool)
			gc.timer = timer.GetVariable().(float32)
			if gc.timer == 0.0 {
				if enemiesAlive.GetVariable().(int) == 0 {
					gc.status = statusInDaggerLobby
				} else {
					gc.status = statusInMainMenu
				}
				return
			}

			if gc.isReplay {
				gc.status = statusIsReplay
			} else {
				if gc.replayPlayerName != "" {
					gc.ResetReplayPlayerVariables()
				}
				gc.status = statusIsPlaying
			}

			gc.gems = gems.GetVariable().(int)
			gc.totalGems = totalGems.GetVariable().(int)
			gc.homing = homing.GetVariable().(int)
			if gc.homing > gc.homingMax {
				gc.homingMax = gc.homing
				gc.homingMaxTime = gc.timer
			}
			if gc.enemiesAlive > gc.enemiesAliveMax {
				gc.enemiesAliveMax = gc.enemiesAlive
				gc.enemiesAliveMaxTime = gc.timer
			}
			gc.enemiesAlive = enemiesAlive.GetVariable().(int)
			gc.enemiesKilled = enemiesKilled.GetVariable().(int)
			gc.daggersFired = daggersFired.GetVariable().(int)
			gc.daggersHit = daggersHit.GetVariable().(int)
			if gc.daggersFired > 0 {
				gc.accuracy = (float64(gc.daggersHit) / float64(gc.daggersFired)) * 100
			} else {
				gc.accuracy = 0.0
			}
			if gc.level2time == 0.0 && gc.gems >= 10 {
				gc.level2time = gc.timer
			}
			if gc.level3time == 0.0 && gc.gems == 70 {
				gc.level3time = gc.timer
			}
			if gc.level4time == 0.0 && gc.gems == 71 {
				gc.level4time = gc.timer
			}
		} else if gc.playerName == "" {
			gc.status = statusConnecting
		} else {
			if gc.replayPlayerName != "" {
				gc.ResetReplayPlayerVariables()
			}
			gc.status = statusIsDead
		}
	}
}
