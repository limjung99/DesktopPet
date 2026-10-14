# CLAUDE.md

이 파일은 Claude Code(claude.ai/code)가 이 저장소에서 작업할 때 참고하는 안내서입니다.

## 응답 언어

- **항상 한국어로 응답한다.** 사용자가 영어(또는 다른 언어)로 질문해도 한국어로 답한다.
- 코드 주석과 사용자에게 보이는 문자열(말풍선 대사, 메뉴 등)도 한국어로 쓴다.

## 작업 규칙

- **목표**: Go로 작성하고 **Windows(10/11)에서 실행되는** 데스크톱 펫을 개발한다.
- **개발 환경**: Apple Silicon(arm64) macOS. 여기서는 exe를 실행할 수 없으므로 항상 `GOOS=windows`로 크로스 컴파일하고, 실제 동작 확인은 Windows에서 해야 한다는 점을 전제로 작업한다.
- **함수 하나는 한 가지 역할만** 하도록 작성한다. 여러 일을 하는 함수(예: 계산 + 렌더링 + Win32 호출이 섞인 함수)는 역할별로 나눈다.
- **모듈별·기능별로 사람이 읽기 쉽게** 작성한다. 이름만으로 역할이 드러나도록 짓는다.
  - 코드는 `package main` 안에서 기능별 파일로 나뉘어 있다. 파일 목록은 `src/main.go` 상단 주석에 있다.
  - 새 코드는 맞는 파일에 넣고, 새 기능 영역이면 새 파일을 만든 뒤 그 목록에 추가한다.
  - Win32 메시지·스타일 값은 숫자 대신 `win32.go`의 이름 있는 상수를 쓴다.

## 프로젝트 개요

Go로 작성한 Windows(10/11) 전용 데스크톱 펫이다. 소스는 `src/*.go`이고, 하나의 `package main`을 기능별 파일로 나눴다. `syscall.NewLazyDLL`로 Win32 API를 직접 호출하므로 cgo, 외부 모듈, 런타임 의존성이 없다.

저장소 루트에는 배포용 묶음(`DesktopPet.exe`, `pet.ini`, `lines.txt`, `README.md`)이 있고, `src/`에 소스가 있다.

git 저장소, 테스트 코드, 린터 설정은 없다.

## 빌드

Apple Silicon macOS에서 개발하므로 항상 크로스 컴파일한다. Windows 전용 syscall을 쓰기 때문에 darwin용으로는 빌드되지 않는다. ARM Windows용을 명시적으로 요청받지 않는 한 `amd64`로 빌드한다(ARM은 `GOARCH=arm64`).

```sh
cd src && GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui -s -w" -o ../DesktopPet.exe .
cd src && GOOS=windows GOARCH=amd64 go vet .
```

- `-H windowsgui`는 콘솔 창이 뜨지 않게 한다.
- README에 있는 명령은 Windows `cmd` 형식(`set GOOS=windows& go build ...`)이고, 결과물이 `src/`에 생긴다. 배포용 exe는 저장소 루트에 둔다.
- 바이너리는 macOS에서 실행할 수 없으므로 실행 테스트는 Windows가 필요하다.

`go vet`은 항상 `win32.go: possible misuse of unsafe.Pointer` 경고를 낸다. `CreateDIBSection`이 돌려준 비트맵 포인터를 `unsafe.Slice`로 감싸는 부분인데, 이 메모리는 Go가 아니라 GDI가 관리하므로 알려진 오탐이다. 이것 외의 vet 경고는 실제 문제로 본다.

Win32를 쓰지 않는 로직(예: `sprite.go`의 스케일링·합성)은 파일을 임시 모듈로 복사해 macOS에서 `go run`으로 확인할 수 있다. 실수 계산 결과를 비교할 때는 `GOARCH=amd64`(Rosetta)로 실행한다. arm64에서는 컴파일러가 곱셈·덧셈을 FMA로 합칠 수 있어서, amd64 Windows 빌드와 반올림이 1만큼 달라질 수 있다.

프로그램이 읽는 테스트용 환경 변수:
- `PET_MONITORS="l,t,r,b;l,t,r,b"`: 모니터 목록 대신 가짜 작업 영역을 쓴다(다중 모니터 배치 테스트용).
- `PET_TEST=1`: 시작하자마자 말풍선을 30초 동안 띄운다.

