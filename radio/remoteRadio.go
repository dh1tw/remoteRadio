package radio

import (
	"fmt"
	"html/template"
	"log"
	"reflect"
	"sync"

	"os"

	"github.com/cskr/pubsub"
	"github.com/dh1tw/remoteRadio/comms"
	"github.com/dh1tw/remoteRadio/events"
	sbRadio "github.com/dh1tw/remoteRadio/sb_radio"
)

type RemoteRadioSettings struct {
	CatResponseCh   chan []byte
	RadioStatusCh   chan []byte
	CatRequestTopic string
	ToWireCh        chan comms.IOMsg
	CapabilitiesCh  chan []byte
	WaitGroup       *sync.WaitGroup
	Events          *pubsub.PubSub
}

type remoteRadio struct {
	state    sbRadio.State
	newState sbRadio.SetState
	caps     sbRadio.Capabilities
	settings RemoteRadioSettings
	cliCmds  map[string]func(r *remoteRadio, args []string)
}

func HandleRemoteRadio(rs RemoteRadioSettings) {
	defer rs.WaitGroup.Done()

	shutdownCh := rs.Events.Sub(events.Shutdown)
	cliCh := rs.Events.Sub(events.Cli)

	r := remoteRadio{}
	r.state.Vfo = &sbRadio.Vfo{}
	r.state.Vfo.Functions = make([]string, 0, 20)
	r.state.Vfo.Levels = make(map[string]float32)
	r.state.Vfo.Parameters = make(map[string]float32)
	r.state.Vfo.Split = &sbRadio.Split{}

	r.settings = rs

	r.cliCmds = make(map[string]func(r *remoteRadio, args []string))
	r.populateCliCmds()

	// newStateInitalized := false

	// rs.Events.Pub(true, events.ForwardCat)

	for {
		select {
		case msg := <-rs.CapabilitiesCh:
			r.deserializeCaps(msg)
			// r.PrintCapabilities()
		case msg := <-rs.CatResponseCh:
			r.deserializeCatResponse(msg)
			// r.PrintState()
		case msg := <-cliCh:
			r.parseCli(msg.([]string))
		case <-shutdownCh:
			log.Println("Disconnecting from Radio")
			return
		}
	}
}

func (r *remoteRadio) sendCatRequest(req sbRadio.SetState) error {
	data, err := req.Marshal()
	if err != nil {
		return err
	}

	msg := comms.IOMsg{}
	msg.Data = data
	msg.Topic = r.settings.CatRequestTopic
	msg.Retain = false
	msg.Qos = 0

	r.settings.ToWireCh <- msg

	return nil
}

var stateTmpl = template.Must(template.New("").Parse(
	`
Current Vfo: {{.CurrentVfo}}
Radio Powered: {{.RadioOn}}
Ptt: {{.Ptt}}
Vfo: {{.Vfo.Vfo}}
  Frequency: {{.Vfo.Frequency}}Hz
  Mode: {{.Vfo.Mode}}
  PBWidth: {{.Vfo.PbWidth}}
  Antenna: {{.Vfo.Ant}}
  Rit: {{.Vfo.Rit}}
  Xit: {{.Vfo.Xit}}
  Split: 
    Enabled: {{.Vfo.Split.Enabled}}
	Vfo: {{.Vfo.Split.Vfo}}
	Frequency: {{.Vfo.Split.Frequency}}
	Mode: {{.Vfo.Split.Mode}}
	PbWidth: {{.Vfo.Split.PbWidth}}
  Tuning Step: {{.Vfo.TuningStep}}
  Functions: {{range $f := .Vfo.Functions}}{{$f}} {{end}}
  Levels: {{range $name, $val := .Vfo.Levels}}
    {{$name}}: {{$val}} {{end}}
  Parameters: {{range $name, $val := .Vfo.Parameters}}
    {{$name}}: {{$val}} {{end}}
`,
))

