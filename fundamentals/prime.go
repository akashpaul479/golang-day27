package fundamentals

import "fmt"

func isprime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true

}
func Prime() {
	var num int
	for {
		fmt.Println("Enter the number (-1 to exit):")
		fmt.Scan(&num)

		if num == -1 {
			fmt.Println("program exicted by user!")
			break
		}

		if isprime(num) {
			fmt.Println(num, "is a prime number")
		} else {
			fmt.Println(num, "is not a prime number")
		}
	}
}
