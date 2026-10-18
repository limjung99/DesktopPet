package main

// ───────────── 대화 상태, 전송, 응답 수신 ─────────────
// LLM 호출은 고루틴에서 하고, 결과는 창 메시지(msgChatReply)로 UI 스레드에 넘긴다.

import (
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

const msgChatReply = WM_APP + 1

type chatMessage struct {
	fromUser bool
	text     string
}

type chatWindow struct {
	hwnd, edit, send, font uintptr
	dpi                    float64
	msgs                   []chatMessage
	pending                bool // 답변을 기다리는 중
	scrollUp               int  // 맨 아래에서 위로 올라간 픽셀
}

// 채팅 창 프로시저에서 접근하는 유일한 채팅 창
var C *chatWindow

// 고루틴이 UI 스레드로 넘기는 답변 보관함
var chatInbox struct {
	sync.Mutex
	replies []chatMessage
}

// 대화 창을 열거나, 이미 있으면 앞으로 가져온다
func (p *pet) openChat() {
	if C == nil {
		C = newChatWindow(p.dpi)
	}
	C.show()
}

func (c *chatWindow) show() {
	pShowWindow.Call(c.hwnd, SW_RESTORE)
	pSetForegroundWindow.Call(c.hwnd)
	pSetFocus.Call(c.edit)
}

// 입력창의 글자를 보낸다. 빈 입력이거나 답변을 기다리는 중이면 무시
func (c *chatWindow) submit() {
	text := strings.TrimSpace(c.editText())
	if text == "" || c.pending {
		return
	}
	c.msgs = append(c.msgs, chatMessage{fromUser: true, text: text})
	c.pending, c.scrollUp = true, 0
	pSetWindowTextW.Call(c.edit, uintptr(unsafe.Pointer(u16(""))))
	pSetFocus.Call(c.edit)
	c.redraw()
	go requestReply(c.hwnd, append([]chatMessage(nil), c.msgs...))
}

func (c *chatWindow) editText() string {
	n, _, _ := pGetWindowTextLengthW.Call(c.edit)
	buf := make([]uint16, n+1)
	pGetWindowTextW.Call(c.edit, uintptr(unsafe.Pointer(&buf[0])), n+1)
	return syscall.UTF16ToString(buf)
}

// (고루틴) LLM 답변을 받아 보관함에 넣고 창에 알린다. 실패하면 오류 문구가 답변이 된다
func requestReply(hwnd uintptr, history []chatMessage) {
	text, err := askLLM(history)
	if err != nil {
		text = "⚠ " + err.Error()
	}
	chatInbox.Lock()
	chatInbox.replies = append(chatInbox.replies, chatMessage{text: text})
	chatInbox.Unlock()
	pPostMessageW.Call(hwnd, msgChatReply, 0, 0)
}

// (UI 스레드) 보관함의 답변을 대화에 붙인다
func (c *chatWindow) onReply() {
	chatInbox.Lock()
	replies := chatInbox.replies
	chatInbox.replies = nil
	chatInbox.Unlock()
	c.msgs = append(c.msgs, replies...)
	c.pending, c.scrollUp = false, 0
	c.redraw()
}

func (c *chatWindow) redraw() {
	pInvalidateRect.Call(c.hwnd, 0, 0)
}

// 메시지 루프에서 호출: 입력창에서 Enter를 누르면 전송하고 true
func handleChatEnter(m *msg) bool {
	if C == nil || m.Message != WM_KEYDOWN || m.WParam != VK_RETURN || m.Hwnd != C.edit {
		return false
	}
	C.submit()
	return true
}
