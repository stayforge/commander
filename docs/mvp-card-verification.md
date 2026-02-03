# MVP Card Verification System

## Overview

This document describes the MVP (Minimum Viable Product) card verification system implemented in Commander. The system validates room access cards against device authorization and time validity constraints using MongoDB as the backend storage.

## Architecture

```
HTTP Request
    ↓
CardVerificationHandler / CardVerificationVguang350Handler
    ↓
CardService (business logic)
    ↓
MongoDB (device & card collections)
    ↓
Verification Result (204 No Content / 200 "0000" / error)
```

## Endpoints

### Standard Card Verification

**Endpoint**: `GET /api/v1/namespaces/:namespace/device/:device_sn/card/:card_number`

**Parameters**:
- `namespace` (string): MongoDB database name (e.g., `org_4e8fb2461d71963a`)
- `device_sn` (string): Device serial number (e.g., `SN20250112001`)
- `card_number` (string): Card number (e.g., `11110011`)

**Success Response**:
- Status: `204 No Content`
- Body: Empty

**Error Responses**:

| Error | Status | Response |
|-------|--------|----------|
| Device not found | 404 | `{"error": "device_not_found", "message": "Device not found", ...}` |
| Device not active | 403 | `{"error": "device_not_active", "message": "Device is not active", ...}` |
| Card not found | 404 | `{"error": "card_not_found", "message": "Card not found", ...}` |
| Card not authorized | 403 | `{"error": "card_not_authorized", "message": "Card is not authorized for this device", ...}` |
| Card expired | 403 | `{"error": "card_expired", "message": "Card has expired", ...}` |
| Card not yet valid | 403 | `{"error": "card_not_yet_valid", "message": "Card is not yet valid", ...}` |

**Example Request**:
```bash
curl -v http://localhost:8080/api/v1/namespaces/org_4e8fb2461d71963a/device/SN20250112001/card/11110011
```

**Example Response (Success)**:
```
HTTP/1.1 204 No Content
```

**Example Response (Error)**:
```json
{
  "error": "device_not_found",
  "message": "Device not found",
  "namespace": "org_4e8fb2461d71963a",
  "device_sn": "SN20250112001",
  "card_number": "11110011",
  "timestamp": "2026-02-03T14:30:00Z"
}
```

### vguang-350 Compatibility Endpoint

**Endpoint**: `GET /api/v1/namespaces/:namespace/device/:device_sn/card/:card_number/vguang-350`

**Parameters**: Same as standard endpoint

**Success Response**:
- Status: `200 OK`
- Content-Type: `text/plain`
- Body: `0000`

**Error Response**: Same as standard endpoint (JSON format)

**Example Request**:
```bash
curl -v http://localhost:8080/api/v1/namespaces/org_4e8fb2461d71963a/device/SN20250112001/card/11110011/vguang-350
```

**Example Response (Success)**:
```
HTTP/1.1 200 OK
Content-Type: text/plain

0000
```

## Verification Logic

### Step 1: Device Validation

- Check if device exists in `devices` collection
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
- `sn`: Serial number (matched against `device_sn` parameter)
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
- `number`: Card number (matched against `card_number` parameter)
- `devices`: Array of device SNs this card is authorized for (empty = not authorized)
- `effective_at`: When the card becomes valid
- `invalid_at`: When the card expires

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

## Logging

All verification operations are logged with level `INFO`. Log entries include:

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
```

## Testing

### Unit Tests

Run model and handler tests:
```bash
go test ./internal/models -v
go test ./internal/services -v
go test ./internal/handlers -v -run Card
```

### Manual Testing with curl

**Test 1: Valid card verification**
```bash
curl -v \
  http://localhost:8080/api/v1/namespaces/org_4e8fb2461d71963a/device/SN20250112001/card/11110011

# Expected: 204 No Content
```

**Test 2: Device not found**
```bash
curl -v \
  http://localhost:8080/api/v1/namespaces/org_4e8fb2461d71963a/device/INVALID_SN/card/11110011

# Expected: 404 with error_code "device_not_found"
```

**Test 3: Card expired**
```bash
# First, insert a card with past invalid_at in MongoDB:
db.cards.insertOne({
  number: "expired_card",
  devices: ["SN20250112001"],
  effective_at: ISODate("2020-01-01T00:00:00Z"),
  invalid_at: ISODate("2020-01-02T00:00:00Z")
})

curl -v \
  http://localhost:8080/api/v1/namespaces/org_4e8fb2461d71963a/device/SN20250112001/card/expired_card

# Expected: 403 with error_code "card_expired"
```

**Test 4: vguang-350 compatibility**
```bash
curl -v \
  http://localhost:8080/api/v1/namespaces/org_4e8fb2461d71963a/device/SN20250112001/card/11110011/vguang-350

# Expected: 200 OK with plain text body "0000"
```

## Troubleshooting

### Common Issues

**Issue**: Card verification always returns 404 (device not found)
- **Solution**: Verify device exists in MongoDB `devices` collection with correct `sn` field
- **Check**: `db.devices.findOne({sn: "SN20250112001"})`

**Issue**: Card verification returns 403 (device not active)
- **Solution**: Update device status to `"active"` in MongoDB
- **Check**: `db.devices.updateOne({sn: "SN20250112001"}, {$set: {status: "active"}})`

**Issue**: Card verification returns 403 (card not authorized)
- **Solution**: Verify device SN is in card's `devices` array
- **Check**: `db.cards.findOne({number: "11110011"})` and confirm `"SN20250112001"` is in `devices` array

**Issue**: Card verification returns 403 (card expired)
- **Solution**: Check `effective_at` and `invalid_at` values are correct
- **Check**: `db.cards.findOne({number: "11110011"})` to view timestamps

### Performance Notes

- Response time typically <100ms for valid cards (single device/card lookup)
- MongoDB connection pooling is managed by the KV layer
- No caching is implemented in MVP (can be added in future versions)

## Future Enhancements

1. **Caching Layer**: Add Redis cache for hot cards and device statuses
2. **Batch Verification**: Support checking multiple cards in single request
3. **Audit Logging**: Track all verification attempts for compliance
4. **Rate Limiting**: Prevent brute force attacks
5. **Extended KV Interface**: Expose card operations through KV abstraction layer
6. **Device Management API**: Add endpoints to manage devices and cards
7. **Webhook Notifications**: Notify external systems of verification results
8. **Multi-tenant Support**: Enforce namespace isolation at application level

## Architecture Notes

The MVP implementation uses direct MongoDB access (not through the KV abstraction layer) for:
- Complex MongoDB queries (array containment, time range queries)
- Better error handling and logging specific to card verification
- Flexibility in data structure handling

This design allows rapid iteration on the MVP without restructuring the KV interface. Future versions can integrate this functionality into the KV layer when the query requirements stabilize.
