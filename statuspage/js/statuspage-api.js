// config.js must be included BEFORE this file!

// Configure access to storage
checkup.storage.setup(checkup.config.storage);

// Initialize notification flag
checkup.notified = 0

// Once the DOM is loaded, go ahead and fetch the timeline.
document.addEventListener('DOMContentLoaded', function() {
	checkup.domReady = true;

	checkup.dom.favicon = document.getElementById("favicon");
	checkup.dom.status = document.getElementById("overall-status");
	checkup.dom.statustext = document.getElementById("overall-status-text");
	checkup.dom.timeframe = document.getElementById("info-timeframe");
	checkup.dom.checkcount = document.getElementById("info-checkcount");
	checkup.dom.lastcheck = document.getElementById("info-lastcheck");
	checkup.dom.availability = document.getElementById("info-availability");
	checkup.dom.timeline = document.getElementById("timeline");

	var thisURL = new URL(window.location.href);
	var days = Number(thisURL.searchParams.get("days"));
	if (days > 0) {
		checkup.config.timeframe = days * time.Day;
	}

	// Immediately begin downloading events
	var firstTimestamp = Math.floor((time.Now() - checkup.config.timeframe) / 1e9);
	checkup.storage.getChecksWithin(firstTimestamp, processNewCheckResult,
		processStatus, allCheckFilesLoaded, firstTimestamp);

	if (Notification && Notification.permission !== "granted")
	    Notification.requestPermission();
}, false);

// Keep page updated
setInterval(function() {
	checkup.storage.getNewChecks(processNewCheckResult, processStatus,
		allCheckFilesLoaded, checkup.config.timeframe);
}, checkup.config.refresh_interval * 1000);

// Update "time ago" tags every so often
setInterval(function() {
	if (checkup.lastStatusTs != 0)
		refreshStatusTime();
}, 5000);

function refreshStatusTime() {
	var mtt = checkup.makeTimeTag(checkup.lastStatusTs * 1e-6) + " ago";
	checkup.dom.lastcheck.innerHTML = mtt;
}


function processStatus(statusString, results, timestamp, statistics) {
	checkup.statusString = statusString;
	checkup.lastStatusTs = timestamp;
	refreshStatusTime();

	statusDiv = document.getElementById('status');

	// Clear status
	while (statusDiv.hasChildNodes()) {
		statusDiv.removeChild(statusDiv.lastChild);
	}

	var table = document.createElement('table');
	var tr = document.createElement('tr');
	tr.innerHTML = "<th>Target</th><th>Status</th>";
	table.appendChild(tr);

	results.forEach(function(r) {
		var tr = document.createElement('tr');
		//var target = document.createElement('div');
		var rStatus = makeTargetStatus(r);
		if (rStatus == "UP")
			tr.className = 'healthy';
		if (rStatus == "DOWN" || rStatus == "DEGRADED")
			tr.className = 'alert';
		tr.innerHTML = "<td><a href='"+r.endpoint+"'>"+r.title+"</a></td><td>" + rStatus + "</td>";
		table.appendChild(tr);
	});

	statusDiv.appendChild(table);

	checkup.dom.availability.innerHTML = statistics.HealthyPC.toFixed(3) + " %";
}

function makeTargetStatus(check) {
	if (check.healthy)
		return "UP";
	if (check.down)
		return "DOWN";
	if (check.degraded)
		return "DEGRADED";
	return "Unknown"
}

function processNewCheckResult(json, timestamp) {
	checkup.checks.push(json);

	if (timestamp > checkup.lastCheckTs) {
		checkup.lastCheckTs = timestamp;
	}

	var process = function(result) {
		checkup.orderedResults.push(result); // will sort later, more efficient that way

		if (!checkup.groupedResults[result.timestamp])
			checkup.groupedResults[result.timestamp] = [result];
		else
			checkup.groupedResults[result.timestamp].push(result);

		if (!checkup.results[result.endpoint])
			checkup.results[result.endpoint] = [result];
		else
			checkup.results[result.endpoint].push(result);

		return;
	};

	process(json);
}

