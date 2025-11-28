package jukebox_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"jukebox"
)

func Example_basic() {
	// Create a new jukebox
	jb := jukebox.New()

	// Load songs from files (actual MP3 files required)
	// song1, err := jb.Load("/path/to/song1.mp3")
	// if err != nil {
	//     log.Fatal(err)
	// }

	// List all songs
	fmt.Printf("Loaded %d songs\n", jb.Count())

	// Start playback
	// jb.Play()

	// Check status
	status := jb.Status()
	fmt.Printf("State: %s, Shuffle: %v\n", status.State, status.Shuffle)

	// Output:
	// Loaded 0 songs
	// State: stopped, Shuffle: false
}

func TestPlayRealSong(t *testing.T) {

	jb := jukebox.New()
	defer jb.Clear()

	// Load an actual MP3 file
	song, err := jb.Load("/mnt/c/Users/gigur/Downloads/Totally Accurate Isekai Simulator.mp3")
	if err != nil {
		log.Printf("Could not load song: %v", err)
		return
	}

	fmt.Printf("Loaded: %s\n", song.Name)
	// Play the song
	if err := jb.Play(); err != nil {
		log.Fatal(err)
	}

	time.Sleep(10 * time.Second)
	/*

		// Let it play for a bit
		time.Sleep(5 * time.Second)

		// Pause
		jb.Pause()
		fmt.Println("Paused")

		// Resume
		jb.Play()

		// Skip to next
		jb.Next()

		// Enable shuffle
		jb.SetShuffle(true)
		fmt.Printf("Shuffle: %v\n", jb.IsShuffled())

		// Stop
		jb.Stop()*/
}

func Example_playback() {
	jb := jukebox.New()

	// Load an actual MP3 file
	song, err := jb.Load("/path/to/your/song.mp3")
	if err != nil {
		log.Printf("Could not load song: %v", err)
		return
	}

	fmt.Printf("Loaded: %s\n", song.Name)

	// Play the song
	if err := jb.Play(); err != nil {
		log.Fatal(err)
	}

	// Let it play for a bit
	time.Sleep(5 * time.Second)

	// Pause
	jb.Pause()
	fmt.Println("Paused")

	// Resume
	jb.Play()

	// Skip to next
	jb.Next()

	// Enable shuffle
	jb.SetShuffle(true)
	fmt.Printf("Shuffle: %v\n", jb.IsShuffled())

	// Stop
	jb.Stop()
}

func Example_shuffle() {
	jb := jukebox.New()

	// Toggle shuffle
	shuffled := jb.ToggleShuffle()
	fmt.Printf("Shuffle enabled: %v\n", shuffled)

	// Reshuffle creates a new random order
	jb.Reshuffle()

	// Output:
	// Shuffle enabled: true
}

func Example_concurrency() {
	jb := jukebox.New()

	// The jukebox is safe for concurrent use from multiple goroutines
	done := make(chan bool, 3)

	// Goroutine 1: Check status repeatedly
	go func() {
		for i := 0; i < 10; i++ {
			_ = jb.Status()
			_ = jb.Count()
		}
		done <- true
	}()

	// Goroutine 2: Control playback
	go func() {
		for i := 0; i < 10; i++ {
			_ = jb.IsPlaying()
			_ = jb.IsPaused()
		}
		done <- true
	}()

	// Goroutine 3: Toggle shuffle
	go func() {
		for i := 0; i < 5; i++ {
			jb.ToggleShuffle()
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done

	fmt.Println("Concurrent operations completed safely")
	// Output:
	// Concurrent operations completed safely
}
