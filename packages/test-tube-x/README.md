# Test Tube X

[`test-tube`](https://github.com/osmosis-labs/test-tube) is a generic library by Osmosis for building testing environments for [CosmWasm](https://cosmwasm.com/) smart contracts. It allows you to test your smart contract logic against the actual Cosmos SDK chain's logic, which is written in Go, using Rust. This eliminates the need to write Go code or learn Go in order to test your smart contracts against the Cosmos SDK.

`test-tube-x` is a fork of test-tube by Persistence made using newer versions of Cosmos SDK (v47). It can be used by anyone to create a testing environments by anyone using the following defaults:
- Cosmos SDK v47
- Persistence SDK v2
- Cosmwasm v1.4


`test-tube-x` is currently used to build [`persistence-test-tube`](https://github.com/persistenceOne/test-tube/tree/main/packages/persistence-test-tube). 