package main

// ───────────── LLM API 키 저장/읽기 ─────────────
// 키는 %APPDATA%\DesktopPet\llm_api_key.bin에 DPAPI로 암호화해 둔다.
// 같은 Windows 사용자 계정에서만 풀 수 있어서, 파일을 복사해 가도 다른 PC·계정에서는 못 읽는다.

import (
	"errors"
	"os"
	"path/filepath"
	"unsafe"
)

const apiKeyFileName = "llm_api_key.bin"

func apiKeyPath() (string, error) {
	dir, err := os.UserConfigDir() // Windows에서는 %APPDATA%
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "DesktopPet", apiKeyFileName), nil
}

// 저장된 키. 없거나 풀 수 없으면 ""
func loadAPIKey() string {
	path, err := apiKeyPath()
	if err != nil {
		return ""
	}
	sealed, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	plain, err := unprotectBytes(sealed)
	if err != nil {
		return ""
	}
	return string(plain)
}

func saveAPIKey(key string) error {
	path, err := apiKeyPath()
	if err != nil {
		return err
	}
	sealed, err := protectBytes([]byte(key))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, sealed, 0o600)
}

// DPAPI 암호화 (현재 사용자 계정 키)
func protectBytes(plain []byte) ([]byte, error) {
	if len(plain) == 0 {
		return nil, errors.New("빈 값은 암호화하지 않는다")
	}
	in := dataBlob{Size: uint32(len(plain)), Data: &plain[0]}
	var out dataBlob
	r, _, err := pCryptProtectData.Call(uintptr(unsafe.Pointer(&in)), 0, 0, 0, 0, CRYPTPROTECT_UI_FORBIDDEN, uintptr(unsafe.Pointer(&out)))
	if r == 0 {
		return nil, err
	}
	return takeBlob(out), nil
}

// DPAPI 복호화
func unprotectBytes(sealed []byte) ([]byte, error) {
	if len(sealed) == 0 {
		return nil, errors.New("빈 파일")
	}
	in := dataBlob{Size: uint32(len(sealed)), Data: &sealed[0]}
	var out dataBlob
	r, _, err := pCryptUnprotectData.Call(uintptr(unsafe.Pointer(&in)), 0, 0, 0, 0, CRYPTPROTECT_UI_FORBIDDEN, uintptr(unsafe.Pointer(&out)))
	if r == 0 {
		return nil, err
	}
	return takeBlob(out), nil
}

// DPAPI가 LocalAlloc으로 돌려준 버퍼를 Go 슬라이스로 복사하고 해제
func takeBlob(b dataBlob) []byte {
	defer pLocalFree.Call(uintptr(unsafe.Pointer(b.Data)))
	return append([]byte(nil), unsafe.Slice(b.Data, b.Size)...)
}
