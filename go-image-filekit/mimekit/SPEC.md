# SPEC: go-image-filekit/mimekit

## 1. Metadata
- **Module**: `go-image-filekit`
- **Package**: `mimekit`
- **Location**: `C:\Library\go-image-filekit\mimekit`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `net/http`, `bytes`, `io`, `strings`, `errors`).

---

## 2. Core Capabilities
1. **Magic Byte Content Dissection**:
   - Inspects the leading 512 bytes of binary payloads to detect genuine file types regardless of spoofed extensions.
2. **Polyglot & Binary Executable Protection**:
   - Identifies binary executables (Windows PE/MZ, Linux ELF, macOS Mach-O, Unix shebang `#!\...`) to prevent malicious upload attacks.
3. **Stream Rewinding via `io.MultiReader`**:
   - `SniffReader(r)` reads 512 bytes for inspection and reconstructs a non-destructive stream allowing downstream handlers (S3 uploaders, DB savers) to read from byte 0.
4. **Convenience Security Checkers**:
   - `IsSafeImage(data)`: Validates that uploaded content is an authentic image and not an executable disguise.

---

## 3. Public API
```go
type ContentType struct {
    MIME         string
    Extension    string
    IsImage      bool
    IsDocument   bool
    IsArchive    bool
    IsAudioVideo bool
    IsExecutable bool
}

func Sniff(data []byte) ContentType
func SniffReader(r io.Reader) (ContentType, io.Reader, error)
func IsSafeImage(data []byte) bool
func IsExecutable(data []byte) bool
```
