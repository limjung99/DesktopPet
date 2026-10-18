package main

// ───────────── 창 생성, 메시지 처리, 타이머 ─────────────

import (
	"syscall"
	"unsafe"
)

const (
	timerFrame       = 1 // 매 프레임 update + render
	timerMaintenance = 2 // 최상위 유지 + 모니터 구성 재확인
)

// 투명 배경, 항상 위, 작업표시줄에 안 보이고 포커스를 가져가지 않는 팝업 창
func createPetWindow(w, h int) uintptr {
	inst, _, _ := pGetModuleHandleW.Call(0)
	cls := registerPetWindowClass(inst)
	const exStyle = WS_EX_LAYERED | WS_EX_TOPMOST | WS_EX_TOOLWINDOW | WS_EX_NOACTIVATE | WS_EX_ACCEPTFILES
	hwnd, _, _ := pCreateWindowExW.Call(exStyle, uintptr(unsafe.Pointer(cls)), uintptr(unsafe.Pointer(u16("Desktop Pet"))),
		WS_POPUP, 0, 0, uintptr(w), uintptr(h), 0, 0, inst, 0)
	return hwnd
}

// 창 클래스 등록 (커서는 손 모양). 반환: 클래스 이름
func registerPetWindowClass(inst uintptr) *uint16 {
	cur, _, _ := pLoadCursorW.Call(0, IDC_HAND)
	cls := u16("DesktopPetWnd")
	wc := wndClassEx{Size: uint32(unsafe.Sizeof(wndClassEx{})), WndProc: syscall.NewCallback(wndProc),
		Style: CS_DBLCLKS, Instance: inst, Cursor: cur, ClassName: cls}
	pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	return cls
}

func startTimers(hwnd uintptr) {
	pSetTimer.Call(hwnd, timerFrame, 1000/fps, 0)
	pSetTimer.Call(hwnd, timerMaintenance, 2000, 0)
}

// 창이 닫힐 때까지 메시지 처리
func runMessageLoop() {
	var m msg
	for {
		r, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		if handleChatEnter(&m) {
			continue
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		pDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func wndProc(hwnd, m, wp, lp uintptr) uintptr {
	switch m {
	case WM_TIMER:
		P.onTimer(wp)
		return 0
	case WM_LBUTTONDOWN:
		P.onDown()
		return 0
	case WM_LBUTTONDBLCLK:
		P.openChat()
		return 0
	case WM_MOUSEMOVE:
		P.onMove()
		return 0
	case WM_LBUTTONUP:
		P.onUp()
		return 0
	case WM_CAPTURECHANGED: // 드래그 중 캡처를 잃으면 놓은 것으로 처리
		if P.pressing {
			P.onUp()
		}
		return 0
	case WM_DROPFILES:
		P.onDropFiles(wp)
		return 0
	case msgTrashRequest:
		P.handleTrashRequest()
		return 0
	case WM_RBUTTONUP:
		P.onMenu()
		return 0
	case WM_MOUSEACTIVATE:
		return MA_NOACTIVATE
	case WM_DISPLAYCHANGE, WM_SETTINGCHANGE:
		P.refreshMonitors() // 기본 처리도 이어서 하도록 return 하지 않음
	case WM_DESTROY:
		pPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(hwnd, m, wp, lp)
	return r
}

func (p *pet) onTimer(id uintptr) {
	switch id {
	case timerFrame:
		p.update()
		p.render()
	case timerMaintenance:
		p.keepOnTop()
		p.refreshMonitors() // 모니터 연결/해제, 작업표시줄 위치 변경 반영
	}
}

// 다른 최상위 창에 가려지지 않도록 다시 맨 위로
func (p *pet) keepOnTop() {
	pSetWindowPos.Call(p.hwnd, HWND_TOPMOST, 0, 0, 0, 0, SWP_NOSIZE|SWP_NOMOVE|SWP_NOACTIVATE)
}
