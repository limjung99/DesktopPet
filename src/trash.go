package main

// ───────────── 휴지통으로 보내기 ─────────────

import (
	"fmt"
	"syscall"
	"unsafe"
)

// 반환: 사용자가 취소했는지, 실패 시 오류
func moveToTrash(paths []string) (aborted bool, err error) {
	op := shFileOp{
		Func:  FO_DELETE,
		From:  doubleNulPaths(paths),
		Flags: FOF_ALLOWUNDO | FOF_NOCONFIRMATION | FOF_WANTNUKEWARNING | FOF_NOERRORUI | FOF_SILENT,
	}
	ret, _, _ := pSHFileOperationW.Call(uintptr(unsafe.Pointer(&op)))
	if ret != 0 {
		return false, fmt.Errorf("오류 코드 0x%X", ret)
	}
	return op.Aborted != 0, nil
}

// "경로1\0경로2\0\0" 형태의 UTF-16 문자열 (SHFileOperationW의 pFrom 형식)
func doubleNulPaths(paths []string) *uint16 {
	var buf []uint16
	for _, p := range paths {
		buf = append(buf, syscall.StringToUTF16(p)...) // 끝의 NUL까지 포함
	}
	buf = append(buf, 0)
	return &buf[0]
}
