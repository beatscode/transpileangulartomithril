package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"com.drleonardo/transpileangulartomithril/angular"

	"github.com/BurntSushi/toml"
	"github.com/robertkrimen/otto"
)

var app angular.App
var vm = otto.New()
var configfile string

/*
angular.module('onBoardExpressApp').controller('onBoardSearchController', function ($scope, $rootScope, $window, onboard_express_service) {...
*/
func main() {
	flag.StringVar(&configfile, "config", "app.conf", "Application Config")

	//Parse Arguments
	flag.Parse()

	masterconfig := ReadConfig(configfile)

	fileBytes, err := ioutil.ReadFile(masterconfig.ExternalMocksFilepath)
	if err != nil {
		log.Fatal(err)
	}
	Start(masterconfig.ScriptsDir, masterconfig.TemplateDir, string(fileBytes))
}

type Config struct {
	TemplateDir           string
	ScriptsDir            string
	ExternalMocksFilepath string
}

func ReadConfig(configfile string) Config {
	_, err := os.Stat(configfile)
	if err != nil {
		log.Fatal("Config file is missing: ", configfile)
	}

	var config Config
	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		log.Fatal(err)
	}
	return config
}
func Start(angularScriptsDir string, angularTemplateDir string, externalMocks string) angular.App {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalln("Start Error:", r)
		}
	}()
	app.VM = vm
	app.TemplateDir = angularTemplateDir
	app.ExternalMocks = externalMocks
	var err error
	var files []os.FileInfo
	fstat, err := os.Stat(angularScriptsDir)
	if err != nil {
		fmt.Println(angularScriptsDir)
		panic(err)
	}
	if fstat.IsDir() {
		file, err := os.Open(angularScriptsDir)
		checkErr(err)

		files, err = file.Readdir(-1)
		checkErr(err)
		for _, file := range files {

			filepath := fmt.Sprintf("%s/%s", angularScriptsDir, file.Name())
			if strings.Contains(file.Name(), ".js") == false {
				continue
			}
			javascriptFileContents, err := ioutil.ReadFile(filepath)
			checkErr(err)
			parse(string(javascriptFileContents))
		}
	} else {
		files = append(files, fstat)
		fileBytes, err := ioutil.ReadFile(angularScriptsDir)
		checkErr(err)
		parse(string(fileBytes))
	}

	return app
}
func parse(javascriptFileContents string) {
	//Set proper mock of angular object
	angularObj, _ := vm.Object(`angular = {}`)
	angularObj.Set("module", app.Module)
	angularObj.Set("controller", app.Controller)
	angularObj.Set("service", app.Service)

	//Run the file/string to build meta data for conversion
	if _, err := vm.Run(javascriptFileContents); err != nil {
		if strings.Contains(err.Error(), "'$scope' is not defined") {
			log.Fatalln("$scope needs to be defined:", string(javascriptFileContents))
		}
		panic(err)
	}

}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
