package main

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestReadInput(t *testing.T) {
	isTesting = true
	input := "test input\n"
	inputChan := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		readInput(ctx, inputChan, strings.NewReader(input), &wg)
	}()

	select {
	case result := <-inputChan:
		if result != "test input" {
			t.Errorf("TestReadInput error, got '%s'", result)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("TestReadInput timeout error")
	}

	cancel()
	wg.Wait()
}

func TestWriteToFile(t *testing.T) {
	inputChan := make(chan string)
	var buf bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		writeToFile(ctx, inputChan, &buf, &wg)
	}()

	inputChan <- "input test text"
	close(inputChan)

	wg.Wait()

	if buf.String() != "input test text\n" {
		t.Errorf("TestWriteToFile err, got '%s'", buf.String())
	}

	cancel()
}
