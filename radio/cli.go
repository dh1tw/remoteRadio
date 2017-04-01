package radio

import (
	"fmt"
	"math"
	"strconv"

	"strings"

	sbRadio "github.com/dh1tw/remoteRadio/sb_radio"
	"github.com/dh1tw/remoteRadio/utils"
)

func (r *remoteRadio) populateCliCmds() {

	r.cliCmds["f"] = getFrequency
	r.cliCmds["get_freq"] = getFrequency
	r.cliCmds["F"] = setFrequency
	r.cliCmds["set_freq"] = setFrequency
	r.cliCmds["m"] = getMode
	r.cliCmds["get_mode"] = getMode
	r.cliCmds["M"] = setMode
	r.cliCmds["set_mode"] = setMode
	r.cliCmds["v"] = getVfo
	r.cliCmds["get_vfo"] = getVfo
	r.cliCmds["V"] = setVfo
	r.cliCmds["set_vfo"] = setVfo
	r.cliCmds["j"] = getRit
	r.cliCmds["get_rit"] = getRit
	r.cliCmds["J"] = setRit
	r.cliCmds["set_rit"] = setRit
	r.cliCmds["z"] = getXit
	r.cliCmds["get_xit"] = getXit
	r.cliCmds["Z"] = setXit
	r.cliCmds["set_xit"] = setXit
	r.cliCmds["y"] = getAnt
	r.cliCmds["get_ant"] = getAnt
	r.cliCmds["Y"] = setAnt
	r.cliCmds["set_ant"] = setAnt
	r.cliCmds["t"] = getPtt
	r.cliCmds["get_ptt"] = getPtt
	r.cliCmds["T"] = setPtt
	r.cliCmds["set_ptt"] = setPtt
	r.cliCmds["G"] = execVfoOp
	r.cliCmds["vfo_op"] = execVfoOp
	r.cliCmds["u"] = getFunction
	r.cliCmds["get_func"] = getFunction
	r.cliCmds["U"] = setFunction
	r.cliCmds["set_func"] = setFunction
	r.cliCmds["l"] = getLevel
	r.cliCmds["get_level"] = getLevel
	r.cliCmds["L"] = setLevel
	r.cliCmds["n"] = getTuningStep
	r.cliCmds["N"] = setTuningStep
	r.cliCmds["set_level"] = setLevel
	r.cliCmds["get_powerstat"] = getPowerStat
	r.cliCmds["set_powerstat"] = setPowerStat
	r.cliCmds["s"] = getSplit
	r.cliCmds["get_split"] = getSplit
	r.cliCmds["S"] = setSplit
	r.cliCmds["set_split"] = setSplit
	r.cliCmds["i"] = getSplit
	r.cliCmds["get_split_freq"] = getSplit
	r.cliCmds["I"] = setSplitFreq
	r.cliCmds["set_split_freq"] = setSplitFreq
	r.cliCmds["x"] = getSplit
	r.cliCmds["get_split_mode"] = getSplit
	r.cliCmds["X"] = setSplitMode
	r.cliCmds["set_split_mode"] = setSplitMode
	r.cliCmds["k"] = getSplit
	r.cliCmds["get_split_freq_mode"] = getSplit
	r.cliCmds["K"] = setSplitFreqMode
	r.cliCmds["set_split_freq_mode"] = setSplitFreqMode
	r.cliCmds["3"] = dumpCaps
	r.cliCmds["5"] = dumpState
	r.cliCmds["?"] = printHelp
	r.cliCmds["help"] = printHelp

}

func (r *remoteRadio) parseCli(cliCmd []string) {

	if len(cliCmd) == 0 {
		fmt.Printf("Rig command: ")
		return
	}

	f, ok := r.cliCmds[cliCmd[0]]
	if !ok {
		fmt.Println("unknown command")
		fmt.Printf("Rig command: ")
		return
	}
	f(r, cliCmd[1:])
}

func getFrequency(r *remoteRadio, args []string) {
	fmt.Println(r.state.Vfo.Frequency)
	fmt.Printf("Rig command: ")
}

func setFrequency(r *remoteRadio, args []string) {

	if ok := checkArgs(args, 1); !ok {
		return
	}

	freq, err := strconv.ParseFloat(args[0], 10)
	if err != nil {
		fmt.Println("frequency must be float")
		return
	}

	// req := r.deepCopyState()
	req := r.initSetState()
	req.Vfo.Frequency = freq
	req.Md.HasFrequency = true
	if err := r.sendCatRequest(req); err != nil {
		fmt.Println(err)
	}
}

