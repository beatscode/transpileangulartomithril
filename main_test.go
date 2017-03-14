package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"testing"

	"github.com/robertkrimen/otto"

	"com.drleonardo/transpileangulartomithril/angular"
)

func TestScopeObject(t *testing.T) {

	var vm = otto.New()
	angularTemplateDir := "."
	var app angular.App
	app.VM = vm
	app.TemplateDir = angularTemplateDir

	componentString := `
    angular.module('myApp',['helloModule','hello2Module']);
    angular.module('myApp').controller('testController',[ '$scope','my_Otherservice', function($scope,my_Otherservice){
        $scope.myvar = "123";
        $scope.list = [1,2,3,4,5];
        $scope.getListItems = function(){
            console.log($items)
            $scope.list = []
        }
    }]);
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

	if len(aModule.ScopeProperties) != 2 {
		t.Error("Failed to parse proper scope list", len(aModule.ScopeProperties))
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
    angular.module('myApp').controller('testController',[ '$scope','my_Otherservice', function($scope,my_Otherservice){
        $scope.myvar = "123";
        $scope.list = [1,2,3,4,5];
        $scope.getListItems = function(){
            console.log($items)
            $scope.list = []
        }
    }]);
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

	if len(aModule.ScopeProperties) != 2 {
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
    angular.module('myApp').controller('testController',[ '$scope','my_Otherservice', function($scope,my_Otherservice){
        $scope.myvar = "123";
        $scope.list = [1,2,3,4,5];
        $scope.getListItems = function(){
            console.log($items)
            $scope.list = []
        }
    }]);
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
	if strings.Contains(functionBody, "console.log($items)") == false {
		t.Error("Function Body is invalid", functionBody)
	}
}
func TestDoctorParse(t *testing.T) {
	var configfile Config
	configfile.TemplateDir = "./test/views/doctor.html"
	configfile.ExternalMocksFilepath = "./externalmocks.js"
	configfile.ScriptsDir = "./test/doctorcontroller.js"
	fileBytes, err := ioutil.ReadFile(configfile.ExternalMocksFilepath)
	if err != nil {
		log.Fatal(err)
	}
	app := Start(configfile.ScriptsDir, configfile.TemplateDir, string(fileBytes))

	var aModule angular.Component
	for _, module := range app.Components {
		fmt.Println(aModule.Name)
		if module.Type == "controller" && module.Name == "Doctorscontroller" {
			aModule = module
		}
	}

	if aModule.Name != "Doctorscontroller" {
		t.Error("Invalid Module Parsing", aModule.Name)
	}

	buf := aModule.ExportController()

	htmlStr := fmt.Sprintf(`<html>
	<head>
    <script src="https://unpkg.com/mithril/mithril.js"></script>
	<script src="https://unpkg.com/babel-standalone@6/babel.min.js"></script>
	</head>
	<body>
		%s
	</body>
	</html>
	`, buf.Bytes())
	_ = ioutil.WriteFile(fmt.Sprintf("%stest.html", aModule.Name), []byte(htmlStr), 0777)

	// if buf.String() != "var DoctorscontrollerComponent" {
	// 	t.Error("Error loading doctor controller component")
	// }
}

func TestConvertAngularHtmlStringToSupportJSX(t *testing.T) {

	var configfile Config
	configfile.TemplateDir = "./test/views/fragment.html"
	configfile.ExternalMocksFilepath = "./externalmocks.js"
	configfile.ScriptsDir = "./test/doctorcontroller.js"
	fileBytes, err := ioutil.ReadFile(configfile.ExternalMocksFilepath)
	if err != nil {
		log.Fatal(err)
	}
	app := Start(configfile.ScriptsDir, configfile.TemplateDir, string(fileBytes))

	var aModule angular.Component
	for _, module := range app.Components {
		fmt.Println(strings.Title("Doctorscontroller"))
		if module.Type == "controller" && module.Name == "Doctorscontroller" {
			aModule = module
		}
	}

	if aModule.Name != "Doctorscontroller" {
		t.Error("Invalid Module Parsing")
	}

	buf := aModule.ExportController()
	_ = ioutil.WriteFile(fmt.Sprintf("./test/testjsx/src/app/%stest.js", aModule.Name), buf.Bytes(), 0777)
	htmlStr := fmt.Sprintf(`<html>
	<head>
	<meta charset="UTF-8">
	</head>
	<body>
		<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" integrity="sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u" crossorigin="anonymous">
		<script src="https://unpkg.com/mithril/mithril.js"></script>
		<script src="%s"></script>
	</body>
	</html>`, `app/bundle.js`)
	_ = ioutil.WriteFile(fmt.Sprintf("./test/testjsx/%stest.html", aModule.Name), []byte(htmlStr), 0777)

}

func TestGetScopeValues(t *testing.T) {
	var vm = otto.New()
	angularTemplateDir := "."
	var app angular.App
	app.VM = vm
	app.TemplateDir = angularTemplateDir

	componentString := `
        angular.module('myApp',['helloModule','hello2Module']);
        angular.module('myApp').controller('testController',[ '$scope', function($scope){
            $scope.myvar = "123";
            $scope.list = [1,2,3,4,5];
			$scope.number = 12341;
			$scope.mybool = true;
			$scope.myobj = { a: 1, b: 2, c: 3 }
            $scope.getListItems = function(){
                console.log($items)
            }
			$scope.$on('SomeEvent/eventThatICareAbout',function(){
				console.log('I do something here');
			})
        }]);
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

	if aModule.ScopeObject.(map[string]interface{})["list"] != "[1,2,3,4,5]" {
		t.Error("Invalid Conversion of Array", aModule.ScopeObject.(map[string]interface{})["list"])
	}
	if aModule.ScopeObject.(map[string]interface{})["myobj"] != `{"a":1,"b":2,"c":3}` {
		t.Error("Invalid Conversion of Object", aModule.ScopeObject.(map[string]interface{})["myobj"])
	}
	if aModule.ScopeObject.(map[string]interface{})["myvar"] != `123` {
		t.Error("Invalid Conversion", aModule.ScopeObject.(map[string]interface{})["myvar"])
	}
	if aModule.ScopeObject.(map[string]interface{})["mybool"] != true {
		t.Error("Invalid Conversion", aModule.ScopeObject.(map[string]interface{})["mybool"])
	}
}

func TestDirectoryScanning(t *testing.T) {

	angularTemplateDir := "."
	angularScriptsDir := "./test"

	Start(angularScriptsDir, angularTemplateDir, "")
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
