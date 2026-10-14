package main

// ───────────── 마우스 입력 ─────────────
// 왼쪽 클릭 = 점프 + 말풍선 / 드래그 = 집어서 던지기 / 오른쪽 클릭 = 메뉴

import (
	"math"
	"unsafe"
)

const (
	menuSay  = 1
	menuJump = 2
	menuQuit = 3
)

func cursor() point { var pt point; pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt))); return pt }

// ── 왼쪽 버튼 ──

func (p *pet) onDown() {
	pt := cursor()
	p.pressing, p.dragging = true, false
	p.pressX, p.pressY = pt.X, pt.Y
	p.grabDX, p.grabDY = float64(pt.X)-p.x, float64(pt.Y)-p.y
	p.hist = []point{pt}
	pSetCapture.Call(p.hwnd)
}

func (p *pet) onMove() {
	if !p.pressing {
		return
	}
	pt := cursor()
	if !p.dragging {
		if !p.movedPastDragThreshold(pt) {
			return
		}
		p.beginDrag()
	}
	p.dragTo(pt)
}

func (p *pet) onUp() {
	if !p.pressing {
		return
	}
	p.pressing = false
	pReleaseCapture.Call()
	if p.dragging {
		p.throwFromDrag()
	} else {
		p.onClick()
	}
}

// 누른 위치에서 일정 거리(6·dpi) 이상 움직였으면 클릭이 아니라 드래그
func (p *pet) movedPastDragThreshold(pt point) bool {
	return abs32(pt.X-p.pressX)+abs32(pt.Y-p.pressY) >= int32(6*p.dpi)
}

func (p *pet) beginDrag() {
	p.dragging = true
	p.setState("drag", -1)
	p.say("으앗!", 1.2)
}

// 잡은 지점을 유지하며 커서를 따라감
func (p *pet) dragTo(pt point) {
	p.x, p.y = float64(pt.X)-p.grabDX, float64(pt.Y)-p.grabDY
	p.recordCursor(pt)
}

// 최근 5개 커서 위치만 보관
func (p *pet) recordCursor(pt point) {
	p.hist = append(p.hist, pt)
	if len(p.hist) > 5 {
		p.hist = p.hist[len(p.hist)-5:]
	}
}

// 놓는 순간의 커서 속도로 던진다
func (p *pet) throwFromDrag() {
	p.vx, p.vy = p.releaseVelocity()
	p.dragging = false
	p.snapInside() // 모니터 사이 빈 곳이나 화면 밖에 놓은 경우
	p.setState("fall", -1)
}

// 최근 커서 이동의 프레임당 평균 속도 (±40·dpi로 제한)
func (p *pet) releaseVelocity() (vx, vy float64) {
	a, b := p.hist[0], p.hist[len(p.hist)-1]
	n := float64(max(1, len(p.hist)-1))
	lim := 40 * p.dpi
	vx = math.Max(-lim, math.Min(lim, float64(b.X-a.X)/n))
	vy = math.Max(-lim, math.Min(lim, float64(b.Y-a.Y)/n))
	return vx, vy
}

func (p *pet) onClick() {
	p.jump()
	p.sayRandom()
}

// ── 오른쪽 버튼 메뉴 ──

func (p *pet) onMenu() {
	p.handleMenuCommand(p.showContextMenu())
}

// 커서 위치에 메뉴를 띄우고 선택된 항목 ID를 반환 (취소 시 0)
func (p *pet) showContextMenu() uintptr {
	m, _, _ := pCreatePopupMenu.Call()
	defer pDestroyMenu.Call(m)
	pAppendMenuW.Call(m, MF_STRING, menuSay, uintptr(unsafe.Pointer(u16("말하기"))))
	pAppendMenuW.Call(m, MF_STRING, menuJump, uintptr(unsafe.Pointer(u16("점프"))))
	pAppendMenuW.Call(m, MF_SEPARATOR, 0, 0)
	pAppendMenuW.Call(m, MF_STRING, menuQuit, uintptr(unsafe.Pointer(u16("종료"))))
	pt := cursor()
	pSetForegroundWindow.Call(p.hwnd)
	cmd, _, _ := pTrackPopupMenu.Call(m, TPM_RETURNCMD|TPM_RIGHTBUTTON, uintptr(pt.X), uintptr(pt.Y), 0, p.hwnd, 0)
	return cmd
}

func (p *pet) handleMenuCommand(cmd uintptr) {
	switch cmd {
	case menuSay:
		p.sayRandom()
	case menuJump:
		p.jump()
	case menuQuit:
		pDestroyWindow.Call(p.hwnd)
	}
}

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}
