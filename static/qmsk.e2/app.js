angular.module('qmsk.e2', [
        'ngRoute',
	    'angular-websocket',
	    'ui.bootstrap'
])

.config(function($routeProvider) {
    $routeProvider
        .when('/screens', {
            templateUrl: '/static/qmsk.e2/screens.html',
            controller: 'ScreensCtrl',
        })
        .otherwise({
            redirectTo: '/screens',
        });
})

.controller('HeaderCtrl', function($scope, $location) {
    $scope.safe = false;

    $scope.isActive = function(prefix) {
        return $location.path().startsWith(prefix);
    };
})

.controller('ScreensCtrl', function($scope) {
    $scope.screens = {
        "0": {
            name:   "test"
        }
    };
})

;
