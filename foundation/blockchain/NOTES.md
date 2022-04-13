Package blockchain is the implementation of a blockchain DB. This overview is
a series of notes that have helped with the development of the code.

Must Read

This technical discussion about blockchains and cryptocurrency by David Rosenthal.

	https://blog.dshr.org/2022/02/ee380-talk.html

This is a technical blog about how digital signatures work by Maarten Zuidhoorn.

	https://medium.com/mycrypto/the-magic-of-digital-signatures-on-ethereum-98fe184dc9c7

Great blockchain class by Patrick McCorry

	https://cryptocurrencyclass.github.io/

Notes

A blockchain is a database for storing things of value that allows anyone to
computer a copy using a cryptographic audit trail. It's immutable and
transparent with a token supply and initial user balances. The global blockchain
settings are defined in a Genesis file.

Consensus is about having nodes come together to reach a decision about the state
of the blockchain. You have to balance Safety with Liveness. Systems like Bitcoin
and Ethereum side on Liveness for high availability. This is why we get
forks at times. We need this since individual mining nodes are not reliable in
terms of being online. A country could shut down thousands of machines.

Blocks are just a batch update to the database.

The blockchain is a conceptual single database with a single mempool.

Merkle tree is for convincing you of something I am telling you. Such
as, this transaction does belong to this block.

Simplified Payment Verification Mode is available thanks to block chains and
merkle trees, I can convince anyone a transaction is in the blockchain. At least
for the most current POW.

Gas is a way to limit the computational cost of processing transactions. There is
a gas limit so a transaction will end and fail without running forever. This is
important when transactions are executing code in a VM, like a smart contract.

The validation functions are the rules to consensus. When the validation rules
change, then blocks can still be accepted on old rules until all nodes are updated.
This can create forks as well until finally all the nodes are using the same rules.

Block time refers to the time it takes to mine a new block. In Ethereum, the
average block time is between 12 to 14 seconds and is evaluated after each block.
The expected block time is set as a constant at the protocol level and is used to
protect the network's security when the miners add more computational power. The
average block time gets compared with the expected block time, and if the average
block time is higher, then the difficulty is decreased in the block header. If
the average block time is smaller, then the difficulty in the block header will
be increased.

After a user initiates a transaction, it is propagated peer-to-peer in each node’s
mempool. Each transaction has a fee attached to it. The fee signals the desire to
purchase blockspace, which allows the transaction to be processed and included in a block.

Every moment there are numerous proposed blocks existing in this “Schrödinger’s
state” between unconfirmed and confirmed, competing to find the first hash output
that satisfies the difficulty target.

Nodes also run a series of validity checks on these transactions. These checks
include ensuring that the funds are still available, the output is not exceeding
the input, the signature is valid, etc

Transactions can be dropped from the pool.

When two nodes want to communicate, they send each other some cryptographic data
(public keys and such) to make sure all of the subsequent data transfer is encrypted.

In Proof of Work, the protocol sets out conditions for what makes a block valid.
It might say, for instance, only a block whose hash begins with 00 will be valid.
The only way for the miner to create one that matches that combination is to
brute-force inputs. They can tweak a parameter in their data to produce a different
outcome for every guess until they get the right hash.

Two directions exist for ASIC resistance; firstly make it sequential memory-hard,
i.e. engineer the function such that the determination of the nonce requires a lot
of memory and bandwidth such that the memory cannot be used in parallel to discover
multiple nonces simultaneously. The second is to make the type of computation it
would need to do general-purpose; the meaning of “specialised hardware” for a
general-purpose task set is, naturally, general purpose hardware and as such
commodity desktop computers are likely to be pretty close to “specialised hardware”
for the task. For Ethereum 1.0 we have chosen the first path.

Before the London Upgrade, Ethereum had fixed-sized blocks. In times of high network
demand, these blocks operated at total capacity. As a result, users often had to wait
for high demand to reduce to get included in a block, which led to a poor user experience.

The London Upgrade introduced variable-size blocks to Ethereum. Each block has a
target size of 15 million gas, but the size of blocks will increase or decrease in
accordance with network demand, up until the block limit of 30 million gas (2x
the target block size). The protocol achieves an equilibrium block size of
15 million on average through the process of tâtonnement. This means if the
block size is greater than the target block size, the protocol will increase
the base fee for the following block. Similarly, the protocol will decrease
the base fee if the block size is less than the target block size. The amount
by which the base fee is adjusted is proportional to how far the current block
size is from the target.
