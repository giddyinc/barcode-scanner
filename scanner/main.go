package main

import (
	"log"

	"github.com/giddyinc/barcode-scanner"
	"github.com/truveris/gousb/usb"
)

func main() {
	ctx := usb.NewContext()
	defer ctx.Close()

	scanners, err := scale.GetScanners(ctx)
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
