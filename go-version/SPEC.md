# Technical Specification: go-version

## 1. Domain & Purpose
`go-version` is a Semantic Versioning (SemVer 2.0.0) specification parser, comparator, and constraint matching engine for Go. It powers API version negotiation, mobile application minimum required version checks (e.g. Flutter client updates), plugin compatibility matrices, and feature flag release toggles.

## 2. Architecture & Design Principles
- **SemVer 2.0.0 Compliance**: Strict implementation of Major.Minor.Patch numbers, dot-separated pre-release identifiers, and build metadata.
- **Accurate Precedence Rules**: Correctly evaluates pre-release precedence (normal > pre-release, numeric vs non-numeric ASCII sorting) while treating build metadata as informational only.
- **Rich Constraint Expressions**: Supports comparison operators (`=`, `!=`, `>`, `>=`, `<`, `<=`), comma-separated conjunctions, Caret ranges (`^1.2.3`), and Tilde ranges (`~1.2.3`).
- **Zero Heavy Dependencies**: Pure Go standard library (`strings`, `strconv`, `fmt`).

## 3. API Surface
- `type Version struct { Major, Minor, Patch int; Prerelease, BuildMetadata string }`
- `func Parse(v string) (*Version, error)`
- `func MustParse(v string) *Version`
- `func (v *Version) Compare(other *Version) int`
- `func (v *Version) GreaterThan(other *Version) bool`
- `func (v *Version) GreaterThanOrEqual(other *Version) bool`
- `func (v *Version) LessThan(other *Version) bool`
- `func (v *Version) LessThanOrEqual(other *Version) bool`
- `func (v *Version) Equal(other *Version) bool`
- `type Constraint struct`
- `func ParseConstraint(expr string) (*Constraint, error)`
- `func (c *Constraint) Check(v *Version) bool`
- `func (c *Constraint) CheckString(vStr string) bool`

## 4. Error Handling
- `ErrInvalidVersion`: Non-compliant SemVer string format.
- `ErrInvalidConstraint`: Invalid comparison expression.

## 5. Verification & Testing
- Unit tests verify SemVer parsing, complete pre-release precedence comparisons, build metadata isolation, and constraint range matching (caret, tilde, and compound bounds).
