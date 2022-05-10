# CHANGELOG

# v0.2.1 (May 10, 2022)

Thi minor release includes some minor bug fixes identified after the previous release

- Fix the `raw cat` command ([#17](https://github.com/informalsystems/multisig/issues/17))
- Fix the `multisig list --all` command ([#23](https://github.com/informalsystems/multisig/issues/23))
- Avoid panic when executing a `tx push` command ([#32](https://github.com/informalsystems/multisig/issues/32)) ([#26](https://github.com/informalsystems/multisig/issues/26))
- Better handling of messages and warnings ([#19](https://github.com/informalsystems/multisig/issues/19)) ([#28](https://github.com/informalsystems/multisig/issues/28))

# v0.2.0 (May 3, 2022)

- Allow multiple txs to be submited for a single (chain, key) pair using the
  `--additional` or `-x` flag. Paths now include a "txIndex" value, starting from 0.
The txIndex is automatically added to the sequence number fetched from the node.
This allows multiple sequential txs to be pushed and signed asynchronously.
`sign` and `broadcast` commands now take a `--index` or `-i` value to specify
the tx to sign or broadcast. Txs must be broadcast in order but can be signed in
any order.
- Don't overwrite a tx without specifying `--force` or `-f`.
- Removed `generate` command, the new `tx push` command has the same effect but a
  better name.
- Added `tx vote` and `tx withdraw` commands that generate the unsigned transaction.
- Added `--denom` flag for fees. Denom can also be set in the configuration,
  under `[[chains]]`. For example `denom=uatom`. If denom is not set anywhere,
  it will be queried from the chain registry.
- Added `tx authz grant` command to create transactions for granting authz persmissions.

# v0.1.0 (Mar 8, 2022)

Basic functionality for managing multisig txs with an AWS bucket!
Only allows one tx per (chain, key) pair at a time.
