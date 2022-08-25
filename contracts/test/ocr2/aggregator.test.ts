import { assert, expect } from 'chai'
import BN from 'bn.js'
import { starknet } from 'hardhat'
import { constants, ec, encode, hash, number, uint256, stark, KeyPair } from 'starknet'
import { BigNumberish } from 'starknet/utils/number'
import { Account, StarknetContract, StarknetContractFactory } from 'hardhat/types/runtime'
import { TIMEOUT } from '../constants'

interface Oracle {
  signer: KeyPair
  transmitter: Account
}

// Required to convert negative values into [0, PRIME) range
function toFelt(int: number | BigNumberish): BigNumberish {
  let prime = number.toBN(encode.addHexPrefix(constants.FIELD_PRIME))
  return number.toBN(int).umod(prime)
}

const CHUNK_SIZE = 31

function encodeBytes(data: Uint8Array): BN[] {
  let felts = []

  // prefix with len
  let len = data.byteLength
  felts.push(number.toBN(len))

  // chunk every 31 bytes
  for (let i = 0; i < data.length; i += CHUNK_SIZE) {
    const chunk = data.slice(i, i + CHUNK_SIZE)
    // cast to int
    felts.push(new BN(chunk, 'be'))
  }
  return felts
}

function decodeBytes(felts: BN[]): Uint8Array {
  let data = []

  // TODO: validate len > 1

  // TODO: validate it fits into 54 bits
  let length = felts.shift()?.toNumber()!

  for (const felt of felts) {
    let chunk = felt.toArray('be', Math.min(CHUNK_SIZE, length))
    data.push(...chunk)

    length -= chunk.length
  }

  return new Uint8Array(data)
}

