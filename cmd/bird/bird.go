package bird

import (
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/gigurra/tofu/cmd/bird/jukebox"
	"github.com/gigurra/tofu/cmd/common"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type Params struct {
	FPS   int `short:"f" optional:"true" help:"Frames per second" default:"15"`
	Speed int `short:"s" optional:"true" help:"Game speed (1-5)" default:"2"`
}

func Cmd() *cobra.Command {
	return boa.CmdT[Params]{
		Use:   "bird",
		Short: "Play Flappy Tofu - a terminal flappy bird game",
		Long: `Play Flappy Tofu in your terminal!

Controls:
  SPACE or ENTER - Flap (jump)
  q or ESC       - Quit

Guide your tofu through the gaps in the chopsticks!
Press Ctrl+C to exit.`,
		ParamEnrich: common.DefaultParamEnricher(),
		RunFunc: func(params *Params, cmd *cobra.Command, args []string) {
			if err := runBird(params); err != nil {
				fmt.Fprintf(os.Stderr, "bird: %v\n", err)
				os.Exit(1)
			}
		},
	}.ToCobra()
}

// Game constants
const (
	gravity      = 0.4
	flapStrength = -2.5
	pipeWidth    = 5
	gapSize      = 10
	pipeSpacing  = 50
)

// Game state
type gameState struct {
	birdY      float64
	birdVel    float64
	pipes      []pipe
	score      int
	highScore  int
	gameOver   bool
	started    bool
	width      int
	height     int
	termHeight int // Full terminal height for padding
	frame      int
	speed      int
}

type pipe struct {
	x      int
	gapY   int
	passed bool
}

// Embed fs under /music
//
//go:embed music
var musicFS embed.FS

func runBird(params *Params) error {

	// initialize jukebox if needed
	slog.Info("Loading music for Flappy Tofu...")
	jb := jukebox.New()
	defer jb.Clear()
	// list files in musicFS
	musicFiles, err := musicFS.ReadDir("music")
	if err == nil && len(musicFiles) > 0 {
		for _, file := range musicFiles {
			if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".mp3") {
				songPath := "music/" + file.Name()
				slog.Info("loading embedded music file", "file", file.Name())
				bytes, err := fs.ReadFile(musicFS, songPath)
				if err != nil {
					slog.Error("failed to read embedded music file", "file", file.Name(), "error", err)
					continue
				}
				_, err = jb.LoadBytes(file.Name(), bytes)
				if err != nil {
					slog.Error("failed to load embedded music file into jukebox", "file", file.Name(), "error", err)
				}
			}
		}
	}

	jb.SetShuffle(true)
	if err != nil {
		slog.Error("failed to start jukebox playback", "error", err)
	}

	slog.Info("Done loading music, starting game")

	// Set terminal to raw mode for input
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Hide cursor
	fmt.Print("\033[?25l")
	// Clear screen
	fmt.Print("\033[2J")

	// Restore cursor and clear screen on exit
	defer func() {
		fmt.Print("\033[?25h") // Show cursor
		fmt.Print("\033[2J")   // Clear screen
		fmt.Print("\033[H")    // Move to home
	}()

	fps := params.FPS
	if fps < 5 {
		fps = 5
	}
	if fps > 60 {
		fps = 60
	}

	speed := params.Speed
	if speed < 1 {
		speed = 1
	}
	if speed > 5 {
		speed = 5
	}

	// Get terminal size
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width, height = 80, 24
	}

	// Initialize game
	game := &gameState{
		width:      width,
		height:     height - 5, // Leave room for score and messages
		termHeight: height,
		speed:      speed,
	}
	resetGame(game, jb)

	// Build frame buffer with extra rows for UI
	// game.height rows for game area + 1 for score + up to 3 for messages
	totalRows := game.height + 4
	screen := make([][]rune, totalRows)
	for i := range screen {
		screen[i] = make([]rune, game.width)
		for j := range screen[i] {
			screen[i][j] = ' '
		}
	}

	frameDuration := time.Second / time.Duration(fps)

	// Input channel
	inputChan := make(chan byte, 10)
	go readInput(inputChan)

	for {

		if game.gameOver || !game.started {
			if jb.IsPlaying() {
				jb.Stop()
			}
		} else {
			if !jb.IsPlaying() {
				_ = jb.Play()
			}
		}

		for len(inputChan) > 0 {
			key := <-inputChan
			if key == 'q' {
				return nil
			}
			if key == 'n' {
				_ = jb.Next()
			}
			if key == ' ' || key == 13 || key == 10 { // space or enter
				if game.gameOver {
					resetGame(game, jb)
				} else {
					game.started = true
					game.birdVel = flapStrength
				}
			}
		}

		level := min(100, max(1, 1+game.score/2))
		frameDuration = time.Second / time.Duration(fps*(9+level)/10)

		updateGame(game, jb)
		renderGame(game, screen, level, jb)
		game.frame++

		time.Sleep(frameDuration)
	}
}

