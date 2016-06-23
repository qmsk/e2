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
        .when('/tally', {
            templateUrl: '/static/qmsk.e2/tally.html',
            controller: 'TallyCtrl',
            reloadOnSearch: false,
        })
        .otherwise({
            redirectTo: '/tally',
        });
})

.factory('Tally', function($http) {
    return $http.get('/api/tally/').then(
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

.controller('TallyCtrl', function($scope, Tally) {
    Tally.then(function(tally){
        $scope.tally = tally;
    });
})
