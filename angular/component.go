package angular

import (
	"bytes"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/robertkrimen/otto"
)

//Component is one of the following: Controller,Directive,Factory,Service,Filter, etc
type Component struct {
	Type            string
	Name            string
	ModelName       string
	FunctionBody    string
	Dependencies    []string
	TemplateStr     string
	Functions       []string
	FunctionBodies  map[string]string
	ScopeProperties []string
	ScopeObject     interface{}
	VM              *otto.Otto
	Module          *App
}

var WatchersMock = `var watchers={}`
var JqueryMock = `var $ = function(){}`
var ScopeObjectMock = `var scopeObj = {
	'$on':function(eventName,func){
		watchers[eventName] = func.toString();
	}
}`

func (aComponent *Component) ParseScopeProperties() {
	//Make sure $scope is first
	//Get Scope Variables that are not functions
	//Normally would use negative lookahead but
	//not supported
	expression := `(?:\$scope\.)(?P<variable>\w+)\s+=\s+[^f]`
	reg := regexp.MustCompile(expression)

	matches := reg.FindAllStringSubmatch(aComponent.FunctionBody, -1)
	found := false
	var scopeProperties []string
	for _, v := range matches {
		for _, tmpV := range scopeProperties {
			if tmpV == v[1] {
				found = true
			}
		}
		if found == false {
			scopeProperties = append(scopeProperties, v[1])
		}
	}

	aComponent.ScopeProperties = scopeProperties
}
func (aComponent *Component) addMocks() {
	aComponent.VM.Object(ScopeObjectMock)
	aComponent.VM.Object(WatchersMock)
	aComponent.VM.Object(JqueryMock)

	var err error
	//Mock dependencies
	for _, dep := range aComponent.Dependencies {
		aComponent.VM.Set(dep, func(call otto.FunctionCall) otto.Value {
			return otto.Value{}
		})
		_, err = aComponent.VM.Run(fmt.Sprintf("new %s()", dep))
		if err != nil {
			panic(err)
		}
	}

}

//ParseScopeValues attempts to export $sccope Values/Properties
func (aComponent *Component) ParseScopeValues() {
	var functionVM *otto.Otto

	functionVM = aComponent.VM //otto.New()
	aComponent.addMocks()

	//Set Function Equal to var
	functionAssignment := fmt.Sprintf(`var func = %s`, aComponent.FunctionBody)

	functionVM.Set("controllerFunction", functionAssignment)
	functionEvalCode := fmt.Sprintf(`
		//External Mocks
		%s 
	    //Set and Run Function
	    eval(controllerFunction)
		$scope = scopeObj
	    func(%s)
	`, aComponent.Module.ExternalMocks, strings.Join(aComponent.Dependencies, ","))
	if _, err := functionVM.Run(functionEvalCode); err != nil {
		//fmt.Println(functionEvalCode)
		//log.Fatal(err.Error())
		panic(err)
	}
	var scopeObjectInterface interface{}
	if scopeobject, err := functionVM.Get("scopeObj"); err == nil {
		scopeObjectInterface, err = scopeobject.Export()
		if err != nil {
			panic(err)
		}
	}
	aComponent.VM = functionVM

	for key, prop := range scopeObjectInterface.(map[string]interface{}) {
		if key == "$on" {
			delete(scopeObjectInterface.(map[string]interface{}), key)
			continue
		}

		//Check if item is a function
		//If so then remove it
		var found = false
		for k := range aComponent.FunctionBodies {
			fmt.Println("Function", k)
			if key == k {
				found = true
			}
		}
		if found {
			delete(scopeObjectInterface.(map[string]interface{}), key)
			continue
		}
		//For Objects/Arrays get json string to represent value
		switch reflect.TypeOf(prop).Kind() {
		case reflect.Map, reflect.Slice:
			if value, err := aComponent.VM.Run(`JSON.stringify(scopeObj.` + key + `)`); err == nil {
				jsonValue := value.String()
				scopeObjectInterface.(map[string]interface{})[key] = jsonValue
			}
			break
		}
	}
	aComponent.ScopeObject = scopeObjectInterface
}

func (aComponent *Component) RemoveScopeFunctionsFromScopeObjectInterface() {
	for key, _ := range aComponent.ScopeObject.(map[string]interface{}) {
		if key == "$on" {
			delete(aComponent.ScopeObject.(map[string]interface{}), key)
			continue
		}

		//Check if item is a function
		//If so then remove it
		var found = false
		for k := range aComponent.FunctionBodies {
			fmt.Println("Function", k)
			if key == k {
				found = true
			}
		}
		if found {
			delete(aComponent.ScopeObject.(map[string]interface{}), key)
			continue
		}
	}
}
func (aComponent *Component) ParseScopeFunctions() {
	//Get Scope Variables that are not functions
	expression := `(?:\$scope\.)(?P<variable>\w+)\s+=\s+function`
	reg := regexp.MustCompile(expression)

	matches := reg.FindAllStringSubmatch(aComponent.FunctionBody, -1)

	var functions []string
	for _, v := range matches {
		functions = append(functions, v[1])
	}

	aComponent.Functions = functions
}

