package util

import "log"

func Check(e error) {
	if e != nil {
		log.Fatal(e)
		panic(e)
	}
}
