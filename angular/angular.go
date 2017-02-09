package angular

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/robertkrimen/otto"
)

//AngularApp is made up of multiple angular.modules
type AngularApp struct {
	Modules     []AngularModule
	Components  []AngularComponent
	VM          *otto.Otto
	TemplateDir string
}

//AngularComponent is one of the following: Controller,Directive,Factory,Service,Filter, etc
type AngularComponent struct {
	Type         string
	Name         string
	FunctionBody string
	TemplateStr  string
}

//AngularModule has dependencies and are tied by the same ng-app directive
type AngularModule struct {
	Name         string
	Dependencies []string
}

func (angular *AngularApp) this() otto.Value {
	value, _ := angular.VM.Run("angular")
	return value
}
func (angular *AngularApp) Module(call otto.FunctionCall) otto.Value {

	modulename := call.Argument(0).String()
	//Find out what type of component this is
	if call.Argument(1).Class() == "Array" {
		deps, _ := call.Argument(1).Export()
		found := false
		for _, v := range angular.Modules {
			if v.Name == modulename {
				found = true
			}
		}
		if !found {
			module := AngularModule{Name: modulename, Dependencies: deps.([]string)}
			angular.Modules = append(angular.Modules, module)
		}
	}
	return angular.this()
}
func (angular *AngularApp) Controller(call otto.FunctionCall) otto.Value {
	controllerName, _ := call.Argument(0).ToString()
	functionBody, _ := call.Argument(1).ToString()

	ctrl := AngularComponent{Name: controllerName, Type: "controller", FunctionBody: functionBody}
	ctrl.FindTemplateString(angular.TemplateDir)
	angular.Components = append(angular.Components, ctrl)

	return otto.Value{}
}

//FindTemplateString searches angular template directory for controller html
func (aComponent *AngularComponent) FindTemplateString(dir string) {
	files, err := ioutil.ReadDir(dir)
	checkErr(err)
	for _, f := range files {
		if filepath.Ext(f.Name()) != ".html" {
			continue
		}

		rawBytes, _ := ioutil.ReadFile(fmt.Sprintf("%s/%s", dir, f.Name()))
		if strings.Contains(string(rawBytes), aComponent.Name) {
			aComponent.TemplateStr = string(rawBytes)
		}
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
