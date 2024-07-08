package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var isTesting = false

func main() {
	inputChan := make(chan string)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	file, err := os.Create("result.txt")
	if err != nil {
		log.Println("Ошибка создания файла:", err)
		return
	}
	defer file.Close()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go readInput(ctx, inputChan, os.Stdin, &wg)

	wg.Add(1)
	go writeToFile(ctx, inputChan, file, &wg)

	go func() {
		<-sigChan
		log.Println("Получен сигнал завершения, завершаем работу")
		cancel()
	}()

	<-ctx.Done()

	wg.Wait()

	fmt.Println("Файл сохранен.")
}

func writeToFile(ctx context.Context, inputChan <-chan string, writer io.Writer, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Println("Завершаем запись в файл")
			return
		case input, ok := <-inputChan:
			if !ok {
				return
			}

			if _, err := writer.Write([]byte(input + "\n")); err != nil {
				log.Println("Ошибка записи в файл: ", err)
				return
			}
		}
	}
}

func readInput(ctx context.Context, inputChan chan<- string, reader io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()

	bufReader := bufio.NewReader(reader)
	inputChanInternal := make(chan string)

	go func() {
		defer close(inputChanInternal)
		for {
			if !isTesting {
				fmt.Print("Введите текст: ")
			}
			line, err := bufReader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					log.Println("Ошибка чтения ввода:", err)
				}
				return
			}
			inputChanInternal <- line[:len(line)-1] // убираем символ новой строки
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("Завершаем чтение ввода")
			close(inputChan)
			return
		case line, ok := <-inputChanInternal:
			if !ok {
				close(inputChan)
				return
			}
			inputChan <- line
		}
	}
}
