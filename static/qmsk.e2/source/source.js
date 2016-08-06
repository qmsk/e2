angular.module('qmsk.e2.source', [
        'ngResource',
])

.directive('e2Source', function() {
    return {
        restrict: 'AE',
        scope: {
            source: '=source',
            input: '=input',
            detail: '=detail',
        },
        templateUrl: '/static/qmsk.e2/source/source.html',
    }
})
