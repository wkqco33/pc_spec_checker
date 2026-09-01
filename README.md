# PC Spec Checker (PC 사양 체커)

**Linux, macOS, Windows** 시스템의 하드웨어 사양 정보를 수집하고 일목요연하게 표시하는 크로스 플랫폼 CLI 도구입니다.

## 기능

이 도구는 다음과 같은 PC 사양 정보를 수집하고 표시합니다:

- **CPU 정보**: 모델명, 물리 코어 수, 논리 코어(스레드) 수, 최대 클럭 속도
- **메모리(RAM) 정보**: 전체 용량, 사용 중인 용량, 사용 가능한 용량, 사용률
- **저장장치 정보**: 장치 경로, 마운트 지점, 파일시스템 타입, 용량 및 사용률
- **GPU 정보**: GPU 이름, 제조사, GPU 메모리, 드라이버 버전 (가능한 경우)
- **출력 모드**: 기본 요약 출력, `--verbose` 또는 `-v`로 전체 상세 출력
- **AI 사양 분석** (`--ai`): LLM이 게임/개발/영상시청/웹서핑/사무 용도별 성능 등급과 업그레이드 제안을 요약 ([LLM_client_go](https://github.com/wkqco33/LLM_client_go) 사용, 기본: 로컬 Ollama)

## 지원 플랫폼

✅ **Linux** - 완벽 지원  
✅ **macOS** - 완벽 지원  
✅ **Windows** - 완벽 지원

프로그램은 실행 시 자동으로 현재 운영체제를 감지하고 적절한 수집 방법을 사용합니다.

## 프로젝트 구조

프로젝트는 레이어드 아키텍처와 OS별 서브패키지로 구성되어 있습니다:

```text
pcsc/
├── model/                  # 데이터 모델 정의
│   ├── system_info.go      # 시스템 정보 구조체
│   └── system_info_test.go # 모델 JSON 테스트
├── collector/              # 시스템 정보 수집 레이어
│   ├── factory.go          # OS별 collector 팩토리
│   ├── collector_test.go   # Linux 통합 테스트
│   ├── linux/              # Linux 전용 패키지
│   │   ├── linux.go        # Linux collector 구현
│   │   └── stub.go         # 다른 OS용 스텁
│   ├── darwin/             # macOS 전용 패키지
│   │   ├── darwin.go       # macOS collector 구현
│   │   ├── vmstat_parse.go # vm_stat 파싱 (OS 무관)
│   │   └── stub.go         # 다른 OS용 스텁
│   └── windows/            # Windows 전용 패키지
│       ├── windows.go      # Windows collector 구현
│       ├── vendor_detect.go# GPU 제조사 판별 (OS 무관)
│       └── stub.go         # 다른 OS용 스텁
├── formatter/              # 출력 포맷팅 레이어
│   ├── formatter.go        # 콘솔 출력 포매터
│   ├── analysis.go         # AI 분석 결과 포매터
│   └── formatter_test.go   # 포매터 유닛 테스트
├── analyzer/               # AI 사양 분석 레이어
│   ├── analyzer.go         # LLM 분석기 (llm.Client 주입)
│   ├── prompt.go           # 프롬프트 빌더 + JSON 스키마
│   ├── provider.go         # 환경변수 기반 LLM 프로바이더 팩토리
│   └── result.go           # 분석 결과 모델
├── main.go                 # 애플리케이션 진입점
├── Taskfile.yml            # 빌드 자동화
├── go.mod                  # Go 모듈 정의
├── ppm.json                # ppm 패키지 매니페스트
├── AGENTS.md               # AI 에이전트 개발 가이드
├── LICENSE                 # MIT 라이선스
├── CONTRIBUTING.md         # 기여 가이드
├── SECURITY.md             # 보안 정책
└── README.md               # 프로젝트 문서
```

### 아키텍처 설명

- **model**: 시스템 정보를 담는 데이터 구조를 정의합니다
- **collector**: 실제 시스템에서 정보를 수집하는 로직을 담당합니다
  - **linux/**: `/proc/cpuinfo`, `/proc/meminfo`, `syscall.Statfs`, `lspci`, `nvidia-smi` 사용
  - **darwin/**: `sysctl`, `vm_stat`, `df`, `system_profiler` 사용
  - **windows/**: `wmic` 명령어와 WMI 사용
  - **조건부 컴파일**: Go의 빌드 태그(`//go:build`)를 사용하여 OS별로 적절한 코드만 컴파일
  - **서브패키지**: 각 OS별 구현을 독립적인 패키지로 분리하여 관리
- **formatter**: 수집된 데이터를 사용자가 보기 좋은 형태로 포맷팅합니다
- **main**: 전체 애플리케이션의 흐름을 조정하고 OS를 자동 감지합니다
- **Taskfile**: 빌드, 테스트, 설치 등을 자동화하는 스크립트 ([Task](https://taskfile.dev))

## 빌드 방법

### Taskfile 사용 (권장)

[Task](https://taskfile.dev) 설치 후 (`go install github.com/go-task/task/v3/cmd/task@latest`):

```bash
# 도움말 보기
task help

# 현재 OS용 빌드
task build

# 빌드 및 실행
task run

# 테스트 실행
task test

# 모든 플랫폼용 빌드
task build-all

# 시스템에 설치 (Linux/macOS, sudo 필요)
task install

# 정리
task clean
```

### 수동 빌드

#### 현재 OS용 빌드

```bash
# 의존성 다운로드 및 정리
go mod tidy

# 실행 파일 빌드
go build -o pcsc
```

#### 크로스 컴파일 (다른 OS용 빌드)

```bash
# Linux용
GOOS=linux GOARCH=amd64 go build -o pcsc_linux

# macOS용 (Intel)
GOOS=darwin GOARCH=amd64 go build -o pcsc_macos_amd64

# macOS용 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o pcsc_macos_arm64

# Windows용
GOOS=windows GOARCH=amd64 go build -o pcsc.exe
```

## 실행 방법

빌드 후 다음과 같이 실행합니다:

```bash
# Linux/macOS
./pcsc

# 모든 세부 정보 표시
./pcsc --verbose
# 짧은 옵션
./pcsc -v

# Windows
pcsc.exe
```

프로그램은 자동으로 현재 운영체제를 감지하고 적절한 방법으로 시스템 정보를 수집합니다.

### AI 사양 분석

`--ai` 플래그를 사용하면 설정된 LLM이 사양을 분석해 게임/개발/영상시청/웹서핑/사무 용도별 성능 등급을 요약해 줍니다.

```bash
./pcsc --ai
```

#### 질문에 맞춘 분석

`--ai` 뒤에 질문을 넣으면 사양을 기반으로 질문에 직접 답변합니다:

```bash
./pcsc --ai "이 PC에서 rust 개발을 할건데 적합할까?"
./pcsc --ai "배틀그라운드 1080p 고사양 옵션으로 돌릴 수 있어?"
./pcsc --ai --ai-model llama3.1 "영상 편집용으로 쓸만한가요?"
```

출력 예시:

```text
┌─ AI 사양 분석
│  6코어 12스레드 CPU와 RTX 3060을 갖춘 균형 잡힌 구성입니다.
│
│  게임       [좋음]     70점 — 인디 및 1080p 게임은 원활합니다.
│  개발       [보통]     50점 — 대형 프로젝트 빌드는 다소 느립니다.
│  영상시청   [좋음]     75점 — 4K 재생이 가능합니다.
│  웹서핑     [매우 좋음] 90점 — 충분합니다.
│  사무       [매우 좋음] 95점 — 사무용으로 문제 없습니다.
│
│  업그레이드 제안: 메모리 추가를 권장합니다.
└─────────────────────────────────────────────────────────────────
```

LLM 프로바이더는 환경변수로 설정합니다 (기본: 로컬 Ollama):

```bash
# 기본: 로컬 Ollama (http://localhost:11434, API 키 불필요)
ollama pull qwen3.4b
./pcsc --ai

# 모델 변경
./pcsc --ai --ai-model llama3.1
# 또는
export PCSC_AI_MODEL=llama3.1

# OpenAI API 사용
export PCSC_AI_PROVIDER=openai
export OPENAI_API_KEY=sk-...
./pcsc --ai

# Azure OpenAI 사용
export PCSC_AI_PROVIDER=azure
export AZURE_API_KEY=...
export AZURE_ENDPOINT=https://<resource-name>.openai.azure.com
./pcsc --ai
```

### 설정 파일 관리

AI 분석 설정은 `pcsc config` 명령으로 관리할 수 있습니다:

```bash
# 기본 설정 파일 생성 (~/.config/pcsc/config.json)
pcsc config init

# 현재 설정 보기
pcsc config show

# 설정 변경
pcsc config set ai.provider openai
pcsc config set ai.model gpt-4o-mini
pcsc config set ai.ollama_base_url http://192.168.1.5:11434/v1
```

설정 파일 예시:

```json
{
  "ai": {
    "provider": "ollama",
    "model": "qwen3.4b",
    "ollama_base_url": ""
  }
}
```

우선순위: **플래그 > 환경변수 > 설정 파일 > 기본값**

| 환경변수 | 설명 | 기본값 |
|----------|------|--------|
| `PCSC_AI_PROVIDER` | LLM 프로바이더 (`ollama`/`openai`/`azure`) | `ollama` |
| `PCSC_AI_MODEL` | 모델명 | `qwen3.4b` (Ollama) / `gpt-4o-mini` (OpenAI) |
| `OLLAMA_BASE_URL` | Ollama 엔드포인트 | `http://localhost:11434/v1` |
| `OPENAI_API_KEY` | OpenAI API 키 (`openai` 시 필수) | — |
| `AZURE_API_KEY` / `AZURE_ENDPOINT` | Azure 설정 (`azure` 시 필수) | — |

AI 분석 중 LLM 연결 실패 시 사양 출력은 정상적으로 유지되고 분석 결과만 생략됩니다.

## Task 주요 태스크

```text
task help          # 사용 가능한 모든 명령어 표시 (기본 실행)
task build         # 현재 OS용 빌드
task run           # 빌드 및 실행
task test          # 유닛 테스트 실행
task test-coverage # 테스트 커버리지 리포트
task verify        # fmt + test + build 전체 검증
task build-all     # 모든 플랫폼용 크로스 컴파일
task install       # 시스템에 설치 (/usr/local/bin)
task uninstall     # 설치된 바이너리 제거
task clean         # 빌드 아티팩트 정리
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

- 기본 시스템 유틸리티: `lspci`
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

이 프로젝트는 [MIT 라이선스](LICENSE)로 배포됩니다.

## 기여

버그 리포트나 기능 제안은 GitHub Issues를 이용해 주세요. 기여 절차와 규칙은 [CONTRIBUTING.md](CONTRIBUTING.md)를 참고하세요. 보안 취약점은 [SECURITY.md](SECURITY.md)에 따라 비공개로 보고해 주세요.
