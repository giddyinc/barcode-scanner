package scanner

import (
	"errors"
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/giddyinc/gousb/usb"
)

const (
	BufferLength = 8
)

var (
	ErrorDeviceNotFound       = errors.New("Device not present")
	ErrorDeviceReadIncomplete = errors.New(
		"Data read from endpoint is not complete")
	ErrorBufferLength = errors.New(fmt.Sprintf("Buffer should have %d bytes", BufferLength))

	// key mapping
	// Ref: http://www.usb.org/developers/hidpage/Hut1_12v2.pdf chapter 10
	Keys = []string{
		"", "", "", "", "a", "b", "c", "d", "e", "f",
		"g", "h", "i", "j", "k", "l", "m", "n", "o", "p",
		"q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
		"1", "2", "3", "4", "5", "6", "7", "8", "9", "0",
		"ENTER", "ESCAPE", "DELETE", "TAB", " ", "-", "=", "[", "]", "\\",
		"", ";", "'", "~", ",", ".", "/",
	}

	UpperKeys = []string{
		" ", " ", " ", " ", "A", "B", "C", "D", "E", "F",
		"G", "H", "I", "J", "K", "L", "M", "N", "O", "P",
		"Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
		"!", "@", "#", "$", "%", "^", "&", "*", "(", ")",
		"ENTER", "ESCAPE", "DELETE", "TAB", " ", "_", "+", "{", "}", "|",
		"", ":", "\"", "~", "<", ">", "?",
	}

	TerminatorStr = "ENTER"
	ShiftKeyStr   = "SHIFT"

	Terminator = []byte{0, 0, 40, 0, 0, 0, 0, 0}
	ShiftKey   = []byte{2, 0, 0, 0, 0, 0, 0, 0}
)

type UsbConfig struct {
	Vendor   usb.ID
	Product  usb.ID
	Config   uint8
	Iface    uint8
	Setup    uint8
	Endpoint uint8
}

type Scanner struct {
	*usb.Device
	Config UsbConfig
}

// CRead deciphers the barcode and pipe it to a channel
func (sc *Scanner) CRead(c chan string) {
	data := make([]byte, BufferLength)
	out := []string{}
	endpoint, err := sc.OpenEndpoint(sc.Config.Config,
		sc.Config.Iface,
		sc.Config.Setup,
		sc.Config.Endpoint|uint8(usb.ENDPOINT_DIR_IN))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("barcode scanner is ready")
	for {
		_, err := endpoint.Read(data)
		if err != nil {
			log.Println(err)
			continue
		}

		d, err := ParseBuffer(data)
		if err != nil {
			log.Println(err)
			continue
		}
		if d != TerminatorStr && d != ShiftKeyStr {
			out = append(out, d)
		}
		if d == TerminatorStr && len(out) > 0 {
			c <- strings.Join(out, "")
			out = []string{}
		}
	}
}

// Read reads buffer via usb port
func (sc *Scanner) Read() ([]string, error) {
	data := make([]byte, BufferLength)
	out := []string{}
	endpoint, err := sc.OpenEndpoint(
		sc.Config.Config,
		sc.Config.Iface,
		sc.Config.Setup,
		sc.Config.Endpoint|uint8(usb.ENDPOINT_DIR_IN))
	if err != nil {
		return out, err
	}
	log.Println("barcode scanner is ready")
	dataLen, err := endpoint.Read(data)
	if err != nil {
		log.Println(err)
	}
	if dataLen != BufferLength {
		return out, ErrorDeviceReadIncomplete
	}

	d, err := ParseBuffer(data)
	if err != nil {
		return out, err
	}
	if d != TerminatorStr && d != ShiftKeyStr {
		out = append(out, d)
	}
	for {
		dataLen, err = endpoint.Read(data)
		if err != nil {
			log.Println(err)
			continue
		}
		d, err = ParseBuffer(data)
		if d != TerminatorStr && d != ShiftKeyStr {
			out = append(out, d)
		}
		if d == TerminatorStr {
			break
		}
	}

	return out, nil
}

// assume they have the same length
func sameSlice(s1 []byte, s2 []byte) bool {
	for i, b := range s1 {
		if b != s2[i] {
			return false
		}
	}
	return true
}

// isTerminator checks if a buffer is a terminator
func isTerminator(buf []byte) bool {
	return sameSlice(buf, Terminator)
}

// isShift checks if a buffer is a shift key
func isShift(buf []byte) bool {
	return sameSlice(buf, ShiftKey)
}

// ParseBuffer parses a 8-byte buffer
func ParseBuffer(buf []byte) (string, error) {
	if len(buf) != BufferLength {
		return "", ErrorBufferLength
	}
	if isTerminator(buf) {
		return TerminatorStr, nil
	} else if isShift(buf) {
		return ShiftKeyStr, nil
	}
	isShift := buf[0] == 2
	key := int(buf[2])
	if key > len(Keys)-1 {
		return "", errors.New(fmt.Sprintf("Unexpected key %d", key))
	}
	if isShift {
		return UpperKeys[key], nil
	} else {
		return Keys[key], nil
	}

}

// GetScanners scans all usb ports to get all scanners
// To omit product ID, set prod to 0.
func GetScanners(ctx *usb.Context, config UsbConfig) ([]*Scanner, error) {
	var scanners []*Scanner
	devices, err := ctx.ListDevices(func(desc *usb.Descriptor) bool {
		var selected = desc.Vendor == config.Vendor
		if config.Product != usb.ID(0) {
			selected = selected && desc.Product == config.Product
		}
		return selected
	})

	if err != nil {
		return scanners, err
	}

	if len(devices) == 0 {
		return scanners, ErrorDeviceNotFound
	}

getDevice:
	for _, dev := range devices {
		if runtime.GOOS == "linux" {
			dev.DetachKernelDriver(0)
		}

		// get devices with IN direction on endpoint
		for _, cfg := range dev.Descriptor.Configs {
			for _, alt := range cfg.Interfaces {
				for _, iface := range alt.Setups {
					for _, end := range iface.Endpoints {
						if end.Direction() == usb.ENDPOINT_DIR_IN {
							config.Config = cfg.Config
							config.Iface = alt.Number
							config.Setup = iface.Number
							config.Endpoint = uint8(end.Number())
							sc := &Scanner{
								dev,
								config,
							}
							// don't timeout reading
							sc.ReadTimeout = 0
							scanners = append(scanners, sc)
							continue getDevice
						}
					}
				}
			}
		}

	}

	return scanners, nil
}
