# Starknet Gauntlet Commands for Chainlink OCR2 Protocol

## Set up

Make sure you have your account details set up in your `.env` file

```bash
# .env
PRIVATE_KEY=0x...
ACCOUNT=0x...
LINK=0x...
```

Note: The [token contract](https://github.com/smartcontractkit/chainlink-starknet/tree/develop/packages-ts/starknet-gauntlet-token) should only be deployed once and the same contract should be used for very aggregator


## Deploying Contracts via Class Hash

If you already know the class hash for a declared contract, all of the deploy commands in this module (ocr2, access_controller, proxy, Example) all have an optional --classHash flag which can be passed in to deploy a contract via the class hash rather than the local contract source code.

e.g
```bash
yarn gauntlet access_controller:deploy --classHash=<CLASS_HASH> --network=<NETWORK>
```

```bash
yarn gauntlet ocr2:deploy --network=<NETWORK> --billingAccessController=<ACCESS_CONTROLLER_CONTRACT> --minSubmissionValue=<MIN_VALUE> --maxSubmissionValue=<MAX_VALUE> --decimals=<DECIMALS> --name=<FEED_NAME> --link=<TOKEN_CONTRACT> --classHash=<CLASS_HASH>
```


## Deploy an Access Controller Contract

Run the following command:

```bash
yarn gauntlet access_controller:deploy --network=<NETWORK>
```

This command will generate a new Access Controller address and will give the details during the deployment. You can then add this contract address in your `.env` files as `BILLING_ACCESS_CONTROLLER`, or pass it into the command directly (as shown below)

## Deploy an OCR2 Contract

Run the following command substituting the following attributes:

1. <TOKEN_CONTRACT> with the [deployed token contract](https://github.com/smartcontractkit/chainlink-starknet/tree/develop/packages-ts/starknet-gauntlet-token)
2. <ACCESS_CONTROLLER_CONTRACT> with the access controller contract you deployed in the previous section

```bash
yarn gauntlet ocr2:deploy --network=<NETWORK> --billingAccessController=<ACCESS_CONTROLLER_CONTRACT> --minSubmissionValue=<MIN_VALUE> --maxSubmissionValue=<MAX_VALUE> --decimals=<DECIMALS> --name=<FEED_NAME> --link=<TOKEN_CONTRACT>
```

This command will generate a new OCR2 address and will give the details during the deployment

## Set the Billing Details on OCR2 Contract

Run the following command substituting <OCR_CONTRACT_ADDRESS> with the OCR2 contract address you received in the deploy step:

```
yarn gauntlet ocr2:set_billing --observationPaymentGjuels=<AMOUNT> --transmissionPaymentGjuels=<AMOUNT> --gasBase=<AMOUNT> --gasPerSignature=<AMOUNT> <CONTRACT_ADDRESS>
```

This Should set the billing details for this feed on contract address

## Set the Config Details on OCR2 Contract

Run the following command substituting <OCR_CONTRACT_ADDRESS> with the OCR2 contract address you received in the deploy step:

```
yarn gauntlet ocr2:set_config --network=<NETWORK> --address=<ADDRESS> --secret=<SECRET> --f=<NUMBER> --signers=[<ACCOUNTS>] --transmitters=[<ACCOUNTS>] --onchainConfig=<CONFIG> --offchainConfig=<CONFIG> --offchainConfigVersion=<NUMBER> <OCR_CONTRACT_ADDRESS>
```

Note: You must include the same random secret to deterministically run the set config multiple times (ex: for multisig proposals among different signers signing the same transaction). This can be achieved by setting a flag or environment variable ``--randomSecret`` or ``RANDOM_SECRET`` respectiviely. 

This Should set the config for this feed on contract address.


## Upgrading

All of the contracts (besides example) can be upgraded by supplying a classHash flag

e.g.

```bash
yarn gautnlet access_controller:upgrade --network=<NETWORK> --classHash=<CLASS_HASH> <CONTRACT_ADDRESS>`,
```

or

```bash
yarn gautnlet ocr2:upgrade --network=<NETWORK> --classHash=<CLASS_HASH> <CONTRACT_ADDRESS>`,
```
