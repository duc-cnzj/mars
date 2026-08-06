package wssender

// MessageChSize 每个连接的通道缓冲大小。
// 必须 >= NSQ MaxInFlight (1000)，否则突发时 handler 会批量丢消息。
// 同时影响 memory/redis/nsq 三个 sender。
const MessageChSize = 1000

// BroadcastRoom 广播频道名：发布到该频道的消息全连接可收（ToAll）。
const BroadcastRoom = "all"
