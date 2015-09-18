// coreos-deploy is a simple server for deploying docker containers into a coreos cluster.
package main

import (
	"flag"
	"runtime"
	"strings"

	"github.com/composer22/coreos-deploy/logger"
	"github.com/composer22/coreos-deploy/server"
)

var (
	log *logger.Logger
)

func init() {
	log = logger.New(logger.UseDefault, false)
}

// main is the main entry point for the application or server launch.
func main() {
	opts := &server.Options{}
	var showVersion bool

	flag.StringVar(&opts.Name, "N", "", "Name of the server.")
	flag.StringVar(&opts.Name, "name", "", "Name of the server.")
	flag.StringVar(&opts.HostName, "H", server.DefaultHostName, "HostName of the server.")
	flag.StringVar(&opts.HostName, "hostname", server.DefaultHostName, "HostName of the server.")
	flag.StringVar(&opts.Domain, "O", "", "Domain of the server.")
	flag.StringVar(&opts.Domain, "domain", "", "Domain of the server.")
	flag.StringVar(&opts.Environment, "E", server.DefaultEnvironment, "Environment of the server.")
	flag.StringVar(&opts.Environment, "environment", server.DefaultEnvironment, "Environment of the server.")
	flag.IntVar(&opts.Port, "p", server.DefaultPort, "Port to listen on.")
	flag.IntVar(&opts.Port, "port", server.DefaultPort, "Port to listen on.")
	flag.IntVar(&opts.ProfPort, "L", server.DefaultProfPort, "Profiler port to listen on.")
	flag.IntVar(&opts.ProfPort, "profiler_port", server.DefaultProfPort, "Profiler port to listen on.")
	flag.IntVar(&opts.MaxProcs, "X", server.DefaultMaxProcs, "Maximum processor cores to use.")
	flag.IntVar(&opts.MaxProcs, "procs", server.DefaultMaxProcs, "Maximum processor cores to use.")
	flag.StringVar(&opts.Etcd2Endpoint, "T", server.DefaultEtcd2Endpoint, "IP address and port to the etcd2 service.")
	flag.StringVar(&opts.Etcd2Endpoint, "etcd2_endpoint", server.DefaultEtcd2Endpoint, "IP address and port to the etcd2 service.")
	flag.StringVar(&opts.DSN, "D", "", "DSN connection string.")
	flag.StringVar(&opts.DSN, "dsn", "", "DSN connection string.")
	flag.BoolVar(&opts.Debug, "d", false, "Enable debugging output.")
	flag.BoolVar(&opts.Debug, "debug", false, "Enable debugging output.")
	flag.BoolVar(&showVersion, "V", false, "Show version.")
	flag.BoolVar(&showVersion, "version", false, "Show version.")
	flag.Usage = server.PrintUsageAndExit
	flag.Parse()

	// Version flag request?
	if showVersion {
		server.PrintVersionAndExit()
	}

	// Check additional params beyond the flags.
	for _, arg := range flag.Args() {
		switch strings.ToLower(arg) {
		case "version":
			server.PrintVersionAndExit()
		case "help":
			server.PrintUsageAndExit()
		}
	}

	// Set thread and proc usage.
	if opts.MaxProcs > 0 {
		runtime.GOMAXPROCS(opts.MaxProcs)
	}
	log.Infof("NumCPU %d GOMAXPROCS: %d\n", runtime.NumCPU(), runtime.GOMAXPROCS(-1))

	s := server.New(opts, log)

	if err := s.Start(); err != nil {
		log.Errorf(err.Error())
	}
}
