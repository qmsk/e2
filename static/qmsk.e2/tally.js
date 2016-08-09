angular.module('qmsk.e2.tally', [
        'qmsk.e2.web',
        'ngResource',
        'ngRoute',
        'ngWebSocket',
        'luegg.directives',
        'jsonFormatter',
])

.config(function($routeProvider) {
    $routeProvider
        .when('/sources', {
            templateUrl: '/static/qmsk.e2/tally/sources.html',
            controller: 'SourcesCtrl',
            reloadOnSearch: false,
        })

        .when('/tally/:id', {
            templateUrl: '/static/qmsk.e2/tally/tally.html',
            controller: 'TallyCtrl',
            resolve: {
                id: function($q, $route) {
                    var d = $q.defer();
                    var id = parseInt($route.current.params.id, 10);

                    if (isNaN(id)) {
                        d.reject("Invalid tally :id");
                    } else {
                        d.resolve(id);
                    }

                    return d.promise;
                },
            },
        })
        .when('/tally', {
            templateUrl: '/static/qmsk.e2/tally/tallys.html',
            controller: 'TallyIndexCtrl',
            reloadOnSearch: false,
        })

        .when('/inputs', {
            templateUrl: '/static/qmsk.e2/tally/inputs.html',
            controller: 'InputsCtrl',
            reloadOnSearch: false,
        })

        .when('/outputs', {
            templateUrl: '/static/qmsk.e2/tally/outputs.html',
            controller: 'OutputsCtrl',
            reloadOnSearch: false,
        })

        .otherwise({
            redirectTo: '/tally',
        });
})

.factory('Tally', function($http) {
    return $http.get('/api/tally').then(
        function success(r) {
            return r.data;
        }
    );
})

.controller('HeaderCtrl', function($scope, $location, httpState) {
    $scope.httpState = httpState;

    $scope.isActive = function(prefix) {
        return $location.path().startsWith(prefix);
    };
})

.controller('StateCtrl', function($scope, Events) {
    $scope.state = Events.state;
})

.controller('TallyIndexCtrl', function($scope) {

})

.controller('TallyCtrl', function($scope, id) {
    $scope.tallyID = id;
    $scope.tally = null;

    $scope.$watch('state.tally', function(tallyState) {
        $scope.tally = null;

        $.each(tallyState.Tally, function(i, tally) {
            if (tally.ID == $scope.tallyID) {
                $scope.tally = tally;
            }
        });
    });
})

.controller('SourcesCtrl', function($scope) {

})
.controller('InputsCtrl', function($scope) {

})
.controller('OutputsCtrl', function($scope) {

})

;
