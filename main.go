package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

type cell struct {
	drawable uint32

	alive     bool
	aliveNext bool

	x int
	y int
}

func (c *cell) draw() {
	if !c.alive {
		return
	}

	gl.BindVertexArray(c.drawable)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
}

func (c *cell) checkState(cells [][]*cell) {
	c.alive = c.aliveNext
	c.aliveNext = c.alive

	liveCount := c.liveNeighbours(cells)
	if c.alive {
		// If less than 2 neibours, the cell dies from underpopulation.
		if liveCount < 2 {
			c.aliveNext = false
		}
		// If 2 or 3 neighbours, the cell lives on to the next generation.
		if liveCount == 2 || liveCount == 3 {
			c.aliveNext = true
		}
		// If more than 3 neighbours, the cell dies from overpopulation.
		if liveCount > 3 {
			c.aliveNext = false
		}
	} else {
		// If 3 neigbours, the cell lives on to the next generation from reproduction.
		if liveCount == 3 {
			c.aliveNext = true
		}
	}
}

func (c *cell) liveNeighbours(cells [][]*cell) int {
	var liveCount int
	add := func(x, y int) {
		if x == len(cells) {
			x = 0
		} else if x == -1 {
			x = len(cells) - 1
		}
		if y == len(cells[x]) {
			y = 1
		} else if y == -1 {
			y = len(cells[x]) - 1
		}

		if cells[x][y].alive {
			liveCount++
		}
	}

	add(c.x-1, c.y) // Left
	add(c.x+1, c.y) // Right
	add(c.x, c.y+1) // Top
	add(c.x, c.y-1) // Bottom

	add(c.x-1, c.y+1) // Top Left
	add(c.x+1, c.y+1) // Top Right
	add(c.x-1, c.y-1) // Bottom Left
	add(c.x+1, c.y-1) // Bottom Right

	return liveCount
}

const (
	width              = 500
	height             = 500
	vertexShaderSource = `
	    #version 410
	    in vec3 vp;
	    void main() {
		gl_Position = vec4(vp, 1.0);
	    }
	` + "\x00"

	fragmentShaderSource = `
	    #version 410
	    out vec4 frag_colour;
	    void main() {
		frag_colour = vec4(1, 1, 1, 1);
	    }
	` + "\x00"
	row     = 100
	columns = 100
)

var (
	fps       int     = 10
	threshold float32 = 0.15
	pattern   string  = "random"

	square = []float32{
		-0.5, 0.5, 0, // Top Left Vertex
		-0.5, -0.5, 0, // Bottom Left Vertex
		0.5, -0.5, 0, // Bottom Right Vertex

		-0.5, 0.5, 0, // Top Left Vertex
		0.5, 0.5, 0, // Top Right Vertex
		0.5, -0.5, 0, // Bottom Right Vertex
	}
	patterns = map[string][][]int{
		"random": nil, // Sentinel value to trigger random generation
		"blinker": {
			{0, 1},
			{1, 1},
			{2, 1},
		},
		"glider": {
			{1, 0},
			{2, 1},
			{0, 2},
			{1, 2},
			{2, 2},
		},
		"lightweightspaceship": {
			{0, 1},
			{0, 3},
			{1, 0},
			{2, 0},
			{3, 0},
			{3, 1},
			{3, 2},
			{2, 3},
		},
		"pulsar": {
			{2, 0}, {3, 0}, {4, 0}, {8, 0}, {9, 0}, {10, 0},
			{0, 2}, {5, 2}, {7, 2}, {12, 2},
			{0, 3}, {5, 3}, {7, 3}, {12, 3},
			{0, 4}, {5, 4}, {7, 4}, {12, 4},
			{2, 5}, {3, 5}, {4, 5}, {8, 5}, {9, 5}, {10, 5},
			{2, 7}, {3, 7}, {4, 7}, {8, 7}, {9, 7}, {10, 7},
			{0, 8}, {5, 8}, {7, 8}, {12, 8},
			{0, 9}, {5, 9}, {7, 9}, {12, 9},
			{0, 10}, {5, 10}, {7, 10}, {12, 10},
			{2, 12}, {3, 12}, {4, 12}, {8, 12}, {9, 12}, {10, 12},
		},
	}
)

