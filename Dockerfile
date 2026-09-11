# 构建阶段
FROM golang:1.21-bookworm AS builder

# 设置构建参数
ARG DEBIAN_FRONTEND=noninteractive
ARG GOPROXY=https://goproxy.cn,direct

# 安装构建依赖
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    cmake \
    git \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 复制源代码
COPY . .

# 1. 编译 stable-diffusion.cpp 动态库
WORKDIR /app/stable-diffusion.cpp
# 初始化 git submodules (ggml 等依赖)
# 如果已经在主项目下载了 submodule 则不需要执行
RUN git submodule update --init --recursive || true
RUN mkdir -p build && cd build && \
    cmake .. \
    -DSD_BUILD_SHARED_LIBS=ON \
    -DCMAKE_BUILD_TYPE=Release \
    -DGGML_NATIVE=OFF && \
    cmake --build . --config Release --parallel $(nproc)

# 2. 编译 Go 应用
WORKDIR /app
RUN go build -mod=vendor -ldflags="-s -w" -o sd-server ./examples/server

# 运行时阶段
FROM debian:bookworm-slim

# 安装运行时依赖
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libgomp1 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 从构建阶段复制动态库
# 注意：CMake 将输出放在了 build/bin 目录下
COPY --from=builder /app/stable-diffusion.cpp/build/bin/libstable-diffusion.so /usr/lib/libstable-diffusion.so
# 复制 Go 二进制文件
COPY --from=builder /app/sd-server .

# 设置库路径
ENV LD_LIBRARY_PATH=/usr/lib

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["./sd-server"]
