package main

// ───────────── 상태 머신과 물리 (매 프레임 update) ─────────────
// idle ↔ walk 는 타이머로 전환, jump/fall 은 중력·튕김을 받는 공중 상태, drag 는 마우스가 잡고 있는 상태.

import (
	"math"
	"math/rand"
)

func (p *pet) update() {
	p.tick++
	switch p.state {
	case "idle":
		p.updateIdle()
	case "walk":
		p.updateWalk()
	case "fall", "jump":
		p.updateAirborne()
	}
}

func (p *pet) walkSpeed() float64    { return cfgWalkSpeed * p.dpi }
func (p *pet) gravityForce() float64 { return gravity * math.Sqrt(p.dpi) }

// ── idle ──

func (p *pet) updateIdle() {
	if !p.grounded() {
		p.setState("fall", -1)
		return
	}
	p.timer--
	if p.timer <= 0 {
		p.startWalking()
	}
}

func (p *pet) startWalking() {
	p.facing = randomFacing()
	p.setState("walk", rnd(fps*2, fps*6))
}

// ── walk ──

func (p *pet) updateWalk() {
	p.walkStep()
	if p.state != "walk" { // tryClimb으로 점프를 시작함
		return
	}
	if !p.grounded() { // 바닥이 더 낮은 옆 모니터로 넘어감 → 떨어짐
		p.vx, p.vy = float64(p.facing)*p.walkSpeed(), 0
		p.setState("fall", -1)
		return
	}
	p.timer--
	if p.timer <= 0 {
		p.stopWalking()
	}
}

// 한 걸음 이동. 막히면 올라갈 곳을 찾고, 없으면 뒤돈다(화면 끝)
func (p *pet) walkStep() {
	var blocked bool
	p.x, blocked = p.moveX(p.x, p.y, float64(p.facing)*p.walkSpeed())
	if blocked && !p.tryClimb(p.facing) {
		p.facing = -p.facing
	}
}

// 멈추고 가끔 혼잣말
func (p *pet) stopWalking() {
	p.setState("idle", rnd(fps, fps*4))
	if rand.Float64() < cfgTalkChance {
		p.sayRandom()
	}
}

// ── jump ──

// 클릭/메뉴로 제자리 점프 (걷거나 서 있을 때만)
func (p *pet) jump() {
	if p.state == "idle" || p.state == "walk" {
		p.vx = float64(p.facing) * cfgWalkSpeed * 1.5 * p.dpi
		p.vy = -14 * math.Sqrt(p.dpi)
		p.setState("jump", -1)
	}
}

// 앞이 막혔을 때 올라설 수 있는 더 높은 바닥(예: 작업표시줄이 있는 옆 모니터)이 있으면 점프해서 넘어감
func (p *pet) tryClimb(dir int) bool {
	up, ok := p.findLedgeHeight(dir)
	if ok {
		p.jumpOntoLedge(up, dir)
	}
	return ok
}

// 진행 방향 앞쪽에서 올라설 수 있는 높이(px)를 아래부터 찾는다 (최대 200·dpi)
func (p *pet) findLedgeHeight(dir int) (float64, bool) {
	ahead := p.x + float64(dir)*p.walkSpeed()*3
	for up := 1.0; up <= 200*p.dpi; up++ {
		if p.valid(ahead, p.y-up) && p.valid(p.x, p.y-up) {
			return up, true
		}
	}
	return 0, false
}

// 높이 up만큼(+여유 12·dpi) 뛰어오를 초속도로 점프
func (p *pet) jumpOntoLedge(up float64, dir int) {
	g := p.gravityForce()
	p.vy = -math.Sqrt(2*g*(up+12*p.dpi)) - g
	p.vx = float64(dir) * p.walkSpeed() * 1.6
	p.setState("jump", -1)
}

// ── fall / jump (공중) ──

func (p *pet) updateAirborne() {
	g := p.gravityForce()
	p.vy += g
	// 세로 먼저 이동 (그래야 점프로 옆 모니터의 더 높은 바닥에 올라설 수 있음)
	var hitVertical, hitWall bool
	p.y, hitVertical = p.moveY(p.x, p.y, p.vy)
	p.x, hitWall = p.moveX(p.x, p.y, p.vx)
	if hitWall && p.state == "fall" { // 던져져서 벽(화면 끝, 더 높은 옆 모니터 바닥의 옆면)에 부딪히면 튕김
		p.vx = -p.vx * bounce
	} // 점프 중에는 튕기지 않고 계속 앞으로 밀어서 턱을 넘어감
	if hitVertical {
		if p.vy > 0 {
			p.hitFloor(g)
		} else {
			p.hitCeiling()
		}
	}
	p.faceMovingDirection()
}

// 빠르게 떨어졌으면 튕기고, 느리면 착지해서 idle
func (p *pet) hitFloor(g float64) {
	if p.vy > 6*g {
		p.vy = -p.vy * bounce
		p.vx *= 0.7
		return
	}
	p.vx, p.vy = 0, 0
	p.setState("idle", rnd(fps, fps*3))
}

func (p *pet) hitCeiling() {
	p.vy = math.Abs(p.vy) * bounce
}

// 가로 속도가 충분하면 그 방향을 바라봄
func (p *pet) faceMovingDirection() {
	if p.vx > 0.5 {
		p.facing = 1
	} else if p.vx < -0.5 {
		p.facing = -1
	}
}
