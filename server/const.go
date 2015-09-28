package server

import "time"

const (
	version              = "0.0.2"        // Application and server version.
	DefaultHostName      = "localhost"    // The hostname of the server.
	DefaultEnvironment   = "development"  // The default environment for the server.
	DefaultPort          = 8080           // Port to receive requests: see IANA Port Numbers.
	DefaultProfPort      = 0              // Profiler port to receive requests.*
	DefaultMaxProcs      = 0              // Maximum number of computer processors to utilize.*
	DefaultEtcd2Endpoint = "0.0.0.0:2379" // Default address and port to etcd2 service.

	suffixSize = 8 // Added to service name to make it unique on deploy.

	// * zeros = no change or no limitations or not enabled.

	// http: routes.
	httpRouteV1Health     = "/v1.0/health"
	httpRouteV1Info       = "/v1.0/info"
	httpRouteV1Metrics    = "/v1.0/metrics"
	httpRouteV1Deploy     = "/v1.0/deploy"
	httpRouteV1Status     = "/v1.0/status/"
	httpRouteV1ClusterMap = "/v1.0/cluster_map"

	// Connections.
	TCPReadTimeout  = 10 * time.Second
	TCPWriteTimeout = 10 * time.Second

	httpGet    = "GET"
	httpPost   = "POST"
	httpPut    = "PUT"
	httpDelete = "DELETE"
	httpHead   = "HEAD"
	httpTrace  = "TRACE"
	httpPatch  = "PATCH"

	// Working directories.
	tmpDir = "/tmp/coreos-deploy/"
	binDir = "/usr/local/bin/"
	appDir = binDir + "coreos-deploy/"

	// Executables.
	fleetctl = appDir + "fleetctl"

	// Error messages.
	InvalidMediaType     = "Invalid Content-Type or Accept header value."
	InvalidMethod        = "Invalid Method for this route."
	InvalidBody          = "Invalid body of text in request."
	InvalidJSONText      = "Invalid JSON format in text of body in request."
	InvalidJSONAttribute = "Invalid - 'text' attribute in JSON not found."
	InvalidAuthorization = "Invalid authorization."
	InvalidQueryString   = "Invalid query string."
)
