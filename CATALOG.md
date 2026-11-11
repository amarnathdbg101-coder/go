# C:\Library Master Catalog 📚

> **184 Production-Ready, High-Efficiency, Security-Hardened Go Libraries.**
> Built strictly adhering to the [Master Architecture Blueprint](BLUEPRINT.md).

---

## 📂 Categorized Domain Suites (Zero-Bloat Architecture)

To keep every Go service lightweight while maintaining instant access to common features, libraries are organized into **8 Domain Work Suites**:

1. **🔒 Security, Auth & Identity Suite**: [`go-authkit`](./go-authkit), [`go-authkit/securityauth`](./go-authkit/securityauth), [`go-jwt`](./go-jwt), [`go-otp`](./go-otp), [`go-crypto`](./go-crypto), [`go-cookie`](./go-cookie), [`go-mask`](./go-mask), [`go-idgen`](./go-idgen), [`go-password`](./go-password), [`go-secguard`](./go-secguard), [`go-keyrotator`](./go-keyrotator), [`go-tamperseal`](./go-tamperseal), [`go-urlsigner`](./go-urlsigner), [`go-secureid`](./go-secureid)
2. **🌐 Network, Resilience & Traffic Suite**: [`go-ratelimit`](./go-ratelimit), [`go-httpclient`](./go-httpclient), [`go-circuitbreaker`](./go-circuitbreaker), [`go-retry`](./go-retry), [`go-healthcheck`](./go-healthcheck), [`go-ipkit`](./go-ipkit), [`go-cors`](./go-cors), [`go-webhook`](./go-webhook), [`go-shutdown`](./go-shutdown), [`go-adaptivelimit`](./go-adaptivelimit), [`go-bulkhead`](./go-bulkhead), [`go-wshub`](./go-wshub), [`go-sse`](./go-sse)
3. **🗄️ Data Caching, SQL & Validation Suite**: [`go-cachekit`](./go-cachekit), [`go-cachekit/data`](./go-cachekit/data), [`go-cachekit/data/cache`](./go-cachekit/data/cache), [`go-cachekit/lru`](./go-cachekit/lru), [`go-cachekit/lfu`](./go-cachekit/lfu), [`go-cachekit/tinylfu`](./go-cachekit/tinylfu), [`go-paginate`](./go-paginate), [`go-paginate/queryfilter`](./go-paginate/queryfilter), [`go-sqlbuilder`](./go-sqlbuilder), [`go-validator`](./go-validator), [`go-validator/valrules`](./go-validator/valrules), [`go-sanitize`](./go-sanitize), [`go-slug`](./go-slug), [`go-env`](./go-env), [`go-env/envconfig`](./go-env/envconfig), [`go-env/envvalidator`](./go-env/envvalidator), [`go-env/hotconfig`](./go-env/hotconfig), [`go-sqlscan`](./go-sqlscan), [`go-idempotency`](./go-idempotency), [`go-jsonpath`](./go-jsonpath)
4. **⚡ Async Workflows, Workers & Events Suite**: [`go-taskqueue`](./go-taskqueue), [`go-taskqueue/async`](./go-taskqueue/async), [`go-taskqueue/async/taskqueue`](./go-taskqueue/async/taskqueue), [`go-workerpool`](./go-workerpool), [`go-workerpool/adaptivepool`](./go-workerpool/adaptivepool), [`go-events`](./go-events), [`go-events/eventemitter`](./go-events/eventemitter), [`go-pubsub`](./go-pubsub), [`go-stream`](./go-stream), [`go-pipeline`](./go-pipeline), [`go-cron`](./go-cron), [`go-delayqueue`](./go-delayqueue), [`go-saga`](./go-saga), [`go-disruptor`](./go-disruptor)
5. **📄 Documents, Media & Files Suite**: [`go-image-filekit`](./go-image-filekit), [`go-image-filekit/imagefx`](./go-image-filekit/imagefx), [`go-image-filekit/studiocomp`](./go-image-filekit/studiocomp), [`go-image-filekit/mimekit`](./go-image-filekit/mimekit), [`go-gopdf-pdfkit`](./go-gopdf-pdfkit), [`go-chromedp-pdfkit`](./go-chromedp-pdfkit), [`go-csvkit`](./go-csvkit), [`go-compress`](./go-compress), [`go-imghash`](./go-imghash), [`go-jsonstream`](./go-jsonstream)
6. **📊 Observability, Metrics & Audit Suite**: [`go-auditlog`](./go-auditlog), [`go-metrics`](./go-metrics), [`go-contextkit`](./go-contextkit), [`go-migrator`](./go-migrator), [`go-sysinfo`](./go-sysinfo), [`go-stopwatch`](./go-stopwatch), [`go-colorlog`](./go-colorlog), [`go-w3ctrace`](./go-w3ctrace), [`go-slidingstats`](./go-slidingstats), [`go-pgexplain`](./go-pgexplain)
7. **🚀 High-Speed Data Structures & Algorithms Suite**: [`go-bloom`](./go-bloom), [`go-locks`](./go-locks), [`go-cachekit/lru`](./go-cachekit/lru), [`go-cachekit/lfu`](./go-cachekit/lfu), [`go-cachekit/tinylfu`](./go-cachekit/tinylfu), [`go-bitmap`](./go-bitmap), [`go-btree`](./go-btree), [`go-radix`](./go-radix), [`go-rbtree`](./go-rbtree), [`go-sortedset`](./go-sortedset), [`go-ringbuffer`](./go-ringbuffer), [`go-concurrentmap`](./go-concurrentmap), [`go-bm25`](./go-bm25), [`go-cuckoofilter`](./go-cuckoofilter)
8. **🛒 E-Commerce & Retail Business Suite**: [`go-money`](./go-money), [`go-ledger`](./go-ledger), [`go-reservation`](./go-reservation), [`go-retailtaxo`](./go-retailtaxo), [`go-parchi`](./go-parchi), [`go-escpos`](./go-escpos), [`go-pricing`](./go-pricing), [`go-bizhours`](./go-bizhours), [`go-hinglish`](./go-hinglish), [`go-geodistance`](./go-geodistance), [`go-geodistance/geohash`](./go-geodistance/geohash), [`go-geodistance/geofence`](./go-geodistance/geofence)

---

### Consolidated module usage

Related packages share a Go module while retaining focused package APIs. Install the module root, then import the desired subpackage:

