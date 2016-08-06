angular.module('qmsk.e2', [
        'qmsk.e2.source',
        'qmsk.e2.web',
        'ngResource',
        'ngRoute',
        'ngWebSocket',
        'luegg.directives',
        'jsonFormatter',
])

.config(function($routeProvider) {
    $routeProvider
        .when('/main', {
            templateUrl: '/static/qmsk.e2/server/main.html',
            controller: 'MainCtrl',
            reloadOnSearch: false,
        })
        .when('/sources', {
            templateUrl: '/static/qmsk.e2/server/sources.html',
            controller: 'SourcesCtrl',
        })
        .when('/screens', {
            templateUrl: '/static/qmsk.e2/server/screens.html',
            controller: 'ScreensCtrl',
        })
        .when('/auxes', {
            templateUrl: '/static/qmsk.e2/server/auxes.html',
            controller: 'AuxesCtrl',
        })
        .when('/presets', {
            templateUrl: '/static/qmsk.e2/server/presets.html',
            controller: 'PresetsCtrl',
        })
        .when('/system', {
            templateUrl: '/static/qmsk.e2/server/system.html',
            controller: 'SystemCtrl',
        })
        .otherwise({
            redirectTo: '/main',
        });
})

.factory('Status', function($http) {
    var Status = {};

    $http.get('/api/status').then(
        function success(r) {
            Status.server = r.data.server;
            Status.mode = r.data.mode;
        }
    );

    return Status;
})

.factory('State', function($rootScope, Events) {
    var State = {
        System: Events.state.System,
    };

    $rootScope.$on('qmsk.e2.event', function($e, event){
        State.System = Events.state.System;
    });

    return State;
})

.factory('Preset', function($resource) {
    return $resource('/api/presets/:id', { }, {
        get: {
            method: 'GET',
            url: '/api/presets',
        },
        all: {
            method: 'GET',
            isArray: true,
        },
        query: {
            method: 'GET',
            isArray: false,
        },
        activate: {
            method: 'POST',
            url: '/api/presets',
        },
    }, {stripTrailingSlashes: false});
})

.filter('dimensions', function() {
    return function(dimensions) {
        if (dimensions && dimensions.width && dimensions.height) {
            return dimensions.width + "x" + dimensions.height;
        } else {
            return null;
        }
    };
})

.controller('HeaderCtrl', function($scope, $location, Status, httpState) {
    $scope.status = Status;
    $scope.state = httpState;

    $scope.isActive = function(prefix) {
        return $location.path().startsWith(prefix);
    };

})

.controller('StatusCtrl', function($scope, Events) {
    $scope.events = Events;
})

.controller('SystemCtrl', function($scope, State) {
    $scope.state = State;
})

.controller('MainCtrl', function($scope, $location, State) {
    $scope.state = State;
    $scope.sources = [];

    $scope.$watch('state.System', function(system) {
        // compute a merged state mapping sources to their active destinations
        // TODO: aux destinations
        $scope.sources = $.map(system.SrcMgr.SourceCol, function(source, sourceID){
            var item = {
                id: sourceID,
                type: source.SrcType,
                name: source.Name,
                source: source,
                active: false,
                preview: [],
                program: [],
            };

            if (source.SrcType == "input") {
                item.input = system.SrcMgr.InputCfgCol[source.InputCfgIndex];
            }

            $.each(system.DestMgr.ScreenDestCol, function(screenID, screen) {
                var output = {
                    type: "screen",
                    id: screenID,
                    name: screen.Name,
                    active: screen.IsActive > 0,
                };

                $.each(screen.LayerCollection, function(layerID, layer) {
                    if (layer.PgmMode > 0 && layer.LastSrcIdx == sourceID) {
                        output.program = true;
                    }

                    if (layer.PvwMode > 0 && layer.LastSrcIdx == sourceID) {
                        output.preview = true;
                    }
                });

                if (output.program) {
                    item.program.push(output);
                }
                if (output.preview) {
                    item.preview.push(output);
                }
                if (output.active && output.preview) {
                    item.active = true;
                }
            });

            return item;
        });
    });

    $scope.selectOrder = function(order) {
        $scope.order = order;
        $scope.orderBy = function(){
            switch (order) {
            case 'source':
                return ['-type', 'name'];
            case 'preview':
                return ['preview_screens', 'program_screens'];
            case 'program':
                return ['program_screens', 'preview_screens'];
            default:
                return [];
            }
        }();

        $location.search('order', order || null);
    };
    $scope.selectOrder($location.search().order || 'source');
})

.controller('SourcesCtrl', function($scope, State) {
    $scope.state = State;
})

.controller('ScreensCtrl', function($scope, State) {
    $scope.state = State;
})

.controller('AuxesCtrl', function($scope, State) {
    $scope.state = State;
})

.controller('PresetsCtrl', function($scope, State, Preset) {
    $scope.state = State;

    // TODO: from URL params
    $scope.collapseGroups = { };

    // group
    $scope.presets = Preset.all(function (presets) {
        var groups = { };

        $.each(presets, function(id, preset) {

            // group it
            var group = groups[preset.group];

            if (!group) {
                group = groups[preset.group] = [];
            }

            group.push(preset);
        });

        $scope.groups = $.map(groups, function(presets, id){
            return {id: id, presets: presets};
        });
    });

    $scope.busy = false;
    $scope.activatePreset = function(preset) {
        if ($scope.busy) {
            return;
        } else {
            $scope.busy = true;
        }

        Preset.activate({id: preset.id},
            function success(r) {
                $scope.busy = false;
            },
            function error(e) {
                $scope.busy = false;
            }
        );
    };
})

;
