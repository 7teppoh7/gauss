package main

import (
	"bufio"
	"fmt"
	"github.com/grafov/bcast"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var fileName string

func init() {
	fileName = "matrix.txt"
}

func test() {
	routinesCount := 2
	n := 5
	rows := make([][]int, routinesCount)
	for i := 0; i < routinesCount; i++ {
		rows[i] = getChunk(n, i, routinesCount)
	}

	messageBcast := bcast.NewGroup()
	iBcast := bcast.NewGroup()
	go messageBcast.Broadcast(0)
	go iBcast.Broadcast(0)

	var firstWg sync.WaitGroup
	for k := 0; k < routinesCount; k++ { // Инициализация go-рутин
		firstWg.Add(1)
		go func(id int, rows []int) {
			defer firstWg.Done()
			memberRow := messageBcast.Join()
			memberI := iBcast.Join()
			row := 0
			// var tmpRow []float64
			for {
				if row == len(rows) {
					fmt.Printf("%d exited (%+v)\n", id, rows)
					break
				}
				i := memberI.Recv().(int)
				var tmpRow string

				//fmt.Printf("routine #%d recieved i = %d\n", id, i)
				if i == rows[row] {
					//fmt.Printf("routine #%d sending row #%d\n", id, i)
					tmpRow = fmt.Sprintf("row#%d from %d", i, id)
					memberRow.Send(tmpRow)
					//fmt.Printf("routine #%d end sending row #%d\n", id, i)
					row++
					fmt.Printf("%d row = %d\n\n", id, row)
				} else {
					fmt.Printf("recieve\n")
					//fmt.Printf("routine #%d start wait row with id #%d\n", id, i)
					tmpRow = memberRow.Recv().(string)
					//fmt.Printf("routine #%d end wait row with id #%d\n", id, i)
				}

				for j := row; j < len(rows); j++ {
					fmt.Printf("process '%s' in %d for row %d\n", tmpRow, id, rows[j])
				}
			}
		}(k, rows[k])
	}
	for {
		if iBcast.MemberCount() == routinesCount {
			break
		}
	}

	for i := 0; i < n-1; i++ {
		iBcast.Send(i)
	}
	firstWg.Wait()
}

func main() {
	routinesCount := 9
	test()
	return

	generateMatrix(fileName, 20)
	matrix := loadData(fileName)
	var beforeP = time.Now()
	parallel(routinesCount, matrix)
	var afterP = time.Now()
	fmt.Printf("\nПараллельное выполнение: %s", afterP.Sub(beforeP).String())
	fmt.Print("\n==================================================================================================\n\n")
	matrix = loadData(fileName)
	var beforeS = time.Now()
	sequence(matrix)
	var afterS = time.Now()
	fmt.Printf("\n\nПоследовательное выполнение: %s", afterS.Sub(beforeS).String())
	fmt.Print("\n==================================================================================================\n\n")

}

func sequence(matrix [][]float64) {
	n := len(matrix)
	matrixClone := make([][]float64, n)

	for i := 0; i < n; i++ {
		matrixClone[i] = make([]float64, n+1)
		for j := 0; j < n+1; j++ {
			matrixClone[i][j] = matrix[i][j]
		}
	}

	// Прямой ход
	for k := 0; k < n; k++ {
		for i := 0; i < n+1; i++ {
			matrixClone[k][i] = matrixClone[k][i] / matrix[k][k]
		}
		for i := k + 1; i < n; i++ {
			K := matrixClone[i][k] / matrixClone[k][k]
			for j := 0; j < n+1; j++ {
				matrixClone[i][j] = matrixClone[i][j] - matrixClone[k][j]*K
			}
		}
		for i := 0; i < n; i++ {
			for j := 0; j < n+1; j++ {
				matrix[i][j] = matrixClone[i][j]
			}
		}
	}

	// Обратный ход
	for k := n - 1; k > -1; k-- {
		for i := n; i > -1; i-- {
			matrixClone[k][i] = matrixClone[k][i] / matrix[k][k]
		}
		for i := k - 1; i > -1; i-- {
			K := matrixClone[i][k] / matrixClone[k][k]
			for j := n; j > -1; j-- {
				matrixClone[i][j] = matrixClone[i][j] - (matrixClone[k][j] * K)
			}
		}
	}

	for i := 0; i < n; i++ {
		fmt.Printf("[%v] ", matrixClone[i][n])
	}
}

func parallel(routinesCount int, matrix [][]float64) {
	n := len(matrix)

	// Распределение строк по процессам
	rows := make([][]int, routinesCount)
	for i := 0; i < routinesCount; i++ {
		rows[i] = getChunk(n, i, routinesCount)
	}

	// Прямой ход
	var firstWg sync.WaitGroup
	for i := 0; i < routinesCount; i++ { // Инициализация go-рутин
		firstWg.Add(1)
		go func(rows []int) {
			defer firstWg.Done()
			row := 0
			// var tmpRow []float64
			for i := 0; i < n-1; i++ {
				if i == rows[row] {
					//send row!
					// tmpRow =
					row++
				} else {
					// tmpRow =
				}
				// for j := 0; j < count; j++ {

				// }
			}
		}(rows[i])
	}
	firstWg.Wait()

	// Обратный ход (Зануление верхнего правого угла)
	// for k := n - 1; k > -1; k-- {
	// 	for i := n; i > -1; i-- {
	// 		matrixClone[k][i] = matrixClone[k][i] / matrix[k][k]
	// 	}
	// 	for i := k - 1; i > -1; i-- {
	// 		K := matrixClone[i][k] / matrixClone[k][k]
	// 		for j := n; j > -1; j-- {
	// 			matrixClone[i][j] = matrixClone[i][j] - (matrixClone[k][j] * K)
	// 		}
	// 	}
	// }

	// for i := 0; i < n; i++ {
	// 	fmt.Printf("[%v] ", matrixClone[i][n])
	// }
}

func doCalculate(useParallel bool, matrix [][]float64) {

	n := len(matrix)

	var i, j, k int

	var tmp float64
	var xx = make([]float64, n)

	for i = 0; i < n; i++ {
		tmp = matrix[i][i]
		for j = n; j >= i; j-- {
			matrix[i][j] /= tmp
		}

		if useParallel {
			var wg sync.WaitGroup

			for j = i + 1; j < n; j++ {
				wg.Add(1)
				j := j
				go func() {
					defer wg.Done()
					tmp = matrix[j][i]
					for k = n; k >= i; k-- {
						matrix[j][k] -= tmp * matrix[i][k]
					}
				}()
			}
			wg.Wait()
		} else {
			for j = i + 1; j < n; j++ {
				tmp = matrix[j][i]
				for k = n; k >= i; k-- {
					matrix[j][k] -= tmp * matrix[i][k]
				}
			}
		}
	}

	xx[n-1] = matrix[n-1][n]
	for i = n - 2; i >= 0; i-- {
		xx[i] = matrix[i][n]

		if useParallel {
			var wg sync.WaitGroup
			for j = i + 1; j < n; j++ {
				wg.Add(1)
				j := j
				go func() {
					defer wg.Done()
					xx[i] -= matrix[i][j] * xx[j]
				}()
			}
			wg.Wait()
		} else {
			for j = i + 1; j < n; j++ {
				xx[i] -= matrix[i][j] * xx[j]
			}
		}
	}

	for i = 0; i < len(xx); i++ {
		fmt.Printf("[%v] ", xx[i])
	}
}

func getChunk(total, routineNum, processesCount int) []int {
	result := make([]int, 0)
	for i := routineNum; i < total; i = i + processesCount {
		result = append(result, i)
	}
	return result
}

func generateMatrix(fileName string, size int) {
	err := os.Truncate(fileName, 0)
	if err != nil {
		panic(err)
	}
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	for i := 0; i < size; i++ {
		line := genLine(size)
		_, err = fmt.Fprintln(file, line)
		if err != nil {
			panic(err)
		}
	}
}

func genLine(size int) string {
	line := ""
	rands := make([]int, size)

	for i := 0; i < size; i++ {
		rands[i] = rand.Intn(20) - 10
	}

	for _, val := range rands {
		line += strconv.Itoa(val) + " "
	}
	line += strconv.Itoa(rand.Intn(20) - 10)
	return line
}

func loadData(fileName string) [][]float64 {
	file, err := os.Open(fileName)

	if err != nil {
		panic(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	var matrix [][]float64

	counter := 0

	for scanner.Scan() {
		line := scanner.Text()
		strNums := strings.Split(line, " ")
		matrix = append(matrix, make([]float64, len(strNums)))

		for i, strNum := range strNums {
			matrix[counter][i], err = strconv.ParseFloat(strNum, 64)
			if err != nil {
				panic(err)
			}
		}
		counter++
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return matrix
}