func main() {
	runtime.LockOSThread()

	window := initGlfw()
	defer glfw.Terminate()

	program := initOpenGL()

	arguments := os.Args[1:]

	// Parse command-line flags for --fps and --threshold, allowing out-of-order specification
	for i := 0; i < len(arguments); i++ {
		switch arguments[i] {
		case "--fps", "-f":
			if i+1 < len(arguments) {
				if fpsArg, err := strconv.Atoi(arguments[i+1]); err == nil && fpsArg > 0 {
					fps = fpsArg
				} else {
					fmt.Fprintf(os.Stderr, "Invalid --fps value (using default %d): %v\n", fps, err)
				}
				i++ // skip value
			} else {
				fmt.Fprintf(os.Stderr, "Missing value for --fps (using default %d)\n", fps)
			}
		case "--threshold", "-t":
			if i+1 < len(arguments) {
				if thresholdArg, err := strconv.ParseFloat(arguments[i+1], 32); err == nil && thresholdArg >= 0 && thresholdArg <= 1 {
					threshold = float32(thresholdArg)
				} else {
					fmt.Fprintf(os.Stderr, "Invalid --threshold value (using default %.2f): %v\n", threshold, err)
				}
				i++ // skip value
			} else {
				fmt.Fprintf(os.Stderr, "Missing value for --threshold (using default %.2f)\n", threshold)
			}
		case "--pattern", "-p":
			if i+1 < len(arguments) {
				patternArg := arguments[i+1]
				if _, ok := patterns[patternArg]; ok {
					pattern = patternArg
				} else {
					fmt.Fprintf(os.Stderr, "Invalid --pattern value (using default %s)\n", pattern)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Missing value for --pattern (using default %s)\n", pattern)
			}
		}
	}

	cells := makecells(pattern)

	for !window.ShouldClose() {
		t := time.Now()

		for x := range cells {
			for _, c := range cells[x] {
				c.checkState(cells)
			}
		}

		draw(cells, window, program)

		time.Sleep(time.Second/time.Duration(fps) - time.Since(t))
	}
}

func initGlfw() *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, "Conway's Game Of Life", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
}

func initOpenGL() uint32 {
	if err := gl.Init(); err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL Version", version)

	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}
	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	prog := gl.CreateProgram()
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	gl.LinkProgram(prog)
	return prog
}

func draw(cells [][]*cell, window *glfw.Window, program uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	for x := range cells {
		for _, c := range cells[x] {
			c.draw()
		}
	}

	glfw.PollEvents()
	window.SwapBuffers()
}

func makeVao(points []float32) uint32 {
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

	return vao
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func makecells(patternName string) [][]*cell {
	rand.Seed(time.Now().UnixNano())

	cells := make([][]*cell, row)
	for x := 0; x < row; x++ {
		for y := 0; y < columns; y++ {
			c := newCell(x, y)
			c.aliveNext = false
			c.alive = false
			cells[x] = append(cells[x], c)
		}
	}

	// Check if a predefined pattern was requested.
	if patternCoords, ok := patterns[patternName]; ok && patternName != "random" {
		// Calculate the offset to center the pattern.
		// Find the max x and y to get the pattern's dimensions.
		maxX, maxY := 0, 0
		for _, coord := range patternCoords {
			if coord[0] > maxX {
				maxX = coord[0]
			}
			if coord[1] > maxY {
				maxY = coord[1]
			}
		}
		offsetX := (row - maxX) / 2
		offsetY := (columns - maxY) / 2

		// Set the cells in the grid according to the pattern.
		for _, coord := range patternCoords {
			x, y := coord[0], coord[1]
			if (x+offsetX) < row && (y+offsetY) < columns {
				cells[x+offsetX][y+offsetY].aliveNext = true
				cells[x+offsetX][y+offsetY].alive = true
			}
		}
	} else {
		// Default to random initialization.
		for x := 0; x < row; x++ {
			for y := 0; y < columns; y++ {
				c := cells[x][y]
				c.aliveNext = rand.Float32() < threshold
				c.alive = c.aliveNext
			}
		}
	}

	return cells
}

func newCell(x, y int) *cell {
	points := make([]float32, len(square), len(square))
	copy(points, square)

	for i := 0; i < len(points); i++ {
		var position float32
		var size float32
		switch i % 3 {
		case 0:
			size = 1.0 / float32(columns)
			position = float32(x) * size
		case 1:
			size = 1.0 / float32(row)
			position = float32(y) * size
		default:
			continue
		}

		if points[i] < 0 {
			points[i] = (position * 2) - 1
		} else {
			points[i] = ((position + size) * 2) - 1
		}
	}

	return &cell{
		drawable: makeVao(points),
		x:        x,
		y:        y,
	}
}
