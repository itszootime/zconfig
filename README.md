# zconfig-daemon

ZConfig is one part daemon, one-to-many parts client library.

ZooKeeper allows us to store configuration values in a distributed manner. 

Accessing and updating these values at runtime isn't always so straightforward, and can be prone to catastrophic knock-on failures (ref airbnb?).

The idea behind ZConfig is to have a single daemon per machine. This daemon is responsible for maintaining a locally-stored mirror of the configuration stored in ZooKeeper.

It's then possible for clients to read from this locally-stored mirror, without having to worry about ZooKeeper connection states and watches. If the clients are interesting in keeping in sync with ZooKeeper, they can setup operating system level file watches. To make things a bit easier, [libraries are available](#).

This enables ZooKeeper configurations to be accessed within any language, without requiring a ZooKeeper client library.

## Benefits

* Requires only one set of ZooKeeper watches per machine, instead of x sets for x worker processes.

## Setup

Some stuff here to setup...

## How to store in ZK?

/zconfig (level 0 -> ChildrenW)

/zconfig/servers (level 1 -> ChildrenW)
/zconfig/servers/db (level 2 -> GetW, ChildrenW)
/zconfig/servers/db/127.0.0.1

/zconfig/search/timeout (level 2 -> GetW, ChildrenW)

Without children, values -> get
With children, values -> children

* Benefit is that services could register ephermeral nodes themselves
but that's a lot of watches.
* Another option would be to increment a version number in /zconfig (ZK does this already?).

/zconfig/servers # returns yaml file

Downside is this is now zconfig specific, changes the whole node when only a single value is modified.
