# 構建階段
FROM golang:1.22-alpine AS builder

# 設置工作目錄
WORKDIR /app

# 複製 go.mod 和 go.sum
COPY go.mod ./

# 下載依賴
RUN go mod download

# 複製源代碼
COPY . .

# 構建應用
RUN CGO_ENABLED=0 GOOS=linux go build -o pochacco

# 運行階段
FROM alpine:latest

# 安裝 ca-certificates，用於 HTTPS 請求
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 從構建階段複製二進制文件
COPY --from=builder /app/pochacco .
# 複製 .env 文件（如果需要的話）
# COPY --from=builder /app/.env .

# 設置執行權限
RUN chmod +x ./pochacco

# 設置入口點
CMD ["bash"] 