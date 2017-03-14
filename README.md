# Transpile angular 1.x to mithril 1.0
Lazy Attempt to scaffold a mithril application from existing angular 1 components

This project attempts to
- Load an angular component (controller,service, factory)
- Parse functions and scope object properties (via regex)  
- Create a mithril 1.0 component(vanilla object with oncreate, view functions) and Scope Model object
- Use the angular component name as Mithril Objects. Appending `Component` and `Model` respectively
- Search for template file containing corresponding controller view.
  -  If found, run html2jsx (install via `npm install html2jsx`)
- Create .js file with both Model and Component 

## TODO
- Import multiple components of the same module to share model properties/functions
- Use state management (redux most likely)

## Caveats
- May take some trial and error to get the Js engine to compile without throwing undefined errors
  - Haven't found a cleaner way other than exposing external Mocks file option
