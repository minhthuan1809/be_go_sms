# SMS Gateway API

SMS Gateway API là một service Go để gửi SMS qua modem USB. API hỗ trợ cả text mode và PDU mode.

## Tính năng

- ✅ Gửi SMS qua modem USB
- ✅ Hỗ trợ text mode và PDU mode
- ✅ Tự động format số điện thoại quốc tế
- ✅ Logging chi tiết
- ✅ Health check và monitoring
- ✅ Quản lý ports và modem info
- ✅ Error handling tốt

## Cài đặt

```bash
# Clone repository
git clone <repository-url>
cd be_go_sms

# Build
go build -o sms-gateway src/cmd/server/main.go

# Chạy server
./sms-gateway
```

## API Endpoints

### 1. Health Check
```bash
GET /api/v1/health
```

Response:
```json
{
  "status": "healthy",
  "version": "2.0.0",
  "timestamp": "2025-08-11T23:38:06+07:00",
  "uptime": "9.916701865s"
}
```

### 2. Gửi SMS
```bash
POST /api/v1/sms/send
```

Request:
```json
{
  "to": "+84325397277",
  "message": "Test message",
  "port": "/dev/ttyUSB0",
  "baud_rate": 115200,
  "mode": "text",
  "timeout": 30
}
```

Response:
```json
{
  "success": true,
  "message_id": "SMS_1754930303",
  "steps": [
    "Opening port /dev/ttyUSB0 at 115200 baud",
    "Port opened successfully",
    "Testing modem with AT command",
    "Disabling echo",
    "Checking network registration",
    "Setting SMS mode to text",
    "Sending SMS to +84325397277",
    "Sending message text",
    "SMS sent successfully"
  ],
  "duration": "816.820393ms",
  "mode": "text",
  "to": "+84325397277",
  "message": "OK",
  "timestamp": "2025-08-11T23:38:23+07:00"
}
```

### 3. Danh sách Ports
```bash
GET /api/v1/ports
```

Response:
```json
{
  "success": true,
  "data": ["/dev/ttyUSB0", "/dev/ttyUSB1"],
  "message": "Available ports retrieved successfully",
  "timestamp": "2025-08-11T23:38:26+07:00"
}
```

### 4. Kiểm tra Port Status
```bash
GET /api/v1/ports/status?port=/dev/ttyUSB0
```

Response:
```json
{
  "port": "/dev/ttyUSB0",
  "available": true,
  "in_use": false
}
```

### 5. Modem Info
```bash
GET /api/v1/modem/info?port=/dev/ttyUSB0
```

## Cấu hình

### Environment Variables

```bash
# Server
SERVER_ADDRESS=:8080
SERVER_READ_TIMEOUT=10
SERVER_WRITE_TIMEOUT=10
SERVER_IDLE_TIMEOUT=120

# Modem
MODEM_DEFAULT_PORT=/dev/ttyUSB0
MODEM_DEFAULT_BAUDRATE=115200
MODEM_TIMEOUT=30

# SMS
SMS_MAX_LENGTH=160
SMS_DEFAULT_TIMEOUT=30
SMS_RETRY_COUNT=3
SMS_RETRY_DELAY=2
```

## Sử dụng

### Gửi SMS cơ bản
```bash
curl -X POST http://localhost:8080/api/v1/sms/send \
  -H "Content-Type: application/json" \
  -d '{
    "to": "+84325397277",
    "message": "Hello from SMS Gateway"
  }'
```

### Gửi SMS với cấu hình đầy đủ
```bash
curl -X POST http://localhost:8080/api/v1/sms/send \
  -H "Content-Type: application/json" \
  -d '{
    "to": "+84325397277",
    "message": "Test message",
    "port": "/dev/ttyUSB0",
    "baud_rate": 115200,
    "mode": "text",
    "timeout": 30
  }'
```

### Gửi SMS với PDU mode
```bash
curl -X POST http://localhost:8080/api/v1/sms/send \
  -H "Content-Type: application/json" \
  -d '{
    "to": "+84325397277",
    "message": "Test PDU",
    "mode": "pdu"
  }'
```

## Troubleshooting

### 1. Port không tồn tại
```json
{
  "success": false,
  "error": "port /dev/ttyUSB0 does not exist"
}
```

### 2. Modem không phản hồi
```json
{
  "success": false,
  "error": "modem not responding: timeout waiting for: OK"
}
```

### 3. Timeout khi gửi SMS
```json
{
  "success": false,
  "error": "failed to send SMS: timeout waiting for: OK"
}
```

## Logs

Server log chi tiết cho mỗi request:

```
2025/08/11 23:38:19 SMS request received - To: +84325397277, Port: /dev/ttyUSB0, Message length: 24
2025/08/11 23:38:19 Starting SMS send process - Port: /dev/ttyUSB0, BaudRate: 0, To: +84325397277, Mode: text
2025/08/11 23:38:19 Using default baud rate: 115200
2025/08/11 23:38:19 Using default timeout: 30
2025/08/11 23:38:19 Calling SMS client SendViaText...
2025/08/11 23:38:19 SMS Client: Starting SendViaText - Port: /dev/ttyUSB0, BaudRate: 115200, To: +84325397277
2025/08/11 23:38:19 SMS Client: Port opened successfully
2025/08/11 23:38:19 SMS Client: Initializing modem...
2025/08/11 23:38:19 SMS Client: Testing modem with AT command...
2025/08/11 23:38:19 SMS Client: Sending AT command: AT, expecting: OK
2025/08/11 23:38:19 SMS Client: Waiting for response with timeout: 5s
2025/08/11 23:38:19 SMS Client: Starting readUntil, timeout: 5s, expected: OK
2025/08/11 23:38:19 SMS Client: Found expected response: OK
2025/08/11 23:38:19 SMS Client: Received response: OK
2025/08/11 23:38:19 SMS Client: Command successful
```

## Yêu cầu hệ thống

- Go 1.19+
- Modem USB hỗ trợ AT commands
- SIM card có credit
- Quyền truy cập serial ports

## License

MIT License
