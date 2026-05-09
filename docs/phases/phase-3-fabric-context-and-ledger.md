# Phase 3: Fabric Context And Ledger

## Scope

Add Fabric-oriented test context and in-memory ledger support.

## Requirements

- Implement mock stub and transaction context accessors.
- Support public state, private data, events, TxID, channel ID, timestamp, and identity.
- Add ledger helpers on `Context`.
- Add event assertion helpers.
- Keep every `It` isolated from other test cases.

## Deliverables

- Fabric mock stub implementation.
- Transaction context implementation.
- Public and private ledger helpers.
- Event log helpers.
- Tests for state isolation and basic chaincode interactions.

