package provider

import (
	"fmt"
	"reflect"

	"github.com/evcc-io/evcc/provider/golang"
	"github.com/evcc-io/evcc/util"
	"github.com/traefik/yaegi/interp"
)

// Go implements Go request provider
type Go struct {
	vm     *interp.Interpreter
	script string
	read   []TransformationConfig
	write  []TransformationConfig
}

//type TransformationConfig struct {
//	Name, Type string
//	Config     Config
//}

func init() {
	registry.Add("go", NewGoProviderFromConfig)
}

// NewGoProviderFromConfig creates a Go provider
func NewGoProviderFromConfig(other map[string]interface{}) (Provider, error) {
	var cc struct {
		VM     string
		Script string
		Read   []TransformationConfig
		Write  []TransformationConfig
	}

	if err := util.DecodeOther(other, &cc); err != nil {
		return nil, err
	}

	vm, err := golang.RegisteredVM(cc.VM, "")
	if err != nil {
		return nil, err
	}

	p := &Go{
		vm:     vm,
		script: cc.Script,
		read:   cc.Read,
		write:  cc.Write,
	}

	return p, nil
}

// FloatGetter parses float from request
func (p *Go) FloatGetter() func() (float64, error) {
	return func() (res float64, err error) {
		if p.read != nil {
			err = transformGetterGo(p)
		}
		if err == nil {
			var v reflect.Value
			v, err = p.vm.Eval(p.script)
			if err == nil {
				if typ := reflect.TypeOf(res); v.CanConvert(typ) {
					res = v.Convert(typ).Float()
				} else {
					err = fmt.Errorf("not a float: %v", v)
				}
			}
		}
		return res, err
	}
}

// IntGetter parses int64 from request
func (p *Go) IntGetter() func() (int64, error) {
	return func() (res int64, err error) {

		if p.read != nil {
			err = transformGetterGo(p)
		}
		if err == nil {
			var v reflect.Value
			v, err = p.vm.Eval(p.script)
			if err == nil {
				if typ := reflect.TypeOf(res); v.CanConvert(typ) {
					res = v.Convert(typ).Int()
				} else {
					err = fmt.Errorf("not an int: %v", v)
				}
			}
		}

		return res, err
	}
}

// StringGetter parses string from request
func (p *Go) StringGetter() func() (string, error) {
	return func() (res string, err error) {
		if p.read != nil {
			err = transformGetterGo(p)
		}
		if err == nil {
			var v reflect.Value
			v, err = p.vm.Eval(p.script)
			if err == nil {
				if typ := reflect.TypeOf(res); v.CanConvert(typ) {
					res = v.Convert(typ).String()
				} else {
					err = fmt.Errorf("not a string: %v", v)
				}
			}
		}
		return res, err
	}
}

// BoolGetter parses bool from request
func (p *Go) BoolGetter() func() (bool, error) {
	return func() (res bool, err error) {
		if p.read != nil {
			err = transformGetterGo(p)
		}
		if err == nil {
			var v reflect.Value
			v, err = p.vm.Eval(p.script)
			if err == nil {
				if typ := reflect.TypeOf(res); v.CanConvert(typ) {
					res = v.Convert(typ).Bool()
				} else {
					err = fmt.Errorf("not a boolean: %v", v)
				}
			}
		}

		return res, err
	}
}

func (p *Go) paramAndEval(param string, val any) error {
	_, err := p.vm.Eval(fmt.Sprintf("%s := %v;", param, val))
	if err == nil {
		_, err = p.vm.Eval(fmt.Sprintf("param := %v;", param))
	}
	if err == nil {
		_, err = p.vm.Eval(fmt.Sprintf("val := %v;", val))
	}
	if err == nil {
		var v reflect.Value
		v, err = p.vm.Eval(p.script)
		if err == nil && p.write != nil {
			err = transformSetterGo(p.write, v)
		}
	}
	return err
}

// IntSetter sends int request
func (p *Go) IntSetter(param string) func(int64) error {
	return func(val int64) error {
		return p.paramAndEval(param, val)
	}
}

// FloatSetter sends float request
func (p *Go) FloatSetter(param string) func(float64) error {
	return func(val float64) error {
		return p.paramAndEval(param, val)
	}
}

// StringSetter sends string request
func (p *Go) StringSetter(param string) func(string) error {
	return func(val string) error {
		return p.paramAndEval(param, val)
	}
}

// BoolSetter sends bool request
func (p *Go) BoolSetter(param string) func(bool) error {
	return func(val bool) error {
		return p.paramAndEval(param, val)
	}
}

func transformGetterGo(p *Go) error {
	for _, cc := range p.read {
		name := cc.Name
		var val any
		if cc.Type == "bool" {
			f, err := NewBoolGetterFromConfig(cc.Config)
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			val, err = f()
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		} else if cc.Type == "int" {
			f, err := NewIntGetterFromConfig(cc.Config)
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			val, err = f()
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		} else if cc.Type == "float" {
			f, err := NewFloatGetterFromConfig(cc.Config)
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			val, err = f()
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		} else {
			f, err := NewStringGetterFromConfig(cc.Config)
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			val, err = f()
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		}
		err := p.paramAndEval(name, val)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	return nil
}
func transformSetterGo(transforms []TransformationConfig, v reflect.Value) error {
	for _, cc := range transforms {
		name := cc.Name
		if cc.Type == "bool" {
			f, err := NewBoolSetterFromConfig(name, cc.Config)
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			if v.CanConvert(reflect.TypeOf(true)) {
				err = f(v.Convert(reflect.TypeOf(true)).Bool())
			} else {
				err = fmt.Errorf("not a int: %s", v)
			}
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		} else if cc.Type == "int" {
			f, err := NewIntSetterFromConfig(name, cc.Config)
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			if v.CanConvert(reflect.TypeOf(0)) {
				err = f(v.Convert(reflect.TypeOf(0)).Int())
			} else {
				err = fmt.Errorf("not a int: %s", v)
			}
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		} else if cc.Type == "float" {
			f, err := NewFloatSetterFromConfig(name, cc.Config)
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			if v.CanConvert(reflect.TypeOf(0.0)) {
				err = f(v.Convert(reflect.TypeOf(0.0)).Float())
			} else {
				err = fmt.Errorf("not a int: %s", v)
			}
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		} else {
			f, err := NewStringSetterFromConfig(name, cc.Config)
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			if v.CanConvert(reflect.TypeOf("")) {
				err = f(v.Convert(reflect.TypeOf("")).String())
			} else {
				err = fmt.Errorf("not a int: %s", v)
			}
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		}
	}
	return nil
}
