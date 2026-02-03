# MVP Card Verification System

## Overview

This document describes the MVP (Minimum Viable Product) card verification system implemented in Commander. The system validates room access cards against device authorization and time validity constraints using MongoDB as the backend storage.

## Architecture

```
HTTP Request (POST)
    ↓
CardVerificationHandler / CardVerificationVguangHandler
    ↓
CardService (business logic)
    ↓
MongoDB (device & card collections)
    ↓
Verification Result (204 / 200 "code=0000" / Status Only)
```

## Endpoints

### Standard Card Verification (New API)

**Endpoint**: `POST /api/v1/namespaces/:namespace`

**Headers**:
- `X-Device-SN` (required): Device serial number (e.g., `SN20250112001`)

**Body**: Plain text card number (e.g., `11110011`)

**Success Response**:
- Status: `204 No Content`
- Body: Empty

**Error Response**:
- Status: `400 Bad Request` (missing header or empty body)
- Status: `403 Forbidden` (not authorized/expired)
- Status: `404 Not Found` (device/card not found)
- Body: Empty (error logged to console)

**Example Request**:
```bash
curl -X POST \
  -H "X-Device-SN: SN20250112001" \
  -d "11110011" \
  http://localhost:8080/api/v1/namespaces/org_4e8fb2461d71963a
```

**Example Response (Success)**:
```
HTTP/1.1 204 No Content
```

**Example Response (Error)**:
```
HTTP/1.1 404 Not Found
(no body)
```

---

### vguang-m350 Compatibility (Legacy API)

**Endpoint**: `POST /api/v1/namespaces/:namespace/device/:device_name`

**Body**: Plain text or binary card number

**Card Number Processing**:
- If alphanumeric: use as-is (converted to uppercase)
- Otherwise: reverse bytes and convert to hex (uppercase)

**Success Response**:
- Status: `200 OK`
- Content-Type: `text/plain`
- Body: `code=0000`

**Error Response**:
- Status: `404 Not Found`
- Body: Empty (error logged to console)

**Example Request (Plain Text)**:
```bash
curl -X POST \
  -d "11110011" \
  http://localhost:8080/api/v1/namespaces/org_4e8fb2461d71963a/device/SN20250112001
```

**Example Response (Success)**:
```
HTTP/1.1 200 OK
Content-Type: text/plain

code=0000
```

**Example Request (Binary)**:
```bash
# Binary card data will be reversed and converted to hex
echo -ne '\x01\x02\x03\x04' | curl -X POST -d @- \
  http://localhost:8080/api/v1/namespaces/org_4e8fb2461d71963a/device/SN20250112001
```

---

## Verification Logic

### Step 1: Device Validation

- Check if device exists in `devices` collection (by `sn` field)
- Verify device's `status` field equals `"active"`
- Abort if device not found or inactive

### Step 2: Card Lookup

- Find card by `number` field in `cards` collection
- Abort if card not found

### Step 3: Device Authorization

- Check if `device_sn` exists in card's `devices` array
- Abort if empty array or device SN not in list

### Step 4: Time-Based Validity

- Check if current time is between `effective_at` and `invalid_at`
- Apply ±60 seconds tolerance (for NTP clock drift)
- Calculation:
  - Valid if: `now > (effective_at - 60s)` AND `now < (invalid_at + 60s)`

**Timeline Example**:
```
effective_at: 2026-08-25 15:00:00
invalid_at:   2026-08-25 16:00:00
tolerance:    ±60 seconds

Valid range: 2026-08-25 14:59:00 to 2026-08-25 16:01:00
```

---

## MongoDB Data Structures

### Devices Collection

```json
{
  "_id": "0bc19267-4785-46df-adb8-c7924db526dd",
  "tenant_id": "7917a81bb17d42eda29e39a0389d2ed9",
  "device_id": "ccc",
  "sn": "SN20250112001",
  "display_name": "一楼温度传感器",
  "price_id": "price_basic_monthly",
  "status": "active",
  "metadata": {
    "model": "TEMP-SENSOR-X1",
    "installation_date": "2025-01-10",
    "last_maintenance": "2025-12-15"
  },
  "created_at": { "$date": "2026-01-12T09:15:47.991Z" },
  "updated_at": { "$date": "2026-01-12T09:15:47.991Z" }
}
```

**Key Fields**:
- `sn`: Serial number (matched against `X-Device-SN` header or `:device_name` URL parameter)
- `status`: Device status (must be `"active"`)

### Cards Collection

```json
{
  "_id": "f7db0bfc-73e5-4888-9355-9f57b0b28d5e",
  "organization_id": "org_4e8fb2461d71963a",
  "number": "11110011",
  "display_name": "aaaacccc",
  "devices": ["device-001", "SN20250112001"],
  "effective_at": { "$date": "2026-08-25T15:00:00.000Z" },
  "invalid_at": { "$date": "2026-08-25T16:00:00.000Z" },
  "barcode_type": "qrcode",
  "created_at": { "$date": "2026-01-12T12:45:32.856Z" },
  "updated_at": { "$date": "2026-01-12T12:45:32.856Z" }
}
```

**Key Fields**:
- `number`: Card number (matched against request body)
- `devices`: Array of device SNs this card is authorized for (empty = not authorized)
- `effective_at`: When the card becomes valid
- `invalid_at`: When the card expires

---

## Configuration

### Environment Variables

