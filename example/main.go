package main

import (
	"log"

	"github.com/giddyinc/barcode-scanner"
	"github.com/truveris/gousb/usb"
)

const (
	Vendor  = 0x0536
	Product = 0x0461
)

func main() {
	ctx := usb.NewContext()
	defer ctx.Close()

	scanners, err := barcode.GetScanners(ctx, Vendor, Product)
	if err != nil {
		log.Fatal(err)
	}
	for _, sc := range scanners {
		defer sc.Close()
	}
	sc := scanners[0]

	chn := make(chan string)
	go sc.CRead(chn)

	for {
		select {
		case bar := <-chn:
			log.Println(bar)
		}
	}
}
