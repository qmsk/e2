angular.module('qmsk.e2.source', [
        'ngResource',
])

.factory('Source', function($resource, Status) {
    return $resource('/api/sources/:id', { }, {
        get: {
            method: 'GET',
            interceptor:    Status.httpInterceptor,
        },
        query: {
            method: 'GET',
            isArray: false,
            interceptor:    Status.httpInterceptor,
        }
    }, {stripTrailingSlashes: true});
})

.directive('e2Source', function() {
    return {
        restrict: 'AE',
        scope: {
            source: '=source',
            detail: '=detail',
        },
        templateUrl: 'qmsk.e2/source/source.html',
    }
})
