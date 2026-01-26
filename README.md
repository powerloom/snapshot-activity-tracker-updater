# snapshot-activity-tracker-updater

An activity tracking and updating service for Powerloom's Decentralized Sequencer-Validator (DSV) Protocol that maintains accurate snapshotter and validator activity metrics and synchronizes with the Data Market and Protocol State contracts.

## Overview

This peer acts as a neutral observer and source for snapshotter and validator activity tracking in the Powerloom protocol. It:

1. **Tracks Validator Finalizations**: Listens to finalized batches from all validators via the validator mesh
2. **Aggregates by Consensus**: Applies majority vote logic to determine canonical state per epoch
3. **Updates On-Chain State**: Keeps protocol contracts synchronized with:
   - Finalized submission counts per epoch
   - EOD (End of Day) counts of slots passing daily snapshot quota
   - EOD final submission counts
4. **Extensible Protocol Coverage**: Designed to track wider protocol activity in future releases

## Architecture

### Core Components

- **Event Monitor**: Watches `EpochReleased` events from Protocol State contract
- **Window Manager**: Manages aggregation windows per `(epochID, dataMarket)` pair
- **Batch Processor**: Receives and tracks finalized batches from validators
- **Consensus Engine**: Aggregates validator outputs using majority vote logic
- **Activity Updater**: Submits on-chain updates for snapshotter activity metrics
- **Relayer Integration**: Queues and broadcasts update transactions via relayer-py

### Data Flow

```
EpochReleased Event
  ↓
Level 1 Finalization Delay (wait for DSV internal aggregation)
  ↓
Aggregation Window Open (collect validator batches)
  ↓
Consensus Aggregation (majority vote per project/CID)
  ↓
Extract Activity Metrics:
  - Finalized submissions per slot
  - Slots passing daily quota
  - EOD final counts
  ↓
On-Chain Update (via relayer-py or direct)
```

## Setup Instructions

### Prerequisites

- Docker and Docker Compose
- Access to libp2p bootstrap peers (validator mesh)
- RPC URL for event monitoring
- Protocol State and Data Market contract addresses
- Trusted updater signer addresses (pre-configured in ProtocolState)

### Docker Setup

The docker-compose setup includes:
- `snapshot-activity-tracker`: Main tracking service
- `relayer-py`: Transaction relayer for on-chain updates
- `redis`: State management
- `rabbitmq`: Transaction queueing

#### 1. Configure Environment Variables

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```bash
# Core networking
BOOTSTRAP_PEERS=/ip4/1.2.3.4/tcp/4001/p2p/QmPeerID1,/ip4/5.6.7.8/tcp/4001/p2p/QmPeerID2
GOSSIPSUB_FINALIZED_BATCH_PREFIX=/powerloom/dsv-devnet-alpha/finalized-batches
GOSSIPSUB_VALIDATOR_PRESENCE_TOPIC=/powerloom/dsv-devnet-alpha/validator/presence

# Contract addresses
PROTOCOL_STATE_CONTRACT=0xYourProtocolStateAddress
DATA_MARKET_ADDRESS=0xYourDataMarketAddress
POWERLOOM_RPC_URL=https://your-powerloom-rpc-url-here
POWERLOOM_RPC_NODES=https://your-rpc-node-1,https://your-rpc-node-2

# Window timing
LEVEL1_FINALIZATION_DELAY_SECONDS=10
AGGREGATION_WINDOW_SECONDS=20

# Update configuration
ENABLE_CONTRACT_UPDATES=true
SUBMISSION_UPDATE_EPOCH_INTERVAL=10
CONTRACT_UPDATE_METHOD=relayer
RELAYER_URL=http://relayer-py:8080
RELAYER_AUTH_TOKEN=your_secure_token_here

# Relayer-py signers (must be in ProtocolState.trustedUpdaters)
VPA_SIGNER_ADDRESSES=0xSigner1,0xSigner2,0xSigner3
VPA_SIGNER_PRIVATE_KEYS=key1,key2,key3
AUTH_TOKEN=${RELAYER_AUTH_TOKEN}

# Tally dumps (optional, for monitoring/debugging)
ENABLE_TALLY_DUMPS=true
TALLY_DUMP_DIR=./tallies
TALLY_RETENTION_DAYS=7
```

