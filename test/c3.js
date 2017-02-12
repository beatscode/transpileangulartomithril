angular.module('myApp').controller('RandomController',function($scope){

    $scope.myvar1 = 'Hello';
    $scope.myvar2 = 'World';

    $scope.doublemyvar = function(){
        $scope.myvar1 = $scope.myvar1 + " " + $scope.myvar1;
    }

});
