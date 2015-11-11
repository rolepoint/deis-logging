package main

import (
    _ "github.com/gliderlabs/logspout/adapters/raw"
    _ "github.com/gliderlabs/logspout/adapters/syslog"
    _ "github.com/gliderlabs/logspout/transports/tcp"
    _ "github.com/gliderlabs/logspout/transports/udp"
    _ "github.com/gliderlabs/logspout/transports/tls"
    _ "github.com/rolepoint/deis-logging/logspout-etcd/etcsrc"
)
