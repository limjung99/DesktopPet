package main

// ───────────── LLM API 키 입력 창 ─────────────
// 첫 실행(저장된 키 없음) 때와 메뉴 "API 키 설정"에서 연다.
// 안내 문구(STATIC), 가려진 입력창(EDIT, ES_PASSWORD), 저장 버튼(BUTTON)으로 이뤄진다.

import (
	"strings"
	"syscall"
	"unsafe"
)

const (
	idKeySave = 201
	keyWidth  = 420
	keyHeight = 170
)

type apiKeyWindow struct {
	hwnd, label, edit, save, font uintptr
	dpi                           float64
}

// 키 입력 창 프로시저에서 접근하는 유일한 키 입력 창
var K *apiKeyWindow

var keyClassName = u16("DesktopPetAPIKeyWnd")

// 키 입력 창을 열거나, 이미 있으면 앞으로 가져온다
func openAPIKeyWindow(dpi float64) {
	if K == nil {
		K = newAPIKeyWindow(dpi)
	}
	K.show()
}

func newAPIKeyWindow(dpi float64) *apiKeyWindow {
	k := &apiKeyWindow{dpi: dpi}
	K = k // 창 생성 중에 오는 메시지도 K를 쓴다
	k.font = createChatFont(dpi)
	inst, _, _ := pGetModuleHandleW.Call(0)
	registerAPIKeyWindowClass(inst)
	k.hwnd = createAPIKeyTopWindow(inst, dpi)
	k.label = createChild(inst, k.hwnd, "STATIC", "사내 LLM API 키를 입력하세요.\r\n이 PC의 현재 사용자 계정으로 암호화해 저장합니다.", 0, 0, 0)
	k.edit = createChild(inst, k.hwnd, "EDIT", "", WS_EX_CLIENTEDGE, WS_TABSTOP|ES_AUTOHSCROLL|ES_PASSWORD, 0)
	k.save = createChild(inst, k.hwnd, "BUTTON", "저장", 0, WS_TABSTOP|BS_DEFPUSHBUTTON, idKeySave)
	for _, h := range []uintptr{k.label, k.edit, k.save} {
		pSendMessageW.Call(h, WM_SETFONT, k.font, 1)
	}
	k.layoutChildren()
	return k
}

func registerAPIKeyWindowClass(inst uintptr) {
	cur, _, _ := pLoadCursorW.Call(0, IDC_ARROW)
	wc := wndClassEx{Size: uint32(unsafe.Sizeof(wndClassEx{})), WndProc: syscall.NewCallback(apiKeyWndProc),
		Instance: inst, Cursor: cur, Background: COLOR_BTNFACE + 1, ClassName: keyClassName}
	pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
}

// 크기 조절이 안 되는 작은 창 (항상 위에 떠서 펫 뒤로 숨지 않는다)
func createAPIKeyTopWindow(inst uintptr, dpi float64) uintptr {
	hwnd, _, _ := pCreateWindowExW.Call(WS_EX_TOPMOST, uintptr(unsafe.Pointer(keyClassName)), uintptr(unsafe.Pointer(u16("데스크톱 펫 API 키 설정"))),
		WS_CAPTION|WS_SYSMENU, CW_USEDEFAULT, CW_USEDEFAULT, uintptr(float64(keyWidth)*dpi), uintptr(float64(keyHeight)*dpi), 0, 0, inst, 0)
	return hwnd
}

// 위에서부터 안내 문구, 입력창, 오른쪽 아래 저장 버튼
func (k *apiKeyWindow) layoutChildren() {
	if k.label == 0 || k.edit == 0 || k.save == 0 {
		return
	}
	var rc rect
	pGetClientRect.Call(k.hwnd, uintptr(unsafe.Pointer(&rc)))
	gap, rowH, btnW := int(12*k.dpi), int(28*k.dpi), int(72*k.dpi)
	w := int(rc.R) - 2*gap
	pMoveWindow.Call(k.label, uintptr(gap), uintptr(gap), uintptr(w), uintptr(2*rowH), 1)
	pMoveWindow.Call(k.edit, uintptr(gap), uintptr(gap+2*rowH), uintptr(w-btnW-gap), uintptr(rowH), 1)
	pMoveWindow.Call(k.save, uintptr(int(rc.R)-gap-btnW), uintptr(gap+2*rowH), uintptr(btnW), uintptr(rowH), 1)
}

func (k *apiKeyWindow) show() {
	pSetWindowTextW.Call(k.edit, uintptr(unsafe.Pointer(u16(""))))
	pShowWindow.Call(k.hwnd, SW_RESTORE)
	pSetForegroundWindow.Call(k.hwnd)
	pSetFocus.Call(k.edit)
}

// 입력한 키를 암호화해 저장하고 바로 쓰기 시작한다. 빈 입력이면 무시
func (k *apiKeyWindow) submit() {
	key := strings.TrimSpace(windowText(k.edit))
	if key == "" {
		pSetFocus.Call(k.edit)
		return
	}
	if err := saveAPIKey(key); err != nil {
		pSetWindowTextW.Call(k.label, uintptr(unsafe.Pointer(u16("저장하지 못했어요: "+err.Error()))))
		return
	}
	llmAPIKey = key
	pSetWindowTextW.Call(k.edit, uintptr(unsafe.Pointer(u16(""))))
	pShowWindow.Call(k.hwnd, SW_HIDE)
	if P != nil {
		P.say("API 키를 저장했어요!", 3)
	}
}

func windowText(hwnd uintptr) string {
	n, _, _ := pGetWindowTextLengthW.Call(hwnd)
	buf := make([]uint16, n+1)
	pGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), n+1)
	return syscall.UTF16ToString(buf)
}

func apiKeyWndProc(hwnd, m, wp, lp uintptr) uintptr {
	k := K
	if k == nil {
		r, _, _ := pDefWindowProcW.Call(hwnd, m, wp, lp)
		return r
	}
	switch m {
	case WM_SIZE:
		k.layoutChildren()
		return 0
	case WM_COMMAND:
		if wp&0xFFFF == idKeySave {
			k.submit()
		}
		return 0
	case WM_SETFOCUS:
		pSetFocus.Call(k.edit)
		return 0
	case WM_CLOSE: // 다시 열 수 있게 숨기기만 한다
		pShowWindow.Call(hwnd, SW_HIDE)
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(hwnd, m, wp, lp)
	return r
}

// 메시지 루프에서 호출: 키 입력창에서 Enter를 누르면 저장하고 true
func handleAPIKeyEnter(m *msg) bool {
	if K == nil || m.Message != WM_KEYDOWN || m.WParam != VK_RETURN || m.Hwnd != K.edit {
		return false
	}
	K.submit()
	return true
}