//ParseFunctionBodies parses controller bodies for function bodies
func (aComponent *Component) ParseFunctionBodies() {
	if len(aComponent.Functions) == 0 {
		return
	}
	aComponent.FunctionBodies = make(map[string]string)
	//functionVM = aComponent.VM
	aComponent.addMocks()
	for _, function := range aComponent.Functions {
		//This function is usually self invoking
		if function == "init" {
			continue
		}
		//Set Function Equal to var
		functionAssignment := fmt.Sprintf(`var func = %s`, aComponent.FunctionBody)
		aComponent.VM.Set("controllerFunction", functionAssignment)
		functionEvalCode := fmt.Sprintf(`
		//External Mocks
		%s 
	    //Set and Run Function
	    eval(controllerFunction)
		$scope = scopeObj
	    func(%s)
	`, aComponent.Module.ExternalMocks, strings.Join(aComponent.Dependencies, ","))
		if _, err := aComponent.VM.Run(functionEvalCode); err != nil {
			//It is hard to get the source of Self invoking functions
			//
			log.Println(functionEvalCode, err.Error())
			continue
			//panic(err)
		}
		var functionBody string
		functionToString := fmt.Sprintf(`scopeObj.%s.toString()`, function)
		if functionString, err := aComponent.VM.Run(functionToString); err != nil {
			//panic(err)
			//It is hard to get the source of Self invoking functions
			//
			log.Println(functionToString, err.Error())
			continue
		} else {
			functionBody = strings.Replace(functionString.String(), "$scope", aComponent.ModelName, -1)
		}
		aComponent.FunctionBodies[function] = functionBody
	}
}

//FindTemplateString searches angular template directory for controller html
func (aComponent *Component) FindTemplateString() {
	dir := aComponent.Module.TemplateDir
	fstat, err := os.Stat(dir)
	if err != nil {
		fmt.Println(dir)
		panic(err)
	}
	if fstat.IsDir() {
		//check if dir is a file or directory
		files, err := ioutil.ReadDir(dir)
		checkErr(err)
		for _, f := range files {
			if filepath.Ext(f.Name()) != ".html" {
				continue
			}

			rawBytes, _ := ioutil.ReadFile(fmt.Sprintf("%s/%s", dir, f.Name()))
			if strings.Contains(string(rawBytes), aComponent.Name) {
				aComponent.TemplateStr = string(rawBytes)
			}
		}
	} else {
		rawBytes, _ := ioutil.ReadFile(dir)
		aComponent.TemplateStr = string(rawBytes)
	}
}

//ExportController exports the controller as a buffer with objects
//and  functions
func (aComponent *Component) ExportController() *bytes.Buffer {

	buf := new(bytes.Buffer)
	if aComponent.TemplateStr == "" {
		return buf
	}
	tmpl := template.New("New Component")
	tmpl = tmpl.Funcs(template.FuncMap{"parseVal": func(args ...interface{}) string {
		var r string
		switch reflect.TypeOf(args[0]).Kind() {
		case reflect.String:
			if args[0].(string) == "undefined" {
				r = "null"
			} else if args[0].(string) == "true" || args[0].(string) == "false" || args[0].(string) == "{}" || args[0].(string) == "[]" {
				r = args[0].(string)
			} else {
				r = fmt.Sprintf(`'%s'`, args[0])
			}
			break
		case reflect.Bool:
			if args[0].(bool) == true {
				r = "true"
			} else {
				r = "false"
			}
			break
		}

		return r
	}})
	aComponent.ConvertAngularElements()
	//TODO: See if we can convert to jsx here
	tmpl, _ = tmpl.Parse(`
	
	var {{.ModelName}} = {
		{{range $key,$el := .ScopeObject}}
			'{{$key}}':{{$el | parseVal}},
		{{end}}
		{{range $key, $func := .FunctionBodies}}
			'{{$key}}': {{$func}},
		{{end}}
	};	
	var {{.Name}}Component = {
		oncreate : function(){

		},

		view : function(){
			return (
				{{.TemplateStr}}	
			)
		}
	};
	m.mount(document.body, {{.Name}}Component)
	`)
	tmpl.Execute(buf, aComponent)
	fmt.Println(buf.String())
	return buf
}