var capsTmpl = template.Must(template.New("").Parse(
	`
Radio Capabilities:

Supported VFOs:{{range $vfo := .Vfos}}{{$vfo}} {{end}}
Supported Modes: {{range $mode := .Modes}}{{$mode}} {{end}}
Supported VFO Operations: {{range $vfoOp := .VfoOps}}{{$vfoOp}} {{end}}
Supported Functions (Get):{{range $getF := .GetFunctions}}{{$getF}} {{end}}
Supported Functions (Set): {{range $setF := .SetFunctions}}{{$setF}} {{end}}
Supported Levels (Get): {{range $val := .GetLevels}}
  {{$val.Name}} ({{$val.Min}}..{{$val.Max}}/{{$val.Step}}){{end}}
Supported Levels (Set): {{range $val := .SetLevels}}
  {{$val.Name}} ({{$val.Min}}..{{$val.Max}}/{{$val.Step}}){{end}}
Supported Parameters (Get): {{range $val := .GetParameters}}
  {{$val.Name}} ({{$val.Min}}..{{$val.Max}}/{{$val.Step}}){{end}}
Supported Parameters (Set): {{range $val := .SetParameters}}
  {{$val.Name}} ({{$val.Min}}..{{$val.Max}}/{{$val.Step}}){{end}}
Max Rit: +-{{.MaxRit}}Hz
Max Xit: +-{{.MaxXit}}Hz
Max IF Shift: +-{{.MaxIfShift}}Hz
Filters [Hz]: {{range $mode, $pbList := .Filters}}
  {{$mode}}:		{{range $pb := $pbList.Value}}{{$pb}} {{end}} {{end}}
Tuning Steps [Hz]: {{range $mode, $tsList := .TuningSteps}}
  {{$mode}}:		{{range $ts := $tsList.Value}}{{$ts}} {{end}} {{end}}
Preamps: {{range $preamp := .Preamps}}{{$preamp}}dB {{end}}
Attenuators: {{range $att := .Attenuators}}{{$att}}dB {{end}} 
`,
))

func (r *remoteRadio) PrintCapabilities() {
	err := capsTmpl.Execute(os.Stdout, r.caps)
	if err != nil {
		fmt.Println(err)
	}
}

func (r *remoteRadio) PrintState() {
	err := stateTmpl.Execute(os.Stdout, r.state)
	if err != nil {
		fmt.Println(err)
	}
}

func (r *remoteRadio) deserializeCaps(msg []byte) error {

	caps := sbRadio.Capabilities{}
	err := caps.Unmarshal(msg)
	if err != nil {
		return err
	}

	r.caps = caps

	return nil
}

