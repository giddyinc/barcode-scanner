package main

import (
	"log"

	"github.com/giddyinc/barcode-scanner"
	"github.com/truveris/gousb/usb"
)

const (
	VENDOR  = 0x0536
	PRODUCT = 0x0461
)

func main() {

	ctx := usb.NewContext()
	defer ctx.Close()

	scanners, err := scale.GetScanners(ctx, VENDOR, PRODUCT)
	if err != nil {
		log.Fatal(err)
	}
	for _, sc := range scanners {
		defer sc.Close()
	}
	sc := scanners[0]
	data, err := sc.Read()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(data)
}
