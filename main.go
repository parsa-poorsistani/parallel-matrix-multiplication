package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type Task struct {
	Row    int
	Col    int
	Result int
}

var matrixData struct {
	MatrixA [][]int `json:"matrixA"`
	MatrixB [][]int `json:"matrixB"`
}

var (
	taskQueue  chan Task
	taskResult map[int]map[int]int
	taskMutex  sync.Mutex
	wg         sync.WaitGroup
)

func worker(id int) {
	for task := range taskQueue {
		//for monitoring the worker threads
		log.Printf("Worker %d started task: Row=%d, Column=%d\n", id, task.Row, task.Col)
		result := 0

		for k := 0; k < len(matrixData.MatrixA[0]); k++ {
			result += matrixData.MatrixA[task.Row][k] * matrixData.MatrixB[k][task.Col]
		}

		taskMutex.Lock()
		if taskResult[task.Row] == nil {
			taskResult[task.Row] = make(map[int]int)
		}
		taskResult[task.Row][task.Col] = result
		taskMutex.Unlock()

		log.Printf("Worker %d completed task: Row=%d, Column=%d\n", id, task.Row, task.Col)

		wg.Done()
	}
}

func multiplyMatrix(c *gin.Context) {

	if err := c.BindJSON(&matrixData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	taskResult = make(map[int]map[int]int)
	numWorkers := len(matrixData.MatrixA) * len(matrixData.MatrixB[0])
	taskQueue = make(chan Task, numWorkers)

	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go worker(i)
	}

	for i := range matrixData.MatrixA {
		for j := range matrixData.MatrixB[0] {
			taskQueue <- Task{Row: i, Col: j}
		}
	}

	wg.Wait()

	c.JSON(http.StatusOK, taskResult)
}
func main() {
	router := gin.Default()
	router.POST("matrix/multiply", multiplyMatrix)
	router.Run(":8080")
}

// type Task struct {
// 	Row    int
// 	Col    int
// 	Result int
// }

// type Worker struct {
// 	ID   int
// 	Task chan Task
// }

// var matrixData struct {
// 	MatrixA [][]int `json:"matrixA"`
// 	MatrixB [][]int `json:"matrixB"`
// }

// var (
// 	workers     []*Worker
// 	taskResult  map[int]map[int]int
// 	taskMutex   sync.Mutex
// 	wg          sync.WaitGroup
// 	numWorkers  int
// 	numTasks    int
// 	workerMutex sync.Mutex
// )

// func workerRoutine(worker *Worker) {
// 	for task := range worker.Task {
// 		log.Printf("Worker %d started task: Row=%d, Column=%d\n", worker.ID, task.Row, task.Col)
// 		result := 0

// 		for k := 0; k < len(matrixData.MatrixA[0]); k++ {
// 			result += matrixData.MatrixA[task.Row][k] * matrixData.MatrixB[k][task.Col]
// 		}


// 		taskMutex.Lock()
// 		if taskResult[task.Row] == nil {
// 			taskResult[task.Row] = make(map[int]int)
// 		}
// 		taskResult[task.Row][task.Col] = result
// 		taskMutex.Unlock()

// 		log.Printf("Worker %d completed task: Row=%d, Column=%d\n", worker.ID, task.Row, task.Col)

// 		wg.Done()
// 	}

// 	workerMutex.Lock()
// 	workers = append(workers[:worker.ID], workers[worker.ID+1:]...)
// 	workerMutex.Unlock()
// }

// func multiplyMatrix(c *gin.Context) {
// 	if err := c.BindJSON(&matrixData); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	numWorkers = len(matrixData.MatrixA) // Set initial number of workers equal to the number of rows in matrixA
// 	numTasks = len(matrixData.MatrixA) * len(matrixData.MatrixB[0])
// 	taskResult = make(map[int]map[int]int)

// 	workers = make([]*Worker, numWorkers)
// 	for i := 0; i < numWorkers; i++ {
// 		worker := &Worker{
// 			ID:   i,
// 			Task: make(chan Task),
// 		}
// 		workers[i] = worker
// 		go workerRoutine(worker)
// 	}

// 	wg.Add(numTasks)

// 	// Assign tasks to available workers
// 	for i := range matrixData.MatrixA {
// 		for j := range matrixData.MatrixB[0] {
// 			workerMutex.Lock()
// 			worker := workers[i%len(workers)] // Round-robin assignment of tasks to workers
// 			worker.Task <- Task{Row: i, Col: j}
// 			workerMutex.Unlock()
// 		}
// 	}

// 	wg.Wait()

// 	c.JSON(http.StatusOK, taskResult)
// }

// func addWorker(c *gin.Context) {
// 	workerMutex.Lock()
// 	defer workerMutex.Unlock()

// 	worker := &Worker{
// 		ID:   len(workers),
// 		Task: make(chan Task),
// 	}
// 	workers = append(workers, worker)
// 	go workerRoutine(worker)

// 	c.JSON(http.StatusOK, gin.H{"message": "Worker added successfully"})
// }

// func removeWorker(c *gin.Context) {
// 	workerMutex.Lock()
// 	defer workerMutex.Unlock()

// 	if len(workers) > 0 {
// 		worker := workers[len(workers)-1]
// 		close(worker.Task)
// 		workers = workers[:len(workers)-1]
// 		c.JSON(http.StatusOK, gin.H{"message": "Worker removed successfully"})
// 	} else {
// 		c.JSON(http.StatusOK, gin.H{"message": "No workers available to remove"})
// 	}
// }

// func main() {
// 	router := gin.Default()
// 	router.POST("/matrix/multiply", multiplyMatrix)
// 	router.POST("/worker/add", addWorker)
// 	router.POST("/worker/remove", removeWorker)
// 	router.Run(":8080")
// }