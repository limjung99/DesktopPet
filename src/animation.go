package main

// ───────────── 애니메이션 프레임 로딩 ─────────────
// 파일 이름 규칙: <애니메이션>_<번호>.png (예: walk_0.png, walk_1.png ...)
// exe 옆 sprites\ 폴더에 PNG가 하나라도 있으면 내장 이미지 대신 그것을 쓴다.

import (
	"bytes"
	"embed"
	"image"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

//go:embed sprites/*.png
var embedded embed.FS

var animNames = []string{"idle", "walk", "drag", "fall", "jump"}

// 애니메이션별 초당 프레임 수
var animFPS = map[string]int{"idle": 3, "walk": 8, "drag": 6, "fall": 6, "jump": 6}

// 이미지가 없는 애니메이션은 순서대로 찾은 다른 애니메이션으로 대체
var animFallbacks = map[string][]string{"idle": {"walk", "drag", "fall", "jump"}, "walk": {"idle"},
	"drag": {"fall", "idle"}, "fall": {"drag", "jump", "idle"}, "jump": {"fall", "idle"}}

// 오른쪽/왼쪽을 보는 프레임 목록
type anim struct{ right, left []*sprite }

var frameNo = regexp.MustCompile(`_(\d+)\.png$`)

func loadSprites(scale float64) map[string]*anim {
	fsys := spriteSource()
	res := map[string]*anim{}
	for _, name := range animNames {
		if a := loadAnim(fsys, name, scale); a != nil {
			res[name] = a
		}
	}
	applyAnimFallbacks(res)
	return res
}

// 사용자 sprites\ 폴더가 있으면 그것, 없으면 내장 이미지
func spriteSource() fs.FS {
	dir := filepath.Join(exeDir(), "sprites")
	if m, _ := filepath.Glob(filepath.Join(dir, "*.png")); len(m) > 0 {
		return os.DirFS(dir)
	}
	sub, _ := fs.Sub(embedded, "sprites")
	return sub
}

// 애니메이션 하나의 프레임을 모두 읽는다. 읽을 수 있는 프레임이 없으면 nil
func loadAnim(fsys fs.FS, name string, scale float64) *anim {
	a := &anim{}
	for _, f := range frameFiles(fsys, name) {
		img, err := decodePNG(fsys, f)
		if err != nil {
			continue
		}
		sp := fromImage(img, scale)
		a.right = append(a.right, sp)
		a.left = append(a.left, sp.mirror())
	}
	if len(a.right) == 0 {
		return nil
	}
	return a
}

// name_*.png 파일을 프레임 번호 순으로 정렬해 반환
func frameFiles(fsys fs.FS, name string) []string {
	files, _ := fs.Glob(fsys, name+"_*.png")
	sort.Slice(files, func(i, j int) bool { return frameNumber(files[i]) < frameNumber(files[j]) })
	return files
}

func frameNumber(f string) int {
	m := frameNo.FindStringSubmatch(f)
	if m == nil {
		return 0
	}
	n, _ := strconv.Atoi(m[1])
	return n
}

func decodePNG(fsys fs.FS, name string) (image.Image, error) {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, err
	}
	return png.Decode(bytes.NewReader(data))
}

func applyAnimFallbacks(res map[string]*anim) {
	for _, name := range animNames {
		if res[name] != nil {
			continue
		}
		for _, alt := range animFallbacks[name] {
			if res[alt] != nil {
				res[name] = res[alt]
				break
			}
		}
	}
}
