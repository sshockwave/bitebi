package utils

import (
	"math/rand"
	"strconv"
	"time"
)

func RandomInit() {
    rand.Seed(time.Now().UnixNano())
}

var randomNames = []string{
	"Alice",
	"Bob",
	"Charlie",
	"David",
	"Eve",
	"Frank",
	"Grace",
	"Hailey",
	"Ivan",
	"Judy",
	"Mike",
	"Nancy",
	"Olivia",
};
func RandomName() string {
	ret := randomNames[rand.Intn(len(randomNames))]
	for i := 0; i < 3; i++ {
		ret += strconv.Itoa(rand.Intn(10))
	}
	return ret
}
