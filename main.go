package main

import (
	"flag"

	"github.com/robertkrimen/otto"
)

//Transpiler will hopfully convert important parts to mithril or some other framework
type Transpiler struct {
	Dependencies []string
}

var angularTemplateDirPtr *string
var angularScriptsDirPtr *string
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
	angularScriptsDirPtr = flag.String("angularTemplateDir", "./test/scripts", "Directory to search for angular scripts")
}
