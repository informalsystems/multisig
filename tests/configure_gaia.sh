#!/bin/bash

gaiad init test_key_1 --chain-id testhub

# cosmos14eadktsf4zzah6har7h7a46tunnj7rq7lmppy5
printf 'luggage rotate orient usage program cloud armed warrior rich erase acquire remember\n' \
  | gaiad keys add test_key_1 --recover --keyring-backend test

# cosmos18eq7nwggdmpxr0d4a23rmdued6jwz5mzwn8vjz
printf 'example ridge wine client logic throw shoulder unknown current artefact donor vague phrase cat final rocket reason leg dizzy insect version cement entry glow\n' \
  | gaiad keys add test_key_2 --recover --keyring-backend test

# cosmos16cfaj6ayxmng766wqkkfyuays796xnd5zdm2re
printf 'clerk picnic pill often soon trigger situate tent amazing private wreck voice shrug scene wrist trash critic silent better cannon carpet raw company cruise\n' \
  | gaiad keys add test_key_3 --recover --keyring-backend test

# cosmos1eeq9uqcrc23lrh9stmagycyvzwtrg0a2sw7kk9
printf 'snack invest february abstract bullet stock repeat clever fiction steel runway donate near catch unique must code rebuild velvet peanut display faculty make flash\n' \
  | gaiad keys add test_key_4 --recover --keyring-backend test

# cosmos1mmq6knly90q9pqfx36srl3gnls3ahdm7s93g2q
gaiad keys add \
  --multisig=test_key_1,test_key_2,test_key_3 \
  --multisig-threshold=2 \
  --keyring-backend test \
  multisig_test_2_of_3

# cosmos19mxmee2hyl23kpa39al2d4ztzzyuh3mletg4vk
gaiad keys add \
  --multisig=test_key_1,test_key_2,test_key_3,test_key_4 \
  --multisig-threshold=3 \
  --keyring-backend test \
  multisig_test_3_of_4

gaiad add-genesis-account cosmos14eadktsf4zzah6har7h7a46tunnj7rq7lmppy5 10000000000stake,1000000000000uatom
gaiad add-genesis-account cosmos18eq7nwggdmpxr0d4a23rmdued6jwz5mzwn8vjz 1000000000000uatom
gaiad add-genesis-account cosmos16cfaj6ayxmng766wqkkfyuays796xnd5zdm2re 1000000000000uatom
gaiad add-genesis-account cosmos1mmq6knly90q9pqfx36srl3gnls3ahdm7s93g2q 1000000000000uatom
gaiad add-genesis-account cosmos19mxmee2hyl23kpa39al2d4ztzzyuh3mletg4vk 1000000000000uatom

gaiad gentx test_key_1 10000000000stake --chain-id testhub --keyring-backend test
gaiad collect-gentxs
