angular.module('qmsk.e2.status', [
    'qmsk.e2.console',
])

.controller('StatusCtrl', function($scope, Console) {
    $scope.console = Console;
})

