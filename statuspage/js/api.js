/**

CheckupAPI Adapter for Checkup.js

**/

var checkup = checkup || {};

checkup.storage = (function() {
	var apiBase;

	// setup prepares this storage unit to operate.
	this.setup = function(cfg) {
		var url = cfg.url_api;
		apiBase = url + '/api/v1';
	};

	// getChecksWithin gets all the checks since starttime (UNIX timestamp),
	// and executes callback for each check result.
	this.getChecksWithin = function(starttime, resultCallback, statusCallback, doneCallback, statsStarttime) {
		var checksLoaded = 0, resultsLoaded = 0;

		var apiQueryCheckup = apiBase + '/checkup?start='+starttime+'&stats_start='+statsStarttime

		checkup.getJSON(apiQueryCheckup, function(checkupResult) {
			statusCallback(checkupResult.status, checkupResult.last_results,
				checkupResult.timestamp, checkupResult.stats);

			var tl = checkupResult.timeline || [];
			tl.forEach(function(checkJSON) {
				checksLoaded++;
				resultsLoaded++;
				if (typeof resultCallback === 'function')
					resultCallback(checkJSON.Result, checkJSON.Timestamp);
			});
			if (typeof doneCallback === 'function')
				doneCallback(checksLoaded, resultsLoaded);
		});

	};

	// getNewChecks gets any checks since the timestamp on the file name
	// of the youngest check file that has been downloaded. If no check
	// files have been downloaded, no new check files will be loaded.
	this.getNewChecks = function(resultCallback, statusCallback, doneCallback, statsTimeframe) {
		var after;
		if (checkup.lastCheckTs == null) {
			// No check has been loaded yet: fall back to the configured
			// timeframe so the page can still discover checks (and set
			// lastCheckTs) when the initial load returned no events.
			after = Math.floor((time.Now() - statsTimeframe) / 1e9);
		} else {
			after = Math.floor(checkup.lastCheckTs / 1e9);
		}
		var statsAfter = Math.floor((time.Now() - statsTimeframe) / 1e9);
		return this.getChecksWithin(after+1, resultCallback, statusCallback, doneCallback, statsAfter);
	};

	return this;
})();
