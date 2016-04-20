package main

import (
	"log"

	"github.com/giddyinc/barcode-scanner"
	"github.com/giddyinc/gousb/usb"
)

const (
	Vendor  = 0x0c2e
	Product = 0x00
)

func main() {
	ctx := usb.NewContext()
	defer ctx.Close()
	config := scanner.UsbConfig{
		Vendor:  Vendor,
		Product: Product,
	}
	scanners, err := scanner.GetScanners(ctx, config)
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