func readInput(ch chan<- byte) {
	reader := bufio.NewReader(os.Stdin)
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return
		}
		ch <- b
	}
}

func resetGame(game *gameState, jb *jukebox.Jukebox) {
	if game.score > game.highScore {
		game.highScore = game.score
	}
	game.birdY = float64(game.height) / 2
	game.birdVel = 0
	game.pipes = []pipe{}
	game.score = 0
	game.gameOver = false
	game.started = false
	game.frame = 0

	// Create initial pipes
	for i := 0; i < 3; i++ {
		game.pipes = append(game.pipes, newPipe(game.width+i*pipeSpacing, game.height))
	}

	_ = jb.Next()
}

func newPipe(x, height int) pipe {
	// Random gap position, leaving room at top and bottom
	minGap := gapSize/2 + 2
	maxGap := height - gapSize/2 - 2
	gapY := minGap + rand.Intn(maxGap-minGap)
	return pipe{x: x, gapY: gapY, passed: false}
}

func updateGame(game *gameState, jb *jukebox.Jukebox) {
	if game.gameOver || !game.started {
		return
	}

	// Apply gravity
	game.birdVel += gravity
	game.birdY += game.birdVel

	// Move pipes
	for i := range game.pipes {
		game.pipes[i].x -= game.speed
	}

	// Check for scoring
	birdX := 10 // Bird's X position (fixed)
	for i := range game.pipes {
		if !game.pipes[i].passed && game.pipes[i].x+pipeWidth < birdX {
			game.pipes[i].passed = true
			game.score++
		}
	}

	// Remove off-screen pipes and add new ones
	if len(game.pipes) > 0 && game.pipes[0].x < -pipeWidth {
		game.pipes = game.pipes[1:]
		// Add new pipe at the end
		lastX := game.pipes[len(game.pipes)-1].x
		game.pipes = append(game.pipes, newPipe(lastX+pipeSpacing, game.height))
	}

	// Check collisions
	if checkCollision(game) {
		game.gameOver = true
		jb.Stop()
	}
}

func checkCollision(game *gameState) bool {
	birdX := 10
	birdY := int(game.birdY)

	// Check floor/ceiling
	if birdY <= 0 || birdY >= game.height-1 {
		return true
	}

	// Check pipe collisions
	for _, p := range game.pipes {
		// Check if bird is in pipe's X range
		if birdX >= p.x-2 && birdX <= p.x+pipeWidth {
			// Check if bird is outside the gap
			if birdY < p.gapY-gapSize/2 || birdY > p.gapY+gapSize/2 {
				return true
			}
		}
	}

	return false
}

