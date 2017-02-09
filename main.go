package main

import (
	"flag"
	"fmt"

	"com.drleonardo/transpileangulartomithril/angular"

	"github.com/robertkrimen/otto"
)

//Transpiler will hopfully convert important parts to mithril or some other framework
type Transpiler struct {
	Dependencies []string
}

var angularTemplateDirPtr *string
var tranpiler Transpiler
var vm = otto.New()

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

/*
angular.module('onBoardExpressApp').controller('onBoardSearchController', function ($scope, $rootScope, $window, onboard_express_service) {...
*/
func main() {
	angularTemplateDirPtr = flag.String("angularTemplateDir", ".", "Directory to search for controller html")
	var angular angular.AngularApp
	angular.VM = vm
	angular.TemplateDir = *angularTemplateDirPtr

	componentString := `
	angular.module('myApp',['helloModule','hello2Module']);
	angular.module('myApp').controller('testController',function($scope){
		$scope.myvar = "123";
		$scope.list = [1,2,3,4,5];
	});
	`
	//Set proper mock of angular object
	angularObj, _ := vm.Object(`angular = {}`)

	angularObj.Set("module", angular.Module)
	angularObj.Set("controller", angular.Controller)

	//Run the file/string to build meta data for transpiling
	if _, err := vm.Run(componentString); err != nil {
		panic(err)
	}

	fmt.Println(angular)
}