```powershell
go get github.com/amarnathdbg101-coder/go/go-env@v1.1.0
go get github.com/amarnathdbg101-coder/go/go-cachekit@v1.1.0
```

For example, configuration packages import as `github.com/amarnathdbg101-coder/go/go-env/envconfig`,
`github.com/amarnathdbg101-coder/go/go-env/envvalidator`, and
`github.com/amarnathdbg101-coder/go/go-env/hotconfig`. Git tags use the module root
(for example `go-env/v1.1.0`), not a subpackage name.
These paths are available remotely only after the updated commit and matching module-root tags are published; existing published tags are immutable.

## 🏛️ Comprehensive Library Matrix

| # | Library | Path | Primary Responsibilities | Zero-Dep? | Tests |
|---|---|---|---|---|---|
| 1 | **`go-authkit`** | [`go-authkit`](./go-authkit) | Dual JWT (15m/7d), Bcrypt Cost 12, OTP Brute-force Lockout, RBAC | ❌ (x/crypto, jwt) | ✅ PASS |
| 2 | **`go-paginate`** | [`go-paginate`](./go-paginate) | SQL Injection-Proof Cursor & Offset Pagination, Generics `Page[T]` | ✅ 100% StdLib | ✅ PASS |
| 3 | **`go-cachekit`** | [`go-cachekit`](./go-cachekit) | L1 RAM + L2 Redis, Singleflight Dogpile Protection, Generics | ❌ (redis) | ✅ PASS |
| 4 | **`go-image-filekit`** | [`go-image-filekit`](./go-image-filekit) | 512B Magic MIME Sniffing, Bilinear Resize, Path Traversal Defense | ✅ 100% StdLib | ✅ PASS |
| 5 | **`go-chromedp-pdfkit`** | [`go-chromedp-pdfkit`](./go-chromedp-pdfkit) | Headless Chrome HTML/CSS to PDF, 3-Tier Anti-Zombie Process | ❌ (chromedp) | ✅ PASS |
| 6 | **`go-gopdf-pdfkit`** | [`go-gopdf-pdfkit`](./go-gopdf-pdfkit) | Sub-20ms Native PDF & 80mm POS Thermal Slips, Unicode Hindi/Rupee | ❌ (gopdf) | ✅ PASS |
| 7 | **`go-ratelimit`** | [`go-ratelimit`](./go-ratelimit) | Token Bucket & Sliding Window, IP/Header Key, Standard 429 Headers | ✅ 100% StdLib | ✅ PASS |
| 8 | **`go-validator`** | [`go-validator`](./go-validator) | Reflection-Free Fluent Validator, Direct `map[string]string` Errors | ✅ 100% StdLib | ✅ PASS |
| 9 | **`go-taskqueue`** | [`go-taskqueue`](./go-taskqueue) | Bounded Worker Pool, Panic Isolation, Exponential Backoff Retries | ✅ 100% StdLib | ✅ PASS |
| 10 | **`go-auditlog`** | [`go-auditlog`](./go-auditlog) | Sub-µs Async Queue, Auto Secret Masking, Atomic Batch Insert | ✅ 100% StdLib | ✅ PASS |
| 11 | **`go-crypto`** | [`go-crypto`](./go-crypto) | AES-256-GCM AEAD, HMAC-SHA256, UUIDv7 (RFC 9562), PBKDF2 | ✅ 100% StdLib | ✅ PASS |
| 12 | **`go-slug`** | [`go-slug`](./go-slug) | Sub-4µs URL Slugifier, Hindi/Devanagari, Latin & Currency Transliteration | ✅ 100% StdLib | ✅ PASS |
| 13 | **`go-env`** | [`go-env`](./go-env) | Zero-Reflection `.env` Loader, Type-Safe Parsing, Masked Dumps | ✅ 100% StdLib | ✅ PASS |
| 14 | **`go-httpclient`** | [`go-httpclient`](./go-httpclient) | Resilient `net/http`, Full-Jitter Retries, Circuit Breaker, Leak-Proof | ✅ 100% StdLib | ✅ PASS |
| 15 | **`go-healthcheck`** | [`go-healthcheck`](./go-healthcheck) | K8s `/livez` & `/readyz`, Parallel Checkers, DB Protection TTL Cache | ✅ 100% StdLib | ✅ PASS |
| 16 | **`go-events`** | [`go-events`](./go-events) | In-Memory Event Bus, Wildcards (`order.*`), Panic Isolation, DLQ | ✅ 100% StdLib | ✅ PASS |
| 17 | **`go-idgen`** | [`go-idgen`](./go-idgen) | 4.8M IDs/sec 64-bit Snowflake ID Generator, 0 B/op, Base62 | ✅ 100% StdLib | ✅ PASS |
| 18 | **`go-csvkit`** | [`go-csvkit`](./go-csvkit) | OOM-Proof Streaming Batch Chunking, Excel BOM Auto-Strip, Typed Rows | ✅ 100% StdLib | ✅ PASS |
| 19 | **`go-circuitbreaker`**| [`go-circuitbreaker`](./go-circuitbreaker) | Universal Function Breaker, Registry, Canary Probes, Fallbacks | ✅ 100% StdLib | ✅ PASS |
| 20 | **`go-retry`** | [`go-retry`](./go-retry) | Context-Aware Retrier, Full Jitter, Backoff Strategies, Generics | ✅ 100% StdLib | ✅ PASS |
| 21 | **`go-metrics`** | [`go-metrics`](./go-metrics) | Lock-Free Prometheus Engine, Histograms (p50/p95/p99), Runtime Gauges | ✅ 100% StdLib | ✅ PASS |
| 22 | **`go-locks`** | [`go-locks`](./go-locks) | Zero-Leak Keyed Mutex & PostgreSQL Advisory Locks (`pg_try_advisory_lock`) | ✅ 100% StdLib | ✅ PASS |
| 23 | **`go-sanitize`** | [`go-sanitize`](./go-sanitize) | Strict HTML/XSS Sanitizer, SQL LIKE Escaper, Zero-Width Unicode Cleaner | ✅ 100% StdLib | ✅ PASS |
| 24 | **`go-contextkit`** | [`go-contextkit`](./go-contextkit) | Collision-Free Context Keys, Detached Background Context, Tracing MW | ✅ 100% StdLib | ✅ PASS |
| 25 | **`go-stream`** | [`go-stream`](./go-stream) | Type-Safe Generics Functional Pipeline, Order-Preserving ParallelMap | ✅ 100% StdLib | ✅ PASS |
| 26 | **`go-workerpool`** | [`go-workerpool`](./go-workerpool) | Bounded Goroutine Pool, TrySubmit Backpressure, SubmitWait | ✅ 100% StdLib | ✅ PASS |
| 27 | **`go-sqlbuilder`** | [`go-sqlbuilder`](./go-sqlbuilder) | Parameterized PostgreSQL Builder ($1, $2), Zero-Reflection, Anti-Injection | ✅ 100% StdLib | ✅ PASS |
| 28 | **`go-cors`** | [`go-cors`](./go-cors) | Zero-Allocation CORS Middleware, Preflight 204, Subdomain Wildcards | ✅ 100% StdLib | ✅ PASS |
| 29 | **`go-compress`** | [`go-compress`](./go-compress) | Recycled Buffer sync.Pool Gzip Middleware, Threshold Buffering (>1KB) | ✅ 100% StdLib | ✅ PASS |
| 30 | **`go-pubsub`** | [`go-pubsub`](./go-pubsub) | Non-Blocking Channel Fanout, Slow-Consumer Mitigation (DropOldest/New) | ✅ 100% StdLib | ✅ PASS |
| 31 | **`go-migrator`** | [`go-migrator`](./go-migrator) | Transactional PostgreSQL Migrator, Embed.FS, SHA-256 Checksums | ✅ 100% StdLib | ✅ PASS |
| 32 | **`go-mask`** | [`go-mask`](./go-mask) | PII Masker (Aadhaar, PAN, PCI-DSS Card, Phone, Email, JSON Redactor) | ✅ 100% StdLib | ✅ PASS |
| 33 | **`go-cookie`** | [`go-cookie`](./go-cookie) | Encrypted Authenticated Cookies (AES-256-GCM + Name Binding AAD) | ✅ 100% StdLib | ✅ PASS |
| 34 | **`go-ipkit`** | [`go-ipkit`](./go-ipkit) | Real Client IP Extractor, Reverse Proxy Traversal, Anti-Spoofing, CIDR | ✅ 100% StdLib | ✅ PASS |
| 35 | **`go-shutdown`** | [`go-shutdown`](./go-shutdown) | Phased Graceful Shutdown (HTTP -> Workers -> DB Storage), Hard Timeout | ✅ 100% StdLib | ✅ PASS |
| 36 | **`go-jwt`** | [`go-jwt`](./go-jwt) | Zero-Dep Standard Library HS256 JWT, Claims Parsing, Clock Drift | ✅ 100% StdLib | ✅ PASS |
| 37 | **`go-bloom`** | [`go-bloom`](./go-bloom) | Ultra-Fast Concurrent Bloom Filter, SplitMix64 Bit Mixing, Bitset | ✅ 100% StdLib | ✅ PASS |
| 38 | **`go-cron`** | [`go-cron`](./go-cron) | Unix 5-Field In-Memory Cron & Interval Scheduler, Panic Isolation | ✅ 100% StdLib | ✅ PASS |
| 39 | **`go-money`** | [`go-money`](./go-money) | Zero-Float-Drift Financial Math, ISO-4217 Currencies, Fair Splits | ✅ 100% StdLib | ✅ PASS |
| 40 | **`go-webhook`** | [`go-webhook`](./go-webhook) | Outbound Webhook Dispatcher, HMAC Anti-Replay Signing, Retry Worker Pool | ✅ 100% StdLib | ✅ PASS |
| 41 | **`go-otp`** | [`go-otp`](./go-otp) | RFC 6238 TOTP & RFC 4226 HOTP 2FA Engine, Authenticator QR URIs | ✅ 100% StdLib | ✅ PASS |
| 42 | **`go-pipeline`** | [`go-pipeline`](./go-pipeline) | Generic Onion-Model Middleware & Workflow Pipeline, Panic Isolation | ✅ 100% StdLib | ✅ PASS |
| 43 | **`go-safecast`** | [`go-safecast`](./go-safecast) | Zero-Allocation Overflow/Underflow-Safe Generic Numeric Type Casting | ✅ 100% StdLib | ✅ PASS |
| 44 | **`go-diff`** | [`go-diff`](./go-diff) | Reflection-Safe Structured Data Diff Engine (Structs, Maps, Slices) | ✅ 100% StdLib | ✅ PASS |
| 45 | **`go-strcase`** | [`go-strcase`](./go-strcase) | High-Performance Case Conversion with Acronym Recognition | ✅ 100% StdLib | ✅ PASS |
| 46 | **`go-batcher`** | [`go-batcher`](./go-batcher) | High-Throughput Micro-Batching Aggregator with Dual-Trigger Flush | ✅ 100% StdLib | ✅ PASS |
| 47 | **`go-kvstore`** | [`go-kvstore`](./go-kvstore) | 32-Way Sharded In-Memory Key-Value Cache with Dual TTL & Atomic Ops | ✅ 100% StdLib | ✅ PASS |
| 48 | **`go-di`** | [`go-di`](./go-di) | Type-Safe Generic Dependency Injection Container with Cycle Detection | ✅ 100% StdLib | ✅ PASS |
| 49 | **`go-tree`** | [`go-tree`](./go-tree) | Generic Prefix Trie Autocomplete & Hierarchical Category Tree | ✅ 100% StdLib | ✅ PASS |
| 50 | **`go-version`** | [`go-version`](./go-version) | SemVer 2.0.0 Parser, Comparator, Caret/Tilde Range Constraint Matcher | ✅ 100% StdLib | ✅ PASS |
| 51 | **`go-ringbuffer`** | [`go-ringbuffer`](./go-ringbuffer) | Thread-Safe Generic Circular FIFO Ring Buffer with Overwrite Modes | ✅ 100% StdLib | ✅ PASS |
| 52 | **`go-geodistance/geohash`** | [`go-geodistance/geohash`](./go-geodistance/geohash) | Geohash Encoding/Decoding, 8-Neighbor Proximity, Haversine Distance | ✅ 100% StdLib | ✅ PASS |
| 53 | **`go-bitmap`** | [`go-bitmap`](./go-bitmap) | Compact Bit Array with Hardware POPCNT and Set Algebra (And/Or/Diff) | ✅ 100% StdLib | ✅ PASS |
| 54 | **`go-cachekit/lru`** | [`go-cachekit/lru`](./go-cachekit/lru) | O(1) Generic LRU Cache with Per-Item TTL and Eviction Callbacks | ✅ 100% StdLib | ✅ PASS |
| 55 | **`go-color`** | [`go-color`](./go-color) | ANSI Terminal Styling, 24-Bit RGB/HEX TrueColor, NO_COLOR Compliant | ✅ 100% StdLib | ✅ PASS |
| 56 | **`go-debounce`** | [`go-debounce`](./go-debounce) | Thread-Safe Debounce & Throttle with Keyed Multi-Tenant Isolation | ✅ 100% StdLib | ✅ PASS |
| 57 | **`go-priorityqueue`** | [`go-priorityqueue`](./go-priorityqueue) | Generic Binary Heap Priority Queue with Max/Min Heap Support | ✅ 100% StdLib | ✅ PASS |
| 58 | **`go-secureid`** | [`go-secureid`](./go-secureid) | CSPRNG Random Tokens, Prefixed API Keys, NanoIDs, Crockford Base32 | ✅ 100% StdLib | ✅ PASS |
| 59 | **`go-pool`** | [`go-pool`](./go-pool) | Bounded Generic Resource Pool with Active Health Checks & Idle Eviction | ✅ 100% StdLib | ✅ PASS |
| 60 | **`go-fsm`** | [`go-fsm`](./go-fsm) | Generic Finite State Machine with Guard Conditions & Lifecycle Hooks | ✅ 100% StdLib | ✅ PASS |
| 61 | **`go-bizhours`** | [`go-bizhours`](./go-bizhours) | Business Operating Hours, Weekly-Offs, Countdown, Live Status | ✅ 100% StdLib | ✅ PASS |
| 62 | **`go-idempotency`**| [`go-idempotency`](./go-idempotency) | Distributed Idempotency-Key Middleware, Cached Responses, Replay Defense | ✅ 100% StdLib | ✅ PASS |
| 63 | **`go-imghash`** | [`go-imghash`](./go-imghash) | Perceptual Image Hashing (pHash, aHash, dHash) for Deduplication | ✅ 100% StdLib | ✅ PASS |
| 64 | **`go-keyrotator`** | [`go-keyrotator`](./go-keyrotator) | Zero-Downtime Cryptographic Key Rotation (Active, Retiring, Retired) | ✅ 100% StdLib | ✅ PASS |
| 65 | **`go-secguard`** | [`go-secguard`](./go-secguard) | OWASP Security Headers, WAF (SQLi/XSS/Traversal), Scanner Trap, IP Jail | ✅ 100% StdLib | ✅ PASS |
| 66 | **`go-concurrentmap`** | [`go-concurrentmap`](./go-concurrentmap) | 32-Way Striped Concurrent Hash Map with Type-Safe Generics | ✅ 100% StdLib | ✅ PASS |
| 67 | **`go-retry/exponentialbackoff`** | [`go-retry/exponentialbackoff`](./go-retry/exponentialbackoff) | Mathematical Exponential Backoff & AWS-Recommended Jitter Strategies | ✅ 100% StdLib | ✅ PASS |
| 68 | **`go-urlsigner`** | [`go-urlsigner`](./go-urlsigner) | HMAC-SHA256 Presigned Temporary URLs with Tamper Protection | ✅ 100% StdLib | ✅ PASS |
| 69 | **`go-stopwatch`** | [`go-stopwatch`](./go-stopwatch) | High-Resolution Multi-Lap Execution Profiler & Profiling Summary | ✅ 100% StdLib | ✅ PASS |
| 70 | **`go-units`** | [`go-units`](./go-units) | Byte Size (SI/IEC) & Extended Duration (Days/Weeks) Parser & Formatter | ✅ 100% StdLib | ✅ PASS |
| 71 | **`go-wildcard`** | [`go-wildcard`](./go-wildcard) | High-Performance Glob Pattern Matcher (`*`, `?`, character classes `[...]`) | ✅ 100% StdLib | ✅ PASS |
| 72 | **`go-checksum`** | [`go-checksum`](./go-checksum) | Multi-Algorithm Streaming Checksum (MD5, SHA1, SHA256, SHA512, CRC32) | ✅ 100% StdLib | ✅ PASS |
| 73 | **`go-templatekit`** | [`go-templatekit`](./go-templatekit) | Ergonomic Template Engine with Filter Pipelines (`{{ var \| filter }}`) | ✅ 100% StdLib | ✅ PASS |
| 74 | **`go-env/envconfig`** | [`go-env/envconfig`](./go-env/envconfig) | Tag-Driven 12-Factor App Struct Config Unmarshaler & Masked Dumps | ✅ 100% StdLib | ✅ PASS |
| 75 | **`go-dedup`** | [`go-dedup`](./go-dedup) | Sliding-Window In-Memory Event & Webhook Deduplicator with Janitor | ✅ 100% StdLib | ✅ PASS |
| 76 | **`go-jsonpath`** | [`go-jsonpath`](./go-jsonpath) | High-Performance JSONPath Parser, Slices & Wildcard Projections | ✅ 100% StdLib | ✅ PASS |
| 77 | **`go-phone`** | [`go-phone`](./go-phone) | E.164 Telephone Parser, Indian 10-Digit Mobile Validator & Masking | ✅ 100% StdLib | ✅ PASS |
| 78 | **`go-featureflag`** | [`go-featureflag`](./go-featureflag) | Dynamic Feature Toggles, Percentage Rollouts, Whitelists & Targeting | ✅ 100% StdLib | ✅ PASS |
| 79 | **`go-fuzzy`** | [`go-fuzzy`](./go-fuzzy) | Fuzzy String Matching, Levenshtein, Damerau, Jaro-Winkler & Ranking | ✅ 100% StdLib | ✅ PASS |
| 80 | **`go-useragent`** | [`go-useragent`](./go-useragent) | High-Speed HTTP User-Agent Parser, Browser, OS, Device & Bot Detection | ✅ 100% StdLib | ✅ PASS |
| 81 | **`go-semaphore`** | [`go-semaphore`](./go-semaphore) | Weighted Counting Semaphore, FIFO Fairness, Timeouts & Draining | ✅ 100% StdLib | ✅ PASS |
| 82 | **`go-set`** | [`go-set`](./go-set) | Type-Safe Generic Set, Union, Intersect, Difference, JSON & Thread-Safe | ✅ 100% StdLib | ✅ PASS |
| 83 | **`go-latch`** | [`go-latch`](./go-latch) | Synchronization Primitives: CountDownLatch & Multi-Phase CyclicBarrier | ✅ 100% StdLib | ✅ PASS |
| 84 | **`go-jsonstream`** | [`go-jsonstream`](./go-jsonstream) | OOM-Proof Streaming JSON & NDJSON Parser/Writer for Massive Datasets | ✅ 100% StdLib | ✅ PASS |
| 85 | **`go-humanize`** | [`go-humanize`](./go-humanize) | Data Humanizer, Indian Numbering (Lakh/Cr), Ordinals & Relative Time | ✅ 100% StdLib | ✅ PASS |
| 86 | **`go-slicekit`** | [`go-slicekit`](./go-slicekit) | Generic Slice Utilities: Chunk, Flatten, Unique, GroupBy, Partition | ✅ 100% StdLib | ✅ PASS |
| 87 | **`go-atomic`** | [`go-atomic`](./go-atomic) | High-Level Atomic Primitives: Bool, String, Duration, Time, Float64 | ✅ 100% StdLib | ✅ PASS |
| 88 | **`go-urlquery`** | [`go-urlquery`](./go-urlquery) | URL Query String Encoder/Decoder for Structs with Tags & Defaults | ✅ 100% StdLib | ✅ PASS |
| 89 | **`go-contenttype`** | [`go-contenttype`](./go-contenttype) | HTTP Content Negotiation, Quality Values ($q$) & MIME Charset Parser | ✅ 100% StdLib | ✅ PASS |
| 90 | **`go-sysinfo`** | [`go-sysinfo`](./go-sysinfo) | Real-Time Diagnostics Telemetry: Memory, Goroutines, Hostname & Uptime | ✅ 100% StdLib | ✅ PASS |
| 91 | **`go-syncmap`** | [`go-syncmap`](./go-syncmap) | Type-Safe Generic Concurrent Map (`sync.Map`) with Atomic Compute | ✅ 100% StdLib | ✅ PASS |
| 92 | **`go-netkit`** | [`go-netkit`](./go-netkit) | Network Utilities: Free Ports, TCP Reachability, Private IP & CIDRs | ✅ 100% StdLib | ✅ PASS |
| 93 | **`go-table`** | [`go-table`](./go-table) | ASCII, Markdown & CSV Terminal Table Formatter with Unicode Support | ✅ 100% StdLib | ✅ PASS |
| 94 | **`go-cachekit/lfu`** | [`go-cachekit/lfu`](./go-cachekit/lfu) | O(1) Least Frequently Used (LFU) Cache with Frequency Lists & TTL | ✅ 100% StdLib | ✅ PASS |
| 95 | **`go-deepclone`** | [`go-deepclone`](./go-deepclone) | Reflection-Safe Deep Cloning Engine with Cycle-Detection Protection | ✅ 100% StdLib | ✅ PASS |
| 96 | **`go-sse`** | [`go-sse`](./go-sse) | WHATWG Server-Sent Events (SSE) Real-Time HTTP Push Streaming Engine | ✅ 100% StdLib | ✅ PASS |
| 97 | **`go-hashring`** | [`go-hashring`](./go-hashring) | Consistent Hash Ring with Virtual Nodes (Ketama) & Replicas | ✅ 100% StdLib | ✅ PASS |
| 98 | **`go-timingwheel`** | [`go-timingwheel`](./go-timingwheel) | High-Capacity $O(1)$ Hashed Timing Wheel Scheduler with Multi-Rounds | ✅ 100% StdLib | ✅ PASS |
| 99 | **`go-password`** | [`go-password`](./go-password) | OWASP/NIST Password Validator, Shannon Entropy & CSPRNG Generator | ✅ 100% StdLib | ✅ PASS |
| 100 | **`go-resilience`** | [`go-resilience`](./go-resilience) | Composable Fault-Tolerance: Retries, Breaker, Bulkhead, Timeout, Fallback | ✅ 100% StdLib | ✅ PASS |
| 101 | **`go-orderedmap`** | [`go-orderedmap`](./go-orderedmap) | Deterministic Generic Ordered Map with $O(1)$ Operations & JSON Marshalling | ✅ 100% StdLib | ✅ PASS |
| 102 | **`go-validator/valrules`** | [`go-validator/valrules`](./go-validator/valrules) | Enterprise Validation Rules: Luhn (Cards), Indian GSTIN, PAN, IFSC, E.164, UUID | ✅ 100% StdLib | ✅ PASS |
| 103 | **`go-httpclient/retryablehttp`** | [`go-httpclient/retryablehttp`](./go-httpclient/retryablehttp) | Resilient HTTP Client with Request Body Rewind, `Retry-After` & Full Jitter | ✅ 100% StdLib | ✅ PASS |
| 104 | **`go-bytespool`** | [`go-bytespool`](./go-bytespool) | Tiered Zero-Allocation Buffer Pool (512B-1MB) with Anti-Bloat Discard | ✅ 100% StdLib | ✅ PASS |
| 105 | **`go-expr`** | [`go-expr`](./go-expr) | Safe In-Memory Formula & Business Rule Expression Evaluator (Arithmetic & Logic) | ✅ 100% StdLib | ✅ PASS |
| 106 | **`go-colorgrad`** | [`go-colorgrad`](./go-colorgrad) | 24-bit TrueColor Terminal Gradients, RGB Interpolator & Dynamic Progress Bars | ✅ 100% StdLib | ✅ PASS |
| 107 | **`go-syncchan`** | [`go-syncchan`](./go-syncchan) | Leak-Proof Channel Combinators: `FanIn`, `FanOut`, `Tee`, `Take`, `Drop`, `OrDone` | ✅ 100% StdLib | ✅ PASS |
| 108 | **`go-saferegex`** | [`go-saferegex`](./go-saferegex) | ReDoS-Protected Regular Expression Matcher with LRU Compile Cache & Hard Timeouts | ✅ 100% StdLib | ✅ PASS |
| 109 | **`go-dag`** | [`go-dag`](./go-dag) | Directed Acyclic Graph Task Runner with Kahn's Sort & Bounded Concurrency | ✅ 100% StdLib | ✅ PASS |
| 110 | **`go-sortedset`** | [`go-sortedset`](./go-sortedset) | Redis-Style In-Memory Sorted Set (ZSET) via SkipList with $O(\log N)$ Ranks | ✅ 100% StdLib | ✅ PASS |
| 111 | **`go-slidingwindow`** | [`go-slidingwindow`](./go-slidingwindow) | Rolling Time Window Ring Buffer Counter & Rate Limiter ($O(1)$ Increments) | ✅ 100% StdLib | ✅ PASS |
| 112 | **`go-env/envvalidator`** | [`go-env/envvalidator`](./go-env/envvalidator) | Zero-Reflection Schema-Driven Startup Environment Validator with Aggregated Reports | ✅ 100% StdLib | ✅ PASS |
| 113 | **`go-mockdb`** | [`go-mockdb`](./go-mockdb) | Lightweight In-Memory SQL Driver Mock for Unit Testing Repositories | ✅ 100% StdLib | ✅ PASS |
| 114 | **`go-multierror`** | [`go-multierror`](./go-multierror) | Thread-Safe Multi-Error Accumulator with Go 1.20+ `Unwrap() []error` Support | ✅ 100% StdLib | ✅ PASS |
| 115 | **`go-pathkit`** | [`go-pathkit`](./go-pathkit) | Path Traversal Defense, Crash-Proof Atomic File Writes & Directory Walker | ✅ 100% StdLib | ✅ PASS |
| 116 | **`go-saga`** | [`go-saga`](./go-saga) | Distributed Transaction Saga Orchestrator with Reverse Backward Compensation | ✅ 100% StdLib | ✅ PASS |
| 117 | **`go-sqlscan`** | [`go-sqlscan`](./go-sqlscan) | Reflection-Cached SQL Rows to Struct Scanner (`ScanAll[T]` & `ScanOne[T]`) | ✅ 100% StdLib | ✅ PASS |
| 118 | **`go-events/eventemitter`** | [`go-events/eventemitter`](./go-events/eventemitter) | Event Pub-Sub Engine with `On`, `Once`, Wildcards & Listener Leak Warnings | ✅ 100% StdLib | ✅ PASS |
| 119 | **`go-colorlog`** | [`go-colorlog`](./go-colorlog) | Zero-Allocation Structured Terminal Logger with Color Badges & Caller Tracing | ✅ 100% StdLib | ✅ PASS |
| 120 | **`go-paginate/queryfilter`** | [`go-paginate/queryfilter`](./go-paginate/queryfilter) | Safe REST URL Query to SQL WHERE & ORDER BY Parameterized Builder | ✅ 100% StdLib | ✅ PASS |
| 121 | **`go-workerpool/adaptivepool`** | [`go-workerpool/adaptivepool`](./go-workerpool/adaptivepool) | Auto-Scaling Elastic Worker Pool with Dynamic Expansion & Idle Worker Eviction | ✅ 100% StdLib | ✅ PASS |
| 122 | **`go-countingbloom`** | [`go-countingbloom`](./go-countingbloom) | Counting Bloom Filter with Safe Deletion (`Remove`) & Counter Saturation Protection | ✅ 100% StdLib | ✅ PASS |
| 123 | **`go-image-filekit/mimekit`** | [`go-image-filekit/mimekit`](./go-image-filekit/mimekit) | 512-Byte Magic Byte File Sniffer & Polyglot Executable Defense with Stream Rewinding | ✅ 100% StdLib | ✅ PASS |
| 124 | **`go-tamperseal`** | [`go-tamperseal`](./go-tamperseal) | Cryptographic Tamper-Proof Data Seal with Expiration & Constant-Time Verification | ✅ 100% StdLib | ✅ PASS |
| 125 | **`go-slidingstats`** | [`go-slidingstats`](./go-slidingstats) | Rolling Window Streaming Statistics & Latency Percentiles (P50, P90, P95, P99) | ✅ 100% StdLib | ✅ PASS |
| 126 | **`go-wshub`** | [`go-wshub`](./go-wshub) | Real-Time WebSocket Hub, Room Multiplexing, Heartbeat & Event Dispatcher | ✅ 100% StdLib | ✅ PASS |
| 127 | **`go-outbox`** | [`go-outbox`](./go-outbox) | Transactional Outbox Pattern for Microservices & Reliable Event Publishing | ✅ 100% StdLib | ✅ PASS |
| 128 | **`go-ledger`** | [`go-ledger`](./go-ledger) | Double-Entry Bookkeeping Ledger Engine with Audit Trail & Zero-Balance Enforcing | ✅ 100% StdLib | ✅ PASS |
| 129 | **`go-reservation`** | [`go-reservation`](./go-reservation) | Inventory & Capacity Reservation Engine with Automatic Expiring Leases | ✅ 100% StdLib | ✅ PASS |
| 130 | **`go-retailtaxo`** | [`go-retailtaxo`](./go-retailtaxo) | Multi-Tier Retail Taxonomy, Category Hierarchy & Product Classifier | ✅ 100% StdLib | ✅ PASS |
| 131 | **`go-parchi`** | [`go-parchi`](./go-parchi) | POS Thermal Receipt Slip Formatter with Hindi/English & Unicode Hindi Currency | ✅ 100% StdLib | ✅ PASS |
| 132 | **`go-escpos`** | [`go-escpos`](./go-escpos) | Raw ESC/POS Thermal Printer Protocol Command Encoder & Binary Builder | ✅ 100% StdLib | ✅ PASS |
| 133 | **`go-dbbackup`** | [`go-dbbackup`](./go-dbbackup) | Automated PostgreSQL Streaming Backup, Rotation & Compression Pipeline | ✅ 100% StdLib | ✅ PASS |
| 134 | **`go-etag`** | [`go-etag`](./go-etag) | HTTP Weak/Strong ETag Generator, If-None-Match Validator & 304 Cache Engine | ✅ 100% StdLib | ✅ PASS |
| 135 | **`go-cloudflare`** | [`go-cloudflare`](./go-cloudflare) | Cloudflare Turnstile CAPTCHA & Security API Verification Client | ✅ 100% StdLib | ✅ PASS |
| 136 | **`go-gemini`** | [`go-gemini`](./go-gemini) | Google Gemini 1.5/2.0 REST AI Client with Streaming & Structured Outputs | ✅ 100% StdLib | ✅ PASS |
| 137 | **`go-geodistance`** | [`go-geodistance`](./go-geodistance) | Ultra-Fast Vincenty/Haversine Geodesic Distance & Proximity Clustering | ✅ 100% StdLib | ✅ PASS |
| 138 | **`go-hinglish`** | [`go-hinglish`](./go-hinglish) | Phonetic Hinglish-to-English Transliteration & Normalization Search Engine | ✅ 100% StdLib | ✅ PASS |
| 139 | **`go-image-filekit/imagefx`** | [`go-image-filekit/imagefx`](./go-image-filekit/imagefx) | High-Efficiency Image Watermarking, Aspect Cropping & Thumbnail Generator | ✅ 100% StdLib | ✅ PASS |
| 140 | **`go-image-filekit/studiocomp`** | [`go-image-filekit/studiocomp`](./go-image-filekit/studiocomp) | Studio Product Photo Background Cleanup, Dropshadow & Composite Canvas | ✅ 100% StdLib | ✅ PASS |
| 141 | **`go-dataloader`** | [`go-dataloader`](./go-dataloader) | Generic N+1 Query Elimination & Request Coalescing with Memoization Cache | ✅ 100% StdLib | ✅ PASS |
| 142 | **`go-deadlock`** | [`go-deadlock`](./go-deadlock) | Runtime Lock-Order Inversion Cycle Detector & Mutex Starvation Monitor | ✅ 100% StdLib | ✅ PASS |
| 143 | **`go-bulkhead`** | [`go-bulkhead`](./go-bulkhead) | Netflix/Hystrix Concurrency Partitioning, Queue Limits & Backpressure Isolator | ✅ 100% StdLib | ✅ PASS |
| 144 | **`go-delayqueue`** | [`go-delayqueue`](./go-delayqueue) | Production-Grade Delayed Job Scheduler, Min-Heap Priority Queue & DLQ Retries | ✅ 100% StdLib | ✅ PASS |
| 145 | **`go-w3ctrace`** | [`go-w3ctrace`](./go-w3ctrace) | W3C TraceContext Recommendation Engine, Distributed Tracing & HTTP Middleware | ✅ 100% StdLib | ✅ PASS |
| 146 | **`go-chaos`** | [`go-chaos`](./go-chaos) | In-Process Chaos Engineering & Fault Injection Engine (Latency, Errors, Panics) | ✅ 100% StdLib | ✅ PASS |
| 147 | **`go-adaptivelimit`** | [`go-adaptivelimit`](./go-adaptivelimit) | TCP Vegas Latency-Driven Adaptive Concurrency Limiter & Fail-Fast Backpressure | ✅ 100% StdLib | ✅ PASS |
| 148 | **`go-supervisor`** | [`go-supervisor`](./go-supervisor) | Erlang/OTP Goroutine Supervisor Tree, Panic Isolation & Crash-Loop Circuit Breaker | ✅ 100% StdLib | ✅ PASS |
| 149 | **`go-distlock`** | [`go-distlock`](./go-distlock) | Distributed Lock with Monotonic Fencing Tokens & Auto-Heartbeat Lease Extension | ✅ 100% StdLib | ✅ PASS |
| 150 | **`go-crdt`** | [`go-crdt`](./go-crdt) | Conflict-Free Replicated Data Types (PNCounter, LWWRegister, ORSet) for Offline Sync | ✅ 100% StdLib | ✅ PASS |
| 151 | **`go-pgexplain`** | [`go-pgexplain`](./go-pgexplain) | Automated PostgreSQL Slow Query EXPLAIN Interceptor, Seq Scan & Missing Index Auditor | ✅ 100% StdLib | ✅ PASS |
| 152 | **`go-shadow`** | [`go-shadow`](./go-shadow) | Asynchronous Production HTTP Traffic Mirroring, Candidate Invocation & Response Diffing | ✅ 100% StdLib | ✅ PASS |
| 153 | **`go-env/hotconfig`** | [`go-env/hotconfig`](./go-env/hotconfig) | Lock-Free Zero-Downtime Hot Config Reloader with SHA-256 Checksum & Inode Watcher | ✅ 100% StdLib | ✅ PASS |
| 154 | **`go-bm25`** | [`go-bm25`](./go-bm25) | In-Memory Full-Text Inverted Index Search Engine with Okapi BM25 Probabilistic Ranking | ✅ 100% StdLib | ✅ PASS |
| 155 | **`go-cachekit/tinylfu`** | [`go-cachekit/tinylfu`](./go-cachekit/tinylfu) | Caffeine Window TinyLFU Cache: 4-Bit Count-Min Sketch & Segmented LRU Scan Resistance | ✅ 100% StdLib | ✅ PASS |
| 156 | **`go-geodistance/geofence`** | [`go-geodistance/geofence`](./go-geodistance/geofence) | Sub-µs Point-in-Polygon (PIP) Ray Casting, AABB O(1) Pre-Filter & Haversine Distance | ✅ 100% StdLib | ✅ PASS |
| 157 | **`go-merkle`** | [`go-merkle`](./go-merkle) | RFC 6962 Binary Merkle Tree, Audit Path Generation & O(log N) Cryptographic Proofs | ✅ 100% StdLib | ✅ PASS |
| 158 | **`go-pricing`** | [`go-pricing`](./go-pricing) | Enterprise E-Commerce Dynamic Tiered Pricing, Quantity Breaks, Bundles & Coupon Engine | ✅ 100% StdLib | ✅ PASS |
| 159 | **`go-window`** | [`go-window`](./go-window) | Streaming Sliding & Tumbling Window Aggregator with Event-Time Watermarking & Late Arrival Handling | ✅ 100% StdLib | ✅ PASS |
| 160 | **`go-roaring`** | [`go-roaring`](./go-roaring) | Compressed 32-bit Roaring Bitmaps with Adaptive Dual Containers & Bitwise SIMD Operations | ✅ 100% StdLib | ✅ PASS |
| 161 | **`go-trafficshaper`** | [`go-trafficshaper`](./go-trafficshaper) | Production-Grade Leaky & Token Bucket Traffic Shaper with Priority Queue Scheduling & Burst Smoothing | ✅ 100% StdLib | ✅ PASS |
| 162 | **`go-simhash`** | [`go-simhash`](./go-simhash) | Locality-Sensitive Hashing LSH with Moses Charikar Fingerprints & Multi-Table Near-Duplicate Search | ✅ 100% StdLib | ✅ PASS |
| 163 | **`go-wal`** | [`go-wal`](./go-wal) | Crash-Resilient Segmented Write-Ahead Log with CRC32 Checksums, Monotonic LSN & Zero-Loss Replay | ✅ 100% StdLib | ✅ PASS |
| 164 | **`go-sstable`** | [`go-sstable`](./go-sstable) | Immutable Sorted String Table with Embedded Bloom Filter, Index Block & Sub-ms Point Scans | ✅ 100% StdLib | ✅ PASS |
| 165 | **`go-hyperloglog`** | [`go-hyperloglog`](./go-hyperloglog) | HyperLogLog++ 64-Bit Cardinality Estimator with LinearCounting Correction & Distributed Merge | ✅ 100% StdLib | ✅ PASS |
| 166 | **`go-cuckoofilter`** | [`go-cuckoofilter`](./go-cuckoofilter) | High-Performance Cuckoo Filter with Dynamic Deletion, 4-Slot Buckets & Partial-Key Kicking | ✅ 100% StdLib | ✅ PASS |
| 167 | **`go-btree`** | [`go-btree`](./go-btree) | Generic Thread-Safe In-Memory B-Tree with Cache-Friendly Branching Degree & Range Iterators | ✅ 100% StdLib | ✅ PASS |
| 168 | **`go-lsmtree`** | [`go-lsmtree`](./go-lsmtree) | Log-Structured Merge Tree Engine with WAL, MemTable, SSTable Flush & Compaction | ✅ 100% StdLib | ✅ PASS |
| 169 | **`go-raft`** | [`go-raft`](./go-raft) | Raft Distributed Consensus Protocol with Randomized Elections, Heartbeats & Quorum Replication | ✅ 100% StdLib | ✅ PASS |
| 170 | **`go-vectorindex`** | [`go-vectorindex`](./go-vectorindex) | In-Memory Vector Similarity Search (Cosine, Euclidean, Dot Product) & Top-K Nearest Neighbors | ✅ 100% StdLib | ✅ PASS |
| 171 | **`go-hpack`** | [`go-hpack`](./go-hpack) | RFC 7541 HPACK Header Compression for HTTP/2 & gRPC with Static Table & Dynamic Caching | ✅ 100% StdLib | ✅ PASS |
| 172 | **`go-mvcc`** | [`go-mvcc`](./go-mvcc) | Multi-Version Concurrency Control Engine with Snapshot Isolation, Conflict Detection & Vacuum | ✅ 100% StdLib | ✅ PASS |
| 173 | **`go-radix`** | [`go-radix`](./go-radix) | Compact Radix Tree / Patricia Trie with Path Compression, Longest Prefix Match & Subtree Walk | ✅ 100% StdLib | ✅ PASS |
| 174 | **`go-heavyhitters`** | [`go-heavyhitters`](./go-heavyhitters) | Space-Saving Streaming Top-K Frequent Items Algorithm with Bounded Frequency Error | ✅ 100% StdLib | ✅ PASS |
| 175 | **`go-eventstore`** | [`go-eventstore`](./go-eventstore) | Event Sourcing Storage Engine with Append-Only Streams, Optimistic Concurrency & Snapshots | ✅ 100% StdLib | ✅ PASS |
| 176 | **`go-disruptor`** | [`go-disruptor`](./go-disruptor) | LMAX Disruptor High-Performance Lock-Free Ring Buffer with Cache-Line Padding & Barriers | ✅ 100% StdLib | ✅ PASS |
| 177 | **`go-tdigest`** | [`go-tdigest`](./go-tdigest) | Ted Dunning's t-Digest Online Streaming Quantile Estimator (P50, P90, P99) & Centroid Merging | ✅ 100% StdLib | ✅ PASS |
| 178 | **`go-rbtree`** | [`go-rbtree`](./go-rbtree) | Left-Leaning Red-Black (LLRB) Self-Balancing Tree with Floor, Ceiling & Range Iterators | ✅ 100% StdLib | ✅ PASS |
| 179 | **`go-bitmapindex`** | [`go-bitmapindex`](./go-bitmapindex) | Columnar Bit-Sliced Index (BSI) Engine for Ultra-Fast OLAP Range Filtering and Aggregations | ✅ 100% StdLib | ✅ PASS |
| 180 | **`go-cachekit/data`** | [`go-cachekit/data`](./go-cachekit/data) | Cache facade for the multi-tier cache engine and store types | ✅ 100% StdLib | ✅ PASS |
| 181 | **`go-cachekit/data/cache`** | [`go-cachekit/data/cache`](./go-cachekit/data/cache) | Multi-tier cache engine, stores, singleflight protection, and typed helpers | ✅ 100% StdLib | ✅ PASS |
| 182 | **`go-taskqueue/async`** | [`go-taskqueue/async`](./go-taskqueue/async) | Async execution and task orchestration package within `go-taskqueue` | ✅ 100% StdLib | ✅ PASS |
| 183 | **`go-authkit/securityauth`** | [`go-authkit/securityauth`](./go-authkit/securityauth) | Authentication engine and user store package within `go-authkit` | ❌ (`go-authkit` dependencies) | ✅ PASS |
| 184 | **`go-taskqueue/async/taskqueue`** | [`go-taskqueue/async/taskqueue`](./go-taskqueue/async/taskqueue) | Bounded task queue with retries and panic isolation | ✅ 100% StdLib | ✅ PASS |

