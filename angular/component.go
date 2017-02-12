package angular

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/robertkrimen/otto"
)

//Component is one of the following: Controller,Directive,Factory,Service,Filter, etc
type Component struct {
	Type           string
	Name           string
	FunctionBody   string
	TemplateStr    string
	Functions      []string
	FunctionBodies map[string]string
	ScopeProperies []string
}

func (aComponent *Component) ParseScopeProperties() {
	//Get Scope Variables that are not functions
	expression := `(?:\$scope\.)(?P<variable>\w+)\s+=\s+[^function]`
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
	aComponent.ScopeProperies = scopeProperties
}

func (aComponent *Component) ParseScopeFunctions() {
	//Get Scope Variables that are not functions
	expression := `(?:\$scope\.)(?P<variable>\w+)\s+=\s+[function]`
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
		functionVM.Object(`scopeObj = {}`)
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
