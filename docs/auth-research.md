# CLI Auth Methods Research

> Research on authentication and licensing methods for paid CLI tools
> Date: 2026-02-09
> Purpose: Evaluate options for file-deduplicator monetization

---

## Recommended Approach: Gumroad License Keys

**Best fit for file-deduplicator because:**
- Already planning to use Gumroad for distribution
- Built-in license key generation
- Simple API verification
- No self-hosted infrastructure needed
- Affordable (free until you make sales)

### Gumroad License Verification Flow

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   User      │────▶│   CLI Tool   │────▶│  Gumroad    │
│  (license)  │◀────│  (verify)    │◀────│    API      │
└─────────────┘     └──────────────┘     └─────────────┘
```

### API Endpoint

```
POST https://api.gumroad.com/v2/licenses/verify
Content-Type: application/x-www-form-urlencoded

product_id=YOUR_PRODUCT_ID
license_key=USER_LICENSE_KEY
increment_uses_count=true
```

### Response Format

```json
{
  "success": true,
  "uses": 3,
  "purchase": {
    "id": "purchase_id",
    "product_id": "product_id",
    "email": "user@example.com",
    "created_at": "2024-01-01T00:00:00Z",
    "test": false
  }
}
```

### Go Implementation Sketch

```go
type LicenseVerifier struct {
    ProductID string
    APIEndpoint string
}

func (v *LicenseVerifier) VerifyLicense(licenseKey string) (bool, error) {
    data := url.Values{}
    data.Set("product_id", v.ProductID)
    data.Set("license_key", licenseKey)
    data.Set("increment_uses_count", "true")
    
    resp, err := http.PostForm(v.APIEndpoint, data)
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()
    
    var result struct {
        Success bool `json:"success"`
        Uses    int  `json:"uses"`
        Purchase struct {
            Test bool `json:"test"`
        } `json:"purchase"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return false, err
    }
    
    // Skip verification limit for test purchases
    if result.Purchase.Test {
        return true, nil
    }
    
    return result.Success, nil
}
```

### Device Activation Pattern

To prevent license sharing, implement device fingerprinting:

```go
func getMachineID() string {
    // Combine hostname + MAC address hash
    hostname, _ := os.Hostname()
    // ... get MAC
    return hash(hostname + mac)
}

func (v *LicenseVerifier) VerifyWithDevice(licenseKey string) (bool, error) {
    machineID := getMachineID()
    // Store activated devices in local config or Gumroad metadata
    // Check if device already activated or under limit
}
```

---

## Alternative Options Compared

### 1. Self-Hosted License Server (Keygen.sh)

**Pros:**
- Full control over licensing logic
- Offline capability possible
- Custom feature flags/entitlements

**Cons:**
- Infrastructure to maintain
- $29/month minimum cost
- Overkill for single CLI tool

**Best for:** Teams with multiple products, complex entitlement needs

### 2. Cryptographic License Keys (f-license)

**Pros:**
- No server required for verification
- Works fully offline
- Open source (github.com/furkansenharputlu/f-license)

**Cons:**
- Keys can be cracked if algorithm exposed
- No revocation capability
- No usage analytics

**Best for:** Offline-only tools, simple one-time purchases

### 3. Payhip License Keys

**Pros:**
- Similar to Gumroad
- Simple API verification
- Good for digital products

**Cons:**
- Less developer-focused than Gumroad
- Smaller ecosystem

**Best for:** Already using Payhip for other products

### 4. Stripe + Custom Backend

**Pros:**
- Full control
- Integrated with existing Stripe

**Cons:**
- Build entire licensing system
- Database + server required
- High maintenance

**Best for:** Existing SaaS with Stripe already integrated

---

## Recommendation

**Use Gumroad License Keys with device activation**

### Implementation Plan

1. **Phase 1: Basic License Check**
   - Add `--license-key` flag
   - Verify against Gumroad on first run
   - Store verified flag in config

2. **Phase 2: Device Activation**
   - Generate machine fingerprint
   - Limit to N devices per license (e.g., 3)
   - Allow deactivation via web dashboard

3. **Phase 3: Graceful Degradation**
   - Cache verification for 7 days
   - Offline mode with cached license
   - Warning when verification needed

### Security Considerations

- **Never** embed product secret in binary
- Use only license key verification endpoint
- Rate limit verification attempts
- Cache results to avoid API spam
- Consider code obfuscation for paid tier

### Pricing Strategy

| Tier | Price | Features |
|------|-------|----------|
| Free | £0 | Basic dedup, no perceptual |
| Pro | £15 | Perceptual hashing, priority support |
| Team | £49 | Multi-device, shared licenses |

---

## References

- [Gumroad License API Docs](https://help.gumroad.com/article/280-license-keys)
- [f-license GitHub](https://github.com/furkansenharputlu/f-license)
- [Keygen.sh](https://keygen.sh/)
- [Payhip License Verification](https://help.payhip.com/article/317-software-license-keys)

---

## Next Steps

1. ✅ Implement `--license-key` flag
2. ✅ Add Gumroad verification function
3. ⬜ Create device fingerprinting
4. ⬜ Add license cache/ offline support
5. ⬜ Set up Gumroad product page
6. ⬜ Implement tier feature gating

**Priority:** HIGH - blocks Pro tier launch
