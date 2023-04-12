package cfgparser

import (
	"github.com/Unknwon/goconfig"
	"strconv"
)

type Cfile struct {
	Path string
	Cfg  *goconfig.ConfigFile
}

func (cf *Cfile) Initialize() {
	var err error
	cf.Cfg, err = goconfig.LoadConfigFile(cf.Path)
	if err != nil {
		panic("Can't Open Config File!" + cf.Path)
	}
}

func (cf *Cfile) InitNonFatal() error {
	var err error
	cf.Cfg, err = goconfig.LoadConfigFile(cf.Path)
	if err != nil {
		panic("Can't Open Config File!" + cf.Path)
	}
	return nil
}

func (cf *Cfile) GetStrVal(section string, key string) string {
	val, err := cf.Cfg.GetValue(section, key)
	if err != nil {
		panic("get config value failed! error:" + err.Error())
	}
	return val
}

func (cf *Cfile) GetBigIntVal(section string, key string) int64 {
	tmpval := cf.GetStrVal(section, key)
	intval, err := strconv.ParseInt(tmpval, 10, 64)
	if err != nil {
		panic("get config value " + key + " failed! error:" + err.Error())
	}
	return intval
}

func (cf *Cfile) GetIntVal(section string, key string) int {
	tmpval := cf.GetStrVal(section, key)
	intval, err := strconv.ParseInt(tmpval, 10, 64)
	if err != nil {
		panic("config value " + key + " not valid int value! error:" + err.Error())
	}
	return int(intval)
}

func (cf *Cfile) GetFltVal(section string, key string) float64 {
	tmpval := cf.GetStrVal(section, key)
	fltval, err := strconv.ParseFloat(tmpval, 64)
	if err != nil {
		panic("config value " + key + " not valid float value! error:" + err.Error())
	}
	return fltval
}

func (cf *Cfile) GetUIntVal(section string, key string) uint64 {
	tmpval := cf.GetStrVal(section, key)
	uintval, err := strconv.ParseUint(tmpval, 10, 64)
	if err != nil {
		panic("config value " + key + " not valid uint value! error:" + err.Error())
	}
	return uintval
}

func (cf *Cfile) SetStrVal(section string, key string, value string) {
	cf.Cfg.SetValue(section, key, value)
}

func (cf *Cfile) SetIntVal(section string, key string, value int) {
	cf.Cfg.SetValue(section, key, strconv.Itoa(value))
}

func (cf *Cfile) SetFltVal(section string, key string, value float64) {
	cf.Cfg.SetValue(section, key, strconv.FormatFloat(value, 'E', -1, 32))
}

func (cf *Cfile) SetUIntVal(section string, key string, value uint64) {
	cf.Cfg.SetValue(section, key, strconv.FormatUint(value, 10))
}

func (cf *Cfile) SaveFile(path string) {
	err := goconfig.SaveConfigFile(cf.Cfg, path)
	if err != nil {
		panic("Config File Save Failed! " + err.Error())
	}
}
