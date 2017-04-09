package cliGui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/cskr/pubsub"
	"github.com/dh1tw/remoteRadio/events"
	sbRadio "github.com/dh1tw/remoteRadio/sb_radio"
	ui "github.com/gizak/termui"
)

func guiLoop(caps sbRadio.Capabilities, evPS *pubsub.PubSub) {

	state := sbRadio.State{}
	var intFreq float64 = 0.0
	lastFreqChange := time.Now()

	latencySpark := ui.NewSparkline()
	latencySpark.Title = ""
	latencySpark.Data = []int{0, 20}
	latencySpark.LineColor = ui.ColorYellow | ui.AttrBold

	latencyWidget := ui.NewSparklines(latencySpark)
	latencyWidget.Height = 5
	latencyWidget.BorderLabel = "Latency"

	powerOnWidget := ui.NewPar("")
	powerOnWidget.Height = 3
	powerOnWidget.BorderLabel = "Power On"

	pttWidget := ui.NewPar("")
	pttWidget.Height = 3
	pttWidget.BorderLabel = "PTT"

	infoWidget := ui.NewList()
	infoWidget.Items = []string{"", ""}
	infoWidget.BorderLabel = "Info"
	infoWidget.Height = 4

	funcData := make([]GuiFunction, len(caps.GetFunctions))
	for i, funcName := range caps.GetFunctions {
		funcData[i].Label = funcName
	}
	funcWidget := ui.NewList()
	funcWidget.Items = SprintFunctions(funcData)
	funcWidget.BorderLabel = "Functions"
	funcWidget.Height = 2 + len(caps.GetFunctions)

	levelData := make([]GuiLevel, len(caps.GetLevels))
	for i, level := range caps.GetLevels {
		levelData[i].Label = level.Name
	}

	levelWidget := ui.NewList()
	levelWidget.Items = SprintLevels(levelData)
	levelWidget.BorderLabel = "Levels"
	levelWidget.Height = 2 + len(caps.GetLevels)

	parameterWidget := ui.NewList()
	parameterWidget.Items = []string{""}
	parameterWidget.BorderLabel = "Parameters"
	parameterWidget.Height = 3 + len(caps.GetParameters)

	frequencyWidget := ui.NewPar("")
	frequencyWidget.BorderLabel = "Frequency"
	frequencyWidget.Height = 9

	sMeterWidget := ui.NewGauge()
	sMeterWidget.Percent = 40
	sMeterWidget.Height = 3
	sMeterWidget.BorderLabel = "S-Meter"
	sMeterWidget.BarColor = ui.ColorGreen
	sMeterWidget.Percent = 0
	sMeterWidget.Label = ""

	swrMeterWidget := ui.NewGauge()
	swrMeterWidget.Percent = 40
	swrMeterWidget.Height = 3
	swrMeterWidget.BorderLabel = "SWR"
	swrMeterWidget.BarColor = ui.ColorYellow
	swrMeterWidget.Percent = 0
	swrMeterWidget.Label = ""

	pMeterWidget := ui.NewGauge()
	pMeterWidget.Percent = 40
	pMeterWidget.Height = 3
	pMeterWidget.BorderLabel = "Power"
	pMeterWidget.BarColor = ui.ColorRed
	pMeterWidget.Percent = 0
	pMeterWidget.Label = ""

	modeWidget := ui.NewPar("")
	modeWidget.Height = 3
	modeWidget.BorderLabel = "Mode"

	vfoWidget := ui.NewPar("")
	vfoWidget.Height = 3
	vfoWidget.BorderLabel = "VFO"

	filterWidget := ui.NewPar("")
	filterWidget.Height = 3
	filterWidget.BorderLabel = "Filter"

	ritWidget := ui.NewPar("")
	ritWidget.Height = 3
	ritWidget.BorderLabel = "RIT"

	xitWidget := ui.NewPar("")
	xitWidget.Height = 3
	xitWidget.BorderLabel = "XIT"

	splitWidget := ui.NewPar("")
	splitWidget.Height = 3
	splitWidget.BorderLabel = "Split"

	splitFrequencyWidget := ui.NewPar("")
	splitFrequencyWidget.Height = 3
	splitFrequencyWidget.BorderLabel = "TX Frequency"

	splitModeWidget := ui.NewPar("")
	splitModeWidget.Height = 3
	splitModeWidget.BorderLabel = "TX Mode"

	splitFilterWidget := ui.NewPar("")
	splitFilterWidget.Height = 3
	splitFilterWidget.BorderLabel = "TX Filter"

	opsWidget := ui.NewList()
	opsWidget.Items = caps.VfoOps
	opsWidget.BorderLabel = "Operations"
	opsWidget.Height = 2 + len(caps.VfoOps)

	logWidgetItems := []string{}

	logWidget := ui.NewList()
	logWidget.Items = logWidgetItems
	logWidget.BorderLabel = "Logging"
	leftColumn := funcWidget.Height + opsWidget.Height
	rightColumn := levelWidget.Height + parameterWidget.Height
	if leftColumn > rightColumn {
		logWidget.Height = leftColumn
	} else {
		logWidget.Height = rightColumn
	}

	cliWidget := ui.NewInput("", false)
	cliWidget.Height = 3
	cliWidget.BorderLabel = "Rig command:"

	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(2, 0, infoWidget, latencyWidget),
			ui.NewCol(8, 0, frequencyWidget),
			ui.NewCol(2, 0, pMeterWidget, swrMeterWidget, sMeterWidget)),
		ui.NewRow(
			ui.NewCol(2, 0, powerOnWidget),
			ui.NewCol(1, 0, vfoWidget),
			ui.NewCol(1, 0, modeWidget),
			ui.NewCol(2, 0, filterWidget),
			ui.NewCol(2, 0, ritWidget),
			ui.NewCol(2, 0, xitWidget),
			ui.NewCol(2, 0, nil)),
		ui.NewRow(
			ui.NewCol(2, 0, pttWidget),
			ui.NewCol(1, 0, splitWidget),
			ui.NewCol(2, 0, splitFrequencyWidget),
			ui.NewCol(1, 0, splitModeWidget),
			ui.NewCol(2, 0, splitFilterWidget)),
		ui.NewRow(
			ui.NewCol(2, 0, funcWidget, opsWidget),
			ui.NewCol(8, 0, logWidget),
			ui.NewCol(2, 0, levelWidget, parameterWidget)),
		ui.NewRow(
			ui.NewCol(12, 0, cliWidget)),
	)

	// calculate layout
	ui.Body.Align()

	ui.Render(ui.Body)

	cliWidget.StartCapture()

	ui.Handle("/sys/kbd/<up>", func(ui.Event) {
		intFreq += float64(state.Vfo.TuningStep)
		freq := intFreq / 1000
		cmd := []string{"set_freq", fmt.Sprintf("%.2f", freq)}
		evPS.Pub(cmd, events.CliInput)
		frequencyWidget.Text = fmt.Sprintf("%.2f kHz", intFreq/1000)
		ui.Render(frequencyWidget)
		lastFreqChange = time.Now()
	})

	ui.Handle("/sys/kbd/<down>", func(ui.Event) {
		intFreq -= float64(state.Vfo.TuningStep)
		freq := intFreq / 1000
		cmd := []string{"set_freq", fmt.Sprintf("%.2f", freq)}
		evPS.Pub(cmd, events.CliInput)
		frequencyWidget.Text = fmt.Sprintf("%.2f kHz", intFreq/1000)
		ui.Render(frequencyWidget)
		lastFreqChange = time.Now()
	})

	ui.Handle("/timer/1s", func(ui.Event) {
		if time.Since(lastFreqChange) > time.Millisecond*300 &&
			intFreq != state.Vfo.Frequency {
			intFreq = state.Vfo.Frequency
			frequencyWidget.Text = fmt.Sprintf("%.2f kHz", intFreq/1000)
			ui.Render(frequencyWidget)
		}
	})

	ui.Handle("/sys/kbd/C-c", func(ui.Event) {
		ui.StopLoop()
		evPS.Pub(true, events.Shutdown)
	})

	ui.Handle("/input/kbd", func(ev ui.Event) {
		evData := ev.Data.(ui.EvtInput)
		if evData.KeyStr == "<enter>" && len(cliWidget.Text()) > 0 {
			cmd := strings.Split(cliWidget.Text(), " ")
			evPS.Pub(cmd, events.CliInput)
			cliWidget.Clear()
			ui.Render(ui.Body)
		}
	})

	ui.Handle("/network/latency", func(e ui.Event) {

		latency := e.Data.(int64) / 1000000 // milli seconds
		if len(latencyWidget.Lines[0].Data) > 20 {
			latencyWidget.Lines[0].Data = latencyWidget.Lines[0].Data[2:]
		}
		latencyWidget.Lines[0].Data = append(latencyWidget.Lines[0].Data, int(latency))
		latencyWidget.Lines[0].Title = fmt.Sprintf("%dms", latency)
		ui.Render(ui.Body)
	})

	freq_loaded := false

	ui.Handle("/network/update", func(e ui.Event) {

		state = e.Data.(sbRadio.State)

		infoWidget.Items[0] = caps.MfgName + " " + caps.ModelName
		infoWidget.Items[1] = caps.Version + " " + caps.Status
		if !freq_loaded {
			intFreq = state.Vfo.Frequency
			freq_loaded = true
			frequencyWidget.Text = fmt.Sprintf("%.2f kHz", intFreq/1000)
		}
		modeWidget.Text = state.Vfo.Mode
		filterWidget.Text = fmt.Sprintf("%v Hz", state.Vfo.PbWidth)
		vfoWidget.Text = state.CurrentVfo
		ritWidget.Text = fmt.Sprintf("%v Hz", state.Vfo.Rit)
		if state.Vfo.Rit != 0 {
			ritWidget.TextBgColor = ui.ColorGreen
		} else {
			ritWidget.TextBgColor = ui.ColorDefault
		}

		xitWidget.Text = fmt.Sprintf("%v Hz", state.Vfo.Xit)
		if state.Vfo.Xit != 0 {
			xitWidget.TextBgColor = ui.ColorRed
		} else {
			xitWidget.TextBgColor = ui.ColorDefault
		}

		if state.Ptt {
			pttWidget.Bg = ui.ColorRed
		} else {
			pttWidget.Bg = ui.ColorDefault
		}
		if state.RadioOn {
			powerOnWidget.Bg = ui.ColorGreen
		} else {
			powerOnWidget.Bg = ui.ColorDefault
		}
		if state.Vfo.Split.Enabled {
			splitWidget.Bg = ui.ColorGreen
			splitFrequencyWidget.Text = fmt.Sprintf("%.2f kHz", state.Vfo.Split.Frequency/1000)
			splitModeWidget.Text = state.Vfo.Split.Mode
			splitFilterWidget.Text = fmt.Sprintf("%v Hz", state.Vfo.Split.PbWidth)
		} else {
			splitWidget.Bg = ui.ColorDefault
			splitFrequencyWidget.Text = ""
			splitModeWidget.Text = ""
			splitFilterWidget.Text = ""
		}
		if state.Ptt {
			sMeterWidget.Percent = 0
			sMeterWidget.Label = ""
			if pValue, ok := state.Vfo.Levels["METER"]; ok {
				pMeterWidget.Label = fmt.Sprintf("%vW", pValue)
			}
			if swrValue, ok := state.Vfo.Levels["SWR"]; ok {
				swrMeterWidget.Label = fmt.Sprintf("1:%.2f", swrValue)
			}
		} else {
			pMeterWidget.Percent = 0
			pMeterWidget.Label = ""
			swrMeterWidget.Percent = 0
			swrMeterWidget.Label = ""
			if sValue, ok := state.Vfo.Levels["STRENGTH"]; ok {
				if sValue < 0 {
					s := int((59 - sValue*-1) / 6)
					sMeterWidget.Label = fmt.Sprintf("S%v", s)
					sMeterWidget.Percent = int((59 - sValue*-1) * 100 / 114)
				} else {
					sMeterWidget.Label = fmt.Sprintf("S9+%vdB", int(sValue))
					sMeterWidget.Percent = int((sValue + 59) * 100 / 114)
				}
			}
		}
		for i, el := range levelData {
			for name, value := range state.Vfo.Levels {
				if el.Label == name {
					levelData[i].Value = value
				}
			}
		}
		levelWidget.Items = SprintLevels(levelData)

		for i, el := range funcData {
			found := false
			for _, funcName := range state.Vfo.Functions {
				if el.Label == funcName {
					funcData[i].Set = true
					found = true
				}
				if !found {
					funcData[i].Set = false
				}
			}
		}
		funcWidget.Items = SprintFunctions(funcData)

		ui.Render(ui.Body)
	})

	ui.Handle("/log/msg", func(e ui.Event) {
		msg := e.Data.(string)
		logWidget.Items = append(logWidget.Items, msg)
		ui.Render(ui.Body)
	})

	ui.Handle("/sys/wnd/resize", func(e ui.Event) {
		ui.Body.Width = ui.TermWidth()
		ui.Body.Align()
		ui.Clear()
		ui.Render(ui.Body)
	})

	ui.Loop()
}

type GuiFunction struct {
	Label string
	Set   bool
}

func SprintFunctions(fs []GuiFunction) []string {
	s := make([]string, 0, len(fs))
	for _, el := range fs {
		item := el.Label
		for i := len(item); i < 8; i++ {
			item = item + " "
		}
		if el.Set {
			item = item + "[X]"
		} else {
			item = item + "[ ]"
		}
		s = append(s, item)
	}
	return s
}

func SprintLevels(lv []GuiLevel) []string {
	s := make([]string, 0, len(lv))
	for _, el := range lv {
		item := el.Label
		for i := len(item); i < 13; i++ {
			item = item + " "
		}
		intr, frac := math.Modf(float64(el.Value))
		if frac > 0 {
			item = item + fmt.Sprintf("%.2f", el.Value)
		} else {
			item = item + fmt.Sprintf("%.0f", intr)
		}
		s = append(s, item)
	}
	return s
}

type GuiLevel struct {
	Label string
	Value float32
}
