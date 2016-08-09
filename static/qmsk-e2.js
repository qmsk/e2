angular.module('qmsk.e2', [
        'qmsk.e2.console',
        'qmsk.e2.web',
        'luegg.directives',
])

.controller('StateCtrl', function($scope, Events) {
    $scope.events = Events;
    $scope.state = Events.state;
})

.controller('HeaderCtrl', function($scope, $location, httpState) {
    $scope.httpState = httpState;
    
    $scope.isActive = function(prefix) {
        return $location.path().startsWith(prefix);
    };
})

.controller('StatusCtrl', function($scope, Console) {
    $scope.console = Console;
})

