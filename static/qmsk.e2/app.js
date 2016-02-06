angular.module('qmsk.e2', [
        'qmsk.e2.source',
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
            reloadOnSearch: false,
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

.controller('MainCtrl', function($scope, $location, Index, $interval) {
    $scope.reload = function() {
        Index().then(function success(index) {
            $scope.screens = index.screens
            $scope.sources = $.map(index.sources, function(source, id){
                return source;
            });
        });
    };

    $scope.selectOrder = function(order) {
        $scope.order = order;
        $scope.orderBy = function(){
            switch (order) {
            case 'source':
                return ['-type', 'name'];
            case 'preview':
                return ['preview_screens', 'program_screens'];
            case 'program':
                return ['program_screens', 'preview_screens'];
            default:
                return [];
            }
        }();

        $location.search('order', order);
    };
    $scope.selectOrder($location.search().order);

    $scope.selectRefresh = function(refresh) {
        $scope.refresh = refresh;

        if ($scope.refreshTimer) {
            $interval.cancel($scope.refreshTimer);
            $scope.refreshTimer = null;
        }

        if (refresh) {
            $scope.refreshTimer = $interval($scope.reload, refresh * 1000);
        }

        $location.search('refresh', refresh || '');
    };
    $scope.selectRefresh($location.search().refresh);

    $scope.reload();
})

.controller('SourcesCtrl', function($scope, Source) {
    $scope.sources = Source.query();
})

.controller('ScreensCtrl', function($scope, Screen) {
    $scope.screens = Screen.query();
})

;
