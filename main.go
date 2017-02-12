package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"com.drleonardo/transpileangulartomithril/angular"

	"github.com/robertkrimen/otto"
)

//Transpiler will hopfully convert important parts to mithril or some other framework
type Transpiler struct {
	Dependencies []string
}

var angularTemplateDirPtr *string
var angularScriptsDirPtr *string
var tranpiler Transpiler
var app angular.App
var vm = otto.New()

/*
angular.module('onBoardExpressApp').controller('onBoardSearchController', function ($scope, $rootScope, $window, onboard_express_service) {...
*/
func main() {
	angularTemplateDirPtr = flag.String("angularTemplateDir", ".", "Directory to search for controller html")
	angularScriptsDirPtr = flag.String("angularTemplateDir", "./test/scripts", "Directory to search for angular scripts")
	Start(*angularScriptsDirPtr, *angularTemplateDirPtr)
}

func Start(angularScriptsDir string, angularTemplateDir string) {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalln("Start Error:", r)
		}
	}()
	app.VM = vm
	app.TemplateDir = angularTemplateDir

	file, err := os.Open(angularScriptsDir)
	checkErr(err)

	files, err := file.Readdir(-1)
	checkErr(err)

	for _, file := range files {

		filepath := fmt.Sprintf("%s/%s", angularScriptsDir, file.Name())
		if strings.Contains(file.Name(), ".js") {
			javascript, err := ioutil.ReadFile(filepath)
			checkErr(err)

			//Set proper mock of angular object
			angularObj, _ := vm.Object(`angular = {}`)
			angularObj.Set("module", app.Module)
			angularObj.Set("controller", app.Controller)
			angularObj.Set("service", app.Service)

			//Run the file/string to build meta data for conversion
			if _, err := vm.Run(javascript); err != nil {
				if strings.Contains(err.Error(), "'$scope' is not defined") {
					log.Fatalln("$scope needs to be defined:", string(javascript))
				}
				panic(err)
			}
		}
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
