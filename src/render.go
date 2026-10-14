package main

// ───────────── 렌더링 (DIB 버퍼에 그린 뒤 레이어드 윈도우로 출력) ─────────────

import (
	"math"
	"time"
	"unsafe"
)

func (p *pet) render() {
	p.clearBuffer()
	p.drawBubble()
	blit(p.buf.pix, p.W, p.H, p.currentFrame(), p.ox, p.oy)
	p.present()
}

// 버퍼 전체를 투명하게
func (p *pet) clearBuffer() {
	pix := p.buf.pix
	for i := range pix {
		pix[i] = 0
	}
}

// 지금 상태·방향·시간에 맞는 애니메이션 프레임
func (p *pet) currentFrame() *sprite {
	a := p.anims[p.state]
	frames := a.right
	if p.facing < 0 {
		frames = a.left
	}
	idx := p.tick * animFPS[p.state] / fps
	return frames[idx%len(frames)]
}

// 말풍선(표시 시간이 지났으면 제거)을 캐릭터 위 가운데에 그린다
func (p *pet) drawBubble() {
	if p.bubble == nil {
		return
	}
	if time.Now().After(p.bubbleUntil) {
		p.bubble = nil
		return
	}
	bx := (p.W - p.bubble.w) / 2
	by := p.oy - p.bubble.h - int(10*p.dpi)
	blit(p.buf.pix, p.W, p.H, p.bubble, bx, by)
	p.drawTail(p.W/2, by+p.bubble.h-1)
}

// 말풍선 꼬리 (흰 삼각형 + 테두리)
func (p *pet) drawTail(cx, top int) {
	th := int(9 * p.dpi)
	for dy := 0; dy <= th; dy++ {
		half := int(float64(th-dy) * 0.7)
		for dx := -half; dx <= half; dx++ {
			c := byte(255)
			if (dx == -half || dx == half) && dy > 0 {
				c = 85
			}
			p.putOpaqueGray(cx+dx, top+dy, c)
		}
	}
}

// 버퍼 (x,y)에 불투명 회색 c를 찍는다 (범위 밖은 무시)
func (p *pet) putOpaqueGray(x, y int, c byte) {
	if x < 0 || y < 0 || x >= p.W || y >= p.H {
		return
	}
	i := (y*p.W + x) * 4
	p.buf.pix[i], p.buf.pix[i+1], p.buf.pix[i+2], p.buf.pix[i+3] = c, c, c, 255
}

// 버퍼를 창에 반영하면서 창 위치도 캐릭터 좌표에 맞춘다
func (p *pet) present() {
	dst := point{int32(math.Round(p.x)) - int32(p.ox), int32(math.Round(p.y)) - int32(p.oy)}
	sz := size{int32(p.W), int32(p.H)}
	src := point{}
	bf := blendFunc{AC_SRC_OVER, 0, 255, AC_SRC_ALPHA}
	scr, _, _ := pGetDC.Call(0)
	pUpdateLayeredWindow.Call(p.hwnd, scr, uintptr(unsafe.Pointer(&dst)), uintptr(unsafe.Pointer(&sz)),
		p.buf.dc, uintptr(unsafe.Pointer(&src)), 0, uintptr(unsafe.Pointer(&bf)), ULW_ALPHA)
	pReleaseDC.Call(0, scr)
}
