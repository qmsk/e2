var phonecatApp = angular.module('e2presets', [
	'ngWebsocket',
	'ui.bootstrap'
]);

var server = document.location.hostname;
var apiPort = parseInt(document.location.port);
var wsPort = apiPort + 1;

var backendUrl = 'http://' + server + ':' + apiPort + '/api/v1/';
var websocketUrl = 'ws://' + server + ':' + wsPort + '';

phonecatApp.controller('PresetsCtrl', function ($scope, $http, $websocket) {

	$scope.base = {};
	$scope.base.inPreview = null;
	$scope.status = [];
	$scope.seq = 0;
	$scope.collapsedGroups = {};
	$scope.server = server;

	$scope.log = function (msg, data) {
		console.log(msg, data);
		$scope.status.unshift({msg: msg, data: data});
	};

	var ws = $websocket.$new({
		url: websocketUrl,
		reconnect: true,

		// workaround https://github.com/wilk/ng-websocket/issues/11
		protocols: []
	});

    ws.$on('$open', function () {
        $scope.log('websocket opened');
        ws.$emit('ping', 'hello');
        $scope.loadPresets(); // reload to get current seq
    });

    ws.$on('$message', function (data) {
        $scope.log('websocket data: ', data);
        $scope.loadPresets();
    });

    ws.$on('$close', function () {
        $scope.log('websocket closed');
    });

    var fff = false;

	$scope.loadPresets = function() {
		$http.get(backendUrl)
			.success(function(data) {
				$scope.data = data;
				$scope.safe = data.safe;
				$scope.seq = data.seq;
			}).error(function(err) {
				$scope.log('error loading preset data');
			});
	};
	$scope.loadPresets();


	$scope.clickPreset = function(id) {
		$http.post(backendUrl + 'preset/' + id, {seq: $scope.seq})
			.success(function(data) {
				$scope.log('preset clicked:', data);
			}).error(function(err) {
				$scope.log('error when selecting preset');
				$scope.loadPresets(); // reload to get current seq
			});
		return false;
	};

	$scope.autotrans = function() {
		return $scope.setInPgm({autotrans: true, seq: $scope.seq});
	}

	$scope.cut = function() {
		return $scope.setInPgm({cut: true, seq: $scope.seq});
	}

	$scope.setInPgm = function(data) {
		$http.post(backendUrl + 'preset/', data)
			.success(function(data) {
				$scope.log('transition clicked: ', data);
			}).error(function(err) {
				$scope.log('error with transition');
				$scope.loadPresets(); // reload to get current seq
			});
		return false;
	}
});
