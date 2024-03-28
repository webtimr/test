package main

import (
	"fmt"
	"sync"
	"time"
)

type Ttype struct {
	id         int
	cT         string // время создания
	fT         string // время выполнения
	taskRESULT []byte
}

func main() {
	taskCreator := func(a chan Ttype, wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			ft := time.Now().Format(time.RFC3339)
			if time.Now().Nanosecond()%2 > 0 { // вот такое условие появления ошибочных тасков
				ft = "Some error occurred"
			}
			a <- Ttype{cT: ft, id: int(time.Now().Unix())} // передаем таск на выполнение
		}
	}

	taskWorker := func(a Ttype, wg *sync.WaitGroup, doneTasks chan<- Ttype, undoneTasks chan<- Ttype) {
		defer wg.Done()
		tt, err := time.Parse(time.RFC3339, a.cT)
		if err != nil {
			undoneTasks <- a
			return
		}
		if tt.After(time.Now().Add(-20 * time.Second)) {
			a.taskRESULT = []byte("task has been succeeded")
		} else {
			a.taskRESULT = []byte("something went wrong")
		}
		a.fT = time.Now().Format(time.RFC3339Nano)

		time.Sleep(time.Millisecond * 150)

		if string(a.taskRESULT[14:]) == "succeeded" {
			doneTasks <- a
		} else {
			undoneTasks <- a
		}
	}

	result := make(map[int]Ttype)
	var doneTasks []Ttype
	var undoneTasks []Ttype
	var wg sync.WaitGroup

	taskChan := make(chan Ttype, 10)
	doneChan := make(chan Ttype)
	undoneChan := make(chan Ttype)

	go taskCreator(taskChan, &wg)

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			for t := range taskChan {
				wg.Add(1)
				go taskWorker(t, &wg, doneChan, undoneChan)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(doneChan)
		close(undoneChan)
	}()

	go func() {
		for r := range doneChan {
			result[r.id] = r
		}
	}()

	go func() {
		for r := range undoneChan {
			undoneTasks = append(undoneTasks, r)
		}
	}()

	time.Sleep(time.Second * 3)

	println("Errors:")
	for _, err := range undoneTasks {
		fmt.Printf("Task id %d time %s, error %s\n", err.id, err.cT, err.taskRESULT)
	}

	println("Done tasks:")
	for id := range doneTasks {
		println(id)
	}
}
