package utils

import (
	"crypto/tls"
	"os"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func LoadCertificate(c, k string) (*tls.Certificate, error) {
	var (
		cert tls.Certificate
		err  error
	)

	if FileExists(c) && FileExists(k) {
		// act them as files in the system.
		cert, err = tls.LoadX509KeyPair(c, k)
	} else {
		// act them as raw contents.
		cert, err = tls.X509KeyPair([]byte(c), []byte(k))
	}

	if err != nil {
		return nil, err
	}

	return &cert, nil
}
