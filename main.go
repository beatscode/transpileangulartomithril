package main

import (
	"fmt"

	"github.com/robertkrimen/otto"
)

type Angular struct{}

func (angular *Angular) Module(name string, deps []string) {
	fmt.Println("Golang Land", name, deps)
}

func main() {
	var angular Angular
	vm := otto.New()
	if err := vm.Set("angular", angular); err != nil {
		panic(err)
	}
	//Set proper mock of angular object
	if _, err := vm.Run(`
		var modules = {};
		var angular = {
			module : function(name,deps){
				console.log(name,deps)
				if(!modules[name]){
					modules[name] = [name,deps]
				}
			}
		};
	`); err != nil {
		panic(err)
	}

	//Run JS File
	if _, err := vm.Run(`
		angular.module('myApp',['widgetModule','pageModule','faceModule']);	
	`); err != nil {
		panic(err)
	}

}
