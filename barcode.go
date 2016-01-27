package barcode

import (
	"errors"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/giddyinc/gousb/usb"
)

const (
	BufferLength = 8

	UsbConfig   = uint8(1)
	UsbIface    = uint8(0)
	UsbSetup    = uint8(0)
	UsbEndpoint = uint8(3)

	SleepDuration = 500 * time.Millisecond

	ErrorIgnore = "libusb: timeout [code -7]"
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

type Scanner struct {
	*usb.Device
}

// CRead deciphers the barcode and pipe it to a channel
func (sc *Scanner) CRead(c chan string) {
	data := make([]byte, BufferLength)
	out := []string{}
	endpoint, err := sc.OpenEndpoint(UsbConfig, UsbIface, UsbSetup,
		UsbEndpoint|uint8(usb.ENDPOINT_DIR_IN))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("barcode scanner is ready")
	for {
		_, err := endpoint.Read(data)
		if err != nil {
			log.Println(err)
			time.Sleep(SleepDuration)
			continue
		}

		d, err := ParseBuffer(data)
		if err != nil {
			log.Println(err)
			time.Sleep(SleepDuration)
			continue
		}
		if d != TerminatorStr && d != ShiftKeyStr {
			out = append(out, d)
		}
		if d == TerminatorStr && len(out) > 0 {
			c <- strings.Join(out, "")
			out = []string{}
		}
		time.Sleep(SleepDuration)
	}
}

// Read reads buffer via usb port
func (sc *Scanner) Read() ([]string, error) {
	data := make([]byte, BufferLength)
	out := []string{}
	endpoint, err := sc.OpenEndpoint(UsbConfig, UsbIface, UsbSetup,
		UsbEndpoint|uint8(usb.ENDPOINT_DIR_IN))
	if err != nil {
		return out, err
	}
	log.Println("barcode scanner is ready")
	dataLen, err := endpoint.Read(data)
	if err != nil {
		if err.Error() != ErrorIgnore {
			log.Println(err)
		}
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
			time.Sleep(SleepDuration)
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
func GetScanners(ctx *usb.Context, v usb.ID, p usb.ID) ([]*Scanner, error) {
	var scanners []*Scanner
	devices, err := ctx.ListDevices(func(desc *usb.Descriptor) bool {
		return desc.Vendor == v && desc.Product == p
	})

	if err != nil {
		return scanners, err
	}

	if len(devices) == 0 {
		return scanners, ErrorDeviceNotFound
	}

	for _, dev := range devices {
		if runtime.GOOS == "linux" {
			dev.DetachKernelDriver(0)
		}

		sc := &Scanner{
			dev,
		}
		scanners = append(scanners, sc)
	}

	return scanners, nil
}
