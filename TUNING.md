## MvIndex Tuning Guide

MvIndex can use additional RAM for caching which has tremendous effects on sync time and query performance. Under normal operations most recent data is automatically kept in memory. If your workload often accesses operations or bigmaps it may be beneficial to follow this guide.

### Database tuning

Our embedded database can use more memory for caching parts of tables. You can observe database stats using the `mvtop` tool or programmatically via the API call `GET /system/tables`. Look for field `packs_count` which is the total number of packs in that table. The table's cache setting controls how many of these packs are kept in memory. To tune for your workload, observe `pack_cache_hits` and `pack_cache_misses` and adjust the cache size. For optimal performance you may want to cache all `ACCOUNT` packs and most/many `OP` & `BIGMAP` table packs.

Default configuration settings are conservative so that the indexer can run on a laptop or low-end server. If you have plenty of RAM you can improve performance considerably. For example, below settings on a 64GB bare-metal server result in \~30GB memory usage with a diverse load on accounts and historic operations.

Set the following options in the config file or as env variables:

```
MV_DB_ACCOUNT_CACHE_SIZE=128
MV_DB_BALANCE_CACHE_SIZE=128
MV_DB_BLOCK_CACHE_SIZE=2
MV_DB_CONTRACT_CACHE_SIZE=8
MV_DB_BIGMAP_TYPES_CACHE_SIZE=32
MV_DB_BIGMAP_VALUES_CACHE_SIZE=4096
MV_DB_CHAIN_CACHE_SIZE=2
MV_DB_SUPPLY_CACHE_SIZE=2
MV_DB_FLOW_CACHE_SIZE=2
MV_DB_OP_CACHE_SIZE=1024
MV_DB_RIGHTS_CACHE_SIZE=2
MV_DB_SNAPSHOT_CACHE_SIZE=2
MV_DB_INCOME_CACHE_SIZE=2
MV_DB_STORAGE_CACHE_SIZE=16
MV_DB_TOKEN_CACHE_SIZE=1024
MV_DB_TOKEN_METADATA_CACHE_SIZE=32
MV_DB_TOKEN_EVENTS_CACHE_SIZE=16
MV_DB_TOKEN_OWNERS_CACHE_SIZE=2048
```

`..CACHE_SIZE` entries are measured in blocks of database rows. Each block contains 32-64k rows and when cached uses 1-20MB RAM depending on table and position in on-chain history. More recent storage updates have become much larger, so be careful with caching the storage tabel too aggressively as it requires an exorbitant amount on memory.


### HTTP server tuning

Each table uses a global lock. We're working on improvements here, but for now if you have very many requests on say `/explorer/op` they all wait for this lock and back-pressure will build which results in a spike in waiting threads and long response times.

To control back-pressure MvIndex v8+ performs request rate limiting by queueing incoming HTTP requests and dispatching them to a pool of worker threads. Queue depth is controlled with `MV_SERVER_QUEUE` and number of workers with `MV_SERVER_WORKERS`. On overflow, new requests immediatly return with HTTP 429 rate limited error response code.

In addition, request processing can be controlled through timeouts:

- `MV_SERVER_HEADER_TIMEOUT` max time to receive all client headers (default 5s)
- `MV_SERVER_READ_TIMEOUT` max time to receive entire client request (default 2s)
- `MV_SERVER_WRITE_TIMEOUT` time to send response (default 15s)

### Integration

**Caching**

MvIndex can return HTTP cache headers (expires, last-modified and cache-control) to interoperate with caches in reverse proxies and browsers.  Set `MV_SERVER_CACHE_ENABLE=true` to enable cache headers in HTTP responses. Data that is subject to change like account info will expire when a new block is expected every minute. Historic data for blocks and operations will expire after 24 hours, but only if the data is outside a reorg-safe window of 64 blocks.

**CORS**

MvIndex can add Cross-Origin Response Headers to its HTTP responses. To enable this feature set `MV_SERVER_CORS_ENABLE=true` or start with `--enable-cors`. You can control CORS headers that are being added with the following config settings:

- `MV_SERVER_CORS_ORIGIN` defaults to `*`, but you can limit this to your own domain
- `MV_SERVER_CORS_ALLOW_HEADERS` sets the list of HTTP headers the client is allowed to use in CORS requests
- `MV_SERVER_CORS_EXPOSE_HEADERS` sets the list of HTTP headers the client can access in responses
- `MV_SERVER_CORS_METHODS` sets the list of HTTP methods the client is allowed to use in CORS requests
- `MV_SERVER_CORS_MAXAGE` (integer, in seconds) sets the maximum time to cache the domains CORS config at clients
- `MV_SERVER_CORS_CREDENTIALS` (bool) enables or disables automatic embedding of client cookies

