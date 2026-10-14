package main

// ───────────── 다중 모니터 & 충돌 판정 ─────────────
// 이동 가능 영역 = 모든 모니터 작업 영역(작업표시줄 제외)의 합집합.
// 캐릭터 "발"(하단 좌/우 끝 두 점)이 합집합 안에 있어야 유효한 위치.

import (
	"math"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

// ── 모니터 목록 ──

// EnumDisplayMonitors 콜백 (NewCallback은 개수 제한이 있어 한 번만 만든다)
var (
	monitorCB   uintptr
	monitorList []rect
)

func monitorProc(hmon, hdc, lprc, lparam uintptr) uintptr {
	var mi struct {
		Size    uint32
		Monitor rect
		Work    rect
		Flags   uint32
	}
	mi.Size = uint32(unsafe.Sizeof(mi))
	if r, _, _ := pGetMonitorInfoW.Call(hmon, uintptr(unsafe.Pointer(&mi))); r != 0 {
		monitorList = append(monitorList, mi.Work)
	}
	return 1
}

// 모든 모니터의 작업 영역. 테스트용 PET_MONITORS 환경 변수가 있으면 그 값을 쓴다
func enumMonitors() []rect {
	if env := os.Getenv("PET_MONITORS"); env != "" {
		return parseMonitorSpec(env)
	}
	if areas := enumSystemMonitors(); len(areas) > 0 {
		return areas
	}
	return []rect{fallbackWorkArea()}
}

// "l,t,r,b;l,t,r,b" 형식 해석 (잘못된 항목은 무시)
func parseMonitorSpec(spec string) []rect {
	var out []rect
	for _, part := range strings.Split(spec, ";") {
		if r, ok := parseRect(part); ok {
			out = append(out, r)
		}
	}
	return out
}

func parseRect(s string) (rect, bool) {
	f := strings.Split(s, ",")
	if len(f) != 4 {
		return rect{}, false
	}
	v := make([]int32, 4)
	for k := range f {
		n, _ := strconv.Atoi(strings.TrimSpace(f[k]))
		v[k] = int32(n)
	}
	return rect{v[0], v[1], v[2], v[3]}, true
}

func enumSystemMonitors() []rect {
	if monitorCB == 0 {
		monitorCB = syscall.NewCallback(monitorProc)
	}
	monitorList = monitorList[:0]
	pEnumDisplayMonitors.Call(0, 0, monitorCB, 0)
	return append([]rect(nil), monitorList...)
}

// 모니터 열거 실패 시: 주 모니터 작업 영역, 그것도 안 되면 고정값
func fallbackWorkArea() rect {
	var wa rect
	pSystemParametersInfo.Call(SPI_GETWORKAREA, 0, uintptr(unsafe.Pointer(&wa)), 0)
	if wa.R > wa.L {
		return wa
	}
	return rect{0, 0, 1280, 760}
}

// 모니터 연결/해제, 해상도·작업표시줄 변경 반영
func (p *pet) refreshMonitors() {
	if a := enumMonitors(); len(a) > 0 {
		p.areas = a
	}
	if p.dragging {
		return
	}
	p.snapInside()
	if (p.state == "idle" || p.state == "walk") && !p.grounded() {
		p.vx, p.vy = 0, 0
		p.setState("fall", -1)
	}
}

// ── 위치 판정 ──

// 화면 좌표 한 점이 작업 영역 중 하나에 들어가는지
func (p *pet) inside(px, py float64) bool {
	x, y := int32(math.Floor(px)), int32(math.Floor(py))
	for _, r := range p.areas {
		if x >= r.L && x < r.R && y >= r.T && y < r.B {
			return true
		}
	}
	return false
}

// (x,y)=캐릭터 좌상단. 발 양 끝이 화면 안이면 유효
func (p *pet) valid(x, y float64) bool {
	fy := y + float64(p.sh) - 1
	return p.inside(x, fy) && p.inside(x+float64(p.sw)-1, fy)
}

func (p *pet) grounded() bool { return !p.valid(p.x, p.y+1) }

// ── 이동 ──

// 막히기 직전까지 가로 이동. 반환: 새 x, 막혔는지
func (p *pet) moveX(x, y, dx float64) (float64, bool) {
	off, blocked := moveUntilBlocked(dx, func(o float64) bool { return p.valid(x+o, y) })
	return x + off, blocked
}

// 막히기 직전까지 세로 이동. 반환: 새 y, 막혔는지
func (p *pet) moveY(x, y, dy float64) (float64, bool) {
	off, blocked := moveUntilBlocked(dy, func(o float64) bool { return p.valid(x, y+o) })
	return y + off, blocked
}

// 한 축으로 d만큼 이동 시도. 한 번에 안 되면 1px씩 전진해 fits가 거짓이 되기 직전에서 멈춘다.
// 반환: 실제 이동량, 막혔는지
func moveUntilBlocked(d float64, fits func(offset float64) bool) (float64, bool) {
	if d == 0 || fits(d) {
		return d, false
	}
	step := math.Copysign(1, d)
	for n := 0.0; math.Abs(n+step) <= math.Abs(d); n += step {
		if !fits(n + step) {
			return n, true
		}
	}
	return 0, true
}

// ── 화면 밖 보정 ──

// 화면 밖(모니터 사이 빈 공간, 분리된 모니터 등)에 있으면 가장 가까운 작업 영역 안으로
func (p *pet) snapInside() {
	if p.valid(p.x, p.y) {
		return
	}
	best := p.nearestArea(p.x+float64(p.sw)/2, p.y+float64(p.sh))
	p.clampInto(best)
}

// 점 (cx,cy)에서 가장 가까운 작업 영역
func (p *pet) nearestArea(cx, cy float64) rect {
	best, bestD := p.areas[0], math.MaxFloat64
	for _, r := range p.areas {
		nx := math.Max(float64(r.L), math.Min(cx, float64(r.R)))
		ny := math.Max(float64(r.T), math.Min(cy, float64(r.B)))
		if d := (nx-cx)*(nx-cx) + (ny-cy)*(ny-cy); d < bestD {
			best, bestD = r, d
		}
	}
	return best
}

// 캐릭터가 영역 r 안에 들어오도록 좌표를 제한
func (p *pet) clampInto(r rect) {
	p.x = math.Max(float64(r.L), math.Min(p.x, float64(r.R)-float64(p.sw)))
	p.y = math.Max(float64(r.T)-float64(p.sh)+1, math.Min(p.y, float64(r.B)-float64(p.sh)))
	if !p.valid(p.x, p.y) { // 모니터가 캐릭터보다 좁은 극단적 경우
		p.x, p.y = float64(r.L), float64(r.B)-float64(p.sh)
	}
}
