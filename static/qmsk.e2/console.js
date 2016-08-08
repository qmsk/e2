angular.module('qmsk.e2.console', [

])

.factory('Console', function($window) {
    var Console = {
        logID: 1,
        logMessages: [],
        log: function(message) {
            var logID = Console.logID++;

            $window.console.log(message)
            Console.logMessages.push({
                id: logID,
                message: message,
            });
        },
    };

    return Console;
})

.controller('ConsoleCtrl', function($scope, Console) {
    $scope.console = Console;
})

