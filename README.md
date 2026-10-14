# 데스크톱 펫

Windows 10/11 화면 위를 돌아다니는 작은 데스크톱 펫입니다.
Go로 만들었고 설치나 별도 런타임이 필요 없습니다.

## 실행

`DesktopPet.exe`를 더블클릭하면 됩니다. 중복 실행은 자동으로 막힙니다.

## 조작

| 동작 | 결과 |
| --- | --- |
| 왼쪽 클릭 | 점프하고 말풍선을 띄웁니다 |
| 드래그 | 펫을 집어서 던질 수 있습니다 |
| 오른쪽 클릭 | 메뉴 (말하기 / 점프 / 종료) |

## 다중 모니터

연결된 모든 모니터를 오가며 걸어 다닙니다.

- 더 낮은 바닥으로 넘어가면 떨어지고, 더 높은 바닥(작업표시줄이 있는 모니터 등)은 점프해서 올라갑니다.
- 모니터 사이가 떨어져 있으면 끝에서 돌아섭니다. 드래그로 옮길 수는 있습니다.
- 모니터 연결/해제, 해상도·작업표시줄 변경은 자동으로 반영됩니다.

## 설정

아래 파일은 모두 선택 사항이며, `DesktopPet.exe`와 같은 폴더에 둡니다. 수정한 뒤에는 다시 실행해야 적용됩니다.

| 파일 | 내용 |
| --- | --- |
| `pet.ini` | 크기(`scale`), 걷는 속도(`walk_speed`), 혼잣말 확률(`talk_chance`) |
| `lines.txt` | 말풍선 대사. 한 줄에 하나, `#`으로 시작하면 무시합니다 |
| `sprites\` | 캐릭터 이미지 폴더. 아래 설명 참고 |

`pet.ini` 기본값은 다음과 같습니다.

```ini
scale=1.0
walk_speed=2
talk_chance=0.15
```

### 캐릭터 이미지 바꾸기

`sprites\` 폴더를 만들고 `<애니메이션>_<번호>.png` 형식의 파일을 넣으면 내장 캐릭터 대신 사용합니다.
애니메이션 이름은 `idle`, `walk`, `drag`, `fall`, `jump`이고, 없는 애니메이션은 다른 것으로 대체됩니다.
예시 파일은 [`src/sprites`](src/sprites)에 있습니다.

## 시작 시 자동 실행

1. `Win + R`을 누르고 `shell:startup`을 입력합니다.
2. 열린 폴더에 `DesktopPet.exe` 바로가기를 넣습니다.

## 직접 빌드하기

Go 1.21 이상이 필요합니다. `src` 폴더에서 실행하세요.

Windows (`cmd`):

```bat
set GOOS=windows& go build -ldflags "-H windowsgui -s -w" -o DesktopPet.exe .
```

macOS / Linux에서 크로스 컴파일:

```sh
GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui -s -w" -o ../DesktopPet.exe .
```

## 소스 구성

`src/`는 하나의 `package main`을 기능별 파일로 나눈 구조입니다.

| 파일 | 역할 |
| --- | --- |
| `main.go` | 시작 순서 |
| `win32.go` | Win32 API 바인딩, 상수, DIB |
| `window.go` | 창 생성, 메시지 처리, 타이머 |
| `pet.go` | 펫 상태 구조체와 초기 배치 |
| `physics.go` | 상태 머신(idle/walk/jump/fall)과 물리 |
| `monitor.go` | 다중 모니터 영역, 충돌 판정, 화면 밖 보정 |
| `input.go` | 마우스 클릭/드래그/메뉴 |
| `render.go` | 프레임 그리기와 화면 출력 |
| `bubble.go` | 말풍선 텍스트 렌더링 |
| `animation.go` | 스프라이트 로딩 |
| `sprite.go` | 이미지 스케일링, 반전, 알파 합성 |
| `config.go` | `pet.ini`, `lines.txt` 읽기 |
