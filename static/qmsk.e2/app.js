angular.module('qmsk.e2', [
        'ngResource',
        'ngRoute',
        'ngWebSocket',
	    'ui.bootstrap'
])

.config(function($routeProvider) {
    $routeProvider
        .when('/main', {
            templateUrl: 'qmsk.e2/main.html',
            controller: 'MainCtrl',
        })
        .when('/sources', {
            templateUrl: 'qmsk.e2/sources.html',
            controller: 'SourcesCtrl',
        })
        .when('/screens', {
            templateUrl: 'qmsk.e2/screens.html',
            controller: 'ScreensCtrl',
        })
        .otherwise({
            redirectTo: '/main',
        });
})

.factory('Status', function($http) {
    var Status = {};

    // Used as a $http.get(...).then(..., Status.onError)
    Status.onError = function(r){
        Status.mode = 'error'
        Status.error = r;
    }
    Status.httpInterceptor = {
        responseError: Status.onError,
    };

    $http.get('/api/status').then(
        function success(r) {
            Status.error = null
            Status.server = r.data.server;
            Status.mode = r.data.mode;
        },
        Status.onError
    );

    return Status;
})

.factory('Index', function($http, Status) {
    return function() {
        return $http.get('/api/').then(
            function success(r) {
                return r.data;
            },
            Status.onError
        );
    };
})

.factory('Source', function($resource, Status) {
    return $resource('/api/sources/:id', { }, {
        get: {
            method: 'GET',
            interceptor:    Status.httpInterceptor,
        },
        query: {
            method: 'GET',
            isArray: false,
            interceptor:    Status.httpInterceptor,
        }
    }, {stripTrailingSlashes: true});
})

.factory('Screen', function($resource, Status) {
    return $resource('/api/screens/:id', { }, {
        get: {
            method: 'GET',
            interceptor:    Status.httpInterceptor,
        },
        query: {
            method: 'GET',
            isArray: false,
            interceptor:    Status.httpInterceptor,
        }
    }, {stripTrailingSlashes: true});
})

.filter('dimensions', function() {
    return function(dimensions) {
        return dimensions.width + "x" + dimensions.height;
    };
})

.controller('HeaderCtrl', function($scope, $location, Status) {
    $scope.safe = false;
    $scope.status = Status

    $scope.isActive = function(prefix) {
        return $location.path().startsWith(prefix);
    };
})

.controller('MainCtrl', function($scope, Index) {
    Index().then(function success(index) {
        $scope.index = index;
    });
})

.controller('SourcesCtrl', function($scope, Source) {
    $scope.sources = Source.query();
})

.controller('ScreensCtrl', function($scope, Screen) {
    $scope.screens = Screen.query();
})

;
