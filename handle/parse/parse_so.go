package parse

import (
	"fmt"
	"plugin"
	"reflect"
	"strings"

	"pressure-test/config"
)

type SoSymbol struct { //插件解析
	VUser int `json:"v_user"`
}

func (s *SoSymbol) ParserSoSymbol(p *plugin.Plugin, caseName string) error {
	t := reflect.TypeOf(s)
	v := reflect.ValueOf(s)
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("not %v", reflect.Ptr)
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("not %v", reflect.Struct)
	}
	t = v.Type()
	var name string
	for i := 0; i < t.NumField(); i++ {
		name = t.Field(i).Tag.Get("json")
		symbol, err := p.Lookup(caseName + strings.Title(name))
		if err != nil {
			continue
		}
		v.Field(i).Set(reflect.ValueOf(symbol))
	}
	return nil
}

var SoObj = new(SoSymbol)

//SoFileParse so文件解析
func SoFileParse() error {
	binaryPath := fmt.Sprintf("%s/%s.so", config.Cfg.SoFilePath, config.Cfg.CaseName)
	p, err := plugin.Open(binaryPath)
	if err != nil {
		return err
	}
	//impl.GetPredefinedLocalVariables(localPredefinedVariablesFunc)
	return SoObj.ParserSoSymbol(p, config.Cfg.CaseName)
}
