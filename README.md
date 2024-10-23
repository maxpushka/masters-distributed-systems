## Requirement

1. Go
2. Python
3. Make
4. Apache Thrift

## Getting started

Run `make server` to start a server.
Run `make client` to start a client.

Note that server must start BEFORE client, otherwise you'll get `TransportException` from Thrift.