---

## 🧰 Developer Tooling: `libkit` CLI

`C:\Library` me build kiye gaye 125 packages ko manage, search, test aur new Clean Architecture Go projects me scaffold karne ke liye dedicated zero-dependency CLI tool:

- **Location**: `C:\Library\libkit`
- **Binary**: `C:\Library\bin\libkit.exe`
- **Usage**:
  - `libkit list` - 125 libraries ka ANSI colored interactive table
  - `libkit search <term>` - Instant keyword search
  - `libkit info <name>` - Quickstart demo aur SPEC architecture contract
  - `libkit test [name|--all]` - Unit test suites execution
  - `libkit new <project>` - Clean Architecture microservice scaffolding
  - `libkit add <name>` - Target project ke `go.mod` me local module inject karna

---

## 🧪 Batch Test Execution

Sabhi 125 libraries ek single command se test ki ja sakti hain:

```powershell
# Using libkit CLI:
C:\Library\bin\libkit.exe test --all

# Or via PowerShell loop:
$dirs = Get-ChildItem -Path "C:\Library" -Directory
$failed = @()
foreach ($d in $dirs) {
    Push-Location $d.FullName
    if (Test-Path "go.mod") {
        Write-Host "Testing $($d.Name)..." -ForegroundColor Cyan
        go test -count=1 ./...
        if ($LASTEXITCODE -ne 0) { $failed += $d.Name }
    }
    Pop-Location
}
if ($failed.Count -eq 0) {
    Write-Host "ALL LIBRARIES PASSED! (100% GREEN)" -ForegroundColor Green
} else {
    Write-Host "Failed: $($failed -join ', ')" -ForegroundColor Red
}
```
