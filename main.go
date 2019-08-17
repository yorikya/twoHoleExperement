package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

const (
	xySize      = 40
	xSize       = 173
	ySize       = xySize
	redColor    = "\033[1;31m%s\033[0m"
	blueColor   = "\033[1;36m%s\033[0m"
	whiteColor  = "\033[1;39m%s\033[0m"
	yellowColor = "\033[1;33m%s\033[0m"
)

var (
	incFuncMaker = func(i int) func(int) int { return func(d int) int { return d + i } }
	incFunc      = incFuncMaker(1)
	debug        = flag.Bool("debug", false, "desable regular work for debug prints")
)

type Point struct {
	x,
	y int
	cellFill string
}

type Board struct {
	maxX,
	maxY int
	cells     [][]*Point
	bariers   []Barier
	filedCell string
	blankCell string
	expPoints []*ExpolisionPoint
}

func (b *Board) init() {
	b.cells = [][]*Point{}
	for y := 0; y < b.maxY; y++ {
		line := []*Point{}
		for x := 0; x < b.maxX; x++ {
			line = append(line, &Point{x: x, y: y, cellFill: b.blankCell})
		}
		b.cells = append(b.cells, line)
	}
}

func (b *Board) addRedPoint(x, y int) {
	b.addFmtPoint(x, y, redColor)
}

func (b *Board) addBluePoint(x, y int) {
	b.addFmtPoint(x, y, blueColor)
}

func (b *Board) addYellowPoint(x, y int) {
	b.addFmtPoint(x, y, yellowColor)
}

func (b *Board) addPoint(x, y int) {
	b.addFmtPoint(x, y, whiteColor)
}

func (b *Board) addFmtPoint(x, y int, fmtstr string) {
	if x < 0 || y < 0 || x >= xSize || y >= ySize {
		return
	}
	b.cells[y][x].cellFill = fmt.Sprintf(fmtstr, b.filedCell)
}

func (b *Board) draw() {
	b.flush()
	for i, row := range b.cells {
		for j := range row {
			fmt.Printf("%s", b.cells[i][j].cellFill)
		}
	}
}

func (b Board) flush() {
	cmd := exec.Command("clear") //Linux example, its tested
	cmd.Stdout = os.Stdout
	cmd.Run()
}

type ExpolisionPoint struct {
	xCurrent,
	yCurrent,
	xNext,
	yNext,
	numMoves,
	maxMoves int
	xFunc,
	yFunc func(int) int
	stopped bool
}

func (e *ExpolisionPoint) move(b *Board) {
	if b.inBorderRange(e.xNext, e.yNext) || b.outOfFiled(e.xNext, e.yNext) {
		e.stopped = true
		return
	}

	cX, cY := e.xCurrent, e.yCurrent
	b.addPoint(e.xNext, e.yNext)
	e.xCurrent = e.xNext
	e.yCurrent = e.yNext
	e.xNext = e.xFunc(cX)
	e.yNext = e.yFunc(cY)
	e.numMoves++
}

func NewExplosionPoint(x, y int, deltaXFunc, deltaYFunc func(d int) int) *ExpolisionPoint {
	e := &ExpolisionPoint{
		xCurrent: x,
		yCurrent: y,
		xNext:    deltaXFunc(x),
		yNext:    deltaYFunc(y),
		maxMoves: 100,
		xFunc:    deltaXFunc,
		yFunc:    deltaYFunc,
	}

	return e
}

func (b *Board) getExposionPoint(x, y int) *ExpolisionPoint {
	for _, p := range b.expPoints {
		if x == p.xCurrent && y == p.yCurrent {
			return p
		}
	}
	return nil
}

func newBoard() *Board {
	b := &Board{
		maxX:      xSize,
		maxY:      ySize,
		blankCell: " ",
		filedCell: "*",
	}
	b.init()
	return b
}

func (b *Board) newExplosionAtom(x, y int) {
	midleX := x / 2

	e := []*ExpolisionPoint{}
	for i := 0; i < y+1; i++ {
		for j := 0; j < x+1; j++ {
			pX := midleX - x
			p := NewExplosionPoint(xSize/2+pX+i, j, incFuncMaker(pX+i), incFunc)
			e = append(e, p)
			b.addPoint(p.xCurrent, p.yCurrent)
		}
	}
	b.expPoints = e
}

type Barier struct {
	startX, endX, startY, endY int
}

func (b Barier) inRange(x, y int) bool {
	return x >= b.startX && x <= b.endX && y >= b.startY && y <= b.endY
}

func (b *Board) inBorderRange(x, y int) bool {
	for _, br := range b.bariers {
		if br.inRange(x, y) {
			return true
		}
	}
	return false
}

func (b *Board) outOfFiled(x, y int) bool {
	return x < 0 || y < 0 || x >= xSize || y >= ySize
}

func (b *Board) addBarier(startX, endX, startY, endY int) {
	deltaX := endX - startX
	deltaY := endY - startY
	for i := 0; i < deltaY; i++ {
		for j := 0; j < deltaX; j++ {
			b.bariers = append(b.bariers, Barier{startX, endX, startY, endY})
			b.addPoint(startX+j, startY+i)
		}
	}
}

func (b *Board) addPunctBarier() {
	row := 16
	b.addBarier(1, 49, row, row+1)
	b.addBarier(61, 112, row, row+1)
	b.addBarier(124, 172, row, row+1)
}

func init() {
	f, err := os.OpenFile("log.out", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println("Starting run")
}

func main() {
	flag.Parse()
	b := newBoard()

	go func() {
		if !*debug {
			for {
				time.Sleep(400 * time.Millisecond)

				b.draw()
			}
		}

	}()

	b.addPunctBarier()
	b.newExplosionAtom(10, 10)
	sensor := 21
	b.addBarier(1, 172, sensor, sensor+1)

	for {
		for _, p := range b.expPoints {
			if p.stopped || p.numMoves == p.maxMoves {
				continue
			}
			p.move(b)
		}

		time.Sleep(300 * time.Millisecond)
		// buf := bufio.NewReader(os.Stdin)
		// _, err := buf.ReadBytes('\n')
		// if err != nil {
		// 	fmt.Println(err)

		// }
	}

}
