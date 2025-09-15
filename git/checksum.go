package git

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/anknetau/orto/fp"
)

type Hasher struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	br     *bufio.Reader
	mu     sync.Mutex // serialize write->read pairs
	stdout io.ReadCloser
}

func RunGetRepoHashFormat(pathToGitBinary string) fp.Algo {
	out, err := runToString(pathToGitBinary, "rev-parse", "--show-object-format")
	if err != nil {
		log.Fatal(err)
	}
	return fp.AlgoOfGitObjectFormat(strings.TrimSpace(out))
}

type WorktreeStatus int

const (
	WorktreeStatusTrue WorktreeStatus = iota
	WorktreeStatusFalse
	WorktreeStatusNotARepo
)

func (env Env) RunGetIsInsideWorktree() WorktreeStatus {
	out, err := runToString(env.PathToBinary, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 128 &&
			strings.Contains(string(exitErr.Stderr), "not a git repository") {
			return WorktreeStatusNotARepo
		} else {
			log.Fatal(err)
		}
	}
	if strings.TrimSpace(out) == "true" {
		return WorktreeStatusTrue
	} else if strings.TrimSpace(out) == "false" {
		return WorktreeStatusFalse
	} else {
		log.Fatal("Could not parse git output: " + out)
		return WorktreeStatusNotARepo
	}
}

func (env Env) RunGetRepoRoot() string {
	// TODO: this returns a path that could not support spaces, escaping and other issues.
	out, err := runToString(env.PathToBinary, "rev-parse", "--show-toplevel")
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(out)
}

func (env Env) RunGetGitDir() string {
	// TODO: this returns a path that could not support spaces, escaping and other issues.
	out, err := runToString(env.PathToBinary, "rev-parse", "--absolute-git-dir")
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(out)
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
	return strings.TrimSpace(oid), nil
}

// Close closes the hasher
func (h *Hasher) Close() error {
	_ = h.stdin.Close()
	return h.cmd.Wait()
}
