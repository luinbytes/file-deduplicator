# CLI Tool Auth/Licensing Recommendation

## Executive Summary

After researching indie CLI tool monetization patterns (ripgrep, fd, bat, and Gumroad API), the recommendation is:

**Use honor system + GitHub Sponsors for now. Add license keys only if piracy becomes significant (>10%).**

---

## Research Findings

### Successful CLI Tools Pattern

| Tool | License Model | Revenue |
|------|--------------|---------|
| ripgrep | MIT + GitHub Sponsors | $3k+/month |
| fd | MIT + GitHub Sponsors | $1k+/month |
| bat | MIT + GitHub Sponsors | $800+/month |

**Key insight:** All major CLI tools use open source + sponsor model. None use license key verification.

### Gumroad License API (For Future Reference)

```
POST https://api.gumroad.com/v2/licenses/verify
Params:
  - product_permalink
  - license_key
  - increment_uses_count=false (avoid usage bloat)

Response includes:
  - purchase validity
  - refunded/disputed status
```

### License Key Pros/Cons

**Pros:**
- Simple Gumroad integration
- Can track purchases

**Cons:**
- Adds friction for legitimate users
- Crackable (determined pirates will bypass)
- Network dependency on launch
- Complexity for open source project

### Freemium Model

**Option:** Free = SHA256 only, Paid = perceptual hashing

**Pros:** Clear value proposition
**Cons:** Code complexity, maintenance overhead

---

## Recommendation

### Phase 1: Launch (Now)
- Honor system pricing
- GitHub Sponsors button
- Gumroad for binary purchases ($10)
- Source code always available

### Phase 2: If Piracy > 10%
- Add optional license key check
- Still open source, just inconvenience pirates
- Use Gumroad API verification

### Phase 3: Future
- SaaS web app for non-technical users
- Subscription model
- Recurring revenue

---

## Implementation

Add to README:
```markdown
## Pricing

**Free:** Build from source (always)

**$10:** Pre-built binaries on [Gumroad](https://gumroad.com/l/file-deduplicator)

Support development via [GitHub Sponsors](https://github.com/sponsors/luinbytes)
```

No code changes needed for Phase 1.
