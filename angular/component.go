package angular

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/robertkrimen/otto"
)

//Component is one of the following: Controller,Directive,Factory,Service,Filter, etc
type Component struct {
	Type            string
	Name            string
	FunctionBody    string
	Dependencies    []string
	TemplateStr     string
	Functions       []string
	FunctionBodies  map[string]string
	ScopeProperties []string
	ScopeObject     interface{}
	VM              *otto.Otto
}

var WatchersMock = `var watchers={}`
var JqueryMock = `var $ = {}`
var ScopeObjectMock = `var scopeObj = {
	'$on':function(eventName,func){
		watchers[eventName] = func.toString();
	}
}`

func (aComponent *Component) ParseScopeProperties() {
	//Make sure $scope is first
	//Get Scope Variables that are not functions
	//Normally would use negative lookahead but
	//not supported
	expression := `(?:\$scope\.)(?P<variable>\w+)\s+=\s+[^f]`
	reg := regexp.MustCompile(expression)

	matches := reg.FindAllStringSubmatch(aComponent.FunctionBody, -1)
	found := false
	var scopeProperties []string
	for _, v := range matches {
		for _, tmpV := range scopeProperties {
			if tmpV == v[1] {
				found = true
			}
		}
		if found == false {
			scopeProperties = append(scopeProperties, v[1])
		}
	}

	aComponent.ScopeProperties = scopeProperties
}
func addMocks(vm *otto.Otto) {
	vm.Object(ScopeObjectMock)
	vm.Object(WatchersMock)
	vm.Object(JqueryMock)
}

//ParseScopeValues attempts to export $sccope Values/Properties
func (aComponent *Component) ParseScopeValues() {
	var functionVM *otto.Otto
	functionVM = otto.New()
	addMocks(functionVM)
	//Set Function Equal to var
	functionAssignment := fmt.Sprintf(`var func = %s`, aComponent.FunctionBody)
	functionVM.Set("controllerFunction", functionAssignment)
	if _, err := functionVM.Run(`
	    //Set and Run Function
	    eval(controllerFunction)
	    func(scopeObj)
	`); err != nil {
		panic(err)
	}
	var scopeObjectInterface interface{}
	if scopeobject, err := functionVM.Get("scopeObj"); err == nil {
		scopeObjectInterface, err = scopeobject.Export()
		if err != nil {
			panic(err)
		}
	}
	aComponent.VM = functionVM

	for key, prop := range scopeObjectInterface.(map[string]interface{}) {
		if key == "$on" {
			delete(scopeObjectInterface.(map[string]interface{}), key)
			continue
		}

		//Check if item is a function
		//If so then remove it
		var found = false
		for k := range aComponent.FunctionBodies {
			fmt.Println("Function", k)
			if key == k {
				found = true
			}
		}
		if found {
			delete(scopeObjectInterface.(map[string]interface{}), key)
			continue
		}
		//For Objects/Arrays get json string to represent value
		switch reflect.TypeOf(prop).Kind() {
		case reflect.Map, reflect.Slice:
			if value, err := aComponent.VM.Run(`JSON.stringify(scopeObj.` + key + `)`); err == nil {
				jsonValue := value.String()
				scopeObjectInterface.(map[string]interface{})[key] = jsonValue
			}
			break
		}
	}
	aComponent.ScopeObject = scopeObjectInterface
}
func (aComponent *Component) ParseScopeFunctions() {
	//Get Scope Variables that are not functions
	expression := `(?:\$scope\.)(?P<variable>\w+)\s+=\s+function`
	reg := regexp.MustCompile(expression)

	matches := reg.FindAllStringSubmatch(aComponent.FunctionBody, -1)

	var functions []string
	for _, v := range matches {
		functions = append(functions, v[1])
	}

	aComponent.Functions = functions
}

//ParseFunctionBodies parses controller bodies for function bodies
func (aComponent *Component) ParseFunctionBodies() {
	if len(aComponent.Functions) == 0 {
		return
	}
	aComponent.FunctionBodies = make(map[string]string)
	var functionVM *otto.Otto

	for _, function := range aComponent.Functions {
		functionVM = otto.New()
		addMocks(functionVM)
		//Set Function Equal to var
		functionAssignment := fmt.Sprintf(`var func = %s`, aComponent.FunctionBody)
		functionVM.Set("controllerFunction", functionAssignment)
		if _, err := functionVM.Run(`
        //Set and Run Function
        eval(controllerFunction)
        func(scopeObj)
    `); err != nil {
			panic(err)
		}
		var functionBody string
		functionToString := fmt.Sprintf(`scopeObj.%s.toString()`, function)
		if functionString, err := functionVM.Run(functionToString); err != nil {
			panic(err)
		} else {
			functionBody = functionString.String()
		}
		aComponent.FunctionBodies[function] = functionBody
	}
}

//FindTemplateString searches angular template directory for controller html
func (aComponent *Component) FindTemplateString(dir string) {
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
