angular.module('myApp2').controller('test1Controller',function($scope){

    $scope.myvar1 = 'Hello';
    $scope.myvar2 = 'World';

    $scope.doublemyvar = function(){
        $scope.myvar1 = $scope.myvar1 + " " + $scope.myvar1;
    }

});
