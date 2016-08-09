angular.module('qmsk.e2', [
        'qmsk.e2.console',
        'qmsk.e2.web',
        'luegg.directives',
])

.controller('StateCtrl', function($scope, Events) {
    $scope.events = Events;
    $scope.state = Events.state;
})

.controller('StatusCtrl', function($scope, Console) {
    $scope.console = Console;
})

