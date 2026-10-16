package internalrpc

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const (
	defaultSSHHostKeyPath    = "./data/internalrpc/ssh_host_ed25519_key"
	defaultSSHKnownHostsPath = "./data/internalrpc/known_hosts"
)

// POSIX record locks are process-scoped, so also serialize callers in this process.
var sshStateGate = make(chan struct{}, 1)

type sshState struct {
	dir  *os.Root
	name string
}

func loadSSHHostSigner(path string) (ssh.Signer, error) {
	if path == "" {
		path = defaultSSHHostKeyPath
	}
	ctx, cancel := context.WithTimeout(context.Background(), authenticationTimeout)
	defer cancel()
	state, err := openSSHState(path)
	if err != nil {
		return nil, err
	}
	defer state.dir.Close()
	release, err := state.lock(ctx)
	if err != nil {
		return nil, err
	}
	defer release()

	file, info, err := state.open(state.name, os.O_RDONLY, true)
	if err == nil {
		data, readErr := io.ReadAll(file)
		closeErr := file.Close()
		if readErr != nil || closeErr != nil {
			clear(data)
			return nil, errors.New("read internal SSH host private key failed")
		}
		defer clear(data)
		if err := state.unchanged(state.name, info); err != nil {
			return nil, err
		}
		signer, err := ssh.ParsePrivateKey(data)
		if err != nil {
			return nil, errors.New("internal SSH host private key must be valid and unencrypted")
		}
		if _, err := nodeKeyAlgorithm(signer.PublicKey()); err != nil {
			return nil, errors.New("internal SSH host private key uses an unsupported algorithm")
		}
		return signer, ctx.Err()
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, errors.New("generate internal SSH host private key failed")
	}
	defer clear(privateKey)
	block, err := ssh.MarshalPrivateKey(privateKey, "")
	if err != nil {
		return nil, errors.New("encode internal SSH host private key failed")
	}
	defer clear(block.Bytes)
	data := pem.EncodeToMemory(block)
	defer clear(data)
	if err := state.replace(ctx, data, nil); err != nil {
		return nil, err
	}
	// The signer may retain the Ed25519 key slice, so parse its own copy before clearing it.
	return ssh.ParsePrivateKey(data)
}

func tofuHostKeyCallback(ctx context.Context, path string) ssh.HostKeyCallback {
	if path == "" {
		path = defaultSSHKnownHostsPath
	}
	return func(address string, _ net.Addr, key ssh.PublicKey) error {
		ctx, cancel := context.WithTimeout(ctx, authenticationTimeout)
		defer cancel()
		if err := ctx.Err(); err != nil {
			return err
		}
		host, port, err := net.SplitHostPort(address)
		if err != nil || host == "" || port == "" || strings.ContainsAny(address, "\x00 \t\r\n,*!?|\\#@") {
			return errors.New("invalid internal SSH gateway host:port")
		}
		state, err := openSSHState(path)
		if err != nil {
			return err
		}
		defer state.dir.Close()
		release, err := state.lock(ctx)
		if err != nil {
			return err
		}
		defer release()

		var data []byte
		file, info, err := state.open(state.name, os.O_RDWR, false)
		if err == nil {
			data, err = io.ReadAll(file)
			closeErr := file.Close()
			if err != nil || closeErr != nil {
				return errors.New("read internal SSH known_hosts failed")
			}
			// knownhosts owns the native parser and its revocation/certificate semantics.
			check, err := knownhosts.New(filepath.Join(state.dir.Name(), state.name))
			if err != nil {
				return errors.New("parse internal SSH known_hosts failed: invalid OpenSSH state")
			}
			if err := state.unchanged(state.name, info); err != nil {
				return err
			}
			err = check(address, sshKnownHostAddress(address), key)
			if err == nil {
				return ctx.Err()
			}
			var keyError *knownhosts.KeyError
			if !errors.As(err, &keyError) || len(keyError.Want) != 0 {
				return sshHostTrustError(address, key, err)
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}

		if len(data) != 0 && data[len(data)-1] != '\n' {
			data = append(data, '\n')
		}
		data = append(data, knownhosts.Line([]string{address}, key)...)
		data = append(data, '\n')
		return state.replace(ctx, data, info)
	}
}

type sshKnownHostAddress string

func (a sshKnownHostAddress) Network() string { return "tcp" }
func (a sshKnownHostAddress) String() string  { return string(a) }

func sshHostTrustError(address string, key ssh.PublicKey, err error) error {
	fingerprint := ssh.FingerprintSHA256(key)
	var keyError *knownhosts.KeyError
	if errors.As(err, &keyError) && len(keyError.Want) != 0 {
		trusted := make([]string, 0, len(keyError.Want))
		for _, want := range keyError.Want {
			trusted = append(trusted, ssh.FingerprintSHA256(want.Key))
		}
		return fmt.Errorf("internal SSH host key mismatch for %q: presented %s, trusted %s: %w", address, fingerprint, strings.Join(trusted, ", "), err)
	}
	if _, ok := errors.AsType[*knownhosts.RevokedError](err); ok {
		return fmt.Errorf("internal SSH host key revoked for %q (%s): %w", address, fingerprint, err)
	}
	return fmt.Errorf("internal SSH host verification failed for %q (%s)", address, fingerprint)
}

func openSSHState(path string) (*sshState, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve internal SSH state path: %w", err)
	}
	name := filepath.Base(absolute)
	if name == "." || name == string(filepath.Separator) || name == ".." {
		return nil, errors.New("internal SSH state path must name a file")
	}
	directory := filepath.Dir(absolute)
	volume := filepath.VolumeName(directory)
	root, err := os.OpenRoot(volume + string(filepath.Separator))
	if err != nil {
		return nil, fmt.Errorf("open internal SSH state filesystem: %w", err)
	}
	parts := strings.FieldsFunc(strings.TrimPrefix(directory, volume), func(r rune) bool {
		return r == rune(filepath.Separator)
	})
	for index, part := range parts {
		info, statErr := root.Lstat(part)
		if errors.Is(statErr, os.ErrNotExist) {
			if err := root.Mkdir(part, 0o700); err != nil && !errors.Is(err, os.ErrExist) {
				root.Close()
				return nil, fmt.Errorf("create internal SSH state directory: %w", err)
			}
			info, statErr = root.Lstat(part)
		}
		if statErr != nil {
			root.Close()
			return nil, fmt.Errorf("inspect internal SSH state directory: %w", statErr)
		}
		final := index == len(parts)-1
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || !safeSSHStateDirectory(info, final) {
			root.Close()
			return nil, errors.New("internal SSH state directory must be a secure non-symlink directory")
		}
		next, err := root.OpenRoot(part)
		if err != nil {
			root.Close()
			return nil, fmt.Errorf("open internal SSH state directory: %w", err)
		}
		opened, statErr := next.Stat(".")
		root.Close()
		if statErr != nil || !os.SameFile(info, opened) {
			next.Close()
			return nil, errors.New("internal SSH state directory changed while opening")
		}
		root = next
	}
	if len(parts) == 0 {
		root.Close()
		return nil, errors.New("internal SSH state requires a dedicated directory")
	}
	return &sshState{dir: root, name: name}, nil
}

