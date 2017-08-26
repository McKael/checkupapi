checkup.config = {
        // How much history to show on the status page. Long durations and
        // frequent checks make for slow loading, so be conservative.
        // This value is in NANOSECONDS to mirror Go's time package.
        "timeframe": 8 * time.Day,

        // How often, in seconds, to query the API and update the page.
        "refresh_interval": 60,

        // Configure access to the checkup API server.
        // It is easier to serve files and API from the same domain,
        // or you will probably have to enable CORS.
        "storage": {
                "url_api": "."
        },

        // The text to display along the top bar depending on overall status.
        "status_text": {
                "healthy": "Situation Normal",
                "degraded": "Degraded Service",
                "down": "Service Disruption"
        }
};
