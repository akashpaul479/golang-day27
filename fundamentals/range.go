package fundamentals

import "fmt"

func Range() {
	numbers := make([]int, 100)

	for i := range numbers {
		fmt.Println(i + 1)
	}
}
