package main

// ───────────── 파일 드롭 → 휴지통 ─────────────
// 탐색기에서 파일을 펫에 놓으면 WM_DROPFILES가 온다.
// 드롭을 처리하는 동안 탐색기가 기다리므로, 경로만 꺼내 두고 확인과 삭제는 메시지로 미룬다.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

const (
	msgTrashRequest  = WM_APP + 2
	maxNamesInPrompt = 5
)

// WM_DROPFILES 처리: 경로를 꺼내 보관하고 본 처리는 나중으로 미룬다
func (p *pet) onDropFiles(hdrop uintptr) {
	paths := queryDroppedPaths(hdrop)
	pDragFinish.Call(hdrop)
	if len(paths) == 0 {
		return
	}
	p.droppedPaths = paths
	pPostMessageW.Call(p.hwnd, msgTrashRequest, 0, 0)
}

// HDROP에서 파일 경로 목록을 꺼낸다
func queryDroppedPaths(hdrop uintptr) []string {
	n, _, _ := pDragQueryFileW.Call(hdrop, DRAGQUERY_COUNT, 0, 0)
	paths := make([]string, 0, n)
	for i := uintptr(0); i < n; i++ {
		size, _, _ := pDragQueryFileW.Call(hdrop, i, 0, 0) // NUL 제외 길이
		buf := make([]uint16, size+1)
		pDragQueryFileW.Call(hdrop, i, uintptr(unsafe.Pointer(&buf[0])), size+1)
		paths = append(paths, syscall.UTF16ToString(buf))
	}
	return paths
}

// 보관해 둔 경로를 검사하고, 사용자 확인을 받은 뒤 휴지통으로 보낸다
func (p *pet) handleTrashRequest() {
	paths := p.droppedPaths
	p.droppedPaths = nil
	if protected := firstProtectedPath(paths); protected != "" {
		p.say("이건 못 버려요!", 2.5)
		return
	}
	if !confirmTrash(paths) {
		p.say("안 버릴게요", 2)
		return
	}
	aborted, err := moveToTrash(paths)
	p.say(trashResultMessage(len(paths), aborted, err), 3)
}

func trashResultMessage(count int, aborted bool, err error) string {
	switch {
	case err != nil:
		return fmt.Sprintf("버리지 못했어요 (%v)", err)
	case aborted:
		return "취소했어요"
	}
	return fmt.Sprintf("%d개를 휴지통에 버렸어요!", count)
}

// 드라이브 루트나 Windows 폴더 안의 경로는 막는다. 해당 경로가 없으면 ""
func firstProtectedPath(paths []string) string {
	for _, p := range paths {
		if isDriveRoot(p) || isInsideWindowsDir(p) {
			return p
		}
	}
	return ""
}

func isDriveRoot(path string) bool {
	clean := filepath.Clean(path)
	return clean == filepath.VolumeName(clean)+`\`
}

func isInsideWindowsDir(path string) bool {
	win := os.Getenv("SystemRoot")
	if win == "" {
		return false
	}
	p, w := strings.ToLower(filepath.Clean(path)), strings.ToLower(filepath.Clean(win))
	return p == w || strings.HasPrefix(p, w+`\`)
}

// "정말 버릴까요?" 예/아니오 창 (기본 선택은 아니오)
func confirmTrash(paths []string) bool {
	text := u16(trashPromptText(paths))
	title := u16("휴지통으로 버리기")
	r, _, _ := pMessageBoxW.Call(0, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(title)),
		MB_YESNO|MB_ICONQUESTION|MB_DEFBUTTON2|MB_SETFOREGROUND|MB_TOPMOST)
	return r == IDYES
}

func trashPromptText(paths []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d개 항목을 휴지통으로 보낼까요?\n\n", len(paths))
	for i, p := range paths {
		if i == maxNamesInPrompt {
			fmt.Fprintf(&b, "… 외 %d개", len(paths)-maxNamesInPrompt)
			break
		}
		b.WriteString(filepath.Base(p) + "\n")
	}
	return b.String()
}
