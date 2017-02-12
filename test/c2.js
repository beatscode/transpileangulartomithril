angular.module('myApp').controller('test2Controller',function($scope){

    $scope.myvar1 = 'I am ';
    $scope.myvar2 = 'Job';

    $scope.doingsomethingrandom = function(){
        return Math.random(100);
    }

});
