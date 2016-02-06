angular.module('qmsk.e2', [
        'ngResource',
        'ngRoute',
        'ngWebSocket',
	    'ui.bootstrap'
])

.config(function($routeProvider) {
    $routeProvider
        .when('/sources', {
            templateUrl: 'qmsk.e2/sources.html',
            controller: 'SourcesCtrl',
        })
        .when('/screens', {
            templateUrl: 'qmsk.e2/screens.html',
            controller: 'ScreensCtrl',
        })
        .otherwise({
            redirectTo: '/screens',
        });
})

.factory('Source', function($resource) {
    return $resource('/api/sources/:id', { }, {
        get: {
            method: 'GET',
        },
        query: {
            method: 'GET',
            isArray: false, // XXX
        }
    }, {stripTrailingSlashes: true});
})

.factory('Screen', function($resource) {
    return $resource('/api/screens/:id', { }, {
        get: {
            method: 'GET',
        },
        query: {
            method: 'GET',
            isArray: false, // XXX
        }
    }, {stripTrailingSlashes: true});
})

.filter('dimensions', function() {
    return function(dimensions) {
        return dimensions.width + "x" + dimensions.height;
    };
})

.controller('HeaderCtrl', function($scope, $location) {
    $scope.safe = false;

    $scope.isActive = function(prefix) {
        return $location.path().startsWith(prefix);
    };
})
.controller('SourcesCtrl', function($scope, Source) {
    $scope.sources = Source.query();
})

.controller('ScreensCtrl', function($scope, Screen) {
    $scope.screens = Screen.query();
})

;
