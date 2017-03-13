package angular

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"text/template"

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
	Module          *App
}

var WatchersMock = `var watchers={}`
var JqueryMock = `var $ = function(){}`
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
func (aComponent *Component) addMocks() {
	aComponent.VM.Object(ScopeObjectMock)
	aComponent.VM.Object(WatchersMock)
	aComponent.VM.Object(JqueryMock)

	var err error
	//Mock dependencies
	for _, dep := range aComponent.Dependencies {
		aComponent.VM.Set(dep, func(call otto.FunctionCall) otto.Value {
			return otto.Value{}
		})
		_, err = aComponent.VM.Run(fmt.Sprintf("new %s()", dep))
		if err != nil {
			panic(err)
		}
	}

}

//ParseScopeValues attempts to export $sccope Values/Properties
func (aComponent *Component) ParseScopeValues() {
	var functionVM *otto.Otto

	functionVM = aComponent.VM //otto.New()
	aComponent.addMocks()

	//Set Function Equal to var
	functionAssignment := fmt.Sprintf(`var func = %s`, aComponent.FunctionBody)

	functionVM.Set("controllerFunction", functionAssignment)
	functionEvalCode := fmt.Sprintf(`
		//External Mocks
		%s 
	    //Set and Run Function
	    eval(controllerFunction)
		$scope = scopeObj
	    func(%s)
	`, aComponent.Module.ExternalMocks, strings.Join(aComponent.Dependencies, ","))
	if _, err := functionVM.Run(functionEvalCode); err != nil {
		fmt.Println(functionEvalCode)
		//log.Fatal(err.Error())
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
	//functionVM = aComponent.VM
	aComponent.addMocks()
	for _, function := range aComponent.Functions {
		//This function is usually self invoking
		if function == "init" {
			continue
		}
		//	functionVM = otto.New()
		//Set Function Equal to var
		functionAssignment := fmt.Sprintf(`var func = %s`, aComponent.FunctionBody)
		aComponent.VM.Set("controllerFunction", functionAssignment)
		functionEvalCode := fmt.Sprintf(`
		//External Mocks
		%s 
	    //Set and Run Function
	    eval(controllerFunction)
		$scope = scopeObj
	    func(%s)
	`, aComponent.Module.ExternalMocks, strings.Join(aComponent.Dependencies, ","))
		if _, err := aComponent.VM.Run(functionEvalCode); err != nil {
			//It is hard to get the source of Self invoking functions
			//
			log.Println(functionEvalCode, err.Error())
			continue
			//panic(err)
		}
		var functionBody string
		functionToString := fmt.Sprintf(`scopeObj.%s.toString()`, function)
		if functionString, err := aComponent.VM.Run(functionToString); err != nil {
			//panic(err)
			//It is hard to get the source of Self invoking functions
			//
			log.Println(functionToString, err.Error())
			continue
		} else {
			functionBody = functionString.String()
		}
		aComponent.FunctionBodies[function] = functionBody
	}
}

//FindTemplateString searches angular template directory for controller html
func (aComponent *Component) FindTemplateString() {
	dir := aComponent.Module.TemplateDir
	fstat, err := os.Stat(dir)
	if err != nil {
		fmt.Println(dir)
		panic(err)
	}
	if fstat.IsDir() {
		//check if dir is a file or directory
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
	} else {
		rawBytes, _ := ioutil.ReadFile(dir)
		aComponent.TemplateStr = string(rawBytes)
	}
}

//ExportController exports the controller as a buffer with objects
//and  functions
func (aComponent *Component) ExportController() *bytes.Buffer {

	buf := new(bytes.Buffer)
	if aComponent.TemplateStr == "" {
		return buf
	}
	tmpl := template.New("New Component")
	tmpl = tmpl.Funcs(template.FuncMap{"parseVal": func(args ...interface{}) string {
		fmt.Println(args[0], reflect.TypeOf(args[0]).Kind())
		var r string
		switch reflect.TypeOf(args[0]).Kind() {
		case reflect.String:
			if args[0].(string) == "undefined" {
				r = "null"
			} else if args[0].(string) == "true" || args[0].(string) == "false" || args[0].(string) == "{}" || args[0].(string) == "[]" {
				r = args[0].(string)
			} else {
				r = fmt.Sprintf(`'%s'`, args[0])
			}
			break
		case reflect.Bool:
			if args[0].(bool) == true {
				r = "true"
			} else {
				r = "false"
			}
			break
		}

		return r
	}})

	tmpl, _ = tmpl.Parse(`	
	<script>
	var {{.Name}}Model = {
		{{range $key,$el := .ScopeObject}}
			'{{$key}}':{{$el | parseVal}},
		{{end}}
		{{range $key, $element := .FunctionBodies}}
			'{{$key}}' : {{$element}},
		{{end}}
	};	
	var {{.Name}}Component = {
		oncreate : function(){

		},

		view : function(){
			return (
				{{.TemplateStr}}	
			)
		}
	};
	m.mount(document.body, <{{.Name}}Component />)
	</script>
	`)
	tmpl.Execute(buf, aComponent)
	fmt.Println(buf.String())
	return buf
}
