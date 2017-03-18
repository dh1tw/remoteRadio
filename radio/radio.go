package radio

import (
	"errors"
	"log"
	"sync"

	"github.com/cskr/pubsub"
	hl "github.com/dh1tw/goHamlib"
	"github.com/dh1tw/remoteRadio/comms"
	sbRadio "github.com/dh1tw/remoteRadio/sb_radio"
)

type RadioSettings struct {
	RigModel         int
	Port             hl.Port
	CatRequestCh     chan []byte
	ToWireCh         chan comms.IOMsg
	CatResponseTopic string
	CapsTopic        string
	WaitGroup        *sync.WaitGroup
	Events           *pubsub.PubSub
}

type radio struct {
	rig      hl.Rig
	state    sbRadio.State
	settings *RadioSettings
}

func HandleRadio(rs RadioSettings) {

	defer rs.WaitGroup.Done()

	r := radio{}
	r.rig = hl.Rig{}
	r.state = sbRadio.State{}
	r.state.Vfo = &sbRadio.Vfo{}
	r.state.Channel = &sbRadio.Channel{}
	r.settings = &rs

	// err := r.rig.SetPort(rs.Port)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	err := r.rig.Init(rs.RigModel)
	if err != nil {
		log.Println(err)
		return
	}

	if err := r.rig.Open(); err != nil {
		log.Println(err)
	}

	if err := r.sendCaps(); err != nil {
		log.Println(err)
	}

	if err := r.queryVfo(); err != nil {
		log.Println(err)
	}

	if err := r.sendState(); err != nil {
		log.Println(err)
	}

	for {
		select {
		case msg := <-rs.CatRequestCh:
			r.deserializeCatRequest(msg)
		}
	}
}

func (r *radio) queryVfo() error {
	vfo, err := r.rig.GetVfo()
	if err != nil {
		return err
	}
	r.state.CurrentVfo = hl.VfoName[vfo]
	r.state.Vfo.Vfo = hl.VfoName[vfo]

	if pwrOn, err := r.rig.GetPowerStat(); err != nil {
		return err
	} else {
		if pwrOn == hl.RIG_POWER_ON {
			r.state.RadioOn = true
		} else {
			r.state.RadioOn = false
		}
	}

	freq, err := r.rig.GetFreq(vfo)
	if err != nil {
		return err
	}
	r.state.Vfo.Frequency = freq

	mode, pbWidth, err := r.rig.GetMode(vfo)
	if err != nil {
		return err
	}
	if modeName, ok := hl.ModeName[mode]; ok {
		r.state.Vfo.Mode = modeName
	} else {
		return errors.New("unknown mode")
	}

	r.state.Vfo.PbWidth = int32(pbWidth)

	ant, err := r.rig.GetAnt(vfo)
	if err != nil {
		return err
	}
	r.state.Vfo.Ant = int32(ant)

	rit, err := r.rig.GetRit(vfo)
	if err != nil {
		return err
	}
	r.state.Vfo.Rit = int32(rit)

	xit, err := r.rig.GetXit(vfo)
	if err != nil {
		return err
	}
	r.state.Vfo.Xit = int32(xit)

	splitOn, txVfo, err := r.rig.GetSplit(vfo)
	if err != nil {
		return err
	}

	txFreq, err := r.rig.GetSplitFreq(vfo)
	if err != nil {
		return err
	}

	txMode, txPbWidth, err := r.rig.GetSplitMode(vfo)
	if err != nil {
		return err
	}

	split := sbRadio.Split{}
	if splitOn == 1 {
		split.Enabled = true
	} else {
		split.Enabled = false
	}

	split.Frequency = txFreq
	if txVfoName, ok := hl.VfoName[txVfo]; ok {
		split.Vfo = txVfoName
	} else {
		return errors.New("unknown Vfo Name")
	}

	if txModeName, ok := hl.ModeName[txMode]; ok {
		split.Mode = txModeName
	} else {
		return errors.New("unknown Mode")
	}

	split.PbWidth = uint32(txPbWidth)

	r.state.Vfo.Split = &split

	tStep, err := r.rig.GetTs(vfo)
	if err != nil {
		return err
	}
	r.state.Vfo.TuningStep = int32(tStep)

	r.state.Vfo.Functions = make([]string, 0, len(hl.FuncName))

	for _, f := range r.rig.Caps.GetFunctions {
		fValue, err := r.rig.GetFunc(vfo, hl.FuncValue[f])
		if err != nil {
			return err
		}
		if fValue {
			r.state.Vfo.Functions = append(r.state.Vfo.Functions, f)
		}
	}

	r.state.Vfo.Levels = make(map[string]float32)
	for _, level := range r.rig.Caps.GetLevels {
		lValue, err := r.rig.GetLevel(vfo, hl.LevelValue[level.Name])
		if err != nil {
			// return err
			log.Println("Warning:", level.Name, "-", err)
		}
		r.state.Vfo.Levels[level.Name] = lValue
	}

	r.state.Vfo.Paramters = make(map[string]float32)
	for _, param := range r.rig.Caps.GetParameters {
		pValue, err := r.rig.GetParm(vfo, hl.ParmValue[param.Name])
		if err != nil {
			return err
		}
		r.state.Vfo.Paramters[param.Name] = pValue
	}

	return nil
}

func (r *radio) sendState() error {

	if state, err := r.state.Marshal(); err == nil {
		stateMsg := comms.IOMsg{}
		stateMsg.Data = state
		stateMsg.Retain = true
		stateMsg.Topic = r.settings.CatResponseTopic
		r.settings.ToWireCh <- stateMsg
	} else {
		return err
	}

	return nil
}

func (r *radio) sendCaps() error {

	if caps, err := r.serializeCaps(); err == nil {
		capsMsg := comms.IOMsg{}
		capsMsg.Data = caps
		capsMsg.Retain = true
		capsMsg.Topic = r.settings.CapsTopic
		r.settings.ToWireCh <- capsMsg
	} else {
		log.Println(err)
	}

	return nil
}