#### 2. Configure Trusted Updaters (REQUIRED)

Before enabling contract updates, signer addresses must be added to ProtocolState:

```solidity
// From contract owner
ProtocolState.addTrustedUpdater(0xSigner1)
ProtocolState.addTrustedUpdater(0xSigner2)
```

Verify:
```solidity
ProtocolState.trustedUpdaters(0xSigner1) // Returns true
```

#### 3. Start Services

```bash
./bootstrap.sh  # Clones relayer-py (optional - start.sh handles this)
./start.sh      # Auto-detects and builds or uses pre-built image
```

The `start.sh` script automatically:
- Checks if `./relayer-py` exists
- If exists: builds from source (development mode)
- If missing: uses pre-built Docker image (production mode)

#### 4. Monitor Operation

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f snapshot-activity-tracker
docker-compose logs -f relayer-py

# Health checks
docker-compose exec relayer-py curl http://localhost:8080/health
docker-compose exec redis redis-cli ping
```

#### 5. Stop Services

```bash
./stop.sh
```

## Activity Tracking Details

### Metrics Tracked

1. **Finalized Submission Counts (Per Epoch)**
   - Aggregated from all validators via consensus
   - Counts unique slots with submissions per project
   - Updated on-chain every N epochs (configurable)

2. **EOD Slot Quota Status**
   - Tracks which slots passed the daily snapshot quota threshold
   - Aggregated at end of each day
   - Enables accurate reward distribution

3. **EOD Final Submission Counts**
   - Final count of all submissions for the day
   - Used for reward calculations and protocol analytics

### Consensus Logic

The service aggregates validator outputs using majority vote:
- Per `(projectID, slotID)` combination
- Selects CID with most validator votes
- Merges submission metadata from all validators
- Produces canonical finalized batch per epoch

## Configuration Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `BOOTSTRAP_PEERS` | *required* | Comma-separated validator peer addresses |
| `PROTOCOL_STATE_CONTRACT` | *required* | Protocol State contract address |
| `DATA_MARKET_ADDRESS` | *required* | Data Market contract address |
| `POWERLOOM_RPC_URL` | *required* | RPC URL for event monitoring |
| `POWERLOOM_RPC_NODES` | *required* | Comma-separated RPC nodes for relayer |
| `LEVEL1_FINALIZATION_DELAY_SECONDS` | `10` | Wait for DSV Level 1 aggregation |
| `AGGREGATION_WINDOW_SECONDS` | `20` | Window to collect validator batches |
| `ENABLE_CONTRACT_UPDATES` | `false` | Enable on-chain updates |
| `SUBMISSION_UPDATE_EPOCH_INTERVAL` | `10` | Update every N epochs |
| `CONTRACT_UPDATE_METHOD` | `relayer` | `relayer` or `direct` |
| `RELAYER_URL` | `http://relayer-py:8080` | Relayer service URL |
| `VPA_SIGNER_ADDRESSES` | *required* | Signer addresses (must be trusted updaters) |
| `VPA_SIGNER_PRIVATE_KEYS` | *required* | Signer private keys (no 0x prefix) |

## Troubleshooting

### Contract updates failing

- Verify signer addresses are in `ProtocolState.trustedUpdaters`
- Check `RELAYER_AUTH_TOKEN` matches `AUTH_TOKEN` in relayer-py
- Ensure signers have sufficient gas balance
- Verify `RELAYER_URL` is correct (`http://relayer-py:8080` for docker-compose)

### No batches received

- Verify bootstrap peers are correct and reachable
- Check topic names match validator mesh configuration
- Ensure network connectivity to libp2p peers

### Events not detected

- Verify `POWERLOOM_RPC_URL` is accessible
- Check `PROTOCOL_STATE_CONTRACT` address is correct
- Ensure contract emits `EpochReleased` events

## Future Extensions

This service is designed to track broader protocol activity:
- Additional snapshotter metrics
- Cross-epoch activity aggregation
- Protocol-wide health monitoring
- Extended reward distribution tracking

## License

MIT License