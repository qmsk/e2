angular.module('qmsk.e2.web', [
        'ngResource',
])

// track global http state
.factory('httpState', function($q) {
    var httpState = {
        error:  null,
        busy:   0,

        request: function(config) {
            httpState.busy++;

            return config;
        },
        requestError: function(err) {
            console.log("Request Error: " + err);

            httpState.busy--;

            return $q.reject(err);
        },

        response: function(r) {
            httpState.busy--;

            return r;
        },
        responseError: function(e) {
            console.log("Response Error: " + e);

            httpState.busy--;
            httpState.error = e;

            return $q.reject(e);
        },
    };

    return httpState
})

.config(function($httpProvider) {
    $httpProvider.interceptors.push('httpState');
})