func getMode(r *remoteRadio, args []string) {
	fmt.Println(r.state.Vfo.Mode)
	fmt.Printf("Rig command: ")
}

func setMode(r *remoteRadio, args []string) {

	if len(args) < 1 || len(args) > 2 {
		fmt.Println("wrong number of arguments")
		return
	}

	mode := strings.ToUpper(args[0])

	if ok := utils.StringInSlice(mode, r.caps.Modes); !ok {
		fmt.Println("unsupported Mode")
		return
	}

	req := r.initSetState()
	req.Vfo.Mode = mode
	req.Md.HasMode = true

	if len(args) == 2 {

		pbWidth, err := strconv.ParseInt(args[1], 10, 32)
		if err != nil {
			fmt.Println("passband width must be integer")
		}

		filters, ok := r.caps.Filters[mode]
		if !ok {
			fmt.Println("WARN: No Filters found for this Mode in Rig Caps")
		}
		if ok := utils.Int32InSlice(int32(pbWidth), filters.Value); !ok {
			fmt.Println("WARN: unspported passband width")
		}
		req.Vfo.PbWidth = int32(pbWidth)
		req.Md.HasPbWidth = true
	}

	if err := r.sendCatRequest(req); err != nil {
		fmt.Println(err)
	}
}

func getVfo(r *remoteRadio, args []string) {
	fmt.Println("Current Vfo:", r.state.CurrentVfo)
	fmt.Printf("Rig command: ")
}

func setVfo(r *remoteRadio, args []string) {
	if ok := checkArgs(args, 1); !ok {
		return
	}

	vfo := strings.ToUpper(args[0])

	if ok := utils.StringInSlice(vfo, r.caps.Vfos); !ok {
		fmt.Println("unsupported VFO")
		return
	}

	req := r.initSetState()
	req.CurrentVfo = vfo

	if err := r.sendCatRequest(req); err != nil {
		fmt.Println(err)
	}
}

func getRit(r *remoteRadio, args []string) {
	fmt.Println("Rit:", r.state.Vfo.Rit)
	fmt.Printf("Rig command: ")
}

func setRit(r *remoteRadio, args []string) {

	if ok := checkArgs(args, 1); !ok {
		return
	}

	rit, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		fmt.Println("rit value must be integer")
		return
	}

	if math.Abs(float64(rit)) > float64(r.caps.MaxRit) {
		fmt.Println("WARN: Rit value larger than supported by Rig")
	}

	req := r.initSetState()
	req.Vfo.Rit = int32(rit)
	req.Md.HasRit = true

	if err := r.sendCatRequest(req); err != nil {
		fmt.Println(err)
	}
}

func getXit(r *remoteRadio, args []string) {
	fmt.Println("Xit:", r.state.Vfo.Xit)
	fmt.Printf("Rig command: ")
}

func setXit(r *remoteRadio, args []string) {

	if !checkArgs(args, 1) {
		return
	}

	xit, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		fmt.Println("xit value must be integer")
		return
	}

	if math.Abs(float64(xit)) > float64(r.caps.MaxXit) {
		fmt.Println("WARN: Xit value larger than supported by Rig")
	}

	req := r.initSetState()

	req.Vfo.Xit = int32(xit)
	req.Md.HasXit = true

	if err := r.sendCatRequest(req); err != nil {
		fmt.Println(err)
	}
}

func getAnt(r *remoteRadio, args []string) {
	fmt.Println("Antenna:", r.state.Vfo.Ant)
	fmt.Println("Rig command: ")
}

func setAnt(r *remoteRadio, args []string) {
	if !checkArgs(args, 1) {
		return
	}

	ant, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		fmt.Println("Antenna value must be integer")
		return
	}

	// check Antenna in CAPS
	req := r.initSetState()
	req.Vfo.Ant = int32(ant)
	req.Md.HasAnt = true

	if err := r.sendCatRequest(req); err != nil {
		fmt.Println(err)
	}
}

func getPowerStat(r *remoteRadio, args []string) {
	fmt.Println("Power On:", r.state.RadioOn)
	fmt.Printf("Rig command: ")
}

func setPowerStat(r *remoteRadio, args []string) {
	if !checkArgs(args, 1) {
		return
	}

	power, err := strconv.ParseBool(args[0])
	if err != nil {
		fmt.Println("Power value must be of type bool (1,t,T,True / 0,f,F,FALSE")
		return
	}

	req := r.initSetState()
	req.RadioOn = power
	req.Md.HasRadioOn = true

	if err := r.sendCatRequest(req); err != nil {
		fmt.Println(err)
	}
}

