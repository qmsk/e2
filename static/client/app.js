var phonecatApp = angular.module('e2presets', ['ngWebsocket']);

var backendUrl = 'http://192.168.0.16:8081/api/v1/'

phonecatApp.controller('PresetsCtrl', function ($scope, $http, $websocket) {

	$scope.base = {};
	$scope.base.inPreview = null;

	var ws = $websocket.$new({
		url: 'ws://192.168.0.16:8082',
		reconnect: true
	});

    ws.$on('$open', function () {
        console.log('websocket opened');
        ws.$emit('ping', 'hello');
    });

    ws.$on('$message', function (data) {
        console.log('websocket data:', data);
        $scope.loadPresets();
    });

    ws.$on('$close', function () {
        console.log('websocket closed');
    });

	$scope.loadPresets = function() {
		$http.get(backendUrl)
			.success(function(data) {
				$scope.data = data;
			}).error(function(err) {
				console.log("error", err);
			});
	};
	$scope.loadPresets();


	$scope.clickPreset = function(id) {
		$http.post(backendUrl + 'preset/' + id)
			.success(function(data) {
				// nothing
			}).error(function(err) {
				console.log("error", err);
			});
		return false;
	};

	$scope.autotrans = function() {
		return $scope.setInPgm({autotrans: true});
	}

	$scope.cut = function() {
		return $scope.setInPgm({cut: true});
	}

	$scope.setInPgm = function(data) {
		$http.post(backendUrl + 'preset/', data)
			.success(function(data) {
				// nothing
			}).error(function(err) {
				console.log("error", err);
			});
		return false;
	}
});