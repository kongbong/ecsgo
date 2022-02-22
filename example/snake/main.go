package main

import (
	"fmt"
	"math/rand"
	"time"
	"log"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/kongbong/ecsgo"
)

const (
	screenWidth  = 640
	screenHeight = 480
	gridSize     = 10
	xNumInScreen = screenWidth / gridSize
	yNumInScreen = screenHeight / gridSize
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Snake (Ebiten Demo)")
	g := newGame()
	defer g.Free()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}


type EbitenGame struct {
	registry *ecsgo.Registry
	renderInfo RenderInfo
}

func newGame() *EbitenGame {
	registry := ecsgo.New()
	g := &EbitenGame{
		registry: registry,
	}

	ecsgo.AddSystem1[Direction](registry, ecsgo.PreTick, inputProcess)
	
	sys := ecsgo.AddSystem4[Direction, Position, Next, Head](registry, ecsgo.OnTick, move)
	ecsgo.AddDependency[GameState](sys)
	
	sys = ecsgo.AddSystem1[Position](registry, ecsgo.OnTick, checkCollision)
	ecsgo.AddDependency[Collision](sys)

	sys = ecsgo.AddSystem5[Direction, Collision, Position, Head, Next](registry, ecsgo.OnTick, processCollsion)
	ecsgo.AddDependency[GameState](sys)

	ecsgo.AddSystem2[Position, Color](registry, ecsgo.OnTick, g.SetRenders)

	ecsgo.AddSystem1[GameState](registry, ecsgo.PostTick, g.checkGameOver)

	g.Reset()
	return g
}

var GlobalGameState ecsgo.Entity

func (g *EbitenGame) Reset() {
	g.Free()
	GlobalGameState = g.registry.Create()
	ecsgo.AddComponent[GameState](g.registry, GlobalGameState, &GameState{
		Speed: 4,
		Level: 1,
		Score: 0,
	})

	// Add Snake Head
	entity := g.registry.Create()
	ecsgo.AddComponent[Position](g.registry, entity, &Position{
		X: xNumInScreen / 2,
		Y: yNumInScreen / 2,
	})
	ecsgo.AddComponent[Direction](g.registry, entity, &Direction{
		Dir: None,
	})
	ecsgo.AddComponent[Color](g.registry, entity, &Color{
		Color: color.RGBA{0x80, 0xa0, 0xc0, 0xff},
	})
	ecsgo.AddComponent[Collision](g.registry, entity, &Collision{})	// Placeholder
	ecsgo.AddComponent[Next](g.registry, entity, &Next{})	// Placehodler
	ecsgo.AddComponent[Head](g.registry, entity, &Head{
		Last: ecsgo.EntityNil,
	})

	// Add Apple
	entity = g.registry.Create()
	ecsgo.AddComponent[Position](g.registry, entity, &Position{
		X: rand.Intn(xNumInScreen-1),
		Y: rand.Intn(yNumInScreen-1),
	})
	ecsgo.AddComponent[Color](g.registry, entity, &Color{
		Color: color.RGBA{0xFF, 0x00, 0x00, 0xff},
	})
	ecsgo.AddComponent[Collision](g.registry, entity, &Collision{})	// Placeholder
	ecsgo.AddTag[Apple](g.registry, entity)
}

func (g *EbitenGame) Free() {
	g.registry.Free()
}

func (g *EbitenGame) Update() error {
	g.registry.Tick(0.01)
	return nil
}

func (g *EbitenGame) Draw(screen *ebiten.Image) {
	for _, v := range g.renderInfo.objs {
		ebitenutil.DrawRect(screen, float64(v.X*gridSize), float64(v.Y*gridSize), gridSize, gridSize, v.Color)
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %0.2f Level: %d Score: %d Best Score: %d", 
		ebiten.CurrentFPS(), g.renderInfo.level, g.renderInfo.score, g.renderInfo.bestScore))
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
	Dir          int
	ElapsedTime int
}

type Collision struct {
	Other ecsgo.Entity
	X     int
	Y     int
}

type Next struct {
	Next ecsgo.Entity
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

type Head struct{
	Last ecsgo.Entity
}

type Apple struct{}
type Body struct{}

type RenderInfo struct {
	objs  []RenderObj
	score int
	level int
	bestScore int
}

type RenderObj struct {
	X int
	Y int
	Color color.Color
}

// Process Input to set Dir
func inputProcess(r *ecsgo.Registry, iter *ecsgo.Iterator) {
	for ; !iter.IsNil(); iter.Next() {
		dir := ecsgo.Get[Direction](iter)
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
	}
}

// check position if there are collision then set collision values
func checkCollision(r *ecsgo.Registry, iter *ecsgo.Iterator) {
	colliders := make(map[int]map[int]ecsgo.Entity)

	for ; !iter.IsNil(); iter.Next() {
		self := iter.Entity()
		pos := ecsgo.Get[Position](iter)
		if colliders[pos.X] == nil {
			colliders[pos.X] = make(map[int]ecsgo.Entity)
		}
		if other, ok := colliders[pos.X][pos.Y]; ok {
			// collision
			ecsgo.Set[Collision](r, self, &Collision{
				Other: other,
				X:     pos.X,
				Y:     pos.Y,
			})
			ecsgo.Set[Collision](r, other, &Collision{
				Other: self,
				X:     pos.X,
				Y:     pos.Y,
			})
		} else {
			colliders[pos.X][pos.Y] = self
			ecsgo.Set[Collision](r, self, nil)
		}
	}
}

func move(r *ecsgo.Registry, iter *ecsgo.Iterator) {
	dir := ecsgo.Get[Direction](iter)
	pos := ecsgo.Get[Position](iter)
	next := ecsgo.Get[Next](iter)
	gameState := ecsgo.Get[GameState](r, GlobalGameState)

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
				
				for next.Next != ecsgo.EntityNil {
					nextPos := ecsgo.Get[Position](r, next.Next)
					nextPos.X, lastX = lastX, nextPos.X
					nextPos.Y, lastY = lastY, nextPos.Y
					next = ecsgo.Get[Next](r, next.Next)
				}
			}
		}
		dir.ElapsedTime = 0
	} else {
		dir.ElapsedTime++
	}
}

func processCollsion(r *ecsgo.Registry, iter *ecsgo.Iterator) {
	for ; !iter.IsNil(); iter.Next() {
		collision := ecsgo.Get[Collision](iter)
		dir := ecsgo.Get[Direction](iter)
		if dir.ElapsedTime != 0 {
			// only check after move
			continue
		}
		if collision != nil && collision.Other != ecsgo.EntityNil {
			if ecsgo.HasTag[Apple](r, collision.Other) {
				// it is apple, eat
				gameState := ecsgo.Get[GameState](r, GlobalGameState)
				gameState.Score++

				if gameState.Score > 20 {
					gameState.Level = 3
					gameState.Speed = 2
				} else if gameState.Score > 10 {
					gameState.Level = 2
					gameState.Speed = 3
				}

				applePos := ecsgo.Get[Position](r, collision.Other)
				applePos.X = rand.Intn(xNumInScreen-1)
				applePos.Y = rand.Intn(yNumInScreen-1)

				head := ecsgo.Get[Head](iter)
				tail := head.Last
				
				// add snake body
				entity := r.Create()
				head.Last = entity
				if tail != ecsgo.EntityNil {
					tailPos, err := ecsgo.GetReadonly[Position](r, tail)
					if err != nil {
						panic(err)
					}
					next := ecsgo.Get[Next](r, tail)
					ecsgo.AddComponent[Position](r, entity, &Position{
						X: tailPos.X,
						Y: tailPos.Y,
					})
					next.Next = entity
				} else {
					pos := ecsgo.Get[Position](iter)
					next := ecsgo.Get[Next](iter)
					ecsgo.AddComponent[Position](r, entity, &Position{
						X: pos.X,
						Y: pos.Y,
					})
					next.Next = entity
				}
				
				ecsgo.AddComponent[Color](r, entity, &Color{
					Color: color.RGBA{0x90, 0xb0, 0xd0, 0xff},
				})
				ecsgo.AddComponent[Collision](r, entity, &Collision{})	// Placeholder
				ecsgo.AddComponent[Next](r, entity, &Next{})	// Placehodler
				ecsgo.AddTag[Body](r, entity)

			} else if ecsgo.HasTag[Body](r, collision.Other) {
				// collision with Body
				// Gameover
				gameState := ecsgo.Get[GameState](r, GlobalGameState)
				gameState.GameOver = true
			}
		}
	}
}

func (g *EbitenGame) SetRenders(r *ecsgo.Registry, iter *ecsgo.Iterator) {
	
	g.renderInfo.objs = g.renderInfo.objs[:0]
	for ; !iter.IsNil(); iter.Next() {
		pos := ecsgo.Get[Position](iter)
		col := ecsgo.Get[Color](iter)

		g.renderInfo.objs = append(g.renderInfo.objs, RenderObj{
			X: pos.X,
			Y: pos.Y,
			Color: col.Color,
		})
	}
}

func (g *EbitenGame) checkGameOver(r *ecsgo.Registry, iter *ecsgo.Iterator) {
	for ; !iter.IsNil(); iter.Next() {
		gameState := ecsgo.Get[GameState](iter)
		if gameState.GameOver {
			g.Reset()
			break
		} else {
			g.renderInfo.score = gameState.Score
			g.renderInfo.level = gameState.Level
			if gameState.Score > g.renderInfo.bestScore {
				g.renderInfo.bestScore = gameState.Score
			}
		}
	}
}