# PC Spec Checker (PC 사양 체커)

**Linux, macOS, Windows** 시스템의 하드웨어 사양 정보를 수집하고 일목요연하게 표시하는 크로스 플랫폼 CLI 도구입니다.

## 기능

이 도구는 다음과 같은 PC 사양 정보를 수집하고 표시합니다:

- **CPU 정보**: 모델명, 물리 코어 수, 논리 코어(스레드) 수, 최대 클럭 속도
- **메모리(RAM) 정보**: 전체 용량, 사용 중인 용량, 사용 가능한 용량, 사용률
- **저장장치 정보**: 장치 경로, 마운트 지점, 파일시스템 타입, 용량 및 사용률
- **GPU 정보**: GPU 이름, 제조사, GPU 메모리, 드라이버 버전 (가능한 경우)

## 지원 플랫폼

✅ **Linux** - 완벽 지원  
✅ **macOS** - 완벽 지원  
✅ **Windows** - 완벽 지원

프로그램은 실행 시 자동으로 현재 운영체제를 감지하고 적절한 수집 방법을 사용합니다.

## 프로젝트 구조

프로젝트는 레이어드 아키텍처와 OS별 서브패키지로 구성되어 있습니다:

```text
pc_spec_checker/
├── model/                  # 데이터 모델 정의
│   └── system_info.go      # 시스템 정보 구조체
├── collector/              # 시스템 정보 수집 레이어
│   ├── factory.go          # OS별 collector 팩토리
│   ├── collector_test.go   # 통합 테스트
│   ├── linux/              # Linux 전용 패키지
│   │   ├── linux.go        # Linux collector 구현
│   │   └── stub.go         # 다른 OS용 스텁
│   ├── darwin/             # macOS 전용 패키지
│   │   ├── darwin.go       # macOS collector 구현
│   │   └── stub.go         # 다른 OS용 스텁
│   └── windows/            # Windows 전용 패키지
│       ├── windows.go      # Windows collector 구현
│       └── stub.go         # 다른 OS용 스텁
├── formatter/              # 출력 포맷팅 레이어
│   ├── formatter.go        # 콘솔 출력 포매터
│   └── formatter_test.go   # 포매터 유닛 테스트
├── main.go                 # 애플리케이션 진입점
├── Makefile                # 빌드 자동화
├── go.mod                  # Go 모듈 정의
└── README.md               # 프로젝트 문서
```

### 아키텍처 설명

