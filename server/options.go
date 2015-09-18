package server

import "encoding/json"

// Options represents parameters that are passed to the application to be used in constructing
// the server.
type Options struct {
	Name          string `json:"name"`          // The name of the server.
	HostName      string `json:"hostName"`      // The hostname of the server.
	Domain        string `json:"domain"`        // The domain of the server.
	Environment   string `json:"environment"`   // The environment of the server (dev, stage, prod, etc).
	Port          int    `json:"port"`          // The default port of the server.
	ProfPort      int    `json:"profPort"`      // The profiler port of the server.
	Etcd2Endpoint string `json:"etcd2Endpoint"` // The IP address and port to the etcd2 service.
	DSN           string `json:"-"`             // The DSN login string to the database.
	MaxProcs      int    `json:"maxProcs"`      // The maximum number of processor cores available.
	Debug         bool   `json:"debugEnabled"`  // Is debugging enabled in the application or server.
}

// String is an implentation of the Stringer interface so the structure is returned as a string
// to fmt.Print() etc.
func (o *Options) String() string {
	b, _ := json.Marshal(o)
	return string(b)
}
