# Specification: HPACK Header Compression (`go-hpack`)

## 1. RFC 7541 Wire Format Encodings

1. **Indexed Header Field (`1xxxxxxx`)**:
   Prefix $N=7$. References static table $[1, 61]$ or dynamic table $[62, \dots]$.
2. **Literal with Incremental Indexing (`01xxxxxx`)**:
   Prefix $N=6$. Encodes header name/value and inserts into dynamic table at index 1.
3. **Literal without Indexing (`0000xxxx`)**:
   Prefix $N=4$. Encodes header without modifying dynamic table.
4. **Literal Never Indexed (`0001xxxx`)**:
   Prefix $N=4$. Critical security invariant: never enters dynamic table across proxies.
5. **Dynamic Table Size Update (`001xxxxx`)**:
   Prefix $N=5$. Notifies peer of dynamic table capacity changes.

## 2. Integer Representation (RFC 7541 Section 5.1)
An integer $I$ in an $N$-bit prefix:
- If $I < 2^N - 1$, store in $N$ bits.
- Else, set $N$ bits to all 1s, subtract $2^N - 1$ from $I$, and emit 7-bit chunks with MSB 1 while $I \ge 128$.
