# CHANGELOG

# unreleased

- Allow multiple txs to be submited for a single (chain, key) pair using the
  `--additional` or `-x` flag. Paths now include a "txIndex" value, starting from 0.
The txIndex is automatically added to the sequence number fetched from the node.
This allows multiple sequential txs to be pushed and signed asynchronously.
`sign` and `broadcast` commands now take a `--index` or `-i` value to specify
the tx to sign or broadcast. Txs must be broadcast in order but can be signed in
any order.
- Don't overwrite a tx without specifying `--force` or `-f`.


# v0.1.0 (Mar 8, 2022)

Basic functionality for managing multisig txs with an AWS bucket!
Only allows one tx per (chain, key) pair at a time.
