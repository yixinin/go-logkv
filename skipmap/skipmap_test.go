package skipmap

import (
	"fmt"
	"testing"
)

func TestSkipMap(t *testing.T) {
	var m = New()

	m.Del("3")
	m.Set("3", "3")
	m.Del("3")

	m.Set("01", "1")
	m.Set("02", "2")
	m.Del("03")
	m.Del("01")
	m.Del("02")
	m.Set("01", "1")
	m.Set("02", "2")
	m.Set("04", "4")
	m.Set("03", "3")
	m.Set("10", "10")
	m.Set("08", "8")
	m.Set("08", "88")
	fmt.Println(m.Get("01"))
	fmt.Println(m.Get("10"))
	fmt.Println(m.Get("06"))
}
