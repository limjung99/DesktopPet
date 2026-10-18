package main

// ───────────── 대화 말풍선 목록 그리기 ─────────────
// 메모리 DC에 전체를 그린 뒤 한 번에 복사해 깜박임을 없앤다.
// 내 말은 오른쪽 노란 말풍선, 상대 말은 왼쪽 흰 말풍선이다.

import (
	"unsafe"
)

// COLORREF = 0x00BBGGRR
const (
	colorChatBg    = 0xD9C7B2
	colorUserBub   = 0x00E5FE
	colorBotBub    = 0xFFFFFF
	colorChatText  = 0x202020
	colorInputBar  = 0xF2F2F2
	chatPendingMsg = "…"
)

type bubbleBox struct {
	fromUser bool
	text     *uint16
	tw, th   int // 글자가 차지하는 크기
}

func (c *chatWindow) paint(hwnd uintptr) {
	var ps paintStruct
	hdc, _, _ := pBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	defer pEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	var rc rect
	pGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
	w, h := int(rc.R), int(rc.B)
	if w <= 0 || h <= 0 {
		return
	}
	mem, _, _ := pCreateCompatibleDC.Call(hdc)
	bmp, _, _ := pCreateCompatibleBitmap.Call(hdc, uintptr(w), uintptr(h))
	old, _, _ := pSelectObject.Call(mem, bmp)
	c.drawChat(mem, w, h)
	pBitBlt.Call(hdc, 0, 0, uintptr(w), uintptr(h), mem, 0, 0, SRCCOPY)
	pSelectObject.Call(mem, old)
	pDeleteObject.Call(bmp)
	pDeleteDC.Call(mem)
}

func (c *chatWindow) drawChat(dc uintptr, w, h int) {
	pSelectObject.Call(dc, c.font)
	pSetBkMode.Call(dc, TRANSPARENT)
	pSetTextColor.Call(dc, colorChatText)
	fillSolid(dc, rect{0, 0, int32(w), int32(h)}, colorChatBg)

	viewH := h - int(chatBarH*c.dpi)
	boxes := c.measureBubbles(dc, w)
	contentH := c.contentHeight(boxes)
	c.scrollUp = clampInt(c.scrollUp, 0, max(0, contentH-viewH))
	y := c.gap() - (max(0, contentH-viewH) - c.scrollUp)
	for _, b := range boxes {
		c.drawBubble(dc, b, w, y)
		y += c.bubbleHeight(b) + c.gap()
	}
	fillSolid(dc, rect{0, int32(viewH), int32(w), int32(h)}, colorInputBar)
}

// 전체 메시지(대기 중이면 "…" 포함)의 글자 크기를 잰다
func (c *chatWindow) measureBubbles(dc uintptr, w int) []bubbleBox {
	maxW := max(int(80*c.dpi), int(float64(w)*0.68)-2*c.pad())
	items := append([]chatMessage(nil), c.msgs...)
	if c.pending {
		items = append(items, chatMessage{text: chatPendingMsg})
	}
	boxes := make([]bubbleBox, 0, len(items))
	for _, m := range items {
		t := u16(m.text)
		tw, th := measureWrapped(dc, t, maxW)
		boxes = append(boxes, bubbleBox{fromUser: m.fromUser, text: t, tw: tw, th: th})
	}
	return boxes
}

func measureWrapped(dc uintptr, t *uint16, maxW int) (w, h int) {
	r := rect{0, 0, int32(maxW), 0}
	pDrawTextW.Call(dc, uintptr(unsafe.Pointer(t)), ^uintptr(0), uintptr(unsafe.Pointer(&r)), DT_LEFT|DT_WORDBREAK|DT_CALCRECT|DT_NOPREFIX)
	return int(r.R), int(r.B)
}

func (c *chatWindow) contentHeight(boxes []bubbleBox) int {
	total := c.gap()
	for _, b := range boxes {
		total += c.bubbleHeight(b) + c.gap()
	}
	return total
}

func (c *chatWindow) drawBubble(dc uintptr, b bubbleBox, w, y int) {
	bw, bh := b.tw+2*c.pad(), c.bubbleHeight(b)
	x := c.margin()
	color := uintptr(colorBotBub)
	if b.fromUser {
		x, color = w-c.margin()-bw, colorUserBub
	}
	fillRoundRect(dc, x, y, bw, bh, int(14*c.dpi), color)
	tr := rect{int32(x + c.pad()), int32(y + c.pad()), int32(x + c.pad() + b.tw), int32(y + c.pad() + b.th)}
	pDrawTextW.Call(dc, uintptr(unsafe.Pointer(b.text)), ^uintptr(0), uintptr(unsafe.Pointer(&tr)), DT_LEFT|DT_WORDBREAK|DT_NOPREFIX)
}

// 휠로 스크롤. delta > 0 이면 위쪽(지난 대화)으로
func (c *chatWindow) scrollBy(delta int) {
	c.scrollUp += delta * int(60*c.dpi) / 120
	c.redraw()
}

func (c *chatWindow) bubbleHeight(b bubbleBox) int { return b.th + 2*c.pad() }
func (c *chatWindow) pad() int                     { return int(10 * c.dpi) }
func (c *chatWindow) gap() int                     { return int(8 * c.dpi) }
func (c *chatWindow) margin() int                  { return int(10 * c.dpi) }

func fillSolid(dc uintptr, r rect, color uintptr) {
	brush, _, _ := pCreateSolidBrush.Call(color)
	pFillRect.Call(dc, uintptr(unsafe.Pointer(&r)), brush)
	pDeleteObject.Call(brush)
}

// 테두리 없는 둥근 사각형
func fillRoundRect(dc uintptr, x, y, w, h, radius int, color uintptr) {
	brush, _, _ := pCreateSolidBrush.Call(color)
	pen, _, _ := pGetStockObject.Call(NULL_PEN)
	oldBrush, _, _ := pSelectObject.Call(dc, brush)
	oldPen, _, _ := pSelectObject.Call(dc, pen)
	pRoundRect.Call(dc, uintptr(x), uintptr(y), uintptr(x+w+1), uintptr(y+h+1), uintptr(radius), uintptr(radius))
	pSelectObject.Call(dc, oldPen)
	pSelectObject.Call(dc, oldBrush)
	pDeleteObject.Call(brush)
}

func clampInt(v, lo, hi int) int { return max(lo, min(hi, v)) }
