package wssender

// MessageChSize per-connection channel buffer.
// Must be >= NSQ MaxInFlight (1000) to avoid drops during normal bursts.
// Affects memory/redis/nsq all three senders.
const MessageChSize = 1000
const BroadcastRoom = "all"
