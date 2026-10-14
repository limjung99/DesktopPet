package main

// ───────────── 말풍선 (GDI로 텍스트를 그려 불투명 스프라이트로 만든다) ─────────────

import (
	"math"
	"math/rand"
	"time"
	"unsafe"
)

func createBubbleFont(dpi float64) uintptr {
	fh := -int32(math.Round(13 * dpi))
	font, _, _ := pCreateFontW.Call(uintptr(fh), 0, 0, 0, 400, 0, 0, 0, DEFAULT_CHARSET, 0, 0, CLEARTYPE_QUALITY, 0,
		uintptr(unsafe.Pointer(u16("Malgun Gothic"))))
	return font
}

// text를 sec초 동안 말풍선으로 표시
func (p *pet) say(text string, sec float64) {
	p.bubble = p.renderBubble(text)
	p.bubbleUntil = time.Now().Add(time.Duration(sec * float64(time.Second)))
}

// 대사 목록에서 하나를 골라 말한다
func (p *pet) sayRandom() {
	p.say(p.lines[rand.Intn(len(p.lines))], bubbleSec)
}

// 텍스트 크기에 맞춘 말풍선 스프라이트 (꼬리 제외)
func (p *pet) renderBubble(text string) *sprite {
	t := u16(text)
	pad := int(10 * p.dpi)
	tw, th := p.measureText(t, int32(p.W-int(30*p.dpi)))
	sp := p.drawTextSprite(t, tw+pad*2, th+pad, pad)
	finishBubble(sp)
	return sp
}

// 최대 너비 maxW에서 줄바꿈했을 때 텍스트가 차지하는 크기
func (p *pet) measureText(t *uint16, maxW int32) (w, h int) {
	tmp := newDIB(1, 1)
	defer tmp.free()
	pSelectObject.Call(tmp.dc, p.font)
	r := rect{0, 0, maxW, 0}
	pDrawTextW.Call(tmp.dc, uintptr(unsafe.Pointer(t)), ^uintptr(0), uintptr(unsafe.Pointer(&r)), DT_CENTER|DT_WORDBREAK|DT_CALCRECT|DT_NOPREFIX)
	return int(r.R), int(r.B)
}

// 흰 바탕 w×h에 텍스트를 가운데 정렬로 그려 스프라이트로 복사
func (p *pet) drawTextSprite(t *uint16, w, h, pad int) *sprite {
	d := newDIB(w, h)
	defer d.free()
	for i := range d.pix {
		d.pix[i] = 255
	}
	pSelectObject.Call(d.dc, p.font)
	pSetBkMode.Call(d.dc, TRANSPARENT)
	pSetTextColor.Call(d.dc, 0x222222)
	tr := rect{int32(pad), int32(pad / 2), int32(w - pad), int32(h - pad/2)}
	pDrawTextW.Call(d.dc, uintptr(unsafe.Pointer(t)), ^uintptr(0), uintptr(unsafe.Pointer(&tr)), DT_CENTER|DT_WORDBREAK|DT_NOPREFIX)
	pGdiFlush.Call()
	sp := newSprite(w, h)
	copy(sp.pix, d.pix)
	return sp
}

// GDI는 알파를 채우지 않으므로 전부 불투명으로 만들고 1px 회색 테두리를 두른다
func finishBubble(sp *sprite) {
	for y := 0; y < sp.h; y++ {
		for x := 0; x < sp.w; x++ {
			i := (y*sp.w + x) * 4
			sp.pix[i+3] = 255
			if x == 0 || y == 0 || x == sp.w-1 || y == sp.h-1 {
				sp.pix[i], sp.pix[i+1], sp.pix[i+2] = 85, 85, 85
			}
		}
	}
}