func getPtt(r *remoteRadio, args []string) {
	fmt.Println("PTT On:", r.state.Ptt)
	fmt.Printf("Rig command: ")
}

func setPtt(r *remoteRadio, args []string) {
	if !checkArgs(args, 1) {
		return
	}

	ptt, err := strconv.ParseBool(args[0])
	if err != nil {
		fmt.Println("PTTr value must be of type bool (1,t,T,True / 0,f,F,FALSE")
		return
	}

	req := r.initSetState()
	req.Ptt = ptt
	req.Md.HasPtt = true

	if err := r.sendCatRequest(req); err != nil {
		fmt.Println(err)
	}
}

func getLevel(r *remoteRadio, args []string) {
	r.PrintLevels()
	fmt.Printf("Rig command: ")
}

func setLevel(r *remoteRadio, args []string) {
	if !checkArgs(args, 2) {
		fmt.Printf("Available Levels: ")
		for _, level := range r.caps.GetSetLevels() {
			fmt.Printf("%s ", level.Name)
		}
		fmt.Printf("\n")
		fmt.Println("Rig command: ")
		return
	}

	levelName := strings.ToUpper(args[0])

	if !valueInValueList(levelName, r.caps.SetLevels) {
		fmt.Println("unknown Value:", levelName)
	}

	levelValue, err := strconv.ParseFloat(args[1], 32)
	if err != nil {
		fmt.Println("Level Value must be of type Float")
		return
	}

	levelMap := make(map[string]float32)

	levelMap[levelName] = float32(levelValue)

	req := r.initSetState()

	req.Vfo.Levels = levelMap
	req.Md.HasLevels = true
	if err := r.sendCatRequest(req); err != nil {
		fmt.Println(err)
	}
}

func getFunction(r *remoteRadio, args []string) {
	fmt.Println("Functions:", r.state.Vfo.Functions)
	fmt.Printf("Rig command: ")
}

func setFunction(r *remoteRadio, args []string) {
	if !checkArgs(args, 1) {
		fmt.Println("Available Functions:", r.caps.SetFunctions)
		fmt.Printf("Rig command: ")
		return
	}

	funcName := args[0]

	req := r.initSetState()
	req.Md.HasFunctions = true

	if !utils.StringInSlice(funcName, req.Vfo.Functions) {
		req.Vfo.Functions = append(req.Vfo.Functions, funcName)
	} else {
		req.Vfo.Functions = utils.RemoveStringFromSlice(funcName, req.Vfo.Functions)
	}

	if err := r.sendCatRequest(req); err != nil {
		fmt.Println(err)
	}

}

func getSplit(r *remoteRadio, args []string) {
	fmt.Println("Split Enabled:", r.state.Vfo.Split.Enabled)
	if r.state.Vfo.Split.Enabled {
		fmt.Println("Split Vfo:", r.state.Vfo.Split.Vfo)
		fmt.Println("Split Freq:", r.state.Vfo.Split.Frequency)
		fmt.Println("Split Mode:", r.state.Vfo.Split.Mode)
		fmt.Println("Split PbWidth:", r.state.Vfo.Split.PbWidth)
	}
	fmt.Printf("Rig command: ")
}

func setSplit(r *remoteRadio, args []string) {
	if !checkArgs(args, 1) {
		return
	}

	splitEnabled, err := strconv.ParseBool(args[0])
	if err != nil {
		fmt.Println("Split Enable/Disable value must be of type bool (1,t,T,True / 0,f,F,FALSE")
		return
	}

	req := r.initSetState()
	req.Md.HasSplit = true

	req.Vfo.Split.Enabled = splitEnabled
	if err := r.sendCatRequest(req); err != nil {
		fmt.Println(err)
	}
}

func setSplitFreq(r *remoteRadio, args []string) {
	if !checkArgs(args, 1) {
		return
	}

	freq, err := strconv.ParseFloat(args[0], 10)
	if err != nil {
		fmt.Println("frequency must be float")
		return
	}

	req := r.initSetState()
	req.Vfo.Split.Enabled = true
	req.Vfo.Split.Frequency = freq
	req.Md.HasSplit = true

	if err := r.sendCatRequest(req); err != nil {
		fmt.Println(err)
	}
}

