# Changelog

## v0.3.5 (2019-10-25)

- fix bug with send empty variables to SyncRequest  

## v0.3.4 (2019-10-16)

- bug fix: probably panic with write to closed channel

## v0.3.3 (2019-10-08)

- field RemoteNodeOption.DialOpts change to slice instead single option
- add tests 

## v0.3.2 (2019-09-20)

- fix possible race condition while connect to remote node
- change some default values

## v0.3.1 (2019-09-11)

- fix version :)

## v0.2.11 (2019-09-11)

- bix bugs, add tests

## v0.2.10 (2019-09-10)

- pass grpc options to rplx.StartReplicationServer method
- add rplx.Stop method

## v0.2.9 (2019-08-27)

- remove option WithNodeMaxBufferSize - move to RemoteNoteOption
- refactoring

## v0.3.0 (2019-08-26)

- fix bug

## v0.2.8 (2019-08-26)

- make RemoteNodesProvider for check remote nodes list

## v0.2.7 (2019-08-23)

- if call node.sync while sync is active, place defer sync

## v0.2.6 (2019-08-22)

- API method 'Upsert' returns new variable value

## v0.2.5 (2019-08-21)

- send all variables to remote node after connect

## v0.2.4 (2019-08-19)

- add api method `All`. Returns maps with NotExpired and Expired variables values

## v0.2.3 (2019-08-15)

- fixes
- use UnixTime as variables version

## v0.2.2 (2019-08-14)

- refactoring add node

## v0.2.1 (2019-08-14)

- AddRemoteNode not returns error if fail send HelloRequest
- HelloRequest sending in infinity loop with interval

## v0.2.0 (2019-08-09)

- massive refactoring

## v0.1.0 (2019-07-23)

- minor release

## v0.0.4 (2019-07-22)

- WIP

## v0.0.3 (2019-07-22)

- add ttlStamp conception

## v0.0.2 (2019-07-22)

- add rplx.Delete method
- add same tests

## v0.0.1b (2019-07-19)

Initial version