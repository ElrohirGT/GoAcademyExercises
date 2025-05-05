package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	newContext, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Infinite loop!
	// select {
	// case <-newContext.Done():
	// 	fmt.Println("DONE!")
	// 	return
	// default:
	// 	for {
	// 		fmt.Println("Waiting...")
	// 	}
	// }

	// No infinite loop
	// Only prints Waiting 2 times...
	// for {
	// 	select {
	// 	case <-newContext.Done():
	// 		fmt.Println("DONE!")
	// 		return
	// 	case <-time.After(time.Second):
	// 		fmt.Println("Waiting...")
	// 	}
	// }

	// No infinite loop
	// Prints Waiting 3 times!
	for {
		select {
		case <-newContext.Done():
			fmt.Println("DONE!")
			return
		default:
			time.Sleep(time.Second)
			fmt.Println("Waiting...")
		}
	}

}