func (r *remoteRadio) deserializeCatResponse(msg []byte) error {

	ns := sbRadio.State{}
	err := ns.Unmarshal(msg)
	if err != nil {
		return err
	}

	fmt.Println(ns)

	if ns.CurrentVfo != r.state.CurrentVfo {
		r.state.CurrentVfo = ns.CurrentVfo
		fmt.Println("Current Vfo:", r.state.CurrentVfo)
	}

	if ns.Vfo != nil {

		if ns.Vfo.GetFrequency() != r.state.Vfo.Frequency {
			r.state.Vfo.Frequency = ns.Vfo.GetFrequency()
			fmt.Println("Frequency:", r.state.Vfo.Frequency)
		}

		if ns.Vfo.GetVfo() != r.state.Vfo.Vfo {
			r.state.Vfo.Vfo = ns.Vfo.GetVfo()
		}

		if ns.Vfo.GetMode() != r.state.Vfo.Mode {
			r.state.Vfo.Mode = ns.Vfo.GetMode()
			fmt.Println("Mode:", r.state.Vfo.Mode)
		}

		if ns.Vfo.GetPbWidth() != r.state.Vfo.PbWidth {
			r.state.Vfo.PbWidth = ns.Vfo.GetPbWidth()
			fmt.Println("Filter:", r.state.Vfo.PbWidth)
		}

		if ns.Vfo.GetAnt() != r.state.Vfo.Ant {
			r.state.Vfo.Ant = ns.Vfo.GetAnt()
			fmt.Println("Antenna:", r.state.Vfo.Ant)
		}

		if ns.Vfo.GetRit() != r.state.Vfo.Rit {
			r.state.Vfo.Rit = ns.Vfo.GetRit()
			fmt.Println("Rit:", r.state.Vfo.Rit)
		}

		if ns.Vfo.GetXit() != r.state.Vfo.Xit {
			r.state.Vfo.Xit = ns.Vfo.GetXit()
			fmt.Println("Xit:", r.state.Vfo.Xit)
		}

		if ns.Vfo.GetSplit() != nil {
			if !reflect.DeepEqual(ns.Vfo.GetSplit(), r.state.Vfo.Split) {
				if err := r.updateSplit(ns.Vfo.Split); err != nil {
					log.Println(err)
				}
			}
		}

		if ns.Vfo.GetTuningStep() != r.state.Vfo.TuningStep {
			r.state.Vfo.TuningStep = ns.Vfo.GetTuningStep()
			fmt.Println("Tuning Step:", r.state.Vfo.TuningStep)
		}

		if ns.Vfo.Functions != nil {
			if !reflect.DeepEqual(ns.Vfo.Functions, r.state.Vfo.Functions) {
				if err := r.updateFunctions(ns.Vfo.GetFunctions()); err != nil {
					log.Println(err)
				}
			}
		}

		if ns.Vfo.Levels != nil {
			if !reflect.DeepEqual(ns.Vfo.Levels, r.state.Vfo.Levels) {
				if err := r.updateLevels(ns.Vfo.GetLevels()); err != nil {
					log.Println(err)
				}
			}
		}

		if ns.Vfo.Parameters != nil {
			if !reflect.DeepEqual(ns.Vfo.Parameters, r.state.Vfo.Parameters) {
				if err := r.updateParams(ns.Vfo.GetParameters()); err != nil {
					log.Println(err)
				}
			}
		}

	}

	if ns.GetRadioOn() != r.state.RadioOn {
		r.state.RadioOn = ns.GetRadioOn()
		fmt.Println("Radio Power On:", r.state.RadioOn)
	}

	if ns.GetPtt() != r.state.Ptt {
		r.state.Ptt = ns.GetPtt()
		fmt.Println("PTT On:", r.state.Ptt)
	}

	return nil
}

func (r *remoteRadio) updateSplit(newSplit *sbRadio.Split) error {

	if newSplit.GetEnabled() != r.state.Vfo.Split.Enabled {
		r.state.Vfo.Split.Enabled = newSplit.GetEnabled()
		fmt.Println("Split Enabled:", r.state.Vfo.Split.Enabled)
	}

	if newSplit.GetFrequency() != r.state.Vfo.Split.Frequency {

		r.state.Vfo.Split.Frequency = newSplit.GetFrequency()
		fmt.Println("Split Frequency:", r.state.Vfo.Split.Frequency)
	}

	if newSplit.GetVfo() != r.state.Vfo.Split.Vfo {

		r.state.Vfo.Split.Vfo = newSplit.GetVfo()
		fmt.Println("Split Vfo:", r.state.Vfo.Split.Vfo)
	}

	if newSplit.GetMode() != r.state.Vfo.Split.Mode {

		r.state.Vfo.Split.Mode = newSplit.GetMode()
		fmt.Println("Split Mode:", r.state.Vfo.Split.Mode)
	}

	if newSplit.GetPbWidth() != r.state.Vfo.Split.PbWidth {

		r.state.Vfo.Split.PbWidth = newSplit.GetPbWidth()
		fmt.Println("Split PbWidth:", r.state.Vfo.Split.PbWidth)
	}

	return nil
}

func (r *remoteRadio) updateFunctions(newFuncs []string) error {

	r.state.Vfo.Functions = newFuncs

	// vfo, _ := hl.VfoValue[r.state.CurrentVfo]

	// functions to be enabled
	// diff := utils.SliceDiff(newFuncs, r.state.Vfo.Functions)
	// for _, f := range diff {
	// 	funcValue, ok := hl.FuncValue[f]
	// 	if !ok {
	// 		return errors.New("unknown function")
	// 	}
	// 	// err := r.rig.SetFunc(vfo, funcValue, true)
	// 	// if err != nil {
	// 	// 	return err
	// 	// }
	// }

	// // functions to be disabled
	// diff = utils.SliceDiff(r.state.Vfo.Functions, newFuncs)
	// for _, f := range diff {
	// 	funcValue, ok := hl.FuncValue[f]
	// 	if !ok {
	// 		return errors.New("unknown function")
	// 	}
	// 	// err := r.rig.SetFunc(vfo, funcValue, false)
	// 	// if err != nil {
	// 	// 	return err
	// 	// }
	// }

	return nil
}

