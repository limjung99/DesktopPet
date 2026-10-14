package main

// ───────────── Win32 바인딩 ─────────────
// cgo 없이 syscall.NewLazyDLL로 필요한 API만 불러온다.

import (
	"syscall"
	"unsafe"
)

var (
	user32 = syscall.NewLazyDLL("user32.dll")
	gdi32  = syscall.NewLazyDLL("gdi32.dll")
	kernel = syscall.NewLazyDLL("kernel32.dll")
	shcore = syscall.NewLazyDLL("shcore.dll")

	pRegisterClassExW     = user32.NewProc("RegisterClassExW")
	pCreateWindowExW      = user32.NewProc("CreateWindowExW")
	pDefWindowProcW       = user32.NewProc("DefWindowProcW")
	pGetMessageW          = user32.NewProc("GetMessageW")
	pTranslateMessage     = user32.NewProc("TranslateMessage")
	pDispatchMessageW     = user32.NewProc("DispatchMessageW")
	pPostQuitMessage      = user32.NewProc("PostQuitMessage")
	pSetTimer             = user32.NewProc("SetTimer")
	pUpdateLayeredWindow  = user32.NewProc("UpdateLayeredWindow")
	pGetDC                = user32.NewProc("GetDC")
	pReleaseDC            = user32.NewProc("ReleaseDC")
	pSetCapture           = user32.NewProc("SetCapture")
	pReleaseCapture       = user32.NewProc("ReleaseCapture")
	pGetCursorPos         = user32.NewProc("GetCursorPos")
	pCreatePopupMenu      = user32.NewProc("CreatePopupMenu")
	pAppendMenuW          = user32.NewProc("AppendMenuW")
	pTrackPopupMenu       = user32.NewProc("TrackPopupMenu")
	pDestroyMenu          = user32.NewProc("DestroyMenu")
	pSetForegroundWindow  = user32.NewProc("SetForegroundWindow")
	pSetWindowPos         = user32.NewProc("SetWindowPos")
	pDestroyWindow        = user32.NewProc("DestroyWindow")
	pShowWindow           = user32.NewProc("ShowWindow")
	pSystemParametersInfo = user32.NewProc("SystemParametersInfoW")
	pLoadCursorW          = user32.NewProc("LoadCursorW")
	pEnumDisplayMonitors  = user32.NewProc("EnumDisplayMonitors")
	pGetMonitorInfoW      = user32.NewProc("GetMonitorInfoW")
	pSetProcessDPIAware   = user32.NewProc("SetProcessDPIAware")
	pDrawTextW            = user32.NewProc("DrawTextW")
	pCreateMutexW         = kernel.NewProc("CreateMutexW")
	pGetModuleHandleW     = kernel.NewProc("GetModuleHandleW")
	pSetDpiAwareness      = shcore.NewProc("SetProcessDpiAwareness")

	pCreateCompatibleDC = gdi32.NewProc("CreateCompatibleDC")
	pCreateDIBSection   = gdi32.NewProc("CreateDIBSection")
	pSelectObject       = gdi32.NewProc("SelectObject")
	pDeleteObject       = gdi32.NewProc("DeleteObject")
	pDeleteDC           = gdi32.NewProc("DeleteDC")
	pCreateFontW        = gdi32.NewProc("CreateFontW")
	pSetTextColor       = gdi32.NewProc("SetTextColor")
	pSetBkMode          = gdi32.NewProc("SetBkMode")
	pGetDeviceCaps      = gdi32.NewProc("GetDeviceCaps")
	pGdiFlush           = gdi32.NewProc("GdiFlush")
)

// 윈도우 메시지
const (
	WM_DESTROY        = 0x0002
	WM_SETTINGCHANGE  = 0x001A
	WM_MOUSEACTIVATE  = 0x0021
	WM_DISPLAYCHANGE  = 0x007E
	WM_TIMER          = 0x0113
	WM_MOUSEMOVE      = 0x0200
	WM_LBUTTONDOWN    = 0x0201
	WM_LBUTTONUP      = 0x0202
	WM_RBUTTONUP      = 0x0205
	WM_CAPTURECHANGED = 0x0215
)

// 창 스타일 / 표시 / 위치
const (
	WS_POPUP                      = 0x80000000
	WS_EX_TOPMOST                 = 0x8
	WS_EX_TOOLWINDOW              = 0x80
	WS_EX_LAYERED                 = 0x80000
	WS_EX_NOACTIVATE              = 0x08000000
	SW_SHOWNOACTIVATE             = 4
	MA_NOACTIVATE                 = 3
	HWND_TOPMOST                  = ^uintptr(0) // (HWND)-1
	SWP_NOSIZE                    = 0x1
	SWP_NOMOVE                    = 0x2
	SWP_NOACTIVATE                = 0x10
	IDC_HAND                      = 32649
	ULW_ALPHA                     = 2
	AC_SRC_OVER                   = 0
	AC_SRC_ALPHA                  = 1
	SPI_GETWORKAREA               = 0x30
	LOGPIXELSY                    = 90
	ERROR_ALREADY_EXISTS          = syscall.Errno(183)
	PROCESS_PER_MONITOR_DPI_AWARE = 2
)

// 메뉴
const (
	MF_STRING       = 0x0
	MF_SEPARATOR    = 0x800
	TPM_RIGHTBUTTON = 0x2
	TPM_RETURNCMD   = 0x100
)

// GDI 텍스트
const (
	DT_CENTER         = 0x1
	DT_WORDBREAK      = 0x10
	DT_CALCRECT       = 0x400
	DT_NOPREFIX       = 0x800
	TRANSPARENT       = 1
	DEFAULT_CHARSET   = 1
	CLEARTYPE_QUALITY = 5
)

type point struct{ X, Y int32 }
type size struct{ CX, CY int32 }
type rect struct{ L, T, R, B int32 }
type msg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      point
	Priv    uint32
}
type wndClassEx struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   uintptr
	Icon       uintptr
	Cursor     uintptr
	Background uintptr
	MenuName   *uint16
	ClassName  *uint16
	IconSm     uintptr
}
type bitmapInfoHeader struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}
type blendFunc struct{ Op, Flags, Alpha, Format byte }

// Go 문자열 → Win32용 UTF-16 포인터
func u16(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }

// ───────────── DIB (GDI 비트맵 버퍼) ─────────────

// 32비트 DIB (top-down BGRA). pix는 비트맵 메모리를 직접 가리킨다.
type dib struct {
	dc, bmp uintptr
	w, h    int
	pix     []byte
}

func newDIB(w, h int) *dib {
	dc := newMemoryDC()
	bi := bitmapInfoHeader{Size: 40, Width: int32(w), Height: -int32(h), Planes: 1, BitCount: 32}
	var bits uintptr
	bmp, _, _ := pCreateDIBSection.Call(dc, uintptr(unsafe.Pointer(&bi)), 0, uintptr(unsafe.Pointer(&bits)), 0, 0)
	pSelectObject.Call(dc, bmp)
	return &dib{dc: dc, bmp: bmp, w: w, h: h, pix: unsafe.Slice((*byte)(unsafe.Pointer(bits)), w*h*4)}
}

func (d *dib) free() { pDeleteObject.Call(d.bmp); pDeleteDC.Call(d.dc) }

// 화면 DC와 호환되는 메모리 DC
func newMemoryDC() uintptr {
	scr, _, _ := pGetDC.Call(0)
	dc, _, _ := pCreateCompatibleDC.Call(scr)
	pReleaseDC.Call(0, scr)
	return dc
}
