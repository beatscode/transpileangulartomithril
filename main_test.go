package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
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
		if module.Type == "controller" && module.Name == "doctorscontroller" {
			aModule = module
		}
	}

	if aModule.Name != "doctorscontroller" {
		t.Error("Invalid Module Parsing")
	}

	buf := aModule.ExportController()

	htmlStr := fmt.Sprintf(`
	<html>
	<head>
    <script src="//unpkg.com/mithril/mithril.js"></script>
	</head>
	<body>
		%s
	</body>
	</html>
	`, buf.Bytes())
	_ = ioutil.WriteFile(fmt.Sprintf("%stest.html", aModule.Name), []byte(htmlStr), 0777)

	// if buf.String() != "var doctorscontrollerComponent" {
	// 	t.Error("Error loading doctor controller component")
	// }
}

func TestConvertAngularHtmlStringToSupportJSX(t *testing.T) {
	var angularStr = `<div class="pull-right">
        <button class="btn btn-primary save-button" type="button" data-ng-click="saveDoctors($event);">Save</button>
        <a data-ng-click="exitForm(currentOption);" class="close-partial">Close</a>
      </div>
	  <button ng-click="saveFullName()"> Save Full Name </button>
      <div class="pull-right">
        <div class="alert alert-error" ng-show="hasError">{{errorMsg}}</div>
        <div class="alert alert-success" ng-show="status" >{{status}}</div>
      </div>

		<select class="span3" ng-model="sDoctor" ng-change="findLocations();" 
		ng-options="s.firstname + ' ' + s.lastname for s in doctors" >
		<option value="">Choose Doctor</option>
		</select>
	  `

	// <button class="btn-rqust-app" onclick={ApptModel.save}>Request Appointment</button>
	// <input type="text" name="Email" value={ApptModel.current.Email}/>

	htmlContent := fmt.Sprintf(`<html><head><body id="view">%s</body></html>`, angularStr)
	reader := strings.NewReader(htmlContent)

	doc, _ := goquery.NewDocumentFromReader(reader)
	componentName := "TestComponentModel"
	var convertedClickFunctions []string
	// ng-click="saveForm()" -> onclick={ComponentModel.saveForm}
	doc.Find("[data-ng-click],[ng-click]").Each(func(i int, s *goquery.Selection) {

		convertHtml := func(funcName string, s *goquery.Selection) {
			onclickSyntax := fmt.Sprintf("%s.%s", componentName, funcName)
			convertedClickFunctions = append(convertedClickFunctions, onclickSyntax)
			s.SetAttr("onclick", onclickSyntax)
			s.RemoveAttr("data-ng-click")
			s.RemoveAttr("ng-click")

			//s.SetHtml()
		}

		if funcName, ok := s.Attr("data-ng-click"); ok {
			convertHtml(funcName, s)
		}

		if funcName, ok := s.Attr("ng-click"); ok {
			convertHtml(funcName, s)
		}
	})

	/*
		<select class="span3" ng-model="sDoctor" ng-change="findDoctorLocation();getDoctorSpecialties();"
		ng-options="s.firstname + ' ' + s.lastname for s in doctors" >
		<option value="">Choose Doctor</option>
		</select>
		----------------------------------

		<select class="slct-loc" style="margin-right:3px" name="Location" onchange={binds}>
				<option>Select Location</option>
				{locations.map((location) => {
					return <option key={location.id}>{location.practice_name}</option>
				})}
				</select>
	*/
	var replaceHTMLMap = make(map[string]string)
	doc.Find("[data-ng-change],[ng-change]").Each(func(i int, s *goquery.Selection) {

		//regex
		//\w+\sin\s+\w+ ng-options="x in array"
		expression := `\w+\s+in\s+(?P<variable>\w+)`
		reg := regexp.MustCompile(expression)

		if ngOptions, ok := s.Attr("ng-options"); ok {
			//Create Map Function
			matches := reg.FindAllStringSubmatch(ngOptions, -1)
			if len(matches) == 0 {
				return
			}
			arrayName := matches[0][1]
			replaceInnerHTML := fmt.Sprintf(`
				{%s.map((obj) => {
					return <option key={%s.id}>{%s}</option>
				})}
			`, arrayName, arrayName, arrayName)

			s.RemoveAttr("ng-options")
			tmpRan := RandomString(8)
			replaceHTMLMap[tmpRan] = replaceInnerHTML
			s.SetAttr("data-parsekey", tmpRan)
		}

		if ngModel, ok := s.Attr("ng-model"); ok {
			s.SetAttr("name", ngModel)
			s.RemoveAttr("ng-model")
		}

		if ngChange, ok := s.Attr("ng-change"); ok {
			s.SetAttr("onchange", "{"+ngChange+"}")
			s.RemoveAttr("ng-change")
		} else {
			s.SetAttr("onchange", "{binds}")
		}

	})
	htmlStr, _ := doc.Find("#view").Html()
	doc, _ = goquery.NewDocumentFromReader(reader)
	for _, fn := range convertedClickFunctions {
		//replace old with new
		//Get what is between parenthesis
		fnWithoutParenthesis := strings.Split(fn, "(")
		htmlStr = strings.Replace(htmlStr, fmt.Sprintf(`"%s"`, fn), fmt.Sprintf(`{%s}`, fnWithoutParenthesis[0]), 1)
	}

	for key, selc := range replaceHTMLMap {
		doc.Find(fmt.Sprintf(`[data-parsekey=%s]`, key)).Each(func(i int, s *goquery.Selection) {
			fmt.Println(selc)
			s.SetText(selc)
		})
	}

	//Replace {{ and }} with { }
	htmlStr = strings.Replace(htmlStr, "{{", "{", -1)
	htmlStr = strings.Replace(htmlStr, "}}", "}", -1)
	fmt.Println(htmlStr)
}
func RandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
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