- **model**: 시스템 정보를 담는 데이터 구조를 정의합니다
- **collector**: 실제 시스템에서 정보를 수집하는 로직을 담당합니다
  - **linux/**: `/proc/cpuinfo`, `/proc/meminfo`, `df`, `lspci`, `nvidia-smi` 사용
  - **darwin/**: `sysctl`, `vm_stat`, `df`, `system_profiler` 사용
  - **windows/**: `wmic` 명령어와 WMI 사용
  - **조건부 컴파일**: Go의 빌드 태그(`//go:build`)를 사용하여 OS별로 적절한 코드만 컴파일
  - **서브패키지**: 각 OS별 구현을 독립적인 패키지로 분리하여 관리
- **formatter**: 수집된 데이터를 사용자가 보기 좋은 형태로 포맷팅합니다
- **main**: 전체 애플리케이션의 흐름을 조정하고 OS를 자동 감지합니다
- **Makefile**: 빌드, 테스트, 설치 등을 자동화하는 Make 스크립트

## 빌드 방법

### Makefile 사용 (권장)

```bash
# 도움말 보기
make help

# 현재 OS용 빌드
make build

# 빌드 및 실행
make run

# 테스트 실행
make test

# 모든 플랫폼용 빌드
make build-all

# 시스템에 설치 (Linux/macOS, sudo 필요)
make install

# 정리
make clean
```

### 수동 빌드

#### 현재 OS용 빌드

```bash
# 프로젝트 디렉토리로 이동
cd /home/wkqco/Project/utils/pc_spec_checker

# 의존성 다운로드 및 정리
go mod tidy

# 실행 파일 빌드
go build -o pc_spec_checker
```

#### 크로스 컴파일 (다른 OS용 빌드)

```bash
# Linux용
GOOS=linux GOARCH=amd64 go build -o pc_spec_checker_linux

# macOS용 (Intel)
GOOS=darwin GOARCH=amd64 go build -o pc_spec_checker_macos_amd64

# macOS용 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o pc_spec_checker_macos_arm64

# Windows용
GOOS=windows GOARCH=amd64 go build -o pc_spec_checker.exe
```

## 실행 방법

빌드 후 다음과 같이 실행합니다:

```bash
# Linux/macOS
./pc_spec_checker

# Windows
pc_spec_checker.exe
```

프로그램은 자동으로 현재 운영체제를 감지하고 적절한 방법으로 시스템 정보를 수집합니다.

## Makefile 주요 타겟

```text
make help        # 사용 가능한 모든 명령어 표시
make build       # 현재 OS용 빌드
make run         # 빌드 및 실행
make test        # 유닛 테스트 실행
make coverage    # 테스트 커버리지 리포트
make build-all   # 모든 플랫폼용 크로스 컴파일
make install     # 시스템에 설치 (/usr/local/bin)
make uninstall   # 설치된 바이너리 제거
make clean       # 빌드 아티팩트 정리
```

## 테스트 실행

유닛 테스트를 실행하려면:

```bash
# 모든 패키지의 테스트 실행
go test ./...

# 상세한 출력과 함께 실행
go test -v ./...

# 특정 패키지만 테스트
go test ./collector
go test ./formatter
```

## 출력 예시

**Linux에서의 실행 결과:**

```text
시스템 정보를 수집 중... (Linux)

╔════════════════════════════════════════════════════════════════╗
║              PC 사양 정보 (System Specifications)              ║
╚════════════════════════════════════════════════════════════════╝

┌─ CPU 정보
│  모델명: AMD Ryzen 5 7530U with Radeon Graphics
│  물리 코어: 1개
│  논리 코어(스레드): 12개
│  최대 클럭: 4541 MHz
└─────────────────────────────────────────────────────────────────

┌─ 메모리 (RAM) 정보
│  전체 용량: 19.39 GB
│  사용 중: 4.55 GB (23.5%)
│  사용 가능: 14.83 GB
└─────────────────────────────────────────────────────────────────
...
```

**macOS와 Windows에서도 유사한 형태로 정보가 표시됩니다.**

## 요구사항

**공통:**

- Go 1.25.5 이상

**Linux:**

- 기본 시스템 유틸리티: `df`, `lspci`
- (선택사항) NVIDIA GPU가 있는 경우: `nvidia-smi`

**macOS:**

- `sysctl` (기본 제공)
- `vm_stat` (기본 제공)
- `system_profiler` (기본 제공)
- `df` (기본 제공)

**Windows:**

- `wmic` (기본 제공)
- (선택사항) NVIDIA GPU가 있는 경우: `nvidia-smi`

## 기술적 특징

### 조건부 컴파일 및 Stub 파일

Go의 빌드 태그(`//go:build`)를 사용하여 OS별로 다른 코드를 컴파일합니다:

**각 OS별 구현 파일:**

- `collector/linux/linux.go` - Linux 전용 구현 (`//go:build linux`)
- `collector/darwin/darwin.go` - macOS 전용 구현 (`//go:build darwin`)
- `collector/windows/windows.go` - Windows 전용 구현 (`//go:build windows`)

**Stub 파일의 역할:**

각 OS 폴더에는 `stub.go` 파일이 있습니다. 이 파일들은 **크로스 컴파일을 가능하게 하기 위해** 필요합니다:

```go
//go:build !linux  // Linux가 아닐 때만 컴파일됨
```

**왜 필요한가?**

1. **크로스 컴파일 지원**: Linux 시스템에서 Windows용 실행 파일을 빌드할 때, Go 컴파일러는 `windows` 패키지를 import하려고 시도합니다. 하지만 `windows.go`는 `//go:build windows` 태그로 인해 컴파일되지 않습니다.
2. **패키지 무결성**: 패키지가 비어있으면 Go 컴파일러가 오류를 발생시킵니다. stub.go는 이를 방지합니다.
3. **타입 호환성**: factory.go에서 모든 OS 패키지를 import하므로, 각 패키지는 항상 컴파일 가능한 코드를 제공해야 합니다.

**동작 방식:**

- Linux에서 빌드 시: `linux.go` 컴파일 ✅, `stub.go` 스킵 (Linux이므로)
- macOS에서 Windows용 빌드 시: `windows/windows.go` 스킵 (macOS이므로), `windows/stub.go` 컴파일 ✅ (Windows가 아니므로)
- 결과: 실제 실행에는 factory.go의 `runtime.GOOS` 체크로 올바른 collector만 사용됩니다

빌드 시 현재 플랫폼에 맞는 파일만 실행 파일에 포함되므로 바이너리 크기가 최적화됩니다.

### 인터페이스 기반 설계

`SystemCollector` 인터페이스를 통해 모든 OS별 구현체를 동일한 방식으로 사용할 수 있습니다:

```go
type SystemCollector interface {
    CollectAll() (*model.SystemInfo, error)
    CollectCPU() (*model.CPUInfo, error)
    CollectMemory() (*model.MemoryInfo, error)
    CollectStorage() ([]model.StorageInfo, error)
    CollectGPU() ([]model.GPUInfo, error)
}
```

## 제한사항

- GPU 정보는 시스템 구성에 따라 부분적으로만 표시될 수 있습니다
- GPU 메모리 정보는 NVIDIA GPU의 경우 가장 정확하게 표시됩니다
- macOS에서 일부 CPU 주파수 정보는 Apple Silicon에서 제한될 수 있습니다
- Windows에서는 관리자 권한이 필요한 정보가 제한될 수 있습니다

## 라이선스

이 프로젝트는 개인 사용을 위한 유틸리티 도구입니다.

## 기여

버그 리포트나 기능 제안은 언제든지 환영합니다!
