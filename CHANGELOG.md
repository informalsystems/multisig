# CHANGELOG

## v0.4.0
*December 20th, 2022*

### Highlights

This major release implements a couple new transaction commands and adds additional flags to improve the flexibility when adding new transactions such as ability to specify gas and fees.

#### FEATURES

- Changed the way `EthAccount` account types are checked  ([#63](https://github.com/informalsystems/multisig/pull/63))
- Added a new command to generate a withdraw-rewards transaction to claim validators' rewards and commission ([#65](https://github.com/informalsystems/multisig/pull/65))
- Added a new command to create a delegate transaction ([#66](https://github.com/informalsystems/multisig/pull/66))
- Allow `fees` and `gas` flags to be specified when creating a new transaction ([#70](https://github.com/informalsystems/multisig/pull/70))
- Added support for '/stride.vesting.StridePeriodicVestingAccount' ([#73](https://github.com/informalsystems/multisig/pull/73))
- Upgraded Cosmos SDK version (v0.45.9) and Golang version (v1.18) ([#80](https://github.com/informalsystems/multisig/pull/80))
- Implemented logic to add a --node flag to the multisig broadcast command ([#82](https://github.com/informalsystems/multisig/pull/82))
- Added logic to allow a user to specify a --key flag to the multisig broadcast command ([#83](https://github.com/informalsystems/multisig/pull/83))

#### BUG FIXES

- Refactored the logic to delete files from S3. Now when pushing (--force) tx files, all existing files in the path will be deleted first ([#74](https://github.com/informalsystems/multisig/pull/74))

## v0.3.0
*September 8th, 2022*
### Highlights

This major release has a few improvements such as new commands, a new config file handling, ability to use an S3 provider other than AWS (e.g. [minio](https://min.io/)), an end-to-end basic test suite that leverages containers and several important bug fixes that were identified during regular use of the tool.

#### FEATURES

- New configuration file handling ([#51](https://github.com/informalsystems/multisig/pull/51))
  - By default, the config file `$HOME/.multisig/config.toml` is used
  - Config file now may be customized via `--config` flag
  - If the file is not specified with the `--config` flag or the file is in the `$HOME/.multisig` folder, then it will look for it in the current folder 
- AWS S3 address can be modified to support S3-compatible self-hosted solutions ([#41](https://github.com/informalsystems/multisig/pull/41))
- Basic end-to-end test suite ([#40](https://github.com/informalsystems/multisig/issues/40))
- New `multisig delete` command that allows the deletion of files in the S3 bucket for a specific chain and key name ([#29](https://github.com/informalsystems/multisig/issues/29))
- Added a new `--description` flag to `multisig tx` commands to allow additional metadata to be added to the `signdata.json` file, e.g a short description about the tx ([#37](https://github.com/informalsystems/multisig/issues/37))
- Implemented logic to use the default values for gas and fees from the configuration file. This replaces the old way where these values were only hard-coded. If values are not specified in the config the hard-coded values will still be used ([#44](https://github.com/informalsystems/multisig/issues/44))
- New `tx authz revoke` command that allows `multisig` to submit a tx to revoke an authz permission previously granted for a particular message type (e.g. vote) ([#18](https://github.com/informalsystems/multisig/issues/18))

#### BUG FIXES

-  Fixed issue that prevented a tx withdraw to be pushed to S3 ([#34](https://github.com/informalsystems/multisig/issues/34))
- Handle different account types that were returned when querying an account to find the account number and the sequence number. This bug prevented signing txs in Regen and Evmos for examples
  - Support for different account types such as `BaseAccount` and `PeriodicVestingAccount` ([#21](https://github.com/informalsystems/multisig/issues/21))
  - Support for the `/ethermint.types.v1.EthAccount` account type (`Evmos` use it) ([#54](https://github.com/informalsystems/multisig/issues/54))
  - Support for the `/cosmos.vesting.v1beta1.ContinuousVestingAccount` account type ([#47](https://github.com/informalsystems/multisig/issues/47))

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
