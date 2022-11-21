# Multisig

---

Disclaimer: Use at your risk, responsibility for damages (if any) to anyone
resulting from the use of this software rest entirely with the user. The author
is not responsible for any damage that its use could cause.

Note: use a released version. The `main` branch is an active development branch.
The tool is still a bit rough around the edges, so its best to use from a clean directory and with a clean S3 bucket.

---

This is a tool for managing multisig txs with Cosmos-SDK based binaries and an
AWS S3 bucket.

See the [github issue #5661 on the
Cosmos-SDK](https://github.com/cosmos/cosmos-sdk/issues/5661) for discussion
about multisig handling. This tool is a multi-chain/multi-key solution to that problem using S3
buckets.

See the [TODO](#todo) list at the bottom of this document for work that needs to be done (please open an issue if you intend to work on something!) :)

## How it Works

Quick summary, with much more below:

- Configure an S3 bucket, some keys, and some chains in a TOML file.
- Create a directory in the bucket for each chain and key, like `/<chain name>/<key name>/`
- All signers have access to the entire s3 bucket, and can read/write at will, so assumption is they are all
  trusted
- `multisig tx push` takes an unsigned tx file and pushes it to the s3 directory along with data needed for signing (eg. account number, sequence number, chain id)
- `multisig tx vote` generate a vote tx and push it to s3 directory
- `multisig tx authz` generate an authz grant tx (delegate, withdraw, commission, vote) or revoke an authz authorization
- `multisig sign` fetches the unsigned tx and signing data for a given chain and key, signs it using the correct binary (eg. `gaiad tx sign unsigned.json ...`), and pushes the signature back to the directory
- `multisig list` lists the files in a directory so you can see who has signed
- `multisig broadcast` fetches all the data from a directory, compiles the signed tx (eg. `gaiad tx multisign unsigned.json ...`), broadcasts it using the configured node, and deletes all the files from the directory so signing can start fresh for a new tx
- `multisig delete` deletes txs from the S3 directory

Everything generally tries to clean up after itself, but files are created and
removed from the present working directory, so you may want to be somewhere
clean. You can also use the `multisig raw` commands to clean-up the s3 bucket individual files 
or the `multisig delete` to clean-up multiple files at once.

Note that s3 doesn't actually have directories, everything is just a file in the
bucket, but files can be prefixed with what looks like directory paths. So the
appearance of a "directory" is just an empty object with a name ending in a `/`.

## Install

```
git clone https://github.com/informalsystems/multisig
cd multisig
go install
```

Make sure your `$HOME/go/bin` or your `$GOPATH/bin` is on your `$PATH`.

Then 

```
multisig help
```

for the list of commands and options.

## Configure

`multisig` uses a simple `config.toml` file. 
Path to config file may be specified via `--config` flag. 
If the config path isn't specified explicitly, multisig will look for it in the current working directory.
If config file is not present in the current working directory, multisig will look for it in `~/.mulitisig/config.toml`.
A documented example file is provided in `data/config.toml`. Copy this example
file to your current directory or to `~/.multisig/` and modify it as necessary.

You will need to:

- Configure your AWS Bucket
- Configure your Keys
- Configure your Chains

### Configure your AWS Bucket

Each user will need an AWS Access Key ID and Secret Access Key that gives them
read/write access to the bucket.

In the `[aws]` section of the multisig config, each user must set the  `bucket`, 
`bucketregion`, `pub`, and `priv` fields 
with the bucket name, AWS region of the bucket, Access Key ID, and Secret Access Key.

```
# aws credentials

[aws]
bucket = "<bucketName>"         # s3 bucket name
bucketregion = "<bucketRegion>" # aws bucket region
pub = "<access key id>"         # Access Key ID
priv = "<secret access key>"    # Secret Access Key
```

If you are setting up the bucket for the first time, you can create an AWS IAM Policy that restricts access to a single bucket and attach
it to a User or Group:

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "ListObjectsInBucket",
            "Effect": "Allow",
            "Action": ["s3:ListBucket"],
            "Resource": ["arn:aws:s3:::bucket-name"]
        },
        {
            "Sid": "AllObjectActions",
            "Effect": "Allow",
            "Action": "s3:*Object",
            "Resource": ["arn:aws:s3:::bucket-name/*"]
        }
    ]
}
```

See [Source](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_examples_s3_rw-bucket.html).

### Configure you Keys

You can specify multiple keys. Each key gets its own `[[keys]]` table. 
Each key must have a `name` used by all signers. The `name` is a high level name for this key
that is independent of any chain or any name in a keystore.
The key must also specify its multisig `address`.
The address can be specified using any bech32 prefix, it doesn't matter if its
`cosmos1` or `osmo1` or whatever. The key itself is chain agnostic.

Each user may also specify a `localname` for the multisig key if they refer to it with a different name in their 
local keystore than the shared `name`. If `localname` is not specified, it defaults to `name`.

Note this means that each key is expected to have the same local name in each binary's keystore (ie.
if its called `mycorp-multisig` in `gaiad`, call it `mycorp-multisig` in `osmod`)!

As an example:

```
[[keys]]
name = "mycorp-main"            # name of this multisig key - same for everyone
address = "cosmos1..."          # bech32 address of the key - same for everyone
localname = "mycorp-multisig"   # name of this key in a signer's local keystore - can be different for everyone
```

### Configure gas and fee

If you specify default values for `gas` and `fee` in the configuration file those will be used instead of the hard-coded 
values coded in the tool.

For example, to set a default fee of `10000` and default gas of `300000` add this to the config file:
```
defaultFee = 10000
defaultGas = 300000
```

> Note: In the near future, multisig will support flags that you will be able to use to specify gas and fees.

### Configure your Chains

You can specify multiple chains. Each chain gets its own `[[chains]]` table.
It should include:

- a `name` for the chain
- the `binary` used to generate, sign, and broadcast txs
- the bech32 `prefix` for addresses
- the chain `id` for signing
- the `denom` for a particular chain (e.g. `uatom`)
- an optional `node` to interact with (for commands that can use/require nodes)

```
[[chains]]
name = "cosmos"                 # name of the chain
binary = "gaiad"                # name of binary
prefix = "cosmos"               # bech32 prefix
id = "cosmoshub-4"              # chain-id
denom = "uatom"                 # denom used to pay fees
node = "http://localhost:26657" # a synced node - only needed for `tx` and `broadcast` commands
```

## Run

Commands:

| Command                                                            | Command Line         |
|--------------------------------------------------------------------|----------------------|
| Broadcast a transaction to the blockchain                          | `multisig broadcast` |
| Delete transaction files from S3                                   | `multisig delete`    |
| Help information                                                   | `multisig help`      |
| List transaction files on S3                                       | `multisig list`      |
| Raw operations commands on S3 and utilities (e.g. convert address) | `multisig raw`       |
| Sign a transaction locally and upload the signature to S3          | `multisig sign`      |
| Create transaction files and upload to S3                          | `multisig tx`        |

## Tx

The multisig tx command allows you to push transactions to S3. You can push `unsigned.json` transactions that are manually generated
or you can use commands to generate the transactions and push it automatically to S3.

### tx push
```
multisig tx push <unsigned tx file> <chain name> <key name>
```

This will push the unsigned tx file (`e.g unsigned.json`) to the directory in the s3 bucket for the specified chain and key (ie. `/<chain name>/<key name>/0`). 

It will also fetch the account number and sequence number from the given `--node <node address>`,
and push a file to the bucket called `signdata.json` containing the account number, sequence number, and chain ID.
The sequence and account number can be overwriten or specified without a node
using the `--sequence` and `--account` flags

> Note: if you use `--node` its shelling out to the `<binary> query account <address>`
command and parsing the response. 

This assumes that the `<binary>` (e.g gaiad) is properly installed on the machine and accessible (can be executed from a command prompt e.g. `$> gaiad`) . The `<binary>` name is retrieved from the config.toml file.

To push multiple txs for the same chain and key, use the `--additional` flag.
Each additional tx will increment the path suffix and the sequence number. For
example, after pushing two txs for the same chain/key pair, you'd have:

```
cosmos/
cosmos/my-key/
cosmos/my-key/0/signdata.json
cosmos/my-key/0/unsigned.json
cosmos/
cosmos/my-key/
cosmos/my-key/1/signdata.json
cosmos/my-key/1/unsigned.json
```

To overwrite the first tx, use `--force`.

### tx vote

```
multisig tx vote <chain name> <key name> <proposal number> <vote option> [flags]
```

This will generate a tx for a governance proposal vote and it will push it to s3 directly. You will need to specify the proposal number and the vote (e.g. yes, no). 
You will also need to specify the denom for the fees (e.g. uatom) if it cannot be retrieved from the configuration file or the 
chain registry.

### tx withdraw

```
multisig tx withdraw <chain name> <key name>
```

This will generate a tx for withdraw all rewards for the account, and it will push it to s3 directly.
You will also need to specify the denom for the fees (e.g. uatom) if it cannot be retrieved from the configuration file or the
chain registry.

### tx claim-validator

```
multisig tx claim-validator <chain name> <key name> <validator_address>
```

This will generate a tx to claim the rewards and commission from a validator account, and it will push it to s3 directly.
You will also need to specify the denom for the fees (e.g. uatom) if it cannot be retrieved from the configuration file or the
chain registry.

### tx authz grant

```
multisig tx authz grant <chain name> <key name> <grantee address> <delegate|withdraw|commission|vote> <expiration in days>
```

This will generate a tx to grant authz permissions to a particular account (grantee). You will also need to specify the message-type that 
you want to grant permission. Currently, only withdraw, commission, delegate, vote are supported. You also need to specify the expiration for
this grant, for example to grant permissions for 30 days please specify '30' as the '<expiration> parameter.'
You will also need to specify the denom for the fees (e.g. uatom) if it cannot be retrieved from the configuration file or the
chain registry.

### tx authz revoke

```
multisig tx authz revoke <chain name> <key name> <grantee address> <withdraw|commission|delegate|vote> [flags]
```

This will generate a tx to revoke a previously granted authz permission

## List

To see the files in the directory of a chain and key:

```
multisig list <chain name> <key name> 
```

To list all the files in the bucket:

```
multisig list --all
```

Example output:k

```
$ multisig list --all
cosmos/
cosmos/mycorp-main/
cosmos/mycorp-validator/
juno/
juno/mycorp-main/
osmosis/
osmosis/mycorp-main/
osmosis/mycorp-main/0/eb.json
osmosis/mycorp-main/0/signdata.json
osmosis/mycorp-main/0/unsigned.json
```

This shows all the chain/key pairs that have been setup. All of them are empty
except `osmosis/mycorp-main` which has one signature (`eb.json`).

## Delete

To delete multiple files from S3 for a particular chain/key pair:

```
 multisig delete <chain name> <key name> [flags]
