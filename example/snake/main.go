package main

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/kongbong/ecsgo"
)

const (
	screenWidth  = 640
	screenHeight = 480
	gridSize     = 10
	xNumInScreen = screenWidth / gridSize
	yNumInScreen = screenHeight / gridSize
)

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Snake (Ebiten Demo)")
	g := newGame()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

type EbitenGame struct {
	registry   *ecsgo.Registry
	renderInfo RenderInfo
	ctx        context.Context
}

func newGame() *EbitenGame {
	g := &EbitenGame{}
	g.ctx = context.Background()
	g.Reset()
	return g
}

var GlobalGameState ecsgo.EntityId

func (g *EbitenGame) Reset() {
	g.registry = ecsgo.NewRegistry()

	sys1 := g.registry.AddSystem("directionSystem", 5, inputProcess)
	q := sys1.NewQuery()
	ecsgo.AddReadWriteComponent[Direction](q)

	sys2 := g.registry.AddSystem("MoveSystem", 4, move)
	q = sys2.NewQuery()
	ecsgo.AddReadWriteComponent[Direction](q)
	ecsgo.AddReadWriteComponent[Position](q)
	ecsgo.AddReadonlyComponent[Next](q)
	ecsgo.AddReadonlyComponent[Head](q)
	q2 := sys2.NewQuery()
	ecsgo.AddReadWriteComponent[GameState](q2)
	q3 := sys2.NewQuery()
	ecsgo.AddReadWriteComponent[Position](q3)
	ecsgo.AddReadonlyComponent[Next](q3)

	sys3 := g.registry.AddSystem("checkCollisionSystem", 3, checkCollision)
	q = sys3.NewQuery()
	ecsgo.AddReadonlyComponent[Position](q)
	ecsgo.AddReadWriteComponent[Collision](q)

	sys4 := g.registry.AddSystem("processCollision", 2, processCollsion)
	q = sys4.NewQuery()
	ecsgo.AddReadonlyComponent[Direction](q)
	ecsgo.AddReadWriteComponent[Collision](q)
	ecsgo.AddReadWriteComponent[Position](q)
	ecsgo.AddReadWriteComponent[Head](q)
	ecsgo.AddReadWriteComponent[Next](q)
	q2 = sys4.NewQuery()
	ecsgo.AddReadWriteComponent[GameState](q2)
	q3 = sys4.NewQuery()
	ecsgo.AddReadWriteComponent[Collision](q3)
	ecsgo.AddReadWriteComponent[Position](q3)
	ecsgo.AddOptionalReadWriteComponent[Next](q)
	ecsgo.AddOptionalReadonlyComponent[Apple](q3)

	sys5 := g.registry.AddSystem("checkGameOverSystem", 1, g.checkGameOver)
	q = sys5.NewQuery()
	ecsgo.AddReadonlyComponent[GameState](q)

	sys6 := g.registry.AddSystem("setRenders", 0, g.SetRenders)
	q = sys6.NewQuery()
	ecsgo.AddReadWriteComponent[Position](q)
	ecsgo.AddReadWriteComponent[Color](q)

	GlobalGameState = g.registry.CreateEntity()
	ecsgo.AddComponent[GameState](g.registry, GlobalGameState, GameState{
		Speed: 4,
		Level: 1,
		Score: 0,
	})

	// Add Snake Head
	snake := g.registry.CreateEntity()
	ecsgo.AddComponent[Position](g.registry, snake, Position{
		X: xNumInScreen / 2,
		Y: yNumInScreen / 2,
	})
	ecsgo.AddComponent[Direction](g.registry, snake, Direction{
		Dir: None,
	})
	ecsgo.AddComponent[Color](g.registry, snake, Color{
		Color: color.RGBA{0x80, 0xa0, 0xc0, 0xff},
	})
	ecsgo.AddComponent[Collision](g.registry, snake, Collision{}) // Placeholder
	ecsgo.AddComponent[Next](g.registry, snake, Next{})           // Placehodler
	ecsgo.AddComponent[Head](g.registry, snake, Head{})

	// Add Apple
	apple := g.registry.CreateEntity()
	ecsgo.AddComponent[Position](g.registry, apple, Position{
		X: rand.Intn(xNumInScreen - 1),
		Y: rand.Intn(yNumInScreen - 1),
	})
	ecsgo.AddComponent[Color](g.registry, apple, Color{
		Color: color.RGBA{0xFF, 0x00, 0x00, 0xff},
	})
	ecsgo.AddComponent[Collision](g.registry, apple, Collision{}) // Placeholder
	ecsgo.AddComponent[Apple](g.registry, apple, Apple{})
}

func (g *EbitenGame) Update() error {
	g.registry.Tick(100*time.Millisecond, g.ctx)
	return nil
}

func (g *EbitenGame) Draw(screen *ebiten.Image) {
	for _, v := range g.renderInfo.objs {
		vector.DrawFilledRect(screen, float32(v.X*gridSize), float32(v.Y*gridSize), gridSize, gridSize, v.Color, false)
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %0.2f Level: %d Score: %d Best Score: %d",
		ebiten.ActualFPS(), g.renderInfo.level, g.renderInfo.score, g.renderInfo.bestScore))
}

func (g *EbitenGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

type Position struct {
	X int
	Y int
}

const (
	None = iota
	Up
	Down
	Left
	Right
)

type Direction struct {
	Dir         int
	ElapsedTime int
}

type Collision struct {
	Other ecsgo.EntityId
	X     int
	Y     int
}

type Next struct {
	Next ecsgo.EntityId
}

type Color struct {
	Color color.RGBA
}

type GameState struct {
	Speed    int
	Level    int
	Score    int
	GameOver bool
}

type Head struct {
	Last ecsgo.EntityId
}

type Apple struct{}
type Body struct{}

type RenderInfo struct {
	objs      []RenderObj
	score     int
	level     int
	bestScore int
}

type RenderObj struct {
	X     int
	Y     int
	Color color.Color
}

// Process Input to set Dir
func inputProcess(ctx *ecsgo.ExecutionContext) error {
	qr := ctx.GetQueryResult(0)
	qr.ForeachEntities(func(accessor *ecsgo.ArcheTypeAccessor) error {
		dir := ecsgo.GetComponentByAccessor[Direction](accessor)
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
			if dir.Dir != Right {
				dir.Dir = Left
			}
		} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
			if dir.Dir != Left {
				dir.Dir = Right
			}
		} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
			if dir.Dir != Up {
				dir.Dir = Down
			}
		} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
			if dir.Dir != Down {
				dir.Dir = Up
			}
		}
		return nil
	})
	return nil
}

// check position if there are collision then set collision values
func checkCollision(ctx *ecsgo.ExecutionContext) error {
	colliders := make(map[int]map[int]ecsgo.EntityId)
	qr := ctx.GetQueryResult(0)
	qr.ForeachEntities(func(accessor *ecsgo.ArcheTypeAccessor) error {
		self := accessor.GetEntityId()
		pos := ecsgo.GetComponentByAccessor[Position](accessor)
		if colliders[pos.X] == nil {
			colliders[pos.X] = make(map[int]ecsgo.EntityId)
		}

		selfCollision := ecsgo.GetComponentByAccessor[Collision](accessor)
		if other, ok := colliders[pos.X][pos.Y]; ok {
			otherCollision := ecsgo.GetComponent[Collision](ctx, other)
			selfCollision.Other = other
			selfCollision.X = pos.X
			selfCollision.Y = pos.Y

			otherCollision.Other = self
			otherCollision.X = pos.X
			otherCollision.Y = pos.Y
		} else {
			colliders[pos.X][pos.Y] = self
			*selfCollision = Collision{}
		}
		return nil
	})
	return nil
}

func move(ctx *ecsgo.ExecutionContext) error {
	var gameState *GameState
	qr := ctx.GetQueryResult(1)
	qr.ForeachEntities(func(accessor *ecsgo.ArcheTypeAccessor) error {
		gameState = ecsgo.GetComponentByAccessor[GameState](accessor)
		return nil
	})

	qr = ctx.GetQueryResult(0)
	qr.ForeachEntities(func(accessor *ecsgo.ArcheTypeAccessor) error {
		dir := ecsgo.GetComponentByAccessor[Direction](accessor)
		pos := ecsgo.GetComponentByAccessor[Position](accessor)
		next := ecsgo.GetComponentByAccessor[Next](accessor)

		if dir.ElapsedTime >= gameState.Speed {
			lastX := pos.X
			lastY := pos.Y

			// Move
			if dir.Dir != None {
				switch dir.Dir {
				case Up:
					pos.Y--
				case Down:
					pos.Y++
				case Left:
					pos.X--
				case Right:
					pos.X++
				}
				if pos.X < 0 || pos.Y < 0 || pos.X >= xNumInScreen || pos.Y >= yNumInScreen {
					// GameOver
					gameState.GameOver = true
				} else {

					for next.Next.NotNil() {
						nextPos := ecsgo.GetComponent[Position](ctx, next.Next)
						nextPos.X, lastX = lastX, nextPos.X
						nextPos.Y, lastY = lastY, nextPos.Y
						next = ecsgo.GetComponent[Next](ctx, next.Next)
					}
				}
			}
			dir.ElapsedTime = 0
		} else {
			dir.ElapsedTime++
		}
		return nil
	})
	return nil
}

