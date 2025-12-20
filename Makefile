# PC Spec Checker Makefile

# 변수 설정
BINARY_NAME=pc_spec_checker
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date +%FT%T%z)
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

# 기본 OS/ARCH
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

# 설치 경로
INSTALL_PATH=/usr/local/bin

# 색상 출력
BLUE=\033[0;34m
GREEN=\033[0;32m
RED=\033[0;31m
NC=\033[0m # No Color

.PHONY: all build clean test install uninstall help run \
        build-linux build-darwin build-windows build-all \
        test-coverage fmt lint

# 기본 타겟
all: clean test build

## help: 사용 가능한 명령어 목록 표시
help:
	@echo "$(BLUE)PC Spec Checker - Makefile 명령어$(NC)"
	@echo ""
	@echo "$(GREEN)기본 명령어:$(NC)"
	@echo "  make build          - 현재 OS용 빌드"
	@echo "  make run            - 프로그램 빌드 및 실행"
	@echo "  make test           - 유닛 테스트 실행"
	@echo "  make clean          - 빌드 산출물 삭제"
	@echo "  make install        - 시스템에 설치 (sudo 필요)"
	@echo "  make uninstall      - 시스템에서 제거 (sudo 필요)"
	@echo ""
	@echo "$(GREEN)크로스 컴파일:$(NC)"
	@echo "  make build-linux    - Linux용 빌드"
	@echo "  make build-darwin   - macOS용 빌드"
	@echo "  make build-windows  - Windows용 빌드"
	@echo "  make build-all      - 모든 플랫폼용 빌드"
	@echo ""
	@echo "$(GREEN)개발:$(NC)"
	@echo "  make fmt            - 코드 포맷팅"
	@echo "  make lint           - 코드 린트"
	@echo "  make test-coverage  - 테스트 커버리지 확인"
	@echo ""

## build: 현재 OS용 바이너리 빌드
build:
	@echo "$(BLUE)Building ${BINARY_NAME} for ${GOOS}/${GOARCH}...$(NC)"
	go build ${LDFLAGS} -o ${BINARY_NAME} .
	@echo "$(GREEN)✓ Build complete: ${BINARY_NAME}$(NC)"

## run: 프로그램 빌드 및 실행
run: build
	@echo "$(BLUE)Running ${BINARY_NAME}...$(NC)"
	@./$(BINARY_NAME)

## test: 유닛 테스트 실행
test:
	@echo "$(BLUE)Running tests...$(NC)"
	go test -v ./...
	@echo "$(GREEN)✓ Tests passed$(NC)"

## test-coverage: 테스트 커버리지 확인
test-coverage:
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report generated: coverage.html$(NC)"

## clean: 빌드 산출물 및 임시 파일 삭제
clean:
	@echo "$(BLUE)Cleaning...$(NC)"
	@rm -f ${BINARY_NAME}
	@rm -f ${BINARY_NAME}_*
	@rm -f *.exe
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ Clean complete$(NC)"

## install: 시스템에 바이너리 설치
install: build
	@echo "$(BLUE)Installing ${BINARY_NAME} to ${INSTALL_PATH}...$(NC)"
	@sudo install -m 755 ${BINARY_NAME} ${INSTALL_PATH}/${BINARY_NAME}
	@echo "$(GREEN)✓ Installed to ${INSTALL_PATH}/${BINARY_NAME}$(NC)"
	@echo "Run '${BINARY_NAME}' from anywhere"

## uninstall: 시스템에서 바이너리 제거
uninstall:
	@echo "$(BLUE)Uninstalling ${BINARY_NAME}...$(NC)"
	@sudo rm -f ${INSTALL_PATH}/${BINARY_NAME}
	@echo "$(GREEN)✓ Uninstalled$(NC)"

## fmt: Go 코드 포맷팅
fmt:
	@echo "$(BLUE)Formatting code...$(NC)"
	go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

## lint: 코드 린트 (golangci-lint 필요)
lint:
	@echo "$(BLUE)Linting code...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
		echo "$(GREEN)✓ Linting complete$(NC)"; \
	else \
		echo "$(RED)golangci-lint not installed. Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
	fi

## build-linux: Linux용 바이너리 빌드
build-linux:
	@echo "$(BLUE)Building for Linux (amd64)...$(NC)"
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}_linux_amd64 .
	@echo "$(GREEN)✓ Built: ${BINARY_NAME}_linux_amd64$(NC)"

## build-darwin: macOS용 바이너리 빌드
build-darwin:
	@echo "$(BLUE)Building for macOS...$(NC)"
	@echo "  - Intel (amd64)"
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}_darwin_amd64 .
	@echo "  - Apple Silicon (arm64)"
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o ${BINARY_NAME}_darwin_arm64 .
	@echo "$(GREEN)✓ Built: ${BINARY_NAME}_darwin_amd64, ${BINARY_NAME}_darwin_arm64$(NC)"

## build-windows: Windows용 바이너리 빌드
build-windows:
	@echo "$(BLUE)Building for Windows (amd64)...$(NC)"
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}_windows_amd64.exe .
	@echo "$(GREEN)✓ Built: ${BINARY_NAME}_windows_amd64.exe$(NC)"

## build-all: 모든 플랫폼용 바이너리 빌드
build-all: build-linux build-darwin build-windows
	@echo ""
	@echo "$(GREEN)✓ All builds complete:$(NC)"
	@ls -lh ${BINARY_NAME}_* 2>/dev/null || true

## deps: 의존성 다운로드 및 정리
deps:
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	go mod download
	go mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(NC)"

## verify: 모든 검증 단계 실행 (fmt, test, build)
verify: fmt test build
	@echo "$(GREEN)✓ All verifications passed$(NC)"

.DEFAULT_GOAL := help
