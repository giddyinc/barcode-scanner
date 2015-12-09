package scale

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/truveris/gousb/usb"
)

const (
	BUFFER_LENGTH = 8

	USB_CONFIG   = uint8(1)
	USB_IFACE    = uint8(0)
	USB_SETUP    = uint8(0)
	USB_ENDPOINT = uint8(3)
)

var (
	ERROR_DEVICE_NOT_FOUND       = errors.New("Device not present")
	ERROR_DEVICE_READ_INCOMPLETE = errors.New(
		"Data read from endpoint is not complete")
	ERROR_BUFFER_LENGTH = errors.New(fmt.Sprintf("Buffer should have %d bytes", BUFFER_LENGTH))

	// key mapping
	KEYS = []string{
		" ", " ", " ", " ",
		"a", "b", "c", "d", "e", "f", "g", "h", "i",
		"j", "k", "l", "m", "n", "o", "p", "q", "r",
		"s", "t", "u", "v", "w", "x", "y", "z",
		"1", "2", "3", "4", "5", "6", "7", "8", "9", "0",
		" ", "-", "=", "[", "]", "\\", ";", "'", "~",
		",", ".", "/",
	}

	UPPER_KEYS = []string{
		" ", " ", " ", " ",
		"A", "B", "C", "D", "E", "F", "G", "H", "I",
		"J", "K", "L", "M", "N", "O", "P", "Q", "R",
		"S", "T", "U", "V", "W", "X", "Y", "Z",
		"!", "@", "#", "$", "%", "^", "&", "*", "(", ")",
		" ", "_", "+", "{", "}", "|", ":", "\"", "~",
		"<", ">", "?",
	}

	TERMINATOR_STR = "尾"
	SHIFT_KEY_STR  = "换"

	TERMINATOR = []byte{0, 0, 0, 0, 0, 0, 0, 0}
	SHIFT_KEY  = []byte{2, 0, 0, 0, 0, 0, 0, 0}
)

type Scanner struct {
	*usb.Device
}

// Read: read buffer via usb port
func (sc *Scanner) Read() ([]string, error) {
	data := make([]byte, BUFFER_LENGTH)
	out := []string{}
	endpoint, err := sc.OpenEndpoint(USB_CONFIG, USB_IFACE, USB_SETUP,
		USB_ENDPOINT|uint8(usb.ENDPOINT_DIR_IN))
	if err != nil {
		return out, err
	}

	dataLen, err := endpoint.Read(data)
	if err != nil {
		return out, err
	}
	if dataLen != BUFFER_LENGTH {
		return out, ERROR_DEVICE_READ_INCOMPLETE
	}

	d, err := ParseBuffer(data)
	if err != nil {
		return out, err
	}
	if d != TERMINATOR_STR && d != SHIFT_KEY_STR {
		out = append(out, d)
	}
	for true {
		dataLen, err = endpoint.Read(data)
		if err != nil {
			break
		}
		d, err = ParseBuffer(data)
		if d != TERMINATOR_STR && d != SHIFT_KEY_STR {
			out = append(out, d)
		}
	}

	return out, nil
}

// isTerminator checks if a buffer is a terminator
func isTerminator(buf []byte) bool {
	for _, b := range buf {
		if b != 0 {
			return false
		}
	}
	return true
}

// isShift checks if a buffer is a shift key
func isShift(buf []byte) bool {
	for i, b := range buf {
		if i == 0 && b != 2 {
			return false
		}
		if i > 0 && b != 0 {
			return false
		}
	}
	return true
}

// ParseBuffer parses a 8-byte buffer
func ParseBuffer(buf []byte) (string, error) {
	if len(buf) != BUFFER_LENGTH {
		return "", ERROR_BUFFER_LENGTH
	}
	if isTerminator(buf) {
		return TERMINATOR_STR, nil
	} else if isShift(buf) {
		return SHIFT_KEY_STR, nil
	}
	isShift := buf[0] == 2
	key := int(buf[2])
	if key > len(KEYS)-1 {
		return "", errors.New(fmt.Sprintf("Unexpected key %d", key))
	}
	if isShift {
		return UPPER_KEYS[key], nil
	} else {
		return KEYS[key], nil
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
		return scanners, ERROR_DEVICE_NOT_FOUND
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