func setSplitMode(r *remoteRadio, args []string) {
	if len(args) < 1 || len(args) > 2 {
		fmt.Println("wrong number of arguments")
		return
	}

	if ok := utils.StringInSlice(args[0], r.caps.Modes); !ok {
		fmt.Println("unsupported Mode")
		return
	}

	req := r.initSetState()
	req.Vfo.Split.Mode = args[0]
	req.Md.HasSplit = true

	if len(args) == 2 {

		pbWidth, err := strconv.ParseInt(args[1], 10, 32)
		if err != nil {
			fmt.Println("passband width must be integer")
		}

		filters, ok := r.caps.Filters[args[0]]
		if !ok {
			fmt.Println("WARN: No Filters found for this Mode in Rig Caps")
		}
		if ok := utils.Int32InSlice(int32(pbWidth), filters.Value); !ok {
			fmt.Println("WARN: unspported passband width")
		}
		req.Vfo.Split.PbWidth = int32(pbWidth)
	}

	req.Vfo.Split.Enabled = true

	if err := r.sendCatRequest(req); err != nil {
		fmt.Println(err)
	}
}

func setSplitFreqMode(r *remoteRadio, args []string) {
	if len(args) < 2 || len(args) > 3 {
		fmt.Println("wrong number of arguments")
		return
	}

	freq, err := strconv.ParseFloat(args[0], 10)
	if err != nil {
		fmt.Println("frequency must be float")
		return
	}

	if ok := utils.StringInSlice(args[1], r.caps.Modes); !ok {
		fmt.Println("unsupported Mode")
		return
	}

	req := r.initSetState()
	req.Vfo.Split.Enabled = true
	req.Vfo.Split.Frequency = freq
	req.Vfo.Split.Mode = args[1]
	req.Md.HasSplit = true

	if len(args) == 3 {

		pbWidth, err := strconv.ParseInt(args[2], 10, 32)
		if err != nil {
			fmt.Println("passband width must be integer")
		}

		filters, ok := r.caps.Filters[args[2]]
		if !ok {
			fmt.Println("WARN: No Filters found for this Mode in Rig Caps")
		}
		if ok := utils.Int32InSlice(int32(pbWidth), filters.Value); !ok {
			fmt.Println("WARN: unspported passband width")
		}
		req.Vfo.Split.PbWidth = int32(pbWidth)
	}

	if err := r.sendCatRequest(req); err != nil {
		fmt.Println(err)
	}
}

func execVfoOp(r *remoteRadio, args []string) {

	for _, vfoOp := range args {
		if !utils.StringInSlice(vfoOp, r.caps.VfoOps) {
			fmt.Println("unknown VFO Operation:", vfoOp)
			return
		}
	}

	req := r.initSetState()
	req.VfoOperations = args

	if err := r.sendCatRequest(req); err != nil {
		fmt.Println(err)
	}

}

func getTuningStep(r *remoteRadio, args []string) {
	fmt.Printf("Tuning Step: %dHz\n", r.state.Vfo.TuningStep)
	fmt.Printf("Rig command: ")
}

func setTuningStep(r *remoteRadio, args []string) {
	if !checkArgs(args, 1) {
		return
	}

	req := r.initSetState()

	ts, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		fmt.Println("tuning step must be integer")
	}

	// check if the given tuning step is supported by the rig
	supportedTs, ok := r.caps.TuningSteps[r.state.Vfo.Mode]
	if !ok {
		fmt.Println("WARN: No Tuning step values registered for this mode")
	}
	if ok := utils.Int32InSlice(int32(ts), supportedTs.Value); !ok {
		fmt.Println("WARN: Tuning step not supported for this Mode")
	}
	req.Vfo.TuningStep = int32(ts)
	req.Md.HasTuningStep = true

	if err := r.sendCatRequest(req); err != nil {
		fmt.Println(err)
	}
}

func dumpCaps(r *remoteRadio, args []string) {
	r.PrintCapabilities()
	fmt.Printf("Rig command: ")
}

func dumpState(r *remoteRadio, args []string) {
	r.PrintState()
	fmt.Printf("Rig command: ")
}

func printHelp(r *remoteRadio, args []string) {
	fmt.Println("Commands:")
	for cmd := range r.cliCmds {
		fmt.Println(cmd)
	}
}

func checkArgs(args []string, length int) bool {
	if len(args) != length {
		fmt.Println("wrong number of arguments")
		return false
	}
	return true
}

func valueInValueList(vName string, vList []*sbRadio.Value) bool {
	for _, value := range vList {
		if value.Name == vName {
			return true
		}
	}
	return false
}
