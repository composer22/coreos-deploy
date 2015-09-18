package server

import (
	"fmt"
	"os"
)

const usageText = `
Description: coreos-deploy is a server for deploying services to a coreos cluster.

Usage: coreos-deploy [options...]

Server options:
    -N, --name NAME                  NAME of the server (default: empty field).
    -H, --hostname HOSTNAME          HOSTNAME of the server (default: localhost).
    -O, --domain DOMAIN              DOMAIN of the site being managed (default: localhost).
    -E, --environment ENVIRONMENT    ENVIRONMENT (development, qa, staging, production).
    -p, --port PORT                  PORT to listen on (default: 6660).
    -L, --profiler_port PORT         *PORT the profiler is listening on (default: off).
    -X, --procs MAX                  *MAX processor cores to use from the machine.
    -T, --etcd2_endpoint IP:PORT     IP:PORT of the etcd2 instance to use.
    -D, --dsn DSN                    DSN string used to connect to database.

    -d, --debug                      Enable debugging output (default: false)

     *  Anything <= 0 is no change to the environment (default: 0).

Common options:
    -h, --help                       Show this message
    -V, --version                    Show version

Examples:

    coreos-deploy -v /var/run/fleet.sock:/var/run/fleet.soc -N "San Francisco" -H 0.0.0.0 -O example.com \
	 -E development --p 8080 \
	 -X 2 -T 0.0.0.0 --dsn "id:password@tcp(your-amazonaws-uri.com:3306)/dbname"

	for DSN usage, see https://github.com/go-sql-driver/mysql
`

// PrintUsageAndExit is used to print out command line options.
func PrintUsageAndExit() {
	fmt.Printf("%s\n", usageText)
	os.Exit(0)
}
