var e2client = angular.module('e2client', [
	'ngWebsocket',
	'ui.bootstrap'
]);

var server = document.location.hostname;
var apiPort = parseInt(document.location.port);
var wsPort = apiPort + 1;

var backendUrl = 'http://' + server + ':' + apiPort + '/api/v1/';
var websocketUrl = 'ws://' + server + ':' + wsPort + '';

e2client.controller('PresetsCtrl', function ($scope, $http, $websocket) {
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
	
	// Websocket
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
		$scope.log('websocket message', data);
		$scope.loadPresets();
	});

	ws.$on('$close', function () {
		$scope.log('websocket closed');
	});
	
	// Presets
	$scope.loadPresets = function() {
		$scope.log("presets load");

		$http.get(backendUrl)
			.success(function(data) {
				$scope.log("presets update", {seq: data.seq, presets_length: Object.keys(data.presets).length});
				$scope.data = data;
				$scope.safe = data.safe;
				$scope.seq = data.seq;
			}).error(function(err) {
				$scope.log('presets error', err);
			});
	};

	$scope.clickPreset = function(id) {
		$scope.log("preset click", {id: id, seq: $scope.seq});

		$http.post(backendUrl + 'preset/' + id, {seq: $scope.seq})
			.success(function(data) {
				$scope.log('preset success', data);
				$scope.seq = data.seq;
			}).error(function(err) {
				$scope.log('preset error', err);
				$scope.loadPresets(); // reload to get current seq
			});
		return false;
	};
	
	// Commands
	$scope.autotrans = function() {
		return $scope.setInPgm({autotrans: true});
	}

	$scope.cut = function() {
		return $scope.setInPgm({cut: true});
	}

	$scope.setInPgm = function(data) {
		data.seq = $scope.seq;
		
		$scope.log("transition click", data);

		$http.post(backendUrl + 'preset/', data)
			.success(function(data) {
				$scope.log('transition success', data);
				$scope.seq = data.seq;
			}).error(function(err) {
				$scope.log('transition error');
				$scope.loadPresets(); // reload to get current seq
			});
		return false;
	}

	// Initialize
	$scope.loadPresets();
});
