package main

import "log"

func main() {
	if err := nvidia(); err != nil {
		log.Fatal(err)
	}
}
