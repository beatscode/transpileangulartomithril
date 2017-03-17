angular.module('myApp').service('testService',['$http',function($http){
    function someFunction(){
        console.log('Hi There');
    }
    var self = this;
    self.dontSayAnything = function(){
        return null;
    }
    return {
        sayHello : function(){
            someFunction();
        },
        dontSayHello : function (){
            self.dontSayAnything();
        }
    }
}]);