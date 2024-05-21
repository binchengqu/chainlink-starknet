import { makeProvider } from '@chainlink/starknet-gauntlet'
import deployCommand from '../../src/commands/ocr2/deploy'
import setBillingCommand from '../../src/commands/ocr2/setBilling'
import setConfigCommand from '../../src/commands/ocr2/setConfig'
import deployACCommand from '../../src/commands/accessController/deploy'
import {
  registerExecuteCommand,
  TIMEOUT,
  LOCAL_URL,
  StarknetAccount,
  fetchAccount,
} from '@chainlink/starknet-gauntlet/test/utils'
import { loadContract } from '@chainlink/starknet-gauntlet'
import { CONTRACT_LIST } from '../../src/lib/contracts'
import { Contract, InvokeTransactionReceiptResponse, RpcProvider } from 'starknet'

const signers = [
  'ocr2on_starknet_04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603730', // ocr2on_starknet_<key>
  'ocr2on_starknet_0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603731', // ocr2on_starknet_0x<key>
  '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603732', // 0x<key>
  '04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603733', // <key>
]

const transmitters = [
  '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603734',
  '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603735',
  '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603736',
  '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603737',
]

const validInput = {
  f: 1,
  signers,
  transmitters,
  onchainConfig: '',
  offchainConfig: {
    deltaProgressNanoseconds: 8000000000,
    deltaResendNanoseconds: 30000000000,
    deltaRoundNanoseconds: 3000000000,
    deltaGraceNanoseconds: 500000000,
    deltaStageNanoseconds: 20000000000,
    rMax: 5,
    s: [1, 2],
    offchainPublicKeys: [
      'af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852090',
      'af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852091',
      'af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852092',
      'af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852093',
    ],
    peerIds: signers,
    reportingPluginConfig: {
      alphaReportInfinite: false,
      alphaReportPpb: 0,
      alphaAcceptInfinite: false,
      alphaAcceptPpb: 0,
      deltaCNanoseconds: 0,
    },
    maxDurationQueryNanoseconds: 2000000000,
    maxDurationObservationNanoseconds: 1000000000,
    maxDurationReportNanoseconds: 200000000,
    maxDurationShouldAcceptFinalizedReportNanoseconds: 200000000,
    maxDurationShouldTransmitAcceptedReportNanoseconds: 200000000,
    configPublicKeys: [
      'af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852094',
      'af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852095',
      'af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852096',
      'af400004fa5d02cd5170b5261032e71f2847ead36159cf8dee68affc3c852097',
    ],
  },
  offchainConfigVersion: 2,
  secret: 'awe accuse polygon tonic depart acuity onyx inform bound gilbert expire',
}