//ConvertAngularElements converts ng-click,ng-change to jsx freindly markey
func (aComponent *Component) ConvertAngularElements() {
	// <button class="btn-rqust-app" onclick={ApptModel.save}>Request Appointment</button>
	// <input type="text" name="Email" value={ApptModel.current.Email}/>

	htmlContent := fmt.Sprintf(`<html><head><body id="view">%s</body></html>`, aComponent.TemplateStr)
	reader := strings.NewReader(htmlContent)

	doc, _ := goquery.NewDocumentFromReader(reader)
	componentName := "TestComponentModel"
	doc.Find("#view").Each(func(i int, s *goquery.Selection) {
		if s.Children().Size() > 1 {
			htmlContent = fmt.Sprintf(`<html><head><body id="view"><div>%s<div></body></html>`, aComponent.TemplateStr)
			reader = strings.NewReader(htmlContent)
			doc, _ = goquery.NewDocumentFromReader(reader)
		}
	})

	var convertedClickFunctions []string
	// ng-click="saveForm()" -> onclick={ComponentModel.saveForm}
	doc.Find("[data-ng-click],[ng-click]").Each(func(i int, s *goquery.Selection) {

		convertHTML := func(funcName string, s *goquery.Selection) {
			onclickSyntax := fmt.Sprintf("%s.%s", componentName, funcName)
			convertedClickFunctions = append(convertedClickFunctions, onclickSyntax)
			s.SetAttr("onclick", onclickSyntax)
			s.RemoveAttr("data-ng-click")
			s.RemoveAttr("ng-click")
		}

		if funcName, ok := s.Attr("data-ng-click"); ok {
			convertHTML(funcName, s)
		}

		if funcName, ok := s.Attr("ng-click"); ok {
			convertHTML(funcName, s)
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
				<option>Select %s</option>
				{%s.%s.map((obj) => {
					return ( <option key={obj.id}>{obj}</option> )
				})}
			`, arrayName, aComponent.ModelName, arrayName)

			s.RemoveAttr("ng-options")
			tmpRan := randomString(8)
			replaceHTMLMap[tmpRan] = html.UnescapeString(replaceInnerHTML)
			s.SetAttr("data-parsekey", tmpRan)
		}

		if ngModel, ok := s.Attr("ng-model"); ok {
			s.SetAttr("name", ngModel)
			s.RemoveAttr("ng-model")
		}

		if ngChange, ok := s.Attr("ng-change"); ok {
			onclickSyntax := fmt.Sprintf("%s.%s", componentName, ngChange)
			convertedClickFunctions = append(convertedClickFunctions, onclickSyntax)
			s.SetAttr("onchange", onclickSyntax)
			s.RemoveAttr("ng-change")
		} else {
			s.SetAttr("onchange", "{binds}")
		}

	})

	for key, selc := range replaceHTMLMap {
		selectKey := fmt.Sprintf(`select[data-parsekey='%s']`, key)
		doc.Find(selectKey).Each(func(i int, s *goquery.Selection) {
			s.SetHtml(selc)
			s.RemoveAttr("data-parsekey")
		})
	}
	htmlStr, _ := doc.Find("#view").Html()

	//Replace {{ and }} with {ModelName.scope_property }
	// {{myvar}} -> {ModelName.myvar}
	htmlStr = strings.Replace(htmlStr, "{{", fmt.Sprintf("{%s.", aComponent.ModelName), -1)
	htmlStr = strings.Replace(htmlStr, "}}", "}", -1)

	htmlStr = html.UnescapeString(htmlStr)
	srcFileName := aComponent.Name + ".html"
	err := ioutil.WriteFile(srcFileName, []byte(htmlStr), 0777)
	if err != nil {
		log.Fatal(err)
	}

	html2jsx, err := exec.LookPath("html2jsx")
	if err != nil {
		log.Fatal("Please install html2jsx (npm install -g html2jsx) to format html")
	}

	out, reterr := exec.Command(html2jsx, srcFileName).CombinedOutput()
	fmt.Println(reterr)
	jsxStr := strings.TrimSpace(string(out))
	//Conversion adds function {Name}() { to the string
	splits := strings.Split(jsxStr, "\n")
	splits[1] = strings.Replace(splits[1], "return", "", 1)
	jsxStr = strings.Join(splits[1:len(splits)-1], "\n")
	jsxStr = strings.Replace(jsxStr, `Â`, "&Acirc;", -1)
	//Remove Temporary file
	err = os.Remove(srcFileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	aComponent.TemplateStr = jsxStr
}

func randomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
