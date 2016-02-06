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

// track global http state
.factory('httpState', function($q) {
    var httpState = {
        error:  null,
        busy:   0,

        request: function(config) {
            httpState.busy++;

            return config;
        },
        requestError: function(err) {
            console.log("Request Error: " + err);

            httpState.busy--;

            return $q.reject(err);
        },

        response: function(r) {
            httpState.busy--;

            return r;
        },
        responseError: function(e) {
            console.log("Response Error: " + e);

            httpState.busy--;
            httpState.error = e;

            return $q.reject(e);
        },
    };

    return httpState
})

.config(function($httpProvider) {
    $httpProvider.interceptors.push('httpState');
})

.factory('Status', function($http) {
    var Status = {};

    $http.get('/api/status').then(
        function success(r) {
            Status.server = r.data.server;
            Status.mode = r.data.mode;
        }
    );

    return Status;
})

.factory('Index', function($http, Status) {
    return function() {
        return $http.get('/api/').then(
            function success(r) {
                return r.data;
            }
        );
    };
})

.factory('Screen', function($resource, Status) {
    return $resource('/api/screens/:id', { }, {
        get: {
            method: 'GET',
        },
        query: {
            method: 'GET',
            isArray: false,
        }
    }, {stripTrailingSlashes: true});
})

.filter('dimensions', function() {
    return function(dimensions) {
        if (dimensions && dimensions.width && dimensions.height) {
            return dimensions.width + "x" + dimensions.height;
        } else {
            return null;
        }
    };
})

.controller('HeaderCtrl', function($scope, $location, Status, httpState) {
    $scope.status = Status;
    $scope.state = httpState;

    $scope.isActive = function(prefix) {
        return $location.path().startsWith(prefix);
    };

})

.controller('MainCtrl', function($scope, $location, Index, $interval) {
    var busy = false;

    $scope.reload = function() {
        if (busy) {
            return;
        } else {
            busy = true;
        }

        Index().then(
            function success(index) {
                busy = false;

                $scope.screens = index.screens
                $scope.sources = $.map(index.sources, function(source, id){
                    return source;
                });
            },
            function error() {
                busy = false;
            }
        );
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
        var refreshInterval = refresh * 1000;

        if ($scope.refreshTimer) {
            $interval.cancel($scope.refreshTimer);
            $scope.refreshTimer = null;
        }

        if (refreshInterval) {
            $scope.refreshTimer = $interval($scope.reload, refreshInterval);
        }

        $location.search('refresh', refresh);
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
