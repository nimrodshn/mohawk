# mohawk

MOck HAWKular is a mock Hawkular server for testing.

## Introduction

Utility server for testing Hawkular clients, the server can mock
a running metrics Hawkular server. It can use different backends for different test use cases.

  - Sqlite backend, persistable read and write, mimics regular behavior of a Hawkular metrics server.
  - Random backend, mimics lots of metrics available only for reading.
  - Timeout backend, mimics errors while query data.


## License and copyright

```
   Copyright 2016 Red Hat, Inc. and/or its affiliates
   and other contributors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
```

## Installation

To install, get the source code, or do ``go install github.com/yaacov/mohawk`` if using go.
To run, users will need the ``server.key`` and ``server.pem`` certification files.

Or if using a Copr repository for Fedora:

```
sudo dnf copr enable yaacov/mohawk
sudo dnf install mohawk
```

## Mock Certifications

The server use mock certification to serve ``https`` requests. Users can use mock
self signed credentials files for this.

To create a self signed credentials use this bash commands:
```
openssl ecparam -genkey -name secp384r1 -out server.key
openssl req -new -x509 -sha256 -key server.key -out server.pem -days 3650
```

## Usage

When installed, run using the command line ``mohawk``, when run from code, use ``go run *.go``

The `-h` flag will print out a help text, that list the command line arguments.

```bash
# run `go run *.go` from the source path, or if installed use:
mohawk -h

Usage of mohawk:
  -api string
      the hawkulr api to mimic [e.g. 0.8.9.Testing, 0.21.2.Final] (default "0.21.0")
  -backend string
      the backend to use [random, sqlite, timeout] (default "sqlite")
  -cert string
      path to TLS cert file (default "server.pem")
  -key string
      path to TLS key file (default "server.key")
  -options string
      specific backend options [e.g. db-dirname (sqlite), max-size (random)]
  -port int
      server port (default 8080)
  -tls
      use TLS server
  -verbose
      more debug output
  -version
      version number
```

## Example of use

Running from the source path using ``go run`` and requesting the help message.

```bash
go run *.go -h
Usage of mohawk:
...
```
Running from system install using ``mohawk`` without ``tls`` and using the ``sqlite`` back end.

```bash
mohawk
2017/01/03 10:06:50 Start server, listen on http://0.0.0.0:8080
```

Running from system install using  ``mohawk`` and the ``random`` back end
[ Remmeber to set up the ``server.key`` and ``server.pem`` files in your path ].

```bash
mohawk -backend random -tls -port 8443
2016/12/01 14:23:48 Start server, listen on https://0.0.0.0:8443
```

## Examples

### Creating the server.pem and server.key files
```bash
openssl ecparam -genkey -name secp384r1 -out server.key
openssl req -new -x509 -sha256 -key server.key -out server.pem -days 3650
```

### Running the tls server on port 8443
```bash
mohawk -tls -port 8443
```

### Reading and writing data
```
# get server status
curl -ks  https://localhost:8443/hawkular/metrics/status

# get a list of all metrics
curl -ks https://localhost:8443/hawkular/metrics/metrics

# post some data (timestamp is in ms)
curl -ks -X POST https://localhost:8443/hawkular/metrics/gauges/raw -d "[{\"id\":\"machine/example.com/test\", \"data\":[{\"timestamp\": 1492434911769, \"value\": 42}]}]"

# read some data (variables can be start, end and bucketDuration)
curl -ks https://localhost:8443/hawkular/metrics/gauges/machine%2Fexample.com%2Ftest/raw?start=1492434911760

# set tags
curl -ks -X PUT https://localhost:8443/hawkular/metrics/gauges/machine%2Fexample.com%2Ftest/tags -d "{\"type\": \"node\", \"hostname\": \"example.com\"}"

# look for metrics by tag value (using a regexp)
curl -ks https://localhost:8443/hawkular/metrics/metrics?tags=hostname:.*\.com

# read multiple data points
curl -ks -X POST https://localhost:8443/hawkular/metrics/gauges/raw/query -d "{\"ids\": [\"machine/example.com/test\"], \"start\": 1492434811769, \"end\": 1492435911769}"

# read multiple data points with aggregation statistics
curl -ks -X POST https://localhost:8443/hawkular/metrics/gauges/raw/query -d "{\"ids\": [\"machine/example.com/test\"], \"start\": 1492434811769, \"end\": 1492435911769, \"bucketDuration\": \"1000s\"}"
```

### Data encoding, using gzip data encoding

```
# using the zcat utility to decode gzip message
curl -k -s -H "Accept-Encoding: gzip" https://localhost:8443/hawkular/metrics/metrics | zcat
```
