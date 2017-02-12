package main

import (
	"strings"
	"testing"

	"github.com/robertkrimen/otto"

	"com.drleonardo/transpileangulartomithril/angular"
)

func TestScopeObjectBuild(t *testing.T) {

	var vm = otto.New()
	angularTemplateDir := "."
	var app angular.App
	app.VM = vm
	app.TemplateDir = angularTemplateDir

	componentString := `
    angular.module('myApp',['helloModule','hello2Module']);
    angular.module('myApp').controller('testController',function($scope,my_Otherservice){
        $scope.myvar = "123";
        $scope.list = [1,2,3,4,5];
        $scope.getListItems = function(){
            console.log($items)
            $scope.list = []
        }
    });
    `
	//Set proper mock of angular object
	angularObj, _ := vm.Object(`angular = {}`)
	angularObj.Set("module", app.Module)
	angularObj.Set("controller", app.Controller)

	//Run the file/string to build meta data for transpiling
	if _, err := vm.Run(componentString); err != nil {
		panic(err)
	}

	var aModule angular.Component
	//	fmt.Println(angular)
	for _, module := range app.Components {
		if module.Type == "controller" && module.Name == "testController" {
			aModule = module
		}
	}

	if len(aModule.ScopeProperies) != 2 {
		t.Error("Failed to parse proper scope list")
	}
}

func TestScopeFunctionBuild(t *testing.T) {

	var vm = otto.New()
	angularTemplateDir := "."
	var app angular.App
	app.VM = vm
	app.TemplateDir = angularTemplateDir

	componentString := `
        angular.module('myApp',['helloModule','hello2Module']);
        angular.module('myApp').controller('testController',function($scope){
            $scope.myvar = "123";
            $scope.list = [1,2,3,4,5];
            $scope.getListItems = function(){
                console.log($items)
            }
        });
        `
	//Set proper mock of angular object
	angularObj, _ := vm.Object(`angular = {}`)
	angularObj.Set("module", app.Module)
	angularObj.Set("controller", app.Controller)

	//Run the file/string to build meta data for transpiling
	if _, err := vm.Run(componentString); err != nil {
		panic(err)
	}

	var aModule angular.Component
	for _, module := range app.Components {
		if module.Type == "controller" && module.Name == "testController" {
			aModule = module
		}
	}

	if len(aModule.ScopeProperies) != 2 {
		t.Error("Failed to parse proper scope list")
	}

	if aModule.Functions[0] != "getListItems" {
		t.Error("Invalid Function Name reading")
	}
}

func TestScopeFunctionBody(t *testing.T) {
	var vm = otto.New()
	angularTemplateDir := "."
	var app angular.App
	app.VM = vm
	app.TemplateDir = angularTemplateDir

	componentString := `
        angular.module('myApp',['helloModule','hello2Module']);
        angular.module('myApp').controller('testController',function($scope){
            $scope.myvar = "123";
            $scope.list = [1,2,3,4,5];
            $scope.getListItems = function(){
                console.log($items)
            }
        });
        `
	//Set proper mock of angular object
	angularObj, _ := vm.Object(`angular = {}`)
	angularObj.Set("module", app.Module)
	angularObj.Set("controller", app.Controller)

	//Run the file/string to build meta data for transpiling
	if _, err := vm.Run(componentString); err != nil {
		panic(err)
	}

	var aModule angular.Component
	//	fmt.Println(angular)
	for _, module := range app.Components {
		if module.Type == "controller" && module.Name == "testController" {
			aModule = module
		}
	}

	functionBody := aModule.FunctionBodies["getListItems"]
	//fmt.Println(functionBody)
	if strings.Contains(functionBody, "console.log($items)") == false {
		t.Error("Function Body is invalid", functionBody)
	}

	//get "getListItems" function Body

}

func TestDirectoryScanning(t *testing.T) {

	angularTemplateDir := "."
	angularScriptsDir := "./test"

	Start(angularScriptsDir, angularTemplateDir)
	noOfComponents := len(app.Components)
	if noOfComponents != 4 {
		t.Error("Incorrent number of components", noOfComponents)
	}

	if len(app.Modules) != 2 {
		t.Error("Invalid number of components")
	}
	//
	// for _, m := range app.Modules {
	// 	fmt.Println(m)
	// }
	//
	// for _, component := range app.Components {
	// 	fmt.Println(component.Name)
	// }
}
