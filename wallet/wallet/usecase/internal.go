package usecase

import (
	"fmt"
	"log"
	"strconv"
)

// generateAccountNo generates new account based on an increment on last account in DB
func generateAccountNo(prev string) (string, bool) {
	num, err := strconv.Atoi(prev)
	if err != nil {
		return "", false
	}
	num++
	if len(strconv.Itoa(num)) <= 10 {
		return fmt.Sprintf("KZT%010d", num), true
	}
	log.Println("Limit exceeded")
	return "", false
}
