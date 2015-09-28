# coreos-deploy
[![License MIT](https://img.shields.io/npm/l/express.svg)](http://opensource.org/licenses/MIT)
[![Build Status](https://travis-ci.org/composer22/coreos-deploy.svg?branch=master)](http://travis-ci.org/composer22/coreos-deploy)
[![Current Release](https://img.shields.io/badge/release-v0.0.1-brightgreen.svg)](https://github.com/composer22/coreos-deploy/releases/tag/v0.0.1)
[![Coverage Status](https://coveralls.io/repos/composer22/coreos-deploy/badge.svg?branch=master)](https://coveralls.io/r/composer22/coreos-deploy?branch=master)

An API server to allow deployment of Docker containers and metadata to a CoreOS cluster written in [Go.](http://golang.org)

## About

This is an API server that provides a means to launch and manage fleet service files on a
coreos cluster.

This service should be launched in a docker container on the service/control part of your cluster.
It is intended to act as a gate to your deployment and management of the cluster.

## Requirements

This service should be run on a CoreOS instance with etcd2 and fleetctl available.
A MySQL database is also required.

For the DB schema, please see ./db/schema.sql

## Usage

```
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

    coreos-deploy -v /var/run/fleet.sock:/var/run/fleet.sock -N "San Francisco" -H 0.0.0.0 \
	 -O example.com -E development -p 8080 \
	 -X 2 -T 0.0.0.0:2379 --dsn "id:password@tcp(your-amazonaws-uri.com:3306)/dbname"

	for DSN usage, see https://github.com/go-sql-driver/mysql
```
Please also see /docker dir for more information on running this service.

## HTTP API

Header for services other than /health should contain:

* Accept: application/json
* Authorization: Bearer with token
* Content-Type: application/json

Example cURL:

```
$ curl -i -H "Accept: application/json" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer S0M3B3EARERTOK3N" \
-X GET "http://0.0.0.0:8080/v1.0/info"

HTTP/1.1 200 OK
Content-Type: application/json;charset=utf-8
Date: Fri, 03 Apr 2015 17:29:17 +0000
Server: San Francisco
X-Request-Id: DC8D9C2E-8161-4FC0-937F-4CA7037970D5
Content-Length: 0
```

Three API routes are provided for service measurement:

* http://localhost:8080/v1.0/health - GET: Is the server alive?
* http://localhost:8080/v1.0/info - GET: What are the params of the server?
* http://localhost:8080/v1.0/metrics - GET: What performance and statistics are from the server?

An additional API is provided for displaying a map of machines and units running within the cluster.
Please see below for more information.

The following is the API Specification for deployment:
```
POST http://localhost:8080/v1.0/deploy

curl -i -H "Accept: application/json" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer S0M3B3EARERTOK3N" \
-X POST "http://0.0.0.0:8080/v1.0/deploy" \
-d "<json payload see below>"
```
json payload:
```
{
  "serviceName":"your-application-name",
  "version":"1.0.0",
  "numInstances":2,
  "serviceTemplate":"[Unit]...place your unit .service file code here",
  "etcd2Keys":{
     "etcd2key1":"value1",
     "etcd2key2":"value2",
     "etcd2key3":"value3",
     "etcd2keyn":"valuen as a string"
   }
}
```
This will return a UUID for the deploy:
```
{
    "deployID": "051A9069-0E3A-41EC-9C98-E6D29E91FBB3"
}
```
...that can be used to check the deploy status:
```
curl -i -H "Accept: application/json" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer S0M3B3EARERTOK3N" \
-X GET "http://0.0.0.0:8080/v1.0/status/051A9069-0E3A-41EC-9C98-E6D29E91FBB3"
```
which returns:
```
{
    "deployID": "051A9069-0E3A-41EC-9C98-E6D29E91FBB3",
    "domain": "example.com",
    "environment": "development",
    "serviceName": "your-application-name",
    "version": "1.0.0",
	"suffix": "abcd1234",
    "numInstances": 2,
    "status": 2,
    "message": "Service deployed successfully.",
    "log": "blabla...\nSUCCESS: Service deployed successfully.\n",
    "updatedAt": "2015-08-27 18:58:30",
    "createdAt": "2015-08-27 18:58:16"
}
```
## Fleet Unit Files and Instantiation

Each deploy should have a unique id assigned as a version.

Fleet templates should be expected to be named on the CoreOS cluster as:
```
<service-name>-<version>-<random 8 char suffix>@.service

ex: my-service-name-1.0.0-ha92kd9x@.service
```

Instantiation occurs in A/B rolling deploy, so services are named with an A/B designator.

When using fleetctl to

ex:
```
major-app-1.0.0-ha20dalf@A1.service
major-app-1.0.0-ha20dalf@A2.service
another-app-2.0.0-ah02kdm4@B1.service
another-app-2.0.0-ah02kdm4@B2.service
yet-another-app-2.0.0-j1dl002m@A1.service
yet-another-app-2.0.0-j1dl002m@A2.service
```
When setting up your unit code, you should include a template similar to this
to avoid collisions:
```
[X-Fleet]
Conflicts=major-app*@*.service
```
coreos-deploy.service is included as a reference.

## Cluster Map API

An additional Restful API is available to provide a display of machines and their unit status in
the cluster.
```
http://localhost:8080/v1.0/cluster_map?mq=<machinequerystring>&uq=<unitquerystring>
```
This call requires a Bearer token.

This returns a json structure of machines and units similar to this:
```
{
  "machines": [{
    "machine": "230be7181dbb4637b8c50773be97eb63",
    "ip": "172.21.1.162",
    "metadata": "role=control",
    "units": [{
      "unit": "example-coreos-deploy.service",
      "hash": "44f1640cfd59e55da0429264e78b4d35b0ff443c",
      "active": "active",
      "load": "loaded",
      "sub": "running"
    }]
  }, {
    "machine": "4b8c11e8ff3c436c80369bdfaec22931",
    "ip": "172.21.3.120",
    "metadata": "role=control",
    "units": [{
      "unit": "example-coreos-deploy.service",
      "hash": "44f1640cfd59e55da0429264e78b4d35b0ff443c",
      "active": "active",
      "load": "loaded",
      "sub": "running"
    }]
  }, {
    "machine": "2d86a7a150904750ad9d38d3a80ea663",
    "ip": "172.21.5.241",
    "metadata": "role=control",
    "units": [{
      "unit": "example-coreos-deploy.service",
      "hash": "44f1640cfd59e55da0429264e78b4d35b0ff443c",
      "active": "active",
      "load": "loaded",
      "sub": "running"
    }]
  }, {
    "machine": "0d636c8fb07e4d1db151d4cc1acd6f4f",
    "ip": "172.21.5.242",
    "metadata": "role=control",
    "units": [{
      "unit": "example-coreos-artifactory-monitor@1.service",
      "hash": "fde5cf3355639457d6fc6d6d4a8ca766d13cc456",
      "active": "active",
      "load": "loaded",
      "sub": "running"
    }, {
      "unit": "example-coreos-deploy.service",
      "hash": "44f1640cfd59e55da0429264e78b4d35b0ff443c",
      "active": "active",
      "load": "loaded",
      "sub": "running"
    }]
  }, {
    "machine": "e7a028794ea3400fb8f46d1856a32e64",
    "ip": "172.21.1.17",
    "metadata": "role=worker",
    "units": []
  }, {
    "machine": "070b415018da47c2848e46c34b3870e6",
    "ip": "172.21.1.18",
    "metadata": "role=worker",
    "units": [{
      "unit": "example-video-mobile-1.0.0-110-dw73yzlw@A1.service",
      "hash": "f70436553c4a43dc0ebefb4efd21472911a58171",
      "active": "active",
      "load": "loaded",
      "sub": "running"
    }]
  }, {
    "machine": "c50e89a9056a4a29944588e4376df464",
    "ip": "172.21.1.19",
    "metadata": "role=worker",
    "units": []
  }, {
    "machine": "9b05056853e342a5abf84c4c15b346dc",
    "ip": "172.21.3.133",
    "metadata": "role=worker",
    "units": []
  }, {
    "machine": "44d42fb7a69e441f86dc33957873507d",
    "ip": "172.21.3.134",
    "metadata": "role=worker",
    "units": [{
      "unit": "example-video-mobile-1.0.0-110-dw73yzlw@A2.service",
      "hash": "f70436553c4a43dc0ebefb4efd21472911a58171",
      "active": "active",
      "load": "loaded",
      "sub": "running"
    }]
  }, {
    "machine": "5bb0f9ef30a04562af047e55fd0fdcd1",
    "ip": "172.21.5.155",
    "metadata": "role=worker",
    "units": []
  }, {
    "machine": "71a80b5560994dbc854349d0efc1fa37",
    "ip": "172.21.5.156",
    "metadata": "role=worker",
    "units": []
  }, {
    "machine": "fe55f79b510542498602daec92144ac6",
    "ip": "172.21.5.157",
    "metadata": "role=worker",
    "units": []
  }]
}
```
* Machines are sorted by: metadata asc, ip asc.
* Units within a machine are sorted by: sub asc, unit asc

Two optional regular expression query string filers are allowed as parameters:

* mq - search all fields within machine for this regexp token.
* uq - search all fields within units of a machine for this regexp token.

## CLI Client Application

A client CLI is available that encapsulates the deploy and status endpoints.
This is useful for submitting and validating remote jobs rather than creating
your own restful client.

Please see the project [coreos-deploy-client](https://github.com/composer22/coreos-deploy-client/)

## Building

This code currently requires version 1.42 or higher of Go.

Information on Golang installation, including pre-built binaries, is available at
<http://golang.org/doc/install>.

Run `go version` to see the version of Go which you have installed.

Run `go build` inside the directory to build.

Run `go test ./...` to run the unit regression tests.

A successful build run produces no messages and creates an executable called `coreos-deploy` in this
directory.

Run `go help` for more guidance, and visit <http://golang.org/> for tutorials, presentations, references and more.

## Docker Images

A prebuilt docker image is available at (http://www.docker.com) [coreos-deploy](https://registry.hub.docker.com/u/composer22/coreos-deploy/)

If you have docker installed, run:
```
docker pull composer22/coreos-deploy:latest

or

docker pull composer22/coreos-deploy:<version>

if available.
```
See /docker directory README for more information on how to run it.

You should run this docker container on the control or service part of your cluster.

## License

(The MIT License)

Copyright (c) 2015 Pyxxel Inc.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to
deal in the Software without restriction, including without limitation the
rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
IN THE SOFTWARE.
