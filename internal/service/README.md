# SMS Service Architecture

Dự án đã được tái cấu trúc để tách logic xử lý SMS thành các module nhỏ, dễ quản lý và bảo trì.

## Cấu trúc File

### 1. `sms_service.go` - Service chính
- **Trách nhiệm**: Interface chính cho SMS operations
- **Chức năng**: 
  - Khởi tạo các module con
  - Cung cấp API chính cho việc gửi SMS
  - Quản lý port availability

### 2. `at_commands.go` - Xử lý AT Commands
- **Trách nhiệm**: Quản lý tất cả các lệnh AT
- **Chức năng**:
  - Test kết nối modem
  - Kiểm tra tín hiệu, SIM, mạng
  - Thiết lập chế độ SMS (text/PDU)
  - Khởi tạo gửi SMS

### 3. `serial_manager.go` - Quản lý Serial Port
- **Trách nhiệm**: Quản lý kết nối serial port
- **Chức năng**:
  - Mở/đóng port
  - Kiểm tra availability
  - Quản lý mutex để tránh xung đột
  - Cấu hình port settings

### 4. `error_handler.go` - Xử lý Lỗi
- **Trách nhiệm**: Phân tích và xử lý lỗi SMS
- **Chức năng**:
  - Phân tích CMS error codes
  - Trích xuất message ID
  - Cung cấp thông báo lỗi thân thiện
  - Phân loại lỗi (hết tiền, mạng, số điện thoại)

### 5. `balance_checker.go` - Kiểm tra Số dư
- **Trách nhiệm**: Kiểm tra số dư SIM
- **Chức năng**:
  - Gửi USSD command kiểm tra số dư
  - Phân tích response
  - Xác định có đủ tiền gửi SMS không

### 6. `sms_sender.go` - Logic Gửi SMS
- **Trách nhiệm**: Điều phối toàn bộ quá trình gửi SMS
- **Chức năng**:
  - Kiểm tra trạng thái modem
  - Kiểm tra số dư
  - Reset modem nếu cần
  - Thực hiện gửi SMS
  - Xử lý fallback PDU mode

## Lợi ích của Cấu trúc Mới

### 1. **Tách biệt trách nhiệm**
- Mỗi file có một trách nhiệm rõ ràng
- Dễ dàng tìm và sửa lỗi
- Code dễ đọc và hiểu

### 2. **Dễ bảo trì**
- Thay đổi logic AT commands không ảnh hưởng đến error handling
- Có thể thay đổi cách kiểm tra số dư mà không ảnh hưởng đến SMS sending
- Dễ dàng thêm tính năng mới

### 3. **Dễ test**
- Có thể test từng module riêng biệt
- Mock các dependency dễ dàng
- Unit test hiệu quả hơn

### 4. **Tái sử dụng**
- Các module có thể được sử dụng độc lập
- Dễ dàng mở rộng cho các tính năng khác

## Cách sử dụng

```go
// Tạo service
smsService := service.NewSMSService()

// Gửi SMS
steps, messageID, err := smsService.SendSMSViaAT(ctx, "/dev/ttyUSB0", 115200, "+84325397277", "Test message")
if err != nil {
    log.Printf("Error: %v", err)
    return
}

log.Printf("SMS sent successfully. Message ID: %s", messageID)
```

## Lưu ý về Lỗi CMS

Dựa trên log lỗi bạn cung cấp, có vẻ như vấn đề là:
- SIM đã được nạp tiền nhưng vẫn gặp lỗi CMS
- Có thể do:
  1. Số điện thoại không đúng định dạng
  2. Mạng chưa ổn định sau khi reset
  3. Cần thời gian để SIM nhận diện số dư mới

Để debug, bạn có thể:
1. Kiểm tra lại số điện thoại đích
2. Đợi thêm vài phút sau khi nạp tiền
3. Thử reset modem và kiểm tra lại
