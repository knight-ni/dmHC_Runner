package cfgparser

import (
	"github.com/Unknwon/goconfig"
	"github.com/sirupsen/logrus"
	"strconv"
)

type Cfile struct {
	Path string
	Cfg  *goconfig.ConfigFile
	Log  *logrus.Logger
}

//func GetCurrentPath() (string, error) {
//	file, err := exec.LookPath(os.Args[0])
//	if err != nil {
//		return "", err
//	}
//	path, err := filepath.Abs(file)
//	if err != nil {
//		return "", err
//	}
//	i := strings.LastIndex(path, "/")
//	if i < 0 {
//		i = strings.LastIndex(path, "\\")
//	}
//	if i < 0 {
//		return "", errors.New("error: Can't find / or \\")
//	}
//	return path[0 : i+1], nil
//}

func (cf *Cfile) Initialize(mylog *logrus.Logger) {
	var err error
	cf.Log = mylog
	cf.Cfg, err = goconfig.LoadConfigFile(cf.Path)

	if err != nil {
		mylog.Fatalf("can't open config file! error:%s", err)
	}
}

func (cf *Cfile) InitNonFatal(mylog *logrus.Logger) error {
	var err error
	cf.Log = mylog
	cf.Cfg, err = goconfig.LoadConfigFile(cf.Path)

	if err != nil {
		return err
	}
	return nil
}

func (cf *Cfile) GetStrVal(section string, key string) string {
	val, err := cf.Cfg.GetValue(section, key)
	if err != nil {
		cf.Log.Warnf("get config value failed! error:%s", err)
		val = ""
	}
	return val
}

func (cf *Cfile) GetIntVal(section string, key string) int64 {
	tmpval := cf.GetStrVal(section, key)
	intval, err := strconv.ParseInt(tmpval, 10, 64)
	if err != nil {
		cf.Log.Warnf("config value %s not valid int value! error:%s", key, err)
		intval = 0
	}
	return intval
}

func (cf *Cfile) GetFltVal(section string, key string) float64 {
	tmpval := cf.GetStrVal(section, key)
	fltval, err := strconv.ParseFloat(tmpval, 64)
	if err != nil {
		cf.Log.Warnf("config value %s not valid float value! error:%s", key, err)
		fltval = 0.00
	}
	return fltval
}

func (cf *Cfile) GetUIntVal(section string, key string) uint64 {
	tmpval := cf.GetStrVal(section, key)
	uintval, err := strconv.ParseUint(tmpval, 10, 64)
	if err != nil {
		cf.Log.Warnf("config value %s not valid int value! error:%s", key, err)
		uintval = 0
	}
	return uintval
}
