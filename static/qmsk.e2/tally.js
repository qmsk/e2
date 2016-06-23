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
            templateUrl: '/static/qmsk.e2/tally-sources.html',
            controller: 'SourcesCtrl',
            reloadOnSearch: false,
        })

        .when('/tally', {
            templateUrl: '/static/qmsk.e2/tally.html',
            controller: 'TallyCtrl',
            reloadOnSearch: false,
        })
        .otherwise({
            redirectTo: '/tally',
        });
})

.factory('Sources', function($http) {
    return $http.get('/api/sources').then(
        function success(r) {
            return r.data;
        }
    );
})

.factory('Tally', function($http) {
    return $http.get('/api/tally').then(
        function success(r) {
            return r.data;
        }
    );
})

.controller('HeaderCtrl', function($scope, $location, httpState) {
    $scope.state = httpState;

    $scope.isActive = function(prefix) {
        return $location.path().startsWith(prefix);
    };
})

.controller('SourcesCtrl', function($scope, Sources) {
    Sources.then(function(sources){
        $scope.sources = sources;
    });
})

.controller('TallyCtrl', function($scope, Events) {
    $scope.tally = Events.state.tally;

    $scope.$on('qmsk.e2.event', function($e, event){
        $scope.tally = Events.state.tally;
    });
})
