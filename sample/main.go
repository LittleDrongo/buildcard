package main

import (
	"time"

	"github.com/LittleDrongo/buildcard"
)

func main() {
	info := buildcard.Snapshot()
	info.Preview()

	time.Sleep(time.Second * 10)

	info = buildcard.Snapshot()
	info.Preview()
}
