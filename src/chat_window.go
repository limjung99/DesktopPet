package main

// ───────────── 채팅 창 생성과 메시지 처리 ─────────────
// 일반 최상위 창 하나에 입력창(EDIT)과 전송 버튼(BUTTON)을 올린다.
// 대화 목록은 이 창이 직접 그린다(chat_view.go).

import (
	"syscall"
	"unsafe"
)

const (
	idChatSend = 101
	chatWidth  = 420
	chatHeight = 620
	chatBarH   = 44 // 아래 입력줄 높이(96dpi 기준)
)

var chatClassName = u16("DesktopPetChatWnd")

func newChatWindow(dpi float64) *chatWindow {
	c := &chatWindow{dpi: dpi, msgs: []chatMessage{{text: "안녕하세요! 무엇을 도와드릴까요?"}}}
	C = c // 창 생성 중에 오는 메시지(WM_SIZE 등)도 C를 쓴다
	c.font = createChatFont(dpi)
	inst, _, _ := pGetModuleHandleW.Call(0)
	registerChatWindowClass(inst)
	c.hwnd = createChatTopWindow(inst, dpi)
	c.edit = createChild(inst, c.hwnd, "EDIT", "", WS_EX_CLIENTEDGE, WS_TABSTOP|ES_AUTOHSCROLL, 0)
	c.send = createChild(inst, c.hwnd, "BUTTON", "전송", 0, WS_TABSTOP|BS_DEFPUSHBUTTON, idChatSend)
	pSendMessageW.Call(c.edit, WM_SETFONT, c.font, 1)
	pSendMessageW.Call(c.send, WM_SETFONT, c.font, 1)
	c.layoutChildren()
	return c
}

func createChatFont(dpi float64) uintptr {
	font, _, _ := pCreateFontW.Call(uintptr(-int32(14*dpi+0.5)), 0, 0, 0, 400, 0, 0, 0, DEFAULT_CHARSET, 0, 0, CLEARTYPE_QUALITY, 0,
		uintptr(unsafe.Pointer(u16("Malgun Gothic"))))
	return font
}

func registerChatWindowClass(inst uintptr) {
	cur, _, _ := pLoadCursorW.Call(0, IDC_ARROW)
	wc := wndClassEx{Size: uint32(unsafe.Sizeof(wndClassEx{})), WndProc: syscall.NewCallback(chatWndProc),
		Instance: inst, Cursor: cur, ClassName: chatClassName}
	pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
}

func createChatTopWindow(inst uintptr, dpi float64) uintptr {
	hwnd, _, _ := pCreateWindowExW.Call(0, uintptr(unsafe.Pointer(chatClassName)), uintptr(unsafe.Pointer(u16("데스크톱 펫 대화"))),
		WS_OVERLAPPEDWINDOW|WS_CLIPCHILDREN, CW_USEDEFAULT, CW_USEDEFAULT, uintptr(float64(chatWidth)*dpi), uintptr(float64(chatHeight)*dpi), 0, 0, inst, 0)
	return hwnd
}

// 시스템 클래스(EDIT, BUTTON) 자식 컨트롤. 위치와 크기는 layoutChildren이 정한다
func createChild(inst, parent uintptr, class, text string, exStyle, style, id uintptr) uintptr {
	h, _, _ := pCreateWindowExW.Call(exStyle, uintptr(unsafe.Pointer(u16(class))), uintptr(unsafe.Pointer(u16(text))),
		WS_CHILD|WS_VISIBLE|style, 0, 0, 10, 10, parent, id, inst, 0)
	return h
}

// 창 크기가 바뀔 때 아래 입력줄에 입력창과 버튼을 배치
func (c *chatWindow) layoutChildren() {
	if c.edit == 0 || c.send == 0 {
		return
	}
	var rc rect
	pGetClientRect.Call(c.hwnd, uintptr(unsafe.Pointer(&rc)))
	gap, btnW := int(8*c.dpi), int(64*c.dpi)
	barH := int(chatBarH * c.dpi)
	y, h := int(rc.B)-barH+gap, barH-2*gap
	pMoveWindow.Call(c.edit, uintptr(gap), uintptr(y), uintptr(int(rc.R)-btnW-3*gap), uintptr(h), 1)
	pMoveWindow.Call(c.send, uintptr(int(rc.R)-btnW-gap), uintptr(y), uintptr(btnW), uintptr(h), 1)
}

func chatWndProc(hwnd, m, wp, lp uintptr) uintptr {
	c := C
	if c == nil {
		r, _, _ := pDefWindowProcW.Call(hwnd, m, wp, lp)
		return r
	}
	switch m {
	case WM_SIZE:
		c.layoutChildren()
		pInvalidateRect.Call(hwnd, 0, 0)
		return 0
	case WM_PAINT:
		c.paint(hwnd)
		return 0
	case WM_ERASEBKGND: // paint에서 전체를 직접 그린다
		return 1
	case WM_COMMAND:
		if wp&0xFFFF == idChatSend {
			c.submit()
		}
		return 0
	case WM_MOUSEWHEEL:
		c.scrollBy(int(int16(wp >> 16)))
		return 0
	case WM_SETFOCUS:
		pSetFocus.Call(c.edit)
		return 0
	case WM_CLOSE: // 닫아도 대화 내용을 유지하려고 숨기기만 한다
		pShowWindow.Call(hwnd, SW_HIDE)
		return 0
	case msgChatReply:
		c.onReply()
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(hwnd, m, wp, lp)
	return r
}
