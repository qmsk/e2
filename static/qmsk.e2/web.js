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

.factory('Events', function($location, $websocket, $rootScope) {
    var Events = {
        url:    'ws://' + window.location.host + '/events',
        open:   false,
        error:  null,

        events: [],
        state:  {},
    }

    var ws = $websocket(Events.url);

    ws.onOpen(function() {
        console.log("WebSocket Open")
        Events.open = true;
    });
    ws.onError(function(error) {
        console.log("WebSocket Error: " + error)
        Events.error = error;
    });
    ws.onClose(function() {
        console.log("WebSocket Closed")
        Events.open = false;
    });

    ws.onMessage(function(message){
        var event = JSON.parse(message.data);
    
        Events.events.push(event);

        $.each(event, function(k, v) {
            Events.state[k] = v;
        });

        $rootScope.$broadcast('qmsk.e2.event', event);
    });

    return Events;
})
