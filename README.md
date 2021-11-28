# Multisig

WARNING: this tool is very new, you probably shouldn't use it!

This is a tool for managing multisig txs with Cosmos-SDK based binaries and an
AWS S3 bucket.

## How it Works

Quick summary, with much more below:

- Configure an S3 bucket, some keys, and some chains in a TOML file.
- Create a directory in the bucket for each chain and key, like `/<chain name>/<key name>`
- `multisig generate` takes a generated unsigned tx file and pushes it to the s3 directory along with data needed for signing (eg. account number, sequence number, chain id)
- `multisig sign` fetches the unsigned tx and signing data for a given chain and key, signs it using the correct binary (eg. `gaiad tx sign unsigned.json ...`), and pushes the signature back to the directory
- `multisig list` lists the files in a directory so you can see who has signed
- `multisig broadcast` fetches all the data from a directory, compiles the signed tx (eg. `gaiad tx multisign unsigned.json ...`), broadcasts it using the configured node, and deletes all the files from the directory so signing can start fresh for a new tx

Everything generally tries to clean up after itself, but files are created and
removed from the present working directory, so you may want to be somewhere
clean.

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

`multisig` uses a simple `config.toml` file expected to be found in the present working directory. 
A documented example file is provided in `data/config.toml`.

You will need to:

- Configure your AWS Bucket
- Configure your Keys
- Configure your Chains

### Configure your AWS Bucket

You will need an s3 bucket and AWS credentials with read and write access to it.

You can setup an AWS IAM Policy that restricts access to a single bucket and attach
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

Each user can be given a shared or personal Access Key ID and Secret Access Key for accessing the bucket.

In the `[aws]` section of the multisig config, update the  `bucket`, `pub`, and `priv` fields 
with the bucket name, Access Key ID, and Secret Access Key.

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

### Configure your Chains

You can specify multiple chains. Each chain gets its own `[[chains]]` table.
It should include:

- a `name` for the chain
- the `binary` used to generate, sign, and broadcast txs
- the bech32 `prefix` for addresses
- the chain `id` for signing
- an optional `node` to interact with (for commands that can use/require nodes)

```
[[chains]]
name = "cosmos"                 # name of the chain
binary = "gaiad"                # name of binary
prefix = "cosmos"               # bech32 prefix
id = "cosmoshub-4"              # chain-id
node = "http://localhost:26657" # a synced node - only needed for `generate` and `broadcast` commands
```

## Run

Commands:

- Generate
- List
- Sign
- Broadcast

### Generate

Generate a `unsigned.json` tx as you normally would for the given multisig address with any of the chain binaries.

Then run

```
multisig generate --tx unsigned.json --sequence <sequence number for account> --account <account number> <chain name> <key name>
```

This will push the `unsigned.json` to the directory in the s3 bucket for the specified chain and key (ie. `/<chain name>/<key name>`). 
It will also push a file called `signdata.json` containing the account number, sequence number, and chain ID.

### List

To see the files in the directory of a chain and key:

```
multisig list <chain name> <key name> 
```

### Sign

To sign a tx:

```
multisig sign --from <local signing key> <chain name> <key name> 
```

Where `--from` is the name of the key in your local keystore, the same as you would provide to `--from` in `gaiad` or other Cosmos-SDK binaries.

### Broadcast

To assemble the signed tx and broadcast it, run:

```
multisig sign --node <node address> <chain name> <key name> 
```

Where the `--node` flag can be used to overwrite what's in the config file.

## TODO

High Priority

- test suite that spins up some local nodes and multisigs for testing
- move to cobra (whoops!). UX showstoppers in urfave:
    - flags have to come before args ?! see https://github.com/urfave/cli/issues/427
    - flags cant have short form aliases
- add `multisig raw` CRUD commands for direct file manipulation in the s3 bucket (eg. for debuging/fixing)
    - should also be able to create directories from the CLI
- `list` command shoud be able to show more detail, eg. which chain/key pairs exist
- add a command for porting a multisig from one binary's keystore to another
  (ie. decoding the bech32 for each key and running `keys add` on the new
  binary)
- `broadcast` should log txs somewhere once they're complete (maybe a log file
  in each top level chain directory?) - should include the key, tx id, and a short description of the tx

Lower Priority

- allow multiple txs to be started at a time per chain/key pair
- Use the https://github.com/cosmos/chain-registry for configuring chains instead of the
  config.toml ?
- other backends besides s3 ?
- other features to better manage multisigs and keystores across binaries ?!

