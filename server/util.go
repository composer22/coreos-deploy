package server

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	mr "math/rand"
	"os/exec"
	"time"
)

// createV4UUID returns a V4 RFC4122 compliant UUID.
func createV4UUID() string {
	u := make([]byte, 16)
	rand.Read(u)
	// 13th char must be 4 and 17th must be in [89AB]
	u[8] = (u[8] | 0x80) & 0xBF
	u[6] = (u[6] | 0x40) & 0x4F
	return fmt.Sprintf("%X-%X-%X-%X-%X", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
}

// randomString returns a random string for n characters.
func randomString(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	mr.Seed(time.Now().UTC().UnixNano())
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		result[i] = chars[mr.Intn(len(chars))]
	}
	return string(result)
}

// execCmd executes an os command and formats any output from stdout/err
func execCmd(cmd *exec.Cmd) (string, error) {
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	result := stdout.String()
	if err := stderr.String(); err != "" {
		return "", errors.New(err)
	}
	return result, err
}
