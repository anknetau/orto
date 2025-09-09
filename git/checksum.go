package git

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// TODO: rename "gitCommand" to "pathToGitBinary" everywhere

type Hasher struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	br     *bufio.Reader
	mu     sync.Mutex // serialize write->read pairs
	stdout io.ReadCloser
}

// NewHasher launches a long-lived git hash-object process.
// Don't forget to call Close() when done!
func NewHasher(pathToGitBinary string) (*Hasher, error) {
	cmd := exec.Command(pathToGitBinary, "hash-object", "--stdin-paths")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		return nil, err
	}

	return &Hasher{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		br:     bufio.NewReader(stdout),
	}, nil
}

// Hash sends one path and returns the corresponding hex OID string.
func (h *Hasher) Hash(path string) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if os.PathSeparator == '\\' {
		path = strings.ReplaceAll(path, "\\", "/")
	}

	// TODO: check if this uses CRLF in windows after each line

	_, err := fmt.Fprintf(h.stdin, "%s\n", path)
	if err != nil {
		return "", err
	}
	oid, err := h.br.ReadString('\n')
	if err != nil {
		return "", err
	}
	return oid, nil
}

// Close closes the hasher
func (h *Hasher) Close() error {
	_ = h.stdin.Close()
	return h.cmd.Wait()
}
