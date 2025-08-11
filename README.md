# SMS Gateway API

SMS Gateway API là một ứng dụng Go cho phép gửi SMS thông qua modem USB. Dự án được tích hợp với Swagger để cung cấp documentation và testing interface.

## 🚀 Tính năng

- ✅ Gửi SMS qua modem USB
- ✅ Kiểm tra trạng thái modem và port
- ✅ Health check API
- ✅ Swagger UI documentation
- ✅ RESTful API endpoints
- ✅ Validation và error handling
- ✅ Logging chi tiết

## 📋 Yêu cầu hệ thống

- Go 1.22 trở lên
- Modem USB hỗ trợ AT commands
- Windows/Linux/macOS

## 🛠️ Cài đặt

### 1. Clone dự án
```bash
git clone <repository-url>
cd sms-gateway
```

### 2. Cài đặt dependencies
```bash
go mod tidy
```

### 3. Cài đặt Swagger CLI (nếu chưa có)
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

## 🏃‍♂️ Cách chạy dự án

### Cách 1: Chạy trực tiếp với go run
```bash
go run src/cmd/server/main.go
```

### Cách 2: Build và chạy binary
```bash
# Build dự án
go build -o sms-gateway.exe src/cmd/server/main.go

# Chạy binary
./sms-gateway.exe
```

### Cách 3: Sử dụng Makefile (nếu có)
```bash
make run
```

## 🌐 Truy cập API

Sau khi chạy thành công, server sẽ khởi động trên port 8080:

### Swagger UI
```
http://localhost:8080/swagger/
```

### API Endpoints

| Method | Endpoint | Mô tả |
|--------|----------|-------|
| GET | `/` | Thông tin API |
| GET | `/api/v1/health` | Health check |
| POST | `/api/v1/sms/send` | Gửi SMS |
| GET | `/api/v1/ports` | Danh sách ports |
| GET | `/api/v1/ports/status` | Trạng thái port |
| GET | `/api/v1/modem/info` | Thông tin modem |


## 📱 Sử dụng API

### 1. Gửi SMS
```bash
curl -X POST http://localhost:8080/api/v1/sms/send \
  -H "Content-Type: application/json" \
  -d '{
    "to": "+84123456789",
    "message": "Hello from SMS Gateway!",
    "port": "COM3"
  }'
```

### 2. Health Check
```bash
curl http://localhost:8080/api/v1/health
```

### 3. Kiểm tra ports
```bash
curl http://localhost:8080/api/v1/ports
```

### 4. Trạng thái port
```bash
curl "http://localhost:8080/api/v1/ports/status?port=COM3"
```

## 🔧 Cấu hình

### File cấu hình
Dự án sử dụng cấu hình mặc định. Bạn có thể tùy chỉnh trong `src/internal/config/`.

### Cấu hình modem
- Default Port: COM3
- Default Baud Rate: 115200
- Timeout: 30 giây

## 📚 Swagger Documentation

### Truy cập Swagger UI
1. Mở trình duyệt web
2. Truy cập: `http://localhost:8080/swagger/`
3. Xem và test các API endpoints

### Cập nhật documentation
Khi thay đổi code, chạy lệnh sau để cập nhật Swagger docs:
```bash
swag init -g src/cmd/server/main.go -o docs
```

## 🧪 Testing

### Test với Swagger UI
1. Mở `http://localhost:8080/swagger/`
2. Chọn endpoint muốn test
3. Click "Try it out"
4. Điền thông tin cần thiết
5. Click "Execute"

### Test với curl
```bash
# Test health check
curl http://localhost:8080/api/v1/health

# Test list ports
curl http://localhost:8080/api/v1/ports
```

## 📁 Cấu trúc dự án

```
sms-gateway/
├── docs/                   # Swagger documentation
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── src/
│   ├── api/               # API layer
│   │   ├── middleware/    # HTTP middleware
│   │   └── router/        # Router configuration
│   ├── cmd/               # Application entry points
│   │   └── server/        # Main server
│   ├── internal/          # Internal packages
│   │   ├── config/        # Configuration
│   │   ├── handler/       # HTTP handlers
│   │   ├── model/         # Data models
│   │   ├── service/       # Business logic
│   │   └── utils/         # Utilities
│   └── pkg/               # Public packages
│       └── validation/    # Validation logic
├── go.mod                 # Go modules
├── go.sum                 # Dependencies checksum
└── README.md              # This file
```

## 🐛 Troubleshooting

### Lỗi thường gặp

1. **Port không tìm thấy**
   - Kiểm tra modem đã kết nối chưa
   - Kiểm tra driver modem
   - Thử port khác

2. **Permission denied**
   - Chạy với quyền admin (Windows)
   - Thêm user vào group dialout (Linux)

3. **Swagger không load**
   - Kiểm tra server đã chạy chưa
   - Chạy lại `swag init` nếu cần

### Logs
Server sẽ in logs chi tiết ra console. Theo dõi logs để debug.

## 🤝 Đóng góp

1. Fork dự án
2. Tạo feature branch
3. Commit changes
4. Push to branch
5. Tạo Pull Request

## 📄 License

MIT License

## 📞 Liên hệ

- Email: [your-email]
- GitHub: [your-github]

---

**Lưu ý**: Đảm bảo modem USB đã được kết nối và cài đặt driver trước khi sử dụng API.