```

## Sign

To sign a tx:

```
multisig sign <chain name> <key name> --from <local signing key>  --index <tx
index>
```

Where `--from` is the name of the key in your local keystore, the same as you would provide to `--from` in `gaiad` or other Cosmos-SDK binaries, and `--index` is the tx index to sign for (default 0).

## Broadcast

To assemble the signed tx and broadcast it, run:

```
multisig broadcast <chain name> <key name> --index <tx index>
```

Where the `--index` is the tx index to sign for (default 0). Note txs must be
broadcast in order.

The `--node` flag can be used to overwrite what's in the config file.

## Raw

There are a set of `raw` subcommands for direct manipulation of bucket objects.
This is mostly for debugging purposes and generally should not need to be used.
See `multisig raw --help` and the help menu for each subcommand for more info.

## Running the tests

The tests are made in an integrational manner to test the things on the real chain.
To bootstrap the tests a local container is launched with an instance of gaia, minio, and multisig tool.
Ensure you have docker installed and running before launching the tests.
To run the tests use the following command in project's directory:

```
./run_tests.sh
```

The command will pull the necessary images and will build the gaia from scratch so the first run can be time-consuming.

## TODO

### High Priority

- add denoms to chains and have `tx push` validate txs are using correct denoms
- tx push should check fees and gas are high enough
- `broadcast` should log the tx once its complete (maybe a log file
  in each top level chain directory?) - should include the key, tx id, and the description 
- need a way to assign local key names (`--from`) to keys (possibly on a per-chain basis)
- use `--broadcast-mode block` ?
- new command to show unclaimed rewards for all addresses on all networks

### Mid Priority

- simulate tx to estimate gas
- add a command for porting a multisig from one binary's keystore to another
  (ie. decoding the bech32 for each key and running `keys add` on the new
  binary)
- proper error handling - sometimes we just print a message and return no error,
  but then the exit code is still 0
- make tx and query response parsing more robust (currently shelling out to CLI
  commands - should we be using the REST server ? maybe 26657 nodes are more
  available than rest ? )

### Lower Priority

- Use the https://github.com/cosmos/chain-registry for configuring chains instead of the
  config.toml ?
- other features to better manage multisigs and keystores across binaries ?!