function allCheckFilesLoaded(numChecksLoaded, numResultsLoaded) {
	// Sort the result lists
	//checkup.orderedResults.sort(function(a, b) { return a.timestamp - b.timestamp; });
	//for (var endpoint in checkup.results)
	//	checkup.results[endpoint].sort(function(a, b) { return a.timestamp - b.timestamp; });

	// Create events for the timeline

	var newEvents = [];
	var statuses = {}; // keyed by endpoint

	// First load the last known status of each endpoint
	for (var i = checkup.events.length-1; i >= 0; i--) {
		var result = checkup.events[i].result;
		if (!statuses[result.endpoint])
			statuses[result.endpoint] = checkup.events[i].status;
	}

	// Then go through the new results and look for new events
	for (var i = checkup.orderedResults.length-numResultsLoaded; i < checkup.orderedResults.length; i++) {
		var result = checkup.orderedResults[i];

		var status = "healthy";
		if (result.degraded) status = "degraded";
		else if (result.down) status = "down";

		if (status != statuses[result.endpoint]) {
			// New event because status changed
			newEvents.push({
				id: checkup.eventCounter++,
				result: result,
				status: status
			});
		}
		if (result.message) {
			// New event because message posted
			newEvents.push({
				id: checkup.eventCounter++,
				result: result,
				status: status,
				message: result.message
			});
		}

		statuses[result.endpoint] = status;
	}

	checkup.events = checkup.events.concat(newEvents);

	function renderTime(ns) {
		var d = new Date(ns * 1e-6);
		var month = d.getMonth()+1;
		var day = d.getDate();
		var dateString = 1900+d.getYear() + "-" + checkup.leftpad(month, 2, "0") + "-" + checkup.leftpad(day, 2, "0");
		var timeString = checkup.leftpad(d.getHours(), 2, "0")+":"+checkup.leftpad(d.getMinutes(), 2, "0");
		return dateString + " " + timeString;
	}

	// Render events
	for (var i = 0; i < newEvents.length; i++) {
		var e = newEvents[i];

		// Render event to timeline
		var evtElem = document.createElement("div");
		evtElem.setAttribute("data-eventid", e.id);
		evtElem.classList.add("event-item");
		evtElem.classList.add("event-id-"+e.id);
		evtElem.classList.add(checkup.color[e.status]);
		if (e.message) {
			evtElem.classList.add("message");
			evtElem.innerHTML = '<div class="message-head">'+checkup.makeTimeTag(e.result.timestamp*1e-6)+' ago</div>';
			evtElem.innerHTML += '<div class="message-body">'+e.message+'</div>';
		} else {
			evtElem.classList.add("event");
			evtElem.innerHTML = '<span class="time">'+renderTime(e.result.timestamp)+'</span> '+e.result.title+" "+e.status;
		}
		checkup.dom.timeline.insertBefore(evtElem, checkup.dom.timeline.childNodes[0]);
	}

	// Update DOM now that we have the whole picture

	// Update overall status
	var overall = "healthy";
	for (var endpoint in checkup.results) {
		if (overall == "down") break;
		var lastResult = checkup.results[endpoint][checkup.results[endpoint].length-1];
		if (lastResult) {
			if (lastResult.down)
				overall = "down";
			else if (lastResult.degraded)
				overall = "degraded";
		}
	}

	if (overall == "healthy" && checkup.statusString == "OK") {
		checkup.dom.favicon.href = "images/status-green.png";
		checkup.dom.status.className = "green";
		checkup.dom.statustext.innerHTML = checkup.config.status_text.healthy || "System Nominal";
		checkup.notified = 0;
	} else if (overall == "down") {
		checkup.dom.favicon.href = "images/status-red.png";
		checkup.dom.status.className = "red";
		checkup.dom.statustext.innerHTML = checkup.config.status_text.down || "Outage";
	        if (checkup.notified < 2 && Notification) {
		    var notification = new Notification('CheckupAPI', {
			body: "Unreachable target(s)!",
		    });
		    checkup.notified = 2;
		}
	} else if (overall == "degraded" || checkup.statusString == "DEGRADED") {
		checkup.dom.favicon.href = "images/status-yellow.png";
		checkup.dom.status.className = "yellow";
		checkup.dom.statustext.innerHTML = checkup.config.status_text.degraded || "Sub-Optimal";
	        if (checkup.notified < 1 && Notification) {
		    var notification = new Notification('CheckupAPI', {
			body: "Degraded system",
		    });
		    checkup.notified = 1;
		}
	} else {
		checkup.dom.favicon.href = "images/status-gray.png";
		checkup.dom.status.className = "gray";
		checkup.dom.statustext.innerHTML = checkup.config.status_text.unknown || "Status Unknown";
	}

	checkup.dom.timeframe.innerHTML = checkup.formatDuration(checkup.config.timeframe);
	//checkup.dom.checkcount.innerHTML = checkup.checks.length;
	checkup.dom.checkcount.innerHTML = checkup.events.length;
}
