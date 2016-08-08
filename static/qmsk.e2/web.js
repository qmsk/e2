angular.module('qmsk.e2.web', [
        'qmsk.e2.console',
        'ngResource',
])

// track global http state
.factory('httpState', function($q, Console) {
    var httpState = {
        error:  null,
        busy:   0,

        request: function(config) {
            httpState.busy++;

            return config;
        },
        requestError: function(err) {
            Console.log("HTTP Request Error: " + err);

            httpState.busy--;

            return $q.reject(err);
        },

        response: function(r) {
            Console.log("HTTP " + r.config.method + " " + r.config.url + ": " + r.status + " " + r.statusText);

            httpState.busy--;

            return r;
        },
        responseError: function(e) {
            Console.log("HTTP Response Error: " + e);

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

.factory('Events', function($location, $websocket, $rootScope, Console) {
    var Events = {
        url:    'ws://' + window.location.host + '/events',
        open:   false,
        error:  null,

        events: [],
        state:  {},
    }

    var ws = $websocket(Events.url);

    ws.onOpen(function() {
        Console.log("WebSocket Open")
        Events.open = true;
    });
    ws.onError(function(error) {
        Console.log("WebSocket Error: " + error)
        Events.error = error;
    });
    ws.onClose(function() {
        Console.log("WebSocket Closed")
        Events.open = false;
    });

    ws.onMessage(function(message){
        var event = JSON.parse(message.data);
    
        Events.events.push(event);

        $.each(event, function(k, v) {
            Events.state[k] = v;
        });
        
        Console.log("WebSocket Update")

        $rootScope.$broadcast('qmsk.e2.event', event);
    });

    return Events;
})