```bash
# Database backend (required)
DATABASE=mongodb

# MongoDB connection URI
MONGODB_URI=mongodb://user:password@localhost:27017/?authSource=admin

# Server port (default: 8080)
SERVER_PORT=8080

# Server environment
ENVIRONMENT=STANDARD
```

### Example .env File

```bash
DATABASE=mongodb
MONGODB_URI=mongodb://admin:password@mongodb.example.com:27017/?authSource=admin
SERVER_PORT=8080
ENVIRONMENT=STANDARD
```

---

## Logging

All verification operations are logged to console with `[CardVerification]` prefix.

**Device Verification**:
```
[CardVerification] Device verified: namespace=org_test, device_sn=SN20250112001, device_id=ccc, status=active
```

**Success**:
```
[CardVerification] SUCCESS: namespace=org_test, card_number=11110011, device_sn=SN20250112001, card_id=f7db0bfc-73e5-4888-9355-9f57b0b28d5e, effective=2026-08-25T15:00:00Z, invalid=2026-08-25T16:00:00Z
```

**Failures**:
```
[CardVerification] Device check failed: namespace=org_test, device_sn=unknown, error=device not found
[CardVerification] Device not active: namespace=org_test, device_sn=SN001, status=inactive
[CardVerification] Card expired: namespace=org_test, card_number=11110011, device_sn=SN001, invalid_at=2026-08-25T16:00:00Z, current_time=2026-08-25T16:02:00Z
[CardVerification:vguang] Failed to read body: namespace=org_test, device_name=SN001, error=<error>
```

---

## Testing

### Unit Tests

Run model and handler tests:
```bash
go test ./internal/models -v
go test ./internal/services -v
go test ./internal/handlers -v -run Card
```

### Manual Testing with curl

**Test 1: Standard API - Valid verification**
```bash
curl -X POST \
  -H "X-Device-SN: SN20250112001" \
  -d "11110011" \
  http://localhost:8080/api/v1/namespaces/org_4e8fb2461d71963a

# Expected: 204 No Content
```

**Test 2: Standard API - Missing header**
```bash
curl -X POST \
  -d "11110011" \
  http://localhost:8080/api/v1/namespaces/org_4e8fb2461d71963a

# Expected: 400 Bad Request (no body)
```

**Test 3: Standard API - Device not found**
```bash
curl -X POST \
  -H "X-Device-SN: INVALID_SN" \
  -d "11110011" \
  http://localhost:8080/api/v1/namespaces/org_4e8fb2461d71963a

# Expected: 404 Not Found (no body)
```

**Test 4: vguang-m350 - Valid verification (plain text)**
```bash
curl -X POST \
  -d "11110011" \
  http://localhost:8080/api/v1/namespaces/org_4e8fb2461d71963a/device/SN20250112001

# Expected: 200 OK
# Body: code=0000
```

**Test 5: vguang-m350 - Card not found**
```bash
curl -X POST \
  -d "nonexistent_card" \
  http://localhost:8080/api/v1/namespaces/org_4e8fb2461d71963a/device/SN20250112001

# Expected: 404 Not Found (no body)
```

**Test 6: vguang-m350 - Binary card data**
```bash
# Simulate binary card reader output
echo -ne '\x01\x02\x03\x04' | curl -X POST -d @- \
  http://localhost:8080/api/v1/namespaces/org_4e8fb2461d71963a/device/SN20250112001

# Card number will be: 04030201 (reversed hex)
# Expected: 200 OK or 404 (depending on card existence)
```

---

## Troubleshooting

### Common Issues

**Issue**: Standard API returns 400 Bad Request
- **Solution**: Verify `X-Device-SN` header is present and body is not empty
- **Check**: `curl -v` to inspect headers and body

**Issue**: vguang API always returns 404
- **Cause**: Device or card not found in MongoDB
- **Check**: Verify device exists with correct SN: `db.devices.findOne({sn: "SN20250112001"})`
- **Check**: Verify card exists with correct number: `db.cards.findOne({number: "11110011"})`

**Issue**: Card verification returns 403 (not authorized)
- **Cause**: Card not authorized for device or time not valid
- **Check**: Verify device SN is in card's `devices` array
- **Check**: Verify current time is within effective/invalid date range (±60s)

**Issue**: Response body is empty instead of JSON
- **Note**: This is expected behavior for errors in the new system - check HTTP status code and console logs

---

## Future Enhancements

1. **Caching Layer**: Add Redis cache for hot cards and device statuses
2. **Batch Verification**: Support checking multiple cards in single request
3. **Audit Logging**: Track all verification attempts for compliance
4. **Rate Limiting**: Prevent brute force attacks
5. **Extended KV Interface**: Expose card operations through KV abstraction layer
6. **Device Management API**: Add endpoints to manage devices and cards
7. **Webhook Notifications**: Notify external systems of verification results
8. **Multi-tenant Support**: Enforce namespace isolation at application level

---

## API Summary

| Method | Endpoint | Purpose | Success | Error |
|--------|----------|---------|---------|-------|
| POST | `/api/v1/namespaces/:namespace` | Standard verification | 204 No Content | Status only |
| POST | `/api/v1/namespaces/:namespace/device/:device_name` | vguang-m350 compatibility | 200 + "code=0000" | 404 |

---

## Notes

- **Error responses contain no body** for security - errors are logged to console
- **vguang-m350 response must be exact**: `"code=0000"` (not just `"0000"`)
- **Card number in request body is always plain text** (even for vguang)
- **Binary handling only in vguang endpoint** - reverses bytes and converts to hex
- **±60 second tolerance** accounts for NTP clock drift on devices and servers
