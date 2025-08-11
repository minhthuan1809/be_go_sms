# 📱 Hướng dẫn sử dụng SMS Gateway

## 🚀 Khởi động

```bash
# Build ứng dụng
go build -o sms-gateway ./cmd/server

# Chạy server
./sms-gateway
```

## 📋 Các bước test

### 1. Kiểm tra trạng thái modem
```bash
curl "http://localhost:8080/port/status?port=/dev/ttyUSB0"
```

### 2. Health check
```bash
curl http://localhost:8080/health
```

### 3. Gửi SMS test
```bash
# Sử dụng script tự động
./test_real_sms.sh

# Hoặc gửi thủ công
curl -X POST http://localhost:8080/send \
  -H "Content-Type: application/json" \
  -d '{
    "port": "/dev/ttyUSB0",
    "baud_rate": 115200,
    "to": "YOUR_REAL_PHONE_NUMBER",
    "message": "Test SMS từ Gateway",
    "timeout": 30
  }'
```

## ⚠️ Lưu ý quan trọng

### ❌ KHÔNG dùng số giả:
- `0123456789` ❌
- `1234567890` ❌
- `9999999999` ❌

### ✅ Dùng số thật:
- `0123456789` (số thật của bạn) ✅
- `+84123456789` (định dạng quốc tế) ✅

## 🔧 Troubleshooting

### Lỗi CMS ERROR:
1. **Kiểm tra số điện thoại** - phải là số thật
2. **Kiểm tra SIM card** - phải có tiền và tín hiệu
3. **Kiểm tra tin nhắn** - không quá 160 ký tự

### Lỗi permission:
```bash
sudo usermod -a -G uucp $USER
# Logout và login lại
```

### Lỗi port không tìm thấy:
```bash
ls /dev/ttyUSB*
```

## 📊 Trạng thái modem

Khi test thành công, bạn sẽ thấy:
- `+CPIN: READY` - SIM sẵn sàng
- `+CREG: 2,1,...` - Đã đăng ký mạng
- `+CSQ: x,y` - Tín hiệu tốt (x > 10)
- `OK` - Gửi SMS thành công

## 🎯 Ví dụ thành công

```json
{
  "steps": [
    "AT -> OK",
    "ATE0 -> OK", 
    "AT+CSQ -> +CSQ: 18,99 OK",
    "AT+CPIN? -> +CPIN: READY OK",
    "AT+CREG? -> +CREG: 2,1,\"32F2\",\"05104201\" OK",
    "AT+CMGF=1 -> OK",
    "AT+CSMP=17,167,0,0 -> OK",
    "AT+CMGS=\"0123456789\" -> >",
    "Final response -> +CMGS: 1 OK"
  ],
  "success": true,
  "message_id": "1",
  "duration": "1.2s"
}
```
