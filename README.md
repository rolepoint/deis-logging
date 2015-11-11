This repository contains some docker images for logging with deis.  It's
packages are designed to completely replace the docker-logspout and
docker-logger components, when running a stateless cluster and using an
external log-hosting service (such as logentries).

It was designed to work with logentries, but there's no reason it couldn't be
adapted to work with other log services.

It was partially inspired by
[rsyslog-deis](https://github.com/everydayhero/rsyslog-deis) but with some
extra functionality.

### Deis-syslog-ng

deis-syslog-ng is designed as a replacement for deis-logger.  It runs a
syslog-ng instance that gets most of it's configuration from etcd.  It uses
confd to monitor etcd for changes and re-generate the config.

It will look for keys under `/deis/log/syslog-ng` in etcd.  Each key in this
directory should contain some JSON with a few keys:

* `url` should contain a URL of the form "hostname:port" that logs should be
  forwarded to.
* `filter` should contain a valid syslog-ng filter expression.  Only logs that
  match this filter will be forwarded to `url`.
* `template` should contain the syslog-ng template to use for formatting log
  messages.  This should include any tokens required by your log service.

For example, to setup a log with etcdctl:

    etcdctl set /deis/log/syslog-ng/my-app '{"url": "data.logentries.com:514", "filter": "program(my-app)", "template": "123456 $MSG"}''

This should forward any messages coming from the docker container named
"my-app" to logentries, with the token "123456" followed by the actual message
contents itself.

### logspout-etcd

logspout-etcd is a custom-build of logspout, that will fetch the host & port of
deis-syslog-ng from etcd, and use that to determine where it should send it's
logs.  This allows it to act like a drop-in replacement for deis-logspout,
provided it is paired with deis-syslog-ng (or some other deis-logger
replacement that expects syslog format logs).

The main reason for this replacement is to get rid of the hardcoding of the
log-format that deis-logspout uses.  Using the standard syslog format makes it
much easier to write custom filtering etc. in deis-syslog-ng rather than trying
to parse the custom log format of deis-logspout.
