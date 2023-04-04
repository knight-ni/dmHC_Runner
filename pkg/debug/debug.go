package debug

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"
)

const dtfmt = "2006-01-02 15:04:05.00000"

var Debug int64 = 0
var SkipLst []string
var TraceName string

type FuncProxy struct {
	realfun func()
}

func NewFuncProxy(myfunc func()) *FuncProxy {
	//返回静态代理类
	return &FuncProxy{
		realfun: myfunc,
	}
}

func (p *FuncProxy) RunFunc() {
	if Debug == 1 {
		funcname := GetFunctionName(p.realfun)
		fmt.Println(errors.New(fmt.Sprintf("%s Enter:%s", time.Now().Format(dtfmt), funcname)))
		if ChkSkip(p.realfun) != true {
			p.realfun()
		}
		fmt.Println(errors.New(fmt.Sprintf("%s Exit:%s", time.Now().Format(dtfmt), funcname)))
	} else {
		p.realfun()
	}
}

func GetFunctionName(i interface{}, seps ...rune) string {
	// 获取函数名称
	fn := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	fields := strings.FieldsFunc(fn, func(sep rune) bool {
		for _, s := range seps {
			if sep == s {
				return true
			}
		}
		return false
	})
	//fmt.Println(fields)
	if size := len(fields); size > 0 {
		return fields[size-1]
	}
	return ""
}

func Trace() func() {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("not found caller")
	}

	fn := runtime.FuncForPC(pc)
	funcname := fn.Name()
	fmt.Println(errors.New(fmt.Sprintf("%s Enter:%s", time.Now().Format(dtfmt), funcname)))
	return func() {
		fmt.Println(errors.New(fmt.Sprintf("%s Exit:%s", time.Now().Format(dtfmt), funcname)))
	}
}

func in(target string, skiparray []string) bool {
	sort.Strings(skiparray)
	index := sort.SearchStrings(skiparray, target)
	if index < len(skiparray) && skiparray[index] == target {
		return true
	}
	return false
}

func ChkSkip(myfunc func()) bool {
	funcname := GetFunctionName(myfunc)
	fulllst := strings.Split(funcname, ".")
	realname := fulllst[len(fulllst)-1]
	if in(realname, SkipLst) == true {
		fmt.Println(errors.New(fmt.Sprintf("%s Skipped:%s", time.Now().Format(dtfmt), funcname)))
		return true
	} else {
		fmt.Println(errors.New(fmt.Sprintf("%s Processing:%s", time.Now().Format(dtfmt), funcname)))
		return false
	}
}