describe('aggregator.cairo', function () {
  this.timeout(TIMEOUT)

  let aggregatorFactory: StarknetContractFactory

  let owner: Account
  let token: StarknetContract
  let aggregator: StarknetContract

  let minAnswer = -10
  let maxAnswer = 1000000000

  let f = 1
  let n = 3 * f + 1
  let oracles: Oracle[] = []
  let config_digest: number

  before(async function () {
    // assumes contract.cairo and events.cairo has been compiled
    aggregatorFactory = await starknet.getContractFactory('ocr2/aggregator')

    // can also be declared as
    // account = (await starknet.deployAccount("OpenZeppelin")) as OpenZeppelinAccount
    // if imported from hardhat/types/runtime"
    owner = await starknet.deployAccount('OpenZeppelin')

    const tokenFactory = await starknet.getContractFactory('link_token')
    token = await tokenFactory.deploy({ owner: owner.starknetContract.address })

    await owner.invoke(token, 'permissionedMint', {
      recipient: owner.starknetContract.address,
      amount: uint256.bnToUint256(100_000_000_000),
    })

    aggregator = await aggregatorFactory.deploy({
      owner: BigInt(owner.starknetContract.address),
      link: BigInt(token.address),
      min_answer: toFelt(minAnswer),
      max_answer: toFelt(maxAnswer),
      billing_access_controller: 0, // TODO: billing AC
      decimals: 8,
      description: starknet.shortStringToBigInt('FOO/BAR'),
    })

    console.log(`contract_address: ${aggregator.address}`)

    let futures = []
    let generateOracle = async () => {
      let transmitter = await starknet.deployAccount('OpenZeppelin')
      return {
        signer: ec.genKeyPair(),
        transmitter,
        // payee
      }
    }
    for (let i = 0; i < n; i++) {
      futures.push(generateOracle())
    }
    oracles = await Promise.all(futures)

    let onchain_config: number[] = []
    let offchain_config_version = 2 // TODO: assert == 2 in contract
    let offchain_config = new Uint8Array([1])

    console.log(encodeBytes(offchain_config))

    let config = {
      oracles: oracles.map((oracle) => {
        return {
          signer: number.toBN(ec.getStarkKey(oracle.signer)),
          transmitter: oracle.transmitter.starknetContract.address,
        }
      }),
      f,
      onchain_config,
      offchain_config_version,
      offchain_config: encodeBytes(offchain_config),
    }

    console.log(config)

    await owner.invoke(aggregator, 'set_config', config)
    // const receipt = await starknet.getTransactionReceipt(txHash);
    // const decodedEvents = await aggregator.decodeEvents(receipt.events);
    // console.log(decodedEvents)

    let result = await aggregator.call('latest_config_details')
    config_digest = result.config_digest

    // Immitate the fetch done by relay to confirm latest_config_details_works
    let block = await starknet.getBlock({ blockNumber: result.block_number })
    let events = block.transaction_receipts[0].events
    const decodedEvents = await aggregator.decodeEvents(events)
    assert.equal(decodedEvents.length, 1)
    let event = decodedEvents[0]
    assert.equal(event.name, 'config_set')

    console.log(`config_digest: ${config_digest.toString(16)}`)
  })

  let transmit = async (epoch_and_round: number, answer: BigNumberish): Promise<any> => {
    let extra_hash = 1
    let observation_timestamp = 1
    let juels_per_fee_coin = 1

    let observers_buf = Buffer.alloc(31)
    let observations = []

    for (const [index, _oracle] of oracles.entries()) {
      observers_buf[index] = index
      observations.push(answer)
    }

    // convert to a single value that will be decoded by toBN
    let observers = `0x${observers_buf.toString('hex')}`

    let msg = hash.computeHashOnElements([
      // report_context
      config_digest,
      epoch_and_round,
      extra_hash,
      // raw_report
      observation_timestamp,
      observers,
      observations.length,
      ...observations,
      juels_per_fee_coin,
    ])

    let signatures = []

    console.log([
      // report_context
      `0x${config_digest.toString(16)}`,
      epoch_and_round,
      extra_hash,
      // raw_report
      observation_timestamp,
      observers,
      observations.length,
      ...observations,
      juels_per_fee_coin,
    ])
    console.log(msg)
    for (let oracle of oracles.slice(0, f + 1)) {
      let [r, s] = ec.sign(oracle.signer, msg)
      console.log(
        `privKey ${oracle.signer.getPrivate()} r ${r} s ${s} pubKey ${number.toBN(
          ec.getStarkKey(oracle.signer),
        )}`,
      )
      signatures.push({
        r,
        s,
        public_key: number.toBN(ec.getStarkKey(oracle.signer)),
      })
    }
    console.log('---')

    let transmitter = oracles[0].transmitter

    return await transmitter.invoke(aggregator, 'transmit', {
      report_context: {
        config_digest,
        epoch_and_round,
        extra_hash,
      },
      observation_timestamp,
      observers,
      observations,
      juels_per_fee_coin,
      signatures,
    })
  }

  it('billing config', async () => {
    await owner.invoke(aggregator, 'set_billing', {
      config: {
        observation_payment_gjuels: 1,
        transmission_payment_gjuels: 5,
      },
    })
  })

  it('transmission', async () => {
    await transmit(1, 99)
    let { round } = await aggregator.call('latest_round_data')
    assert.equal(round.round_id, 1)
    assert.equal(round.answer, 99)

    await transmit(2, toFelt(-10))
    ;({ round } = await aggregator.call('latest_round_data'))
    assert.equal(round.round_id, 2)
    assert.equal(round.answer, -10)

    try {
      await transmit(3, -100)
      expect.fail()
    } catch (err: any) {
      // Round should be unchanged
      let { round: new_round } = await aggregator.call('latest_round_data')
      assert.deepEqual(round, new_round)
    }
  })

  it('payee management', async () => {
    let payees = oracles.map((oracle) => ({
      transmitter: oracle.transmitter.starknetContract.address,
      payee: oracle.transmitter.starknetContract.address, // reusing transmitter acocunts as payees for simplicity
    }))
    // call set_payees, should succeed because all payees are zero
    await owner.invoke(aggregator, 'set_payees', { payees })
    // call set_payees, should succeed because values are unchanged
    await owner.invoke(aggregator, 'set_payees', { payees })

    let oracle = oracles[0].transmitter
    let transmitter = oracle.starknetContract.address
    let payee = transmitter

    let proposed_oracle = oracles[1].transmitter
    let proposed_transmitter = proposed_oracle.starknetContract.address
    let proposed_payee = proposed_transmitter

    // can't transfer to self
    try {
      await oracle.invoke(aggregator, 'transfer_payeeship', {
        transmitter,
        proposed: payee,
      })
      expect.fail()
    } catch (err: any) {
      // TODO: expect(err.message).to.contain("");
    }

    // only payee can transfer
    try {
      await proposed_oracle.invoke(aggregator, 'transfer_payeeship', {
        transmitter,
        proposed: proposed_payee,
      })
      expect.fail()
    } catch (err: any) {}

    // successful transfer
    await oracle.invoke(aggregator, 'transfer_payeeship', {
      transmitter,
      proposed: proposed_payee,
    })

    // only proposed payee can accept
    try {
      await oracle.invoke(aggregator, 'accept_payeeship', { transmitter })
      expect.fail()
    } catch (err: any) {}

    // successful accept
    await proposed_oracle.invoke(aggregator, 'accept_payeeship', {
      transmitter,
    })
  })

  it('payments and withdrawals', async () => {
    let oracle = oracles[0]
    // NOTE: previous test changed oracle0's payee to oracle1
    let payee = oracles[1].transmitter

    let { amount: owed } = await payee.call(aggregator, 'owed_payment', {
      transmitter: oracle.transmitter.starknetContract.address,
    })
    // several rounds happened so we are owed payment
    assert.ok(owed > 0)

    // no funds on contract, so no LINK available for payment
    let { available } = await aggregator.call('link_available_for_payment')
    assert.ok(available < 0) // should be negative: we owe payments

    // deposit LINK to contract
    await owner.invoke(token, 'transfer', {
      recipient: aggregator.address,
      amount: uint256.bnToUint256(100_000_000_000),
    })

    // we have enough funds available now
    available = (await aggregator.call('link_available_for_payment')).available
    assert.ok(available > 0)

    // attempt to withdraw the payment
    await payee.invoke(aggregator, 'withdraw_payment', {
      transmitter: oracle.transmitter.starknetContract.address,
    })

    // balance as transferred to payee
    let { balance } = await token.call('balanceOf', {
      account: payee.starknetContract.address,
    })

    assert.ok(number.toBN(owed).eq(uint256.uint256ToBN(balance)))

    // owed payment is now zero
    owed = (
      await payee.call(aggregator, 'owed_payment', {
        transmitter: oracle.transmitter.starknetContract.address,
      })
    ).amount
    assert.ok(owed == 0)
  })
})
