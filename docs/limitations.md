# Limitations

`cctest` is a unit-test helper, not a Fabric network simulator.

Out of scope:

- endorsement policy simulation
- MVCC conflict simulation
- validation phase behavior
- orderer, peer, channel, lifecycle, or CouchDB infrastructure
- exact CouchDB query planner behavior
- complete Fabric history database behavior; cctest records in-memory public state updates, but does not model peer history database configuration or block height
- complete private data collection semantics

The rich query implementation is intentionally conservative. It supports a documented selector subset and returns errors for unsupported operators instead of guessing.

The core package does not import cc-tools. cc-tools support lives in the nested `cctools` module so consumers of the core package do not pull cc-tools unless they opt into that module.
