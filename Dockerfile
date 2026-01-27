FROM golang:1.24-bookworm AS builder

# Install dependencies for cross-compilation (Ebitengine needs CGo on Linux)
RUN apt-get update && apt-get install -y \
    libgl1-mesa-dev \
    libxcursor-dev \
    libxrandr-dev \
    libxinerama-dev \
    libxi-dev \
    libxxf86vm-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build Linux binary
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /out/spacehole-linux-amd64 ./cmd/spacehole

# Build Windows binary (no CGo needed for Windows cross-compile with purego)
RUN CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o /out/spacehole-windows-amd64.exe ./cmd/spacehole

# Minimal output stage
FROM scratch
COPY --from=builder /out/ /out/
