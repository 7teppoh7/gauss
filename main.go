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
)

var (
	fileName    string
	matrixInRam [][]float64
)

const withAnswer = true
const withDebug = true
const useFsForMatrix = true

func init() {
	fileName = "matrix.txt"
}

func main() {
	var matrix [][]float64

	matrixSize := 3

	if useFsForMatrix {
		generateMatrix(fileName, matrixSize)
		matrix = loadData(fileName)
	} else {
		generateMatrixInRAM(matrixSize)

		matrix = make([][]float64, matrixSize)
		for i := 0; i < matrixSize; i++ {
			matrix[i] = make([]float64, matrixSize+1)
			for j := 0; j < matrixSize+1; j++ {
				matrix[i][j] = matrixInRam[i][j]
			}
		}
	}

	var beforeS = time.Now()
	sequence(matrix)
	var afterS = time.Now()
	fmt.Printf("\n\nПоследовательное выполнение: %v (sec), %.5f (min)", afterS.Sub(beforeS).Seconds(), afterS.Sub(beforeS).Minutes())
	fmt.Print("\n==================================================================================================\n\n")
	timeSequence := afterS.Sub(beforeS).Seconds()

	routinesCount := 16
	if useFsForMatrix {
		matrix = loadData(fileName)
	} else {
		matrix = matrixInRam
	}

	var beforeP = time.Now()
	parallel(routinesCount, matrix)
	var afterP = time.Now()
	fmt.Printf("\nПараллельное выполнение: %v (sec), %.5f (min)", afterP.Sub(beforeP).Seconds(), afterP.Sub(beforeP).Minutes())
	fmt.Print("\n==================================================================================================\n\n")
	timeParallel := afterP.Sub(beforeP).Seconds()

	fmt.Printf("Ускорение составило: %.4f", (timeSequence / timeParallel))
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

	if withAnswer {
		for i := 0; i < n; i++ {
			fmt.Printf("[%v] ", matrixClone[i][n])
		}
	}
}

func parallel(routinesCount int, matrix [][]float64) {
	n := len(matrix)
	matrixClone := make([][]float64, n)

	for i := 0; i < n; i++ {
		matrixClone[i] = make([]float64, n+1)
		for j := 0; j < n+1; j++ {
			matrixClone[i][j] = matrix[i][j]
		}
	}

	// Распределение строк по go-рутинам
	rowsPerRoutine := make([][]int, routinesCount)
	for i := 0; i < routinesCount; i++ {
		rowsPerRoutine[i] = getChunk(n, i, routinesCount)
	}

	// Прямой ход
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

	if withAnswer {
		for i := 0; i < n; i++ {
			fmt.Printf("[%v] ", matrixClone[i][n])
		}
	}

}

func getChunk(total, routineNum, processesCount int) []int {
	result := make([]int, 0)
	for i := routineNum; i < total; i = i + processesCount {
		result = append(result, i)
	}
	return result
}

func generateMatrixInRAM(size int) {
	if withDebug {
		fmt.Println("Начало генерации матрицы в оперативную память")
	}
	matrixInRam = make([][]float64, size)
	for i := 0; i < size; i++ {
		matrixInRam[i] = make([]float64, size+1)
		for j := 0; j < size+1; j++ {
			matrixInRam[i][j] = float64(rand.Intn(20) - 10)
		}
	}
	if withDebug {
		fmt.Println("Матрица сгенерирована")
	}
}

func generateMatrix(fileName string, size int) {
	if withDebug {
		fmt.Println("Начало генерации матрицы в файл")
	}
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
	if withDebug {
		fmt.Println("Матрица сгенерирована")
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
	if withDebug {
		fmt.Println("Начало закрузки матрицы")
	}
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
