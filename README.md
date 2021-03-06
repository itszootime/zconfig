# zconfig

[![Build Status](https://travis-ci.org/itszootime/zconfig.svg?branch=master)](https://travis-ci.org/itszootime/zconfig)

A daemon for maintaining a locally-cached copy of a configuration stored in ZooKeeper.

## Overview

[ZooKeeper](http://zookeeper.apache.org/) (ZK) enables distributed and reliable storage of configuration values as a series of nodes. Typically, each component requiring access to these values will need to communicate with ZK. This has a number of disadvantages:

* A client library may not be available for the language of choice.
* Even if client libraries are available, handling ZK connection states can be complex.
* If every component process connects to ZK, there is a potential for a flooding of watches and mass (re)connects causing knock-on failures.

[ZConfig](https://github.com/itszootime/zconfig) is a daemon for maintaining a locally-cached copy of a configuration stored in ZK. Clients are able to read from this locally-cached copy without concern for ZK connection states and watches. To keep in sync with ZK, clients can watch OS-level file events. This approach can alleviate scalability issues, as only one set of ZK watches is required per daemon/machine, as opposed to `x` sets for `x` component processes.

## Build

1. Clone repository.

1. Install or update [godep](https://github.com/tools/godep) and fetch dependencies:

  ```
  $ go get -u github.com/tools/godep
  $ godep restore
  ```

1. Install:

  ```
  $ go install
  ```

## Usage

`zconfig` can be started with the following flags:

Flag          | Purpose
--------------|----------
`-zk`        | ZK connection string (default `localhost:2181`)
`-base-path` | Path where the locally-cached configuration will be stored (default `.`)
`-zk-root`   | ZK path to the configuration (default `/zconfig`)

### Behaviour

Once running, the daemon will recursively setup watches for children (and their values) of the root. During initialisation, and when changes are detected, it'll fetch children and values to build a configuration, which is then serialized as a series of YAML files stored in the base path.

For example, consider a series of nodes created like so:

```
$ ./zookeeper/bin/zkCli.sh
[zk] create /zconfig/servers
[zk] create /zconfig/servers/db
[zk] create /zconfig/servers/db/192.168.0.1
[zk] create /zconfig/servers/db/192.168.0.2
[zk] create /zconfig/settings
[zk] create /zconfig/settings/timeout 1000
```

The locally-cached configuration would be stored as:

```yaml
# servers.yml
db:
- 192.168.0.1
- 192.168.0.2
```

```yaml
# settings.yml
timeout: "1000"
```

The daemon will retrieve all values stored in ZK as strings. Empty values are converted to null.

Additional logic is required to determine if the value for a key should contain an array or a map. Only if **all** children of a node have empty values, it'll be an array. Consider [the previous example](#usage), but now we want to clear the timeout value:

```
$ ./zookeeper/bin/zkCli.sh
[zk] set /zconfig/settings/timeout ""
```

The stored configuration will be modified like so:

```yaml
# settings.yml
- timeout
```

As you can see, the timeout node is no longer treated as a key-value pair, but as a value in an array. For this reason, clients should return null for **any** key that doesn't exist in the locally-cached files.

## Q&A

**Why not just store a YAML/JSON serialized configuration directly in ZK?**
This complicates matters when storing the configuration in ZK, and means we can't take advantage of ephemeral nodes for service discovery.

**Won't this create lots of ZK watches?**
Yes, it can do. Due to the fact that nesting is allowed within the configuration, a single node requires watches for both the children and the value.

**Is it production ready?**
I'd currently consider this an alpha release, so unfortunately not at the moment. You're always welcome to try.

## Acknowledgements

Thanks to the [engineers at Pinterest](https://engineering.pinterest.com/): their article [ZooKeeper Resilience at Pinterest](https://engineering.pinterest.com/blog/zookeeper-resilience-pinterest) provided much of the inspiration for this approach. During the development of ZConfig, their [KingPin](https://github.com/pinterest/kingpin) toolset was open-sourced, which includes the 'ZK Update Monitor'.