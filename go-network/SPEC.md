# SPEC: go-network

## 1. Scope & Purpose
Unified, high-performance networking, resilience, and traffic management suite for Go microservices.
Consolidates Rate Limiting, HTTP Client with Retries & Circuit Breaker, CORS, Healthchecks, IP Extraction, Webhooks, and Graceful Shutdown under one clean, modular module.

## 2. Public API Sub-packages
- `github.com/shopsilo/go-network/ratelimit`: Token Bucket & Sliding Window Rate Limiter
- `github.com/shopsilo/go-network/httpclient`: Resilient net/http client with full-jitter retries
- `github.com/shopsilo/go-network/circuitbreaker`: Universal function & HTTP circuit breaker
- `github.com/shopsilo/go-network/ipkit`: Real client IP extractor & anti-spoofing
- `github.com/shopsilo/go-network/cors`: Zero-allocation CORS middleware

## 3. Concurrency & Memory Guarantees
- Sub-microsecond evaluation.
- Thread-safe memory stores with background ticker cleanup.
