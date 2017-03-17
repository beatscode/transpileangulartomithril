package angular

import (
	"fmt"
	"log"
	"strings"

	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/parser"

	"reflect"

	"github.com/robertkrimen/otto"
	"github.com/wolfgarnet/walker"
)

//App is made up of multiple angular.modules
type App struct {
	Modules       []AngularModule
	Components    []Component
	VM            *otto.Otto
	TemplateDir   string
	ExternalMocks string
}

//AngularModule has dependencies and are tied by the same ng-app directive
type AngularModule struct {
	Name         string
	Dependencies []string
}

func (angular *App) this() otto.Value {
	value, _ := angular.VM.Run("angular")
	return value
}

//IntrospectScope will try to study
func (app *App) IntrospectScope() {

	vh := &walker.Hook{}
	var programMetadata []walker.Metadata
	vh.OnNode = func(node ast.Node, metadata []walker.Metadata) error {
		//fmt.Println("node", node.Idx0())
		//fmt.Println("metadata", metadata)
		programMetadata = metadata
		return nil
	}

	filename := "" // A filename is optional
	src := `
	// Sample xyzzy example
	(function(){
		if (3.14159 > 0) {
			console.log("Hello, World.");
			return;
		}

		var xyzzy = NaN;
		console.log("Nothing happens.");
		return xyzzy;
	})();
	`

	// Parse some JavaScript, yielding a *ast.Program and/or an ErrorList
	program, _ := parser.ParseFile(nil, filename, src, 0)
	//fmt.Println(program.Body)
	for k, v := range program.Body {
		fmt.Println(k, v)
	}
	visitor := &walker.VisitorImpl{}
	visitor.AddHook(vh)
	_walker := walker.NewWalker(visitor)
	_walker.Begin(program)

	//	fmt.Println(programMetadata)
	/*
		for _, y := range programMetadata {
			fmt.Println(y.(ast.Node))
		}*/
}

func (angular *App) Module(call otto.FunctionCall) otto.Value {
	found := false
	modulename := call.Argument(0).String()

	//Find out what type of component this is
	if call.Argument(1).Class() == "Array" {
		deps, _ := call.Argument(1).Export()

		for _, v := range angular.Modules {
			if v.Name == modulename {
				found = true
			}
		}
		if !found {
			module := AngularModule{Name: modulename, Dependencies: deps.([]string)}
			angular.Modules = append(angular.Modules, module)
		}
	} else {
		for _, v := range angular.Modules {
			if v.Name == modulename {
				found = true
			}
		}
		if !found {
			module := AngularModule{Name: modulename, Dependencies: nil}
			angular.Modules = append(angular.Modules, module)
		}
	}
	return angular.this()
}
func (angular *App) Controller(call otto.FunctionCall) otto.Value {
	controllerName, _ := call.Argument(0).ToString()
	argumentsStr, _ := call.Argument(1).ToString()

	functionBody := getFunctionBodyFromString(argumentsStr)
	var dependencies []string
	//Just In Case I need to get arguments
	argList, _ := call.Argument(1).Export()
	switch reflect.TypeOf(argList).Kind() {
	case reflect.Slice:
		argListCount := len(argList.([]interface{}))
		for k, v := range argList.([]interface{}) {
			if k == argListCount-1 {
				continue
			}
			dependencies = append(dependencies, v.(string))
		}

		break
	default:
		log.Fatal(controllerName + " is not properly formed")
	}
	modelname := strings.Title(strings.Replace(strings.ToLower(controllerName), "controller", "", -1)) + "Model"

	//Create Controller Object
	ctrl := Component{Name: strings.Title(controllerName), Type: "controller",
		ModelName: modelname, FunctionBody: functionBody, Dependencies: dependencies, Module: angular, VM: call.Otto}
	ctrl.FindTemplateString()
	ctrl.ParseScopeProperties()
	ctrl.ParseScopeValues()
	ctrl.ParseScopeFunctions()
	ctrl.ParseFunctionBodies()
	ctrl.RemoveScopeFunctionsFromScopeObjectInterface()

	//Add to controller object
	angular.Components = append(angular.Components, ctrl)

	return otto.Value{}
}
func (angular *App) RegisterRoute(call otto.FunctionCall) otto.Value {
	route, _ := call.Argument(0).ToString()
	config, _ := call.Argument(1).Export()
	// if value, ok := config.(map[string]interface{})["controller"]; ok {
	// 	//fmt.Println("Controller Name", value.(string))
	// }
	if value, ok := config.(map[string]interface{})["templateUrl"]; ok {
		fmt.Println("Template URL", value.(string))
	}
	fmt.Println(route)
	v, _ := angular.VM.Get(`$routeProvider`)
	return v
}

//Config parses app config function
/*
.....config(['$routeProvider',
	function($routeProvider) {
        "use strict";

        $routeProvider.when('/account', {
                templateUrl: admin_url + 'partials/office_info.html',
                controller: accountcontroller
            })..
*/

func (angular *App) Config(call otto.FunctionCall) otto.Value {
	argumentsStr, _ := call.Argument(0).ToString()
	var deps interface{}
	var dependencies []string
	fmt.Println(call.Argument(0))
	if call.Argument(0).Class() == "Array" {
		deps, _ = call.Argument(0).Export()

		for _, el := range deps.([]interface{}) {
			switch reflect.TypeOf(el).Kind() {
			case reflect.String:
				dependencies = append(dependencies, el.(string))
				break
			}
		}

	}
	if strings.Contains(argumentsStr, "$routeProvider") {
		//fmt.Println(argumentsStr)
		functionBody := getFunctionBodyFromString(argumentsStr)
		functionAssignment := fmt.Sprintf(`var func = %s`, functionBody)
		angular.VM.Set("configFunction", functionAssignment)
		//fmt.Println(functionBody)
		functionEvalCode := fmt.Sprintf(`
		//external mocks
        %s
		eval(configFunction)
		func(%s)
	`, angular.ExternalMocks, strings.Join(dependencies, ","))
		fmt.Println(functionEvalCode)
		if _, err := angular.VM.Run(functionEvalCode); err != nil {
			//fmt.Println(functionEvalCode)
			//log.Fatal(err.Error())
			panic(err)
		}
	}

	return otto.Value{}
}
func (angular *App) Service(call otto.FunctionCall) otto.Value {
	serviceName, _ := call.Argument(0).ToString()
	argumentsStr, _ := call.Argument(1).ToString()

	//TODO service are different than controllers
	//parse the usual syntax properly
	//May I should just extract the function from the service.
	//Won't know what the style is to get more information
	functionBody := getFunctionBodyFromString(argumentsStr)
	//fmt.Println(functionBody)
	service := Component{
		Name:         serviceName,
		Type:         "service",
		FunctionBody: functionBody,
	}
	angular.Components = append(angular.Components, service)

	return otto.Value{}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func getFunctionBodyFromString(argumentsStr string) string {
	var functionBody string
	funcSplit := strings.SplitN(argumentsStr, "function(", 2)

	if len(funcSplit) == 1 {
		funcSplit = strings.SplitN(argumentsStr, "function (", 2)
	}
	if len(funcSplit) >= 2 {
		functionBody = fmt.Sprintf("function(%s", funcSplit[1])
	}
	return functionBody
}