describe('OCR2 Contract', () => {
  let account: StarknetAccount
  let contractAddress: string
  let accessController: string

  beforeAll(async () => {
    account = await fetchAccount()
  })

  it(
    'Deploy AC',
    async () => {
      // TODO: owner can't be 0 anymore
      const command = await registerExecuteCommand(deployACCommand).create({}, [])

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')
      accessController = report.responses[0].contract
    },
    TIMEOUT,
  )

  it(
    'Deployment',
    async () => {
      const command = await registerExecuteCommand(deployCommand).create(
        {
          input: {
            owner: account.address,
            maxAnswer: 10000,
            minAnswer: 1,
            decimals: 18,
            description: 'Test Feed',
            billingAccessController: accessController,
            linkToken: '0x04cc1bfa99e282e434aef2815ca17337a923cd2c61cf0c7de5b326d7a8603730',
          },
        },
        [],
      )

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      contractAddress = report.responses[0].contract
    },
    TIMEOUT,
  )

  it(
    'Set billing using --batch',
    async () => {
      // the number of calls to batch into one transaction
      const batchCount = 3

      // transfer overflow on set billing
      const command = await registerExecuteCommand(setBillingCommand).create(
        {
          input: {
            observationPaymentGjuels: 1,
            transmissionPaymentGjuels: 1,
            gasBase: 14951,
            gasPerSignature: 13,
          },
          batch: true,
        },
        // Calls set_billing on the same contract `batchCount` times (with the
        // same input) using one transaction. We use the same contract address
        // here for simplicity, but a more realistic example would involve using
        // different contract addresses (where each of them has the same ABI)
        // and passing different inputs to each of them.
        Array.from({ length: batchCount }).map(() => contractAddress),
      )

      const report = await command.execute()
      const maybeRes = report.responses.at(0)
      expect(maybeRes).not.toBeFalsy()

      const res = maybeRes!
      expect(res.tx.status).toEqual('ACCEPTED')
      expect(res.tx.hash).not.toBeNull()

      // Checks that the update was successful
      const provider: RpcProvider = makeProvider(LOCAL_URL).provider
      const { contract } = loadContract(CONTRACT_LIST.OCR2)
      const ocr2Contract = new Contract(contract.abi, contractAddress, provider)
      const billing = await ocr2Contract.billing()
      expect(billing.observation_payment_gjuels).toEqual(BigInt(1))
      expect(billing.transmission_payment_gjuels).toEqual(BigInt(1))

      // Checks that the correct number of batch calls were made
      const receipt = await provider.waitForTransaction(res.tx.hash)
      expect(receipt.isSuccess()).toBeTruthy()
      if (receipt.isSuccess()) {
        const invocations = receipt.events.filter((ev) => ev.from_address === contractAddress)
        expect(invocations.length).toEqual(batchCount)
      }
    },
    TIMEOUT,
  )

  it(
    'Set config using --input and --batch',
    async () => {
      // the number of calls to batch into one transaction
      const batchCount = 3

      const command = await registerExecuteCommand(setConfigCommand).create(
        {
          input: validInput,
          batch: true,
        },
        // Calls set_config on the same contract `batchCount` times (with the
        // same input) using one transaction. We use the same contract address
        // here for simplicity, but a more realistic example would involve using
        // different contract addresses (where each of them has the same ABI)
        // and passing different inputs to each of them.
        Array.from({ length: batchCount }).map(() => contractAddress),
      )

      const report = await command.execute()
      const maybeRes = report.responses.at(0)
      expect(maybeRes).not.toBeFalsy()

      const res = maybeRes!
      expect(res.tx.status).toEqual('ACCEPTED')
      expect(res.tx.hash).not.toBeNull()

      const provider: RpcProvider = makeProvider(LOCAL_URL).provider
      const { contract } = loadContract(CONTRACT_LIST.OCR2)
      const ocr2Contract = new Contract(contract.abi, contractAddress, provider)
      const resultTransmitters = await ocr2Contract.transmitters()

      // retrieve signer keys from transaction event
      // based on event struct: https://github.com/smartcontractkit/chainlink-starknet/blob/develop/contracts/src/chainlink/ocr2/aggregator.cairo#L260
      const receipt = (await provider.getTransactionReceipt(
        res.tx.hash,
      )) as InvokeTransactionReceiptResponse

      // Checks that the correct number of batch calls were made
      const invocations = receipt.events.filter((ev) => ev.from_address === contractAddress)
      expect(invocations.length).toEqual(batchCount)

      // TODO: use StarknetContract decodeEvents from starknet-hardhat-plugin instead
      const eventData = receipt.events[0].data
      // reconstruct signers array from event
      const eventSigners: bigint[] = []
      for (let i = 0; i < signers.length; i++) {
        const signer = BigInt(eventData[2 + 2 * i]) // split according to event structure
        eventSigners.push(signer)
      }

      expect(eventSigners).toEqual(
        // eaiser to remove prefix and 0x and then add 0x back
        signers.map((s) => BigInt(`0x${s.replace('ocr2on_starknet_', '').replace('0x', '')}`)),
      ) // remove all prefixes
      expect(resultTransmitters).toEqual(transmitters.map((transmitter) => BigInt(transmitter)))
    },
    TIMEOUT,
  )
})
