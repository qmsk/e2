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

        .when('/tally', {
            templateUrl: '/static/qmsk.e2/tally/tally.html',
            controller: 'TallyCtrl',
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
    $scope.tally = Events.state.tally;

    $scope.$on('qmsk.e2.event', function($e, event){
        $scope.tally = Events.state.tally;
    });
})

.controller('TallyCtrl', function($scope) {

})
.controller('SourcesCtrl', function($scope) {

})
.controller('InputsCtrl', function($scope) {

})
.controller('OutputsCtrl', function($scope) {

})

;
