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
	shell  = syscall.NewLazyDLL("shell32.dll")

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
	pPostMessageW         = user32.NewProc("PostMessageW")
	pSendMessageW         = user32.NewProc("SendMessageW")
	pGetClientRect        = user32.NewProc("GetClientRect")
	pMoveWindow           = user32.NewProc("MoveWindow")
	pSetFocus             = user32.NewProc("SetFocus")
	pGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
	pGetWindowTextW       = user32.NewProc("GetWindowTextW")
	pSetWindowTextW       = user32.NewProc("SetWindowTextW")
	pBeginPaint           = user32.NewProc("BeginPaint")
	pEndPaint             = user32.NewProc("EndPaint")
	pInvalidateRect       = user32.NewProc("InvalidateRect")
	pFillRect             = user32.NewProc("FillRect")
	pMessageBoxW          = user32.NewProc("MessageBoxW")

	pDragQueryFileW   = shell.NewProc("DragQueryFileW")
	pDragFinish       = shell.NewProc("DragFinish")
	pSHFileOperationW = shell.NewProc("SHFileOperationW")
	pCreateMutexW     = kernel.NewProc("CreateMutexW")
	pGetModuleHandleW = kernel.NewProc("GetModuleHandleW")
	pSetDpiAwareness  = shcore.NewProc("SetProcessDpiAwareness")

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

	pCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	pBitBlt                 = gdi32.NewProc("BitBlt")
	pCreateSolidBrush       = gdi32.NewProc("CreateSolidBrush")
	pRoundRect              = gdi32.NewProc("RoundRect")
	pGetStockObject         = gdi32.NewProc("GetStockObject")
)

// 윈도우 메시지
const (
	WM_DESTROY        = 0x0002
	WM_SIZE           = 0x0005
	WM_SETFOCUS       = 0x0007
	WM_PAINT          = 0x000F
	WM_CLOSE          = 0x0010
	WM_ERASEBKGND     = 0x0014
	WM_SETFONT        = 0x0030
	WM_SETTINGCHANGE  = 0x001A
	WM_MOUSEACTIVATE  = 0x0021
	WM_DISPLAYCHANGE  = 0x007E
	WM_KEYDOWN        = 0x0100
	WM_COMMAND        = 0x0111
	WM_DROPFILES      = 0x0233
	WM_TIMER          = 0x0113
	WM_MOUSEMOVE      = 0x0200
	WM_LBUTTONDOWN    = 0x0201
	WM_LBUTTONUP      = 0x0202
	WM_LBUTTONDBLCLK  = 0x0203
	WM_RBUTTONUP      = 0x0205
	WM_MOUSEWHEEL     = 0x020A
	WM_CAPTURECHANGED = 0x0215
	WM_APP            = 0x8000
	VK_RETURN         = 0x0D
)

// 창 스타일 / 표시 / 위치
const (
	WS_POPUP                      = 0x80000000
	WS_EX_TOPMOST                 = 0x8
	WS_EX_ACCEPTFILES             = 0x10
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

// 일반 창(채팅 창)과 자식 컨트롤
const (
	CS_DBLCLKS          = 0x8
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_CLIPCHILDREN     = 0x02000000
	WS_CHILD            = 0x40000000
	WS_VISIBLE          = 0x10000000
	WS_TABSTOP          = 0x00010000
	WS_EX_CLIENTEDGE    = 0x200
	ES_AUTOHSCROLL      = 0x80
	BS_DEFPUSHBUTTON    = 0x1
	CW_USEDEFAULT       = 0x80000000
	SW_HIDE             = 0
	SW_SHOW             = 5
	SW_RESTORE          = 9
	IDC_ARROW           = 32512
	SRCCOPY             = 0x00CC0020
	NULL_PEN            = 8
)

// 파일 드롭, 휴지통, 메시지 박스
const (
	DRAGQUERY_COUNT     = 0xFFFFFFFF // DragQueryFileW: 파일 개수를 묻는 인덱스
	FO_DELETE           = 0x3
	FOF_SILENT          = 0x4    // 진행률 창 없음
	FOF_NOCONFIRMATION  = 0x10   // 확인 창 없음
	FOF_ALLOWUNDO       = 0x40   // 영구 삭제 대신 휴지통으로
	FOF_NOERRORUI       = 0x400  // 오류 창 없음
	FOF_WANTNUKEWARNING = 0x4000 // 휴지통에 못 넣고 영구 삭제하게 되면 그때는 경고
	MB_YESNO            = 0x4
	MB_ICONQUESTION     = 0x20
	MB_DEFBUTTON2       = 0x100
	MB_SETFOREGROUND    = 0x10000
	MB_TOPMOST          = 0x40000
	IDYES               = 6
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
	DT_LEFT           = 0x0
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
type paintStruct struct {
	Hdc       uintptr
	Erase     int32
	RcPaint   rect
	Restore   int32
	IncUpdate int32
	Reserved  [32]byte
}

// SHFILEOPSTRUCTW (64비트 배치)
type shFileOp struct {
	Hwnd         uintptr
	Func         uint32
	From         *uint16 // 경로마다 NUL로 끝나고 전체는 NUL 두 개로 끝난다
	To           *uint16
	Flags        uint16
	Aborted      int32 // 사용자가 중간에 취소했는지
	NameMappings uintptr
	Title        *uint16
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
