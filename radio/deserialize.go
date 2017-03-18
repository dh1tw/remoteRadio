package radio

import (
	"errors"
	"log"
	"reflect"

	hl "github.com/dh1tw/goHamlib"
	sbRadio "github.com/dh1tw/remoteRadio/sb_radio"
	"github.com/dh1tw/remoteRadio/utils"
)

func (r *radio) deserializeCatRequest(request []byte) error {

	ns := sbRadio.SetState{}
	if err := ns.Unmarshal(request); err != nil {
		return err
	}

	if ns.CurrentVfo != r.state.CurrentVfo {
		if err := r.updateCurrentVfo(ns.CurrentVfo); err != nil {
			log.Println(err)
		}
	}

	if ns.VfoOperations != nil {
		if err := r.execVfoOperations(ns.GetVfoOperations()); err != nil {
			log.Println(err)
		}
	}

	if ns.Vfo != nil {
		if ns.Vfo.GetMode() != r.state.Vfo.Mode {
			if err := r.updateMode(ns.Vfo.GetMode()); err != nil {
				log.Println(err)
			}
		}

		if ns.Vfo.GetPbWidth() != r.state.Vfo.PbWidth {
			if err := r.updatePbWidth(ns.Vfo.GetPbWidth()); err != nil {
				log.Println(err)
			}
		}

		if ns.Vfo.GetAnt() != r.state.Vfo.Ant {
			if err := r.updateAntenna(ns.Vfo.GetAnt()); err != nil {
				log.Println(err)
			}
		}

		if ns.Vfo.GetRit() != r.state.Vfo.Rit {
			if err := r.updateRit(ns.Vfo.GetRit()); err != nil {
				log.Println(err)
			}
		}

		if ns.Vfo.GetXit() != r.state.Vfo.Xit {
			if err := r.updateXit(ns.Vfo.GetXit()); err != nil {
				log.Println(err)
			}
		}

		if ns.Vfo.Split != nil {
			if !reflect.DeepEqual(ns.Vfo.Split, r.state.Vfo.Split) {
				if err := r.updateSplit(ns.Vfo.Split); err != nil {
					log.Println(err)
				}
			}
		}

		if ns.Vfo.GetTuningStep() != r.state.Vfo.TuningStep {
			if err := r.updateTs(ns.Vfo.GetTuningStep()); err != nil {
				log.Println(err)
			}
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

		if ns.Vfo.Paramters != nil {
			if !reflect.DeepEqual(ns.Vfo.Paramters, r.state.Vfo.Paramters) {
				if err := r.updateParams(ns.Vfo.GetParamters()); err != nil {
					log.Println(err)
				}
			}
		}
	}

	if ns.GetRadioOn() != r.state.RadioOn {
		if err := r.updatePowerOn(ns.GetRadioOn()); err != nil {
			log.Println(err)
		}
	}

	if ns.GetPtt() != r.state.Ptt {
		if err := r.updatePtt(ns.GetPtt()); err != nil {
			log.Println(err)
		}
	}

	return nil
}

func (r *radio) updateCurrentVfo(newVfo string) error {
	if vfo, ok := hl.VfoValue[newVfo]; ok {
		err := r.rig.SetVfo(vfo)
		if err != nil {
			return err
		}
		r.state.CurrentVfo = newVfo
		r.state.Vfo.Vfo = newVfo
	} else {
		return errors.New("unknown Vfo")
	}
	return nil
}

func (r *radio) updateFrequency(newFreq float64) error {
	vfo, _ := hl.VfoValue[r.state.CurrentVfo]
	err := r.rig.SetFreq(vfo, newFreq)
	if err != nil {
		return err
	}
	r.state.Vfo.Frequency = newFreq
	return nil
}

func (r *radio) execVfoOperations(vfoOps []string) error {
	vfo, _ := hl.VfoValue[r.state.CurrentVfo]
	for _, v := range vfoOps {
		vfoOpValue, ok := hl.VfoValue[v]
		if !ok {
			return errors.New("unknown VFO Operation")
		}
		err := r.rig.VfoOp(vfo, vfoOpValue)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *radio) updateMode(newMode string) error {
	vfo, _ := hl.VfoValue[r.state.CurrentVfo]
	pbWidth := int(r.state.Vfo.PbWidth)
	newModeValue, ok := hl.ModeValue[newMode]
	if !ok {
		return errors.New("unknown mode")
	}
	err := r.rig.SetMode(vfo, newModeValue, pbWidth)
	if err != nil {
		return err
	}

	r.state.Vfo.Mode = newMode

	return nil
}

func (r *radio) updatePbWidth(newPbWidth int32) error {
	vfo, _ := hl.VfoValue[r.state.CurrentVfo]
	modeValue := hl.ModeValue[r.state.Vfo.Mode]
	err := r.rig.SetMode(vfo, modeValue, int(newPbWidth))
	if err != nil {
		return err
	}
	r.state.Vfo.PbWidth = newPbWidth

	return nil
}

func (r *radio) updateAntenna(newAnt int32) error {
	vfo, _ := hl.VfoValue[r.state.CurrentVfo]
	err := r.rig.SetAnt(vfo, int(newAnt))
	if err != nil {
		return err
	}
	r.state.Vfo.Ant = newAnt

	return nil
}

func (r *radio) updateRit(newRit int32) error {
	vfo, _ := hl.VfoValue[r.state.CurrentVfo]
	err := r.rig.SetRit(vfo, int(newRit))
	if err != nil {
		return err
	}
	r.state.Vfo.Rit = newRit

	return nil
}

func (r *radio) updateXit(newXit int32) error {
	vfo, _ := hl.VfoValue[r.state.CurrentVfo]
	err := r.rig.SetXit(vfo, int(newXit))
	if err != nil {
		return err
	}
	r.state.Vfo.Xit = newXit

	return nil
}

func (r *radio) updateSplit(newSplit *sbRadio.Split) error {
	vfo, _ := hl.VfoValue[r.state.CurrentVfo]
	if newSplit.GetEnabled() != r.state.Vfo.Split.Enabled {
		err := r.rig.SetSplit(vfo, utils.Btoi(newSplit.GetEnabled()))
		if err != nil {
			return err
		}
	}

	if newSplit.GetEnabled() {
		if newSplit.GetFrequency() != r.state.Vfo.Split.Frequency {
			err := r.rig.SetSplitFreq(vfo, newSplit.GetFrequency())
			if err != nil {
				return err
			}
			r.state.Vfo.Split.Enabled = newSplit.GetEnabled()
		}

		if newSplit.GetMode() != r.state.Vfo.Split.Mode {
			newSplitModeValue, ok := hl.ModeValue[newSplit.GetMode()]
			if !ok {
				return errors.New("unknown split mode")
			}
			err := r.rig.SetSplitMode(vfo, newSplitModeValue, int(r.state.Vfo.Split.PbWidth))
			if err != nil {
				return err
			}
			r.state.Vfo.Split.Mode = newSplit.GetMode()
		}

		if newSplit.GetPbWidth() != r.state.Vfo.Split.PbWidth {
			splitModeValue := hl.ModeValue[newSplit.GetMode()]
			err := r.rig.SetSplitMode(vfo, splitModeValue, int(newSplit.GetPbWidth()))
			if err != nil {
				return err
			}
			r.state.Vfo.Split.PbWidth = newSplit.GetPbWidth()
		}
	}

	return nil
}

func (r *radio) updateTs(newTs int32) error {
	vfo, _ := hl.VfoValue[r.state.CurrentVfo]
	err := r.rig.SetTs(vfo, int(newTs))
	if err != nil {
		return err
	}
	r.state.Vfo.TuningStep = newTs

	return nil
}

func (r *radio) updateFunctions(newFuncs []string) error {
	vfo, _ := hl.VfoValue[r.state.CurrentVfo]

	// functions to be enabled
	diff := utils.SliceDiff(newFuncs, r.state.Vfo.Functions)
	for _, f := range diff {
		funcValue, ok := hl.FuncValue[f]
		if !ok {
			return errors.New("unknown function")
		}
		err := r.rig.SetFunc(vfo, funcValue, true)
		if err != nil {
			return err
		}
	}

	// functions to be disabled
	diff = utils.SliceDiff(r.state.Vfo.Functions, newFuncs)
	for _, f := range diff {
		funcValue, ok := hl.FuncValue[f]
		if !ok {
			return errors.New("unknown function")
		}
		err := r.rig.SetFunc(vfo, funcValue, false)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *radio) updateLevels(newLevels map[string]float32) error {
	vfo, _ := hl.VfoValue[r.state.CurrentVfo]

	for k, v := range newLevels {
		levelValue, ok := hl.LevelValue[k]
		if !ok {
			return errors.New("unknown Level")
		}
		if _, ok := r.state.Vfo.Levels[k]; !ok {
			return errors.New("unsupported Level for this rig")
		}

		if r.state.Vfo.Levels[k] != v {
			err := r.rig.SetLevel(vfo, levelValue, v)
			if err != nil {
				return nil
			}
		}
	}

	return nil
}

func (r *radio) updateParams(newParams map[string]float32) error {
	vfo, _ := hl.VfoValue[r.state.CurrentVfo]

	for k, v := range newParams {
		paramValue, ok := hl.ParmValue[k]
		if !ok {
			return errors.New("unknown Parameter")
		}
		if _, ok := r.state.Vfo.Paramters[k]; !ok {
			return errors.New("unsupported Parameter for this rig")
		}
		if r.state.Vfo.Levels[k] != v {
			err := r.rig.SetLevel(vfo, paramValue, v)
			if err != nil {
				return nil
			}
		}
	}

	return nil
}

func (r *radio) updatePowerOn(pwrOn bool) error {

	var pwrStat int
	if pwrOn {
		pwrStat = hl.RIG_POWER_ON
	} else {
		pwrStat = hl.RIG_POWER_OFF
	}

	err := r.rig.SetPowerStat(pwrStat)
	if err != nil {
		return err
	}

	return nil
}

func (r *radio) updatePtt(ptt bool) error {

	var pttValue int
	if ptt {
		pttValue = hl.RIG_PTT_ON
	} else {
		pttValue = hl.RIG_PTT_OFF
	}

	err := r.rig.SetPowerStat(pttValue)
	if err != nil {
		return err
	}

	return nil
}
