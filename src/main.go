// 데스크톱 펫 (Windows 전용, 외부 런타임 불필요)
//
// 조작: 왼쪽 클릭 = 점프 + 말풍선 / 드래그 = 집어서 던지기 / 오른쪽 클릭 = 메뉴(말하기, 점프, 종료)
// 사용자 설정(선택, exe와 같은 폴더):
//
//	sprites\idle_0.png ...  캐릭터 이미지 교체
//	lines.txt               말풍선 대사 (UTF-8, 한 줄에 하나)
//	pet.ini                 scale=1.0 / walk_speed=2 / talk_chance=0.15
//
// 파일 구성:
//
//	main.go       시작 순서
//	win32.go      Win32 API 바인딩, 상수, DIB
//	window.go     창 생성, 메시지 처리, 타이머
//	pet.go        펫 상태 구조체와 초기 배치
//	physics.go    상태 머신(idle/walk/jump/fall)과 물리
//	monitor.go    다중 모니터 영역, 충돌 판정, 화면 밖 보정
//	input.go      마우스 클릭/드래그/메뉴
//	render.go     프레임 그리기와 화면 출력
//	bubble.go     말풍선 텍스트 렌더링
//	animation.go  스프라이트 로딩
//	sprite.go     이미지 스케일링, 반전, 알파 합성
//	config.go     pet.ini, lines.txt 읽기
package main

import (
	"os"
	"runtime"
	"unsafe"
)

func main() {
	runtime.LockOSThread() // Win32 창과 메시지 루프는 같은 OS 스레드에 있어야 한다
	if alreadyRunning() {
		return
	}
	enableDPIAwareness()
	loadIni()

	p := newPet(systemDPIScale())
	if p == nil { // 캐릭터 이미지를 찾지 못함
		return
	}
	P = p
	p.hwnd = createPetWindow(p.W, p.H)
	p.render()
	pShowWindow.Call(p.hwnd, SW_SHOWNOACTIVATE)
	startTimers(p.hwnd)
	if os.Getenv("PET_TEST") != "" { // 테스트용: 시작하자마자 말풍선
		p.say("테스트 말풍선입니다", 30)
	}
	runMessageLoop()
}

// 중복 실행 방지: 이름 있는 뮤텍스가 이미 있으면 다른 인스턴스가 실행 중
func alreadyRunning() bool {
	_, _, err := pCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(u16("Local\\DesktopPetSingleInstance"))))
	return err == ERROR_ALREADY_EXISTS
}

// 모니터별 DPI 인식 (Windows 8.1 미만이면 시스템 DPI 인식)
func enableDPIAwareness() {
	if pSetDpiAwareness.Find() == nil {
		pSetDpiAwareness.Call(PROCESS_PER_MONITOR_DPI_AWARE)
	} else {
		pSetProcessDPIAware.Call()
	}
}

// 시스템 DPI 배율 (96dpi = 1.0)
func systemDPIScale() float64 {
	scr, _, _ := pGetDC.Call(0)
	dpiY, _, _ := pGetDeviceCaps.Call(scr, LOGPIXELSY)
	pReleaseDC.Call(0, scr)
	if dpi := float64(dpiY) / 96; dpi > 0 {
		return dpi
	}
	return 1
}
