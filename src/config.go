package main

// ───────────── 사용자 설정 (exe와 같은 폴더의 pet.ini, lines.txt) ─────────────

import (
	"bufio"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	cfgScale      = 1.0
	cfgWalkSpeed  = 2.0
	cfgTalkChance = 0.15
)

var defaultLines = []string{"안녕!", "오늘도 화이팅!", "심심해~", "커피 한 잔 어때요?",
	"물 마실 시간!", "스트레칭 해요", "뭐 하고 있어?", "졸려...", "정환쿤 사랑해!"}

// 실행 파일이 있는 폴더 (설정·스프라이트 탐색 기준)
func exeDir() string {
	p, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(p)
}

// exe 폴더의 텍스트 파일을 줄 단위로 읽는다 (UTF-8 BOM 제거). 파일이 없으면 ok=false
func readConfigLines(name string) (lines []string, ok bool) {
	f, err := os.Open(filepath.Join(exeDir(), name))
	if err != nil {
		return nil, false
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, strings.TrimPrefix(sc.Text(), "\uFEFF"))
	}
	return lines, true
}

// lines.txt → 말풍선 대사 목록. 비어 있거나 없으면 기본 대사
func loadLines() []string {
	raw, ok := readConfigLines("lines.txt")
	if !ok {
		return defaultLines
	}
	var out []string
	for _, l := range raw {
		l = strings.TrimSpace(l)
		if isDialogueLine(l) {
			out = append(out, l)
		}
	}
	if len(out) == 0 {
		return defaultLines
	}
	return out
}

// 빈 줄과 # 주석은 대사가 아니다
func isDialogueLine(l string) bool {
	return l != "" && !strings.HasPrefix(l, "#")
}

// pet.ini → cfg* 전역값 갱신. 파일이 없으면 기본값 유지
func loadIni() {
	raw, ok := readConfigLines("pet.ini")
	if !ok {
		return
	}
	for _, l := range raw {
		if key, val, ok := parseIniLine(l); ok {
			applyIniValue(key, val)
		}
	}
}

// "key=숫자" 한 줄 해석. 형식이 맞지 않으면 ok=false (; 주석 줄도 여기서 걸러짐)
func parseIniLine(line string) (key string, val float64, ok bool) {
	k, v, found := strings.Cut(line, "=")
	if !found {
		return "", 0, false
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	if err != nil {
		return "", 0, false
	}
	return strings.ToLower(strings.TrimSpace(k)), val, true
}

func applyIniValue(key string, val float64) {
	switch key {
	case "scale":
		cfgScale = math.Max(0.2, math.Min(5, val))
	case "walk_speed":
		cfgWalkSpeed = val
	case "talk_chance":
		cfgTalkChance = val
	}
}
