package angular

import (
	"fmt"

	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/parser"

	"github.com/robertkrimen/otto"
	"github.com/wolfgarnet/walker"
)

//App is made up of multiple angular.modules
type App struct {
	Modules     []AngularModule
	Components  []Component
	VM          *otto.Otto
	TemplateDir string
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
	fmt.Println(program.Body)
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
	functionBody, _ := call.Argument(1).ToString()

	ctrl := Component{Name: controllerName, Type: "controller", FunctionBody: functionBody}
	ctrl.FindTemplateString(angular.TemplateDir)
	ctrl.ParseScopeProperties()
	ctrl.ParseScopeFunctions()
	ctrl.ParseFunctionBodies()
	angular.Components = append(angular.Components, ctrl)

	return otto.Value{}
}
func (angular *App) Service(call otto.FunctionCall) otto.Value {
	serviceName, _ := call.Argument(0).ToString()
	functionBody, _ := call.Argument(1).ToString()

	service := Component{Name: serviceName, Type: "service", FunctionBody: functionBody}
	//TODO service are different than controllers
	//parse the usual syntax properly
	angular.Components = append(angular.Components, service)

	return otto.Value{}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
