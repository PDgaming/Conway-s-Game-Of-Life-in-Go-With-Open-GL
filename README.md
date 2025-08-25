# Conway's Game of Life (OpenGL/Go)

This project is an interactive and visual implementation of [Conway's Game of Life](https://en.wikipedia.org/wiki/Conway%27s_Game_of_Life) using Go and OpenGL. It features real-time rendering, customizable simulation parameters, and several classic patterns.

## Features

- **Real-time visualization** of the Game of Life grid using OpenGL.
- **Configurable grid size** and simulation speed.
- **Pattern selection**: Choose from random, blinker, glider, lightweight spaceship, and pulsar.
- **Command-line arguments** for FPS, threshold, and pattern.
- **Toroidal (wrap-around) grid**: Edges connect for continuous simulation.

## Requirements

- Go 1.25 or newer
- OpenGL 4.1 compatible GPU/driver
- [go-gl/gl](https://github.com/go-gl/gl) and [go-gl/glfw](https://github.com/go-gl/glfw) (see `go.mod`)

## Installation

1. **Clone the repository:**

   ```
   git clone https://github.com/PDgaming/Conway-s-Game-Of-Life-in-Go-With-Open-GL
   cd Conway-s-Game-Of-Life-in-Go-With-Open-GL
   ```

2. **Install dependencies:**

   ```
   go mod download
   ```

3. **Build the project:**
   ```
   go build
   ```

## Usage

To run the program, simply execute the following command:

```
./ConwaysGameOfLife
```

Set optional command-line arguments to customize the simulation:

- **FPS**: Set the simulation speed in frames per second (default: 10).

  ```
  ./ConwaysGameOfLife --fps 10
  ```

  or

  ```
    ./ConwaysGameOfLife -f 10
  ```

- **Threshold**: Set the probability threshold for cell state change (default: 0.15).

  ```
  ./ConwaysGameOfLife --threshold 0.15
  ```

  or

  ```bash
    ./ConwaysGameOfLife -t 0.15
  ```

- **Pattern**: Set the initial pattern (default: random). Available patterns: random, blinker, glider, lightweightspaceship, pulsar.
  ```
  ./ConwaysGameOfLife --pattern random
  ```
  or
  ```bash
    ./ConwaysGameOfLife -p random
  ```
