package main

// ───────────── 스프라이트 (프리멀티플라이드 BGRA 이미지) ─────────────

import (
	"image"
	"math"
)

type sprite struct {
	w, h int
	pix  []byte
}

func newSprite(w, h int) *sprite {
	return &sprite{w: w, h: h, pix: make([]byte, w*h*4)}
}

// 채널별 실수 색상 (0~255, 프리멀티플라이드)
type rgba struct{ r, g, b, a float64 }

func (c rgba) add(o rgba) rgba    { return rgba{c.r + o.r, c.g + o.g, c.b + o.b, c.a + o.a} }
func (c rgba) div(n float64) rgba { return rgba{c.r / n, c.g / n, c.b / n, c.a / n} }
func lerp(c, o rgba, t float64) rgba {
	return rgba{c.r*(1-t) + o.r*t, c.g*(1-t) + o.g*t, c.b*(1-t) + o.b*t, c.a*(1-t) + o.a*t}
}

func (s *sprite) setPixel(x, y int, c rgba) {
	i := (y*s.w + x) * 4
	s.pix[i], s.pix[i+1], s.pix[i+2], s.pix[i+3] = byte(c.b+.5), byte(c.g+.5), byte(c.r+.5), byte(c.a+.5)
}

// image.Image → scale배 크기의 스프라이트 (축소는 영역 평균, 확대는 쌍선형)
func fromImage(src image.Image, scale float64) *sprite {
	b := src.Bounds()
	s := newSprite(scaledLength(b.Dx(), scale), scaledLength(b.Dy(), scale))
	sample := sampleBilinear
	if scale < 1 {
		sample = sampleArea
	}
	for y := 0; y < s.h; y++ {
		for x := 0; x < s.w; x++ {
			s.setPixel(x, y, sample(src, scale, x, y))
		}
	}
	return s
}

func scaledLength(n int, scale float64) int {
	return int(math.Max(1, math.Round(float64(n)*scale)))
}

// 원본 픽셀 (범위 밖 좌표는 가장자리로 고정). image.Color.RGBA()는 이미 프리멀티플라이드(16bit)
func pixelAt(src image.Image, x, y int) rgba {
	b := src.Bounds()
	x = clamp(x, 0, b.Dx()-1)
	y = clamp(y, 0, b.Dy()-1)
	R, G, B, A := src.At(b.Min.X+x, b.Min.Y+y).RGBA()
	return rgba{float64(R) / 257, float64(G) / 257, float64(B) / 257, float64(A) / 257}
}

// 축소: 대상 픽셀 (x,y)가 덮는 원본 영역의 평균
func sampleArea(src image.Image, scale float64, x, y int) rgba {
	x0, x1 := int(float64(x)/scale), int(float64(x+1)/scale)
	y0, y1 := int(float64(y)/scale), int(float64(y+1)/scale)
	var sum rgba
	n := 0.0
	for yy := y0; yy < max(y1, y0+1); yy++ {
		for xx := x0; xx < max(x1, x0+1); xx++ {
			sum = sum.add(pixelAt(src, xx, yy))
			n++
		}
	}
	return sum.div(n)
}

// 확대: 주변 4픽셀 쌍선형 보간
func sampleBilinear(src image.Image, scale float64, x, y int) rgba {
	fx, fy := (float64(x)+0.5)/scale-0.5, (float64(y)+0.5)/scale-0.5
	ix, iy := int(math.Floor(fx)), int(math.Floor(fy))
	tx, ty := fx-float64(ix), fy-float64(iy)
	top := lerp(pixelAt(src, ix, iy), pixelAt(src, ix+1, iy), tx)
	bottom := lerp(pixelAt(src, ix, iy+1), pixelAt(src, ix+1, iy+1), tx)
	return lerp(top, bottom, ty)
}

// 좌우 반전 사본 (왼쪽을 볼 때 사용)
func (s *sprite) mirror() *sprite {
	m := newSprite(s.w, s.h)
	for y := 0; y < s.h; y++ {
		for x := 0; x < s.w; x++ {
			copy(m.pix[(y*s.w+x)*4:][:4], s.pix[(y*s.w+s.w-1-x)*4:][:4])
		}
	}
	return m
}

// dst(dw×dh) 위 (ox,oy)에 s를 알파 합성. 범위 밖은 잘라낸다
func blit(dst []byte, dw, dh int, s *sprite, ox, oy int) {
	for y := 0; y < s.h; y++ {
		dy := oy + y
		if dy < 0 || dy >= dh {
			continue
		}
		for x := 0; x < s.w; x++ {
			dx := ox + x
			if dx < 0 || dx >= dw {
				continue
			}
			blendPixel(dst[(dy*dw+dx)*4:], s.pix[(y*s.w+x)*4:])
		}
	}
}

// 프리멀티플라이드 over 합성: dst = src + dst × (1 - srcAlpha)
func blendPixel(dst, src []byte) {
	a := int(src[3])
	if a == 0 {
		return
	}
	inv := 255 - a
	for k := 0; k < 4; k++ {
		dst[k] = byte(int(src[k]) + int(dst[k])*inv/255)
	}
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