func renderGame(game *gameState, backBuffer [][]rune, level int, jb *jukebox.Jukebox) {

	// Clear back buffer
	for i := range backBuffer {
		for j := range backBuffer[i] {
			backBuffer[i][j] = ' '
		}
	}

	// Draw ground
	for x := 0; x < game.width; x++ {
		if game.height-1 >= 0 && game.height-1 < len(backBuffer) {
			backBuffer[game.height-1][x] = '='
		}
	}

	// Draw ceiling
	for x := 0; x < game.width; x++ {
		if len(backBuffer) > 0 {
			backBuffer[0][x] = '='
		}
	}

	// Draw pipes (chopsticks!)
	for _, p := range game.pipes {
		for x := p.x; x < p.x+pipeWidth && x < game.width; x++ {
			if x < 0 {
				continue
			}
			// Draw top part
			for y := 1; y < p.gapY-gapSize/2; y++ {
				if y < game.height {
					backBuffer[y][x] = chopstickChar(y)
				}
			}
			// Draw bottom part
			for y := p.gapY + gapSize/2 + 1; y < game.height-1; y++ {
				if y >= 0 {
					backBuffer[y][x] = chopstickChar(y)
				}
			}
		}
	}

	// Draw bird (tofu!)
	birdY := int(game.birdY)
	birdX := 10
	if birdY > 0 && birdY < game.height-1 && birdX+3 < game.width {
		// Tofu bird with animation
		var tofuTop, tofuMid, tofuBot string
		if game.birdVel < 0 {
			// Flapping up
			tofuTop = "^_^"
			tofuMid = "[#]"
			tofuBot = "\\~/"
		} else {
			// Falling
			tofuTop = "o_o"
			tofuMid = "[#]"
			tofuBot = "/_\\"
		}

		if game.gameOver {
			tofuTop = "x_x"
			tofuMid = "[#]"
			tofuBot = "/|\\"
		}

		// Draw tofu
		if birdY-1 > 0 && birdY-1 < game.height {
			for i, c := range tofuTop {
				if birdX+i < game.width {
					backBuffer[birdY-1][birdX+i] = c
				}
			}
		}
		if birdY > 0 && birdY < game.height {
			for i, c := range tofuMid {
				if birdX+i < game.width {
					backBuffer[birdY][birdX+i] = c
				}
			}
		}
		if birdY+1 > 0 && birdY+1 < game.height {
			for i, c := range tofuBot {
				if birdX+i < game.width {
					backBuffer[birdY+1][birdX+i] = c
				}
			}
		}
	}

	// Draw score line into buffer
	currentSongStr := ""
	if song := jb.CurrentSong(); song != nil {
		currentSongStr = strings.TrimSuffix(song.Name, ".mp3")
	}
	scoreText := fmt.Sprintf(" Score: %d  |  High Score: %d  |  Level: %d  |  Song: %s  |  SPACE=Flap  Q=Quit  N=Next Song ", game.score, game.highScore, level, currentSongStr)
	drawTextToRow(backBuffer[game.height], scoreText, 0)

	// Draw game state messages into buffer
	msgRow := game.height + 1
	if game.gameOver {
		drawCenteredText(backBuffer[msgRow], game.width, "  GAME OVER!  ")
		drawCenteredText(backBuffer[msgRow+1], game.width, "  Press SPACE to restart  ")
	} else if !game.started {
		drawCenteredText(backBuffer[msgRow], game.width, "  FLAPPY TOFU  ")
		drawCenteredText(backBuffer[msgRow+1], game.width, "  Press SPACE to start  ")
		drawCenteredText(backBuffer[msgRow+2], game.width, "  Guide the tofu through the chopsticks!  ")
	}

	fmt.Print("\033[H") // Move cursor to top
	for i, row := range backBuffer {
		fmt.Print(string(row))
		if i < len(backBuffer)-1 {
			fmt.Print("\r\n") // Need \r in raw mode
		}
	}
}

func drawTextToRow(row []rune, text string, startX int) {
	for i, c := range text {
		if startX+i < len(row) {
			row[startX+i] = c
		}
	}
}

func drawCenteredText(row []rune, width int, text string) {
	padding := (width - len(text)) / 2
	if padding < 0 {
		padding = 0
	}
	drawTextToRow(row, text, padding)
}

func chopstickChar(y int) rune {
	// Static alternating pattern for chopstick texture
	if y%2 == 0 {
		return '#'
	}
	return ':'
}
