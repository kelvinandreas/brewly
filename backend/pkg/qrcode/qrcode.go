// Package qrcode provides a thin wrapper around go-qrcode for generating PNG QR codes.
package qrcode

import (
	"fmt"

	goqr "github.com/skip2/go-qrcode"
)

// Generate encodes content as a 256×256 PNG QR code and returns the raw bytes.
func Generate(content string) ([]byte, error) {
	png, err := goqr.Encode(content, goqr.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("qrcode.Generate: %w", err)
	}
	return png, nil
}