func (r *remoteRadio) updateLevels(newLevels map[string]float32) error {

	r.state.Vfo.Levels = newLevels

	// vfo, _ := hl.VfoValue[r.state.CurrentVfo]

	// for k, v := range newLevels {
	// 	levelValue, ok := hl.LevelValue[k]
	// 	if !ok {
	// 		return errors.New("unknown Level")
	// 	}
	// 	if _, ok := r.state.Vfo.Levels[k]; !ok {
	// 		return errors.New("unsupported Level for this rig")
	// 	}

	// if r.state.Vfo.Levels[k] != v {
	// 	err := r.rig.SetLevel(vfo, levelValue, v)
	// 	if err != nil {
	// 		return nil
	// 	}
	// }
	// }

	return nil
}

func (r *remoteRadio) updateParams(newParams map[string]float32) error {

	r.state.Vfo.Parameters = newParams

	// vfo, _ := hl.VfoValue[r.state.CurrentVfo]

	// for k, v := range newParams {
	// 	paramValue, ok := hl.ParmValue[k]
	// 	if !ok {
	// 		return errors.New("unknown Parameter")
	// 	}
	// if _, ok := r.state.Vfo.Parameters[k]; !ok {
	// 	return errors.New("unsupported Parameter for this rig")
	// }
	// if r.state.Vfo.Levels[k] != v {
	// 	err := r.rig.SetLevel(vfo, paramValue, v)
	// 	if err != nil {
	// 		return nil
	// 	}
	// }
	// }

	return nil
}

func (r *remoteRadio) deepCopyState() sbRadio.SetState {

	request := sbRadio.SetState{}

	request.CurrentVfo = r.state.CurrentVfo
	request.RadioOn = r.state.RadioOn
	request.Ptt = r.state.Ptt
	request.Vfo = &sbRadio.Vfo{}
	request.Vfo.Vfo = r.state.Vfo.Vfo
	request.Vfo.Frequency = r.state.Vfo.Frequency
	request.Vfo.Mode = r.state.Vfo.Mode
	request.Vfo.PbWidth = r.state.Vfo.PbWidth
	request.Vfo.Ant = r.state.Vfo.Ant
	request.Vfo.Rit = r.state.Vfo.Rit
	request.Vfo.Xit = r.state.Vfo.Xit
	request.Vfo.Split = &sbRadio.Split{}
	request.Vfo.Split.Enabled = r.state.Vfo.Split.Enabled
	request.Vfo.Split.Frequency = r.state.Vfo.Split.Frequency
	request.Vfo.Split.Mode = r.state.Vfo.Split.Mode
	request.Vfo.Split.Vfo = r.state.Vfo.Split.Vfo
	request.Vfo.Split.PbWidth = r.state.Vfo.Split.PbWidth
	request.Vfo.TuningStep = r.state.Vfo.TuningStep
	request.Vfo.Functions = make([]string, 0, len(r.state.Vfo.Functions))
	for _, f := range r.state.Vfo.Functions {
		request.Vfo.Functions = append(request.Vfo.Functions, f)
	}
	request.Vfo.Levels = make(map[string]float32)
	for key, value := range r.state.Vfo.Levels {
		if ok := valueInSlice(key, r.caps.SetLevels); ok {
			request.Vfo.Levels[key] = value
		}
	}
	request.Vfo.Parameters = make(map[string]float32)
	for key, value := range r.state.Vfo.Parameters {
		if ok := valueInSlice(key, r.caps.SetParameters); ok {
			request.Vfo.Parameters[key] = value
		}
	}

	return request
}

func valueInSlice(a string, list []*sbRadio.Value) bool {
	for _, b := range list {
		if b.Name == a {
			return true
		}
	}
	return false
}