func processCollsion(ctx *ecsgo.ExecutionContext) error {
	var gameState *GameState
	qr := ctx.GetQueryResult(1)
	qr.ForeachEntities(func(accessor *ecsgo.ArcheTypeAccessor) error {
		gameState = ecsgo.GetComponentByAccessor[GameState](accessor)
		return nil
	})

	qr = ctx.GetQueryResult(0)
	qr.ForeachEntities(func(accessor *ecsgo.ArcheTypeAccessor) error {
		collision := ecsgo.GetComponentByAccessor[Collision](accessor)
		dir := ecsgo.GetComponentByAccessor[Direction](accessor)
		if dir.ElapsedTime != 0 {
			// only check after move
			return nil
		}
		if collision != nil && collision.Other.NotNil() {
			if ecsgo.HasComponent[Apple](ctx, collision.Other) {
				// it is apple, eat
				gameState.Score++

				if gameState.Score > 20 {
					gameState.Level = 3
					gameState.Speed = 2
				} else if gameState.Score > 10 {
					gameState.Level = 2
					gameState.Speed = 3
				}

				applePos := ecsgo.GetComponent[Position](ctx, collision.Other)
				applePos.X = rand.Intn(xNumInScreen - 1)
				applePos.Y = rand.Intn(yNumInScreen - 1)

				head := ecsgo.GetComponentByAccessor[Head](accessor)
				tail := head.Last

				// add snake body
				snakeBody := ctx.CreateEntity()
				head.Last = snakeBody
				if tail.NotNil() {
					tailPos := ecsgo.GetComponent[Position](ctx, tail)
					next := ecsgo.GetComponent[Next](ctx, tail)
					ecsgo.AddComponent[Position](ctx.GetResgiry(), snakeBody, Position{
						X: tailPos.X,
						Y: tailPos.Y,
					})
					next.Next = snakeBody
				} else {
					pos := ecsgo.GetComponentByAccessor[Position](accessor)
					next := ecsgo.GetComponentByAccessor[Next](accessor)
					ecsgo.AddComponent[Position](ctx.GetResgiry(), snakeBody, Position{
						X: pos.X,
						Y: pos.Y,
					})
					next.Next = snakeBody
				}

				ecsgo.AddComponent[Color](ctx.GetResgiry(), snakeBody, Color{
					Color: color.RGBA{0x90, 0xb0, 0xd0, 0xff},
				})
				ecsgo.AddComponent[Collision](ctx.GetResgiry(), snakeBody, Collision{}) // Placeholder
				ecsgo.AddComponent[Next](ctx.GetResgiry(), snakeBody, Next{})           // Placehodler
				ecsgo.AddComponent[Body](ctx.GetResgiry(), snakeBody, Body{})

			} else if ecsgo.HasComponent[Body](ctx, collision.Other) {
				// collision with Body
				// Gameover
				gameState.GameOver = true
			}
		}
		return nil
	})
	return nil
}

func (g *EbitenGame) SetRenders(ctx *ecsgo.ExecutionContext) error {

	g.renderInfo.objs = g.renderInfo.objs[:0]
	qr := ctx.GetQueryResult(0)
	qr.ForeachEntities(func(accessor *ecsgo.ArcheTypeAccessor) error {
		pos := ecsgo.GetComponentByAccessor[Position](accessor)
		col := ecsgo.GetComponentByAccessor[Color](accessor)

		g.renderInfo.objs = append(g.renderInfo.objs, RenderObj{
			X:     pos.X,
			Y:     pos.Y,
			Color: col.Color,
		})
		return nil
	})
	return nil
}

func (g *EbitenGame) checkGameOver(ctx *ecsgo.ExecutionContext) error {
	qr := ctx.GetQueryResult(0)
	qr.ForeachEntities(func(accessor *ecsgo.ArcheTypeAccessor) error {
		gameState := ecsgo.GetComponentByAccessor[GameState](accessor)
		if gameState.GameOver {
			g.Reset()
		} else {
			g.renderInfo.score = gameState.Score
			g.renderInfo.level = gameState.Level
			if gameState.Score > g.renderInfo.bestScore {
				g.renderInfo.bestScore = gameState.Score
			}
		}
		return nil
	})
	return nil
}
