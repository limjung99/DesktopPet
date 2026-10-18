package main

// ───────────── 펫 상태와 초기화 ─────────────

import (
	"math"
	"math/rand"
	"time"
)

const (
	fps       = 30
	gravity   = 1.0
	bounce    = 0.45
	bubbleH   = 70 // 캐릭터 위 말풍선 공간(96dpi 기준)
	minW      = 220
	bubbleSec = 2.5
)

type pet struct {
	hwnd  uintptr
	anims map[string]*anim
	lines []string

	// 크기: sw×sh = 캐릭터, W×H = 창(말풍선 공간 포함), (ox,oy) = 창 안 캐릭터 위치
	sw, sh, W, H, ox, oy int
	buf                  *dib
	dpi                  float64
	areas                []rect // 이동 가능한 모니터 작업 영역들

	// 물리/상태. (x,y) = 캐릭터 좌상단 화면 좌표
	x, y, vx, vy float64
	facing       int // 1 = 오른쪽, -1 = 왼쪽
	state        string
	timer, tick  int

	// 마우스 입력
	pressing, dragging bool
	pressX, pressY     int32
	grabDX, grabDY     float64
	hist               []point // 던지기 속도 계산용 최근 커서 위치

	// 드롭된 파일 중 휴지통 확인을 기다리는 경로
	droppedPaths []string

	// 말풍선
	bubble      *sprite
	bubbleUntil time.Time
	font        uintptr
}

// 창 프로시저(wndProc)에서 접근하는 유일한 펫
var P *pet

// 펫 생성. 캐릭터 이미지를 하나도 찾지 못하면 nil
func newPet(dpi float64) *pet {
	p := &pet{dpi: dpi, lines: loadLines(), facing: 1, state: "idle", timer: fps * 2}
	p.anims = loadSprites(cfgScale * dpi)
	if p.anims["idle"] == nil {
		return nil
	}
	p.initLayout()
	p.buf = newDIB(p.W, p.H)
	p.font = createBubbleFont(dpi)
	p.areas = enumMonitors()
	p.placeAtRandomStart()
	p.facing = randomFacing()
	return p
}

// 첫 idle 프레임 크기를 기준으로 창 크기와 캐릭터 위치를 정한다
func (p *pet) initLayout() {
	f := p.anims["idle"].right[0]
	p.sw, p.sh = f.w, f.h
	p.W = max(p.sw, int(minW*p.dpi))
	p.H = p.sh + int(bubbleH*p.dpi)
	p.ox, p.oy = (p.W-p.sw)/2, int(bubbleH*p.dpi)
}

// 아무 모니터의 바닥 임의 위치에서 시작
func (p *pet) placeAtRandomStart() {
	start := p.areas[rand.Intn(len(p.areas))]
	p.x = float64(start.L) + rand.Float64()*math.Max(0, float64(start.R-start.L)-float64(p.sw))
	p.y = float64(start.B) - float64(p.sh)
	p.snapInside()
}

// 상태 전환. frames >= 0이면 상태 지속 타이머도 설정
func (p *pet) setState(s string, frames int) {
	p.state, p.tick = s, 0
	if frames >= 0 {
		p.timer = frames
	}
}

func randomFacing() int { return []int{-1, 1}[rand.Intn(2)] }

// a 이상 b 이하 임의 정수
func rnd(a, b int) int { return a + rand.Intn(b-a+1) }