## 구조

- **렌더링**
  - `WS_EX_LAYERED|TOPMOST|TOOLWINDOW|NOACTIVATE` 팝업 창 하나를 32비트 top-down DIB에 그린 뒤 `UpdateLayeredWindow`로 출력한다.
  - 모든 픽셀 버퍼는 **프리멀티플라이드 BGRA**이고, `blit`은 프리멀티플라이드 "over" 합성을 한다.
  - 창은 말풍선 자리를 두기 위해 스프라이트보다 크다(`minW`, `bubbleH`). `p.ox/p.oy`는 창 안에서 스프라이트의 위치이고, `p.x/p.y`는 스프라이트 좌상단의 화면 좌표다.
- **스프라이트**
  - `sprites/*.png`는 `//go:embed`로 내장된다. exe 옆 `sprites\` 폴더에 PNG가 하나라도 있으면 내장 이미지를 통째로 대체한다.
  - 파일 이름은 `<애니메이션>_<번호>.png`이고, 애니메이션은 `idle, walk, drag, fall, jump`다. 없는 애니메이션은 다른 것으로 대체한다(`animation.go`의 `animFallbacks`).
  - 로드할 때 `cfgScale * dpi`로 한 번 크기를 바꾸고, 왼쪽을 볼 때 쓸 반전 이미지도 미리 만든다.
- **말풍선**: GDI `DrawTextW`(맑은 고딕)로 임시 DIB에 글자를 그린 뒤 불투명 스프라이트로 복사한다. 꼬리는 매 프레임 `drawTail`에서 그린다.
- **메인 루프**
  - OS 스레드에 고정된 Win32 메시지 루프다.
  - 타이머 1(30fps)은 `update()` + `render()`를 실행한다. 타이머 2(2초마다)는 최상위 창 상태를 다시 지정하고 `refreshMonitors()`를 호출한다.
  - `WM_DISPLAYCHANGE`/`WM_SETTINGCHANGE`를 받을 때도 모니터 정보를 갱신한다.
- **상태 머신**(`p.state` 문자열)
  - `idle` ↔ `walk`는 타이머로 전환한다.
  - `jump`/`fall`은 중력과 튕김을 받는 물리 상태다.
  - `drag`는 마우스로 잡고 있는 상태다. 놓을 때의 던지기 속도는 `p.hist`에 있는 최근 커서 위치 약 5개로 계산한다.
- **다중 모니터 충돌 모델**
  - 이동 가능 영역은 모든 모니터 *작업 영역*(작업표시줄 제외)의 합집합이다. 스프라이트 하단 양 끝("발")이 합집합 안에 있어야 유효한 위치다(`valid`).
  - `moveX`/`moveY`는 막히기 직전까지 1px 단위로 전진한다. `grounded()`는 아래로 1px도 못 움직이는 상태다.
  - 더 낮은 바닥으로 걸어 나가면 `fall`이 된다. 벽에 막혔을 때 200·dpi px 이내에 더 높은 바닥이 있으면 `tryClimb`으로 뛰어오른다.
  - `snapInside`는 모니터 사이 빈 공간이나 화면 밖에 있을 때 가장 가까운 작업 영역 안으로 끌어온다.
  - `fall/jump` 중에는 의도적으로 Y를 X보다 먼저 움직인다. 그래야 점프로 옆 모니터의 더 높은 바닥에 올라설 수 있다.
- **DPI**: 프로세스는 모니터별 DPI 인식 모드다. `p.dpi`(시스템 DPI / 96)로 크기, 속도, 중력을 조절한다. 중력은 `sqrt(dpi)`에 비례한다.
- **설정**: exe 폴더의 `pet.ini`(`scale`은 0.2~5로 제한, `walk_speed`, `talk_chance`)와 `lines.txt`(UTF-8, 한 줄에 하나, `#` 주석, BOM 제거)를 읽는다. 파일이 없으면 기본값을 쓴다.
- **단일 실행**: 이름 있는 뮤텍스 `Local\DesktopPetSingleInstance`로 중복 실행을 막는다.
- `syscall.NewCallback`은 만들 수 있는 개수에 전역 제한이 있다. 콜백(예: `monitorCB`)은 한 번만 만들어 재사용하고, 호출할 때마다 새로 만들지 않는다.
