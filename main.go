package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grafov/bcast"
)

var fileName string

func init() {
	fileName = "matrix.txt"
}

func main() {
	generateMatrix(fileName, 25)
	matrix := loadData(fileName)

	// fmt.Printf("%+v\n\n", matrix)

	var beforeS = time.Now()
	sequence(matrix)
	var afterS = time.Now()
	fmt.Printf("\n\nПоследовательное выполнение: %s", afterS.Sub(beforeS).String())
	fmt.Print("\n==================================================================================================\n\n")

	routinesCount := 2
	matrix = loadData(fileName)

	var beforeP = time.Now()
	// doCalculate(false, matrix)
	parallelV2(routinesCount, matrix)
	// doCalculate(false, matrix)
	var afterP = time.Now()
	fmt.Printf("\nПараллельное выполнение: %s", afterP.Sub(beforeP).String())
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
				matrixClone[i][j] = matrixClone[i][j] - (matrixClone[k][j] * K)
			}
		}
		for i := 0; i < n; i++ {
			for j := 0; j < n+1; j++ {
				matrix[i][j] = matrixClone[i][j]
			}
		}
	}
	fmt.Printf("%+v\n", matrix)

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
		// fmt.Printf("[%v] ", matrixClone[i][n])
	}
}

func parallelV2(routinesCount int, matrix [][]float64) {
	n := len(matrix)
	matrixClone := make([][]float64, n)

	for i := 0; i < n; i++ {
		matrixClone[i] = make([]float64, n+1)
		for j := 0; j < n+1; j++ {
			matrixClone[i][j] = matrix[i][j]
		}
	}

	rowsPerRoutine := make([][]int, routinesCount)
	for i := 0; i < routinesCount; i++ {
		rowsPerRoutine[i] = getChunk(n, i, routinesCount)
	}

	for matrixLine := 0; matrixLine < n; matrixLine++ {
		for i := 0; i < n+1; i++ {
			matrixClone[matrixLine][i] = matrixClone[matrixLine][i] / matrix[matrixLine][matrixLine]
		}
		var wg sync.WaitGroup
		for routine := 0; routine < routinesCount; routine++ {
			wg.Add(1)
			go func(k int, rows []int) {
				defer wg.Done()
				for row := 0; row < len(rows); row++ {
					i := rows[row] + 1
					if i == n || i < k+1 {
						continue
					}
					K := matrixClone[i][k] / matrixClone[k][k]
					for j := 0; j < n+1; j++ {
						matrixClone[i][j] = matrixClone[i][j] - (matrixClone[k][j] * K)
					}
				}
				for row := 0; row < len(rows); row++ {
					i := rows[row] + 1
					if i == n {
						break
					}
					for j := 0; j < n+1; j++ {
						matrix[i][j] = matrixClone[i][j]
					}
				}
			}(matrixLine, rowsPerRoutine[routine])
		}
		wg.Wait()
		for i := 0; i < n+1; i++ {
			matrix[matrixLine][i] = matrixClone[matrixLine][i]
		}
	}
	fmt.Printf("%+v\n", matrix)

}

//====================================

func parallel(routinesCount int, matrix [][]float64) {
	n := len(matrix)

	// Распределение строк по процессам
	rows := make([][]int, routinesCount)
	for i := 0; i < routinesCount; i++ {
		rows[i] = getChunk(n, i, routinesCount)
	}
	// messageBcast := bcast.NewGroup()
	iBcast := bcast.NewGroup()
	// go messageBcast.Broadcast(0)
	go iBcast.Broadcast(0)

	var firstWg sync.WaitGroup
	for k := 0; k < routinesCount; k++ { // Инициализация go-рутин
		firstWg.Add(1)
		go func(id int, rows []int, matrixClone [][]float64) {
			defer firstWg.Done()
			// memberRow := messageBcast.Join()
			memberI := iBcast.Join()
			row := 0
			var tmpRow []float64
			for {
				i := memberI.Recv().(int)
				// fmt.Printf("routine #%d recieved i = %d\n", id, i)
				if row < len(rows) && i == rows[row] {
					// fmt.Printf("routine #%d sending row #%d\n", id, i)
					// tmpRow = fmt.Sprintf("row#%d from %d", i, id)
					// memberRow.Send(tmpRow)
					// fmt.Printf("routine #%d end sending row #%d\n", id, i)
					row++
					// fmt.Printf("%d row = %d\n\n", id, row)
				} else {
					// fmt.Printf("routine #%d start wait row with id #%d\n", id, i)
					// tmpRow = memberRow.Recv().(string)
					// fmt.Printf("routine #%d end wait row with id #%d\n", id, i)
				}
				for i := 0; i < n+1; i++ {
					matrixClone[k][i] = matrixClone[k][i] / matrix[k][k]
				}
				for i := k + 1; i < n; i++ {
					K := matrixClone[i][k] / matrixClone[k][k]
					fmt.Printf("THIS %v THIS (RIGHT)\n", K)
					for j := 0; j < n+1; j++ {
						matrixClone[i][j] = matrixClone[i][j] - (matrixClone[k][j] * K)
					}
				}
				tmpRow = matrix[i]
				for j := row; j < len(rows); j++ {
					K := matrixClone[rows[j]][i] / tmpRow[i]
					fmt.Printf("%d %d %v\n", rows[j], i, K)
					for p := i; p < n+1; p++ {
						matrixClone[p][rows[j]] = matrixClone[p][rows[j]] - (K * tmpRow[p])
					}
				}
				for i := 0; i < n; i++ {
					for j := 0; j < n+1; j++ {
						matrix[i][j] = matrixClone[i][j]
					}
				}
				if i == n-2 {
					// fmt.Printf("%d exited (%+v)\n", id, rows)
					break
				}
			}
		}(k, rows[k], matrix)
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
	// fmt.Printf("%+v", matrix)

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
