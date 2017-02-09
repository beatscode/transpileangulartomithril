package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/robertkrimen/otto"
)

//AngularApp is made up of multiple angular.modules
type AngularApp struct {
	Modules    []AngularModule
	Components []AngularComponent
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

//Transpiler will hopfully convert important parts to mithril or some other framework
type Transpiler struct {
	Dependencies []string
}

var angular AngularApp
var angularTemplateDirPtr *string
var tranpiler Transpiler
var vm = otto.New()

func (angular *AngularApp) this() otto.Value {
	value, _ := vm.Run("angular")
	return value
}
func (angular *AngularApp) module(call otto.FunctionCall) otto.Value {

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
func (angular *AngularApp) controller(call otto.FunctionCall) otto.Value {
	controllerName, _ := call.Argument(0).ToString()
	functionBody, _ := call.Argument(1).ToString()

	ctrl := AngularComponent{Name: controllerName, Type: "controller", FunctionBody: functionBody}
	ctrl.FindTemplateString()
	angular.Components = append(angular.Components, ctrl)

	return otto.Value{}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

//FindTemplateString searches angular template directory for controller html
func (aComponent *AngularComponent) FindTemplateString() {
	files, err := ioutil.ReadDir(*angularTemplateDirPtr)
	checkErr(err)
	for _, f := range files {
		if filepath.Ext(f.Name()) != ".html" {
			continue
		}

		rawBytes, _ := ioutil.ReadFile(fmt.Sprintf("%s/%s", *angularTemplateDirPtr, f.Name()))
		if strings.Contains(string(rawBytes), aComponent.Name) {
			aComponent.TemplateStr = string(rawBytes)
		}
	}
}

func (tx *Transpiler) transpile(angular AngularApp) {

}

/*
angular.module('onBoardExpressApp').controller('onBoardSearchController', function ($scope, $rootScope, $window, onboard_express_service) {...
*/
func main() {
	angularTemplateDirPtr = flag.String("angularTemplateDir", ".", "Directory to search for controller html")

	componentString := `
	angular.module('myApp',['helloModule','hello2Module']);
	angular.module('myApp').controller('testController',function($scope){
		$scope.myvar = "123";
		$scope.list = [1,2,3,4,5];
	});
	`
	//Set proper mock of angular object
	angularObj, _ := vm.Object(`angular = {}`)

	angularObj.Set("module", angular.module)
	angularObj.Set("controller", angular.controller)

	//Run the file/string to build meta data for transpiling
	if _, err := vm.Run(componentString); err != nil {
		panic(err)
	}

	fmt.Println(angular)
}