func (s *sshState) open(name string, flags int, private bool) (*os.File, os.FileInfo, error) {
	info, err := s.dir.Lstat(name)
	if err != nil {
		return nil, nil, err
	}
	if !info.Mode().IsRegular() || !safeSSHStateFile(info, private, flags&os.O_RDWR != 0) {
		return nil, nil, errors.New("internal SSH state must be a secure regular non-symlink file")
	}
	file, err := s.dir.OpenFile(name, flags|sshStateOpenFlags(), 0)
	if err != nil {
		return nil, nil, fmt.Errorf("open internal SSH state: %w", err)
	}
	opened, err := file.Stat()
	if err != nil || !os.SameFile(info, opened) {
		file.Close()
		return nil, nil, errors.New("internal SSH state file changed while opening")
	}
	return file, opened, nil
}

func (s *sshState) unchanged(name string, info os.FileInfo) error {
	current, err := s.dir.Lstat(name)
	if info == nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
	} else if err == nil && os.SameFile(info, current) && current.Mode() == info.Mode() && current.Size() == info.Size() && current.ModTime() == info.ModTime() {
		return nil
	}
	return errors.New("internal SSH state changed during operation")
}

func (s *sshState) lock(ctx context.Context) (func(), error) {
	select {
	case sshStateGate <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	file, _, err := s.open(s.name+".lock", os.O_RDWR, true)
	if errors.Is(err, os.ErrNotExist) {
		file, err = s.dir.OpenFile(s.name+".lock", os.O_RDWR|os.O_CREATE|os.O_EXCL|sshStateOpenFlags(), 0o600)
		if errors.Is(err, os.ErrExist) {
			file, _, err = s.open(s.name+".lock", os.O_RDWR, true)
		}
	}
	if err != nil {
		<-sshStateGate
		return nil, fmt.Errorf("open internal SSH state lock: %w", err)
	}
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err = ctx.Err(); err != nil {
			break
		}
		var locked bool
		locked, err = trySSHStateLock(file)
		if err != nil {
			break
		}
		if locked {
			info, statErr := file.Stat()
			if statErr == nil {
				statErr = s.unchanged(s.name+".lock", info)
			}
			if statErr != nil {
				err = statErr
				break
			}
			return func() {
				file.Close()
				<-sshStateGate
			}, nil
		}
		select {
		case <-ctx.Done():
			err = ctx.Err()
		case <-ticker.C:
			continue
		}
		break
	}
	file.Close()
	<-sshStateGate
	return nil, fmt.Errorf("lock internal SSH state: %w", err)
}

func (s *sshState) replace(ctx context.Context, data []byte, previous os.FileInfo) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return errors.New("create internal SSH state temporary name failed")
	}
	name := "." + s.name + ".tmp-" + hex.EncodeToString(nonce[:])
	file, err := s.dir.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL|sshStateOpenFlags(), 0o600)
	if err != nil {
		return fmt.Errorf("create internal SSH state temporary file: %w", err)
	}
	defer s.dir.Remove(name)
	written, writeErr := file.Write(data)
	if writeErr == nil && written != len(data) {
		writeErr = io.ErrShortWrite
	}
	if writeErr == nil {
		writeErr = file.Sync()
	}
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		return errors.New("persist internal SSH state temporary file failed")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := s.unchanged(s.name, previous); err != nil {
		return err
	}
	if err := s.dir.Rename(name, s.name); err != nil {
		return fmt.Errorf("publish internal SSH state: %w", err)
	}
	return syncSSHStateDirectory(s.dir)
}
