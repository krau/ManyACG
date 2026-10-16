package internalrpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

const (
	sshRPCUser             = "manyacg"
	sshRPCChannel          = "manyacg-rpc@manyacg"
	sshPublicKeyPermission = "manyacg-public-key"
	sshNodeIDPermission    = "manyacg-node-id"
)

type sshTransportCredentials struct {
	server             bool
	hostSigner         ssh.Signer
	authorizedKeysFile string
	address            string
	knownHostsPath     string
	signer             ssh.Signer
}

func newSSHServerCredentials(hostKeyPath, authorizedKeysFile string) (credentials.TransportCredentials, error) {
	if _, err := readAuthorizedKeys(authorizedKeysFile); err != nil {
		return nil, err
	}
	signer, err := loadSSHHostSigner(hostKeyPath)
	if err != nil {
		return nil, err
	}
	return &sshTransportCredentials{
		server:             true,
		hostSigner:         signer,
		authorizedKeysFile: authorizedKeysFile,
	}, nil
}

func newSSHClientCredentials(address, knownHostsPath string, signer ssh.Signer) credentials.TransportCredentials {
	return &sshTransportCredentials{address: address, knownHostsPath: knownHostsPath, signer: signer}
}

type sshAuthInfo struct {
	credentials.CommonAuthInfo
	publicKey []byte
	nodeID    string
}

func (sshAuthInfo) AuthType() string { return "ssh" }

func sshAuthenticatedKey(ctx context.Context) (publicKey []byte, nodeID string, ok bool) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, "", false
	}
	auth, ok := p.AuthInfo.(sshAuthInfo)
	if !ok || len(auth.publicKey) == 0 || auth.nodeID == "" {
		return nil, "", false
	}
	return auth.publicKey, auth.nodeID, true
}

func (*sshTransportCredentials) Info() credentials.ProtocolInfo {
	return credentials.ProtocolInfo{SecurityProtocol: "ssh", SecurityVersion: "2.0"}
}

func (c *sshTransportCredentials) Clone() credentials.TransportCredentials {
	clone := *c
	return &clone
}

func (*sshTransportCredentials) OverrideServerName(string) error {
	return errors.New("SSH host identity is fixed by the configured gateway address")
}

func (c *sshTransportCredentials) ClientHandshake(ctx context.Context, _ string, raw net.Conn) (net.Conn, credentials.AuthInfo, error) {
	if c.server {
		raw.Close()
		return nil, nil, errors.New("server SSH credentials cannot authenticate a client connection")
	}
	guard, err := startSSHHandshake(ctx, raw)
	if err != nil {
		return nil, nil, err
	}
	defer guard.stop()
	var conn *sshChannelConn
	complete := false
	defer func() {
		if !complete {
			if conn != nil {
				conn.Close()
			} else {
				raw.Close()
			}
		}
	}()
	if c.address == "" || c.signer == nil {
		return nil, nil, errors.New("SSH gateway address and node signer are required")
	}
	algorithm, err := nodeKeyAlgorithm(c.signer.PublicKey())
	if err != nil {
		return nil, nil, err
	}
	signer := c.signer
	if signer.PublicKey().Type() == ssh.KeyAlgoRSA {
		algorithmSigner, ok := signer.(ssh.AlgorithmSigner)
		if !ok {
			return nil, nil, errors.New("RSA node signer must support RSA-SHA512")
		}
		signer, err = ssh.NewSignerWithAlgorithms(algorithmSigner, []string{algorithm})
		if err != nil {
			return nil, nil, fmt.Errorf("restrict SSH node signature algorithm: %w", err)
		}
	}
	hostContext, cancelHost := context.WithCancel(context.WithoutCancel(ctx))
	stopHostCancellation := context.AfterFunc(guard.ctx, cancelHost)
	defer func() {
		stopHostCancellation()
		if !complete {
			cancelHost()
		}
	}()
	algorithms := ssh.SupportedAlgorithms()
	config := &ssh.ClientConfig{
		Config: ssh.Config{
			KeyExchanges: algorithms.KeyExchanges,
			Ciphers:      algorithms.Ciphers,
			MACs:         algorithms.MACs,
		},
		User:              sshRPCUser,
		Auth:              []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback:   tofuHostKeyCallback(hostContext, c.knownHostsPath),
		HostKeyAlgorithms: algorithms.HostKeys,
	}
	sshConn, channels, requests, err := ssh.NewClientConn(raw, c.address, config)
	if err != nil {
		return nil, nil, fmt.Errorf("SSH client handshake: %w", err)
	}
	conn = &sshChannelConn{raw: raw, done: make(chan struct{}), cancelHost: cancelHost}
	conn.rejectRequests(requests)
	conn.rejectChannels(channels)
	result := conn.openRPCChannel(sshConn)
	select {
	case opened := <-result:
		if opened.err != nil {
			return nil, nil, fmt.Errorf("open SSH RPC channel: %w", opened.err)
		}
		conn.channel = opened.channel
	case <-guard.ctx.Done():
		return nil, nil, guard.ctx.Err()
	}
	if !stopHostCancellation() {
		err := guard.ctx.Err()
		if err == nil {
			err = context.Canceled
		}
		return nil, nil, err
	}
	if err := guard.finish(); err != nil {
		return nil, nil, err
	}
	complete = true
	return conn, sshAuthInfo{CommonAuthInfo: credentials.CommonAuthInfo{SecurityLevel: credentials.PrivacyAndIntegrity}}, nil
}

func (c *sshTransportCredentials) ServerHandshake(raw net.Conn) (net.Conn, credentials.AuthInfo, error) {
	if !c.server {
		raw.Close()
		return nil, nil, errors.New("client SSH credentials cannot authenticate a server connection")
	}
	guard, err := startSSHHandshake(context.Background(), raw)
	if err != nil {
		return nil, nil, err
	}
	defer guard.stop()
	var conn *sshChannelConn
	complete := false
	defer func() {
		if !complete {
			if conn != nil {
				conn.Close()
			} else {
				raw.Close()
			}
		}
	}()
	algorithms := ssh.SupportedAlgorithms()
	var userAlgorithms []string
	for _, algorithm := range algorithms.PublicKeyAuths {
		switch algorithm {
		case ssh.KeyAlgoED25519, ssh.KeyAlgoECDSA256, ssh.KeyAlgoECDSA384, ssh.KeyAlgoECDSA521, ssh.KeyAlgoRSASHA512:
			userAlgorithms = append(userAlgorithms, algorithm)
		}
	}
	config := &ssh.ServerConfig{
		Config: ssh.Config{
			KeyExchanges: algorithms.KeyExchanges,
			Ciphers:      algorithms.Ciphers,
			MACs:         algorithms.MACs,
		},
		PublicKeyAuthAlgorithms: userAlgorithms,
		PublicKeyCallback: func(metadata ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if metadata.User() != sshRPCUser {
				return nil, errors.New("invalid SSH RPC user")
			}
			if _, err := nodeKeyAlgorithm(key); err != nil {
				return nil, err
			}
			keys, err := readAuthorizedKeys(c.authorizedKeysFile)
			if err != nil {
				return nil, err
			}
			wire := string(key.Marshal())
			if keys[wire] == nil {
				return nil, errors.New("SSH node public key is not authorized")
			}
			return &ssh.Permissions{Extensions: map[string]string{
				sshPublicKeyPermission: wire,
				sshNodeIDPermission:    ssh.FingerprintSHA256(key),
			}}, nil
		},
		VerifiedPublicKeyCallback: func(_ ssh.ConnMetadata, key ssh.PublicKey, permissions *ssh.Permissions, algorithm string) (*ssh.Permissions, error) {
			expected, err := nodeKeyAlgorithm(key)
			if err != nil || algorithm != expected {
				return nil, errors.New("invalid SSH node signature algorithm")
			}
			return permissions, nil
		},
	}
	config.AddHostKey(c.hostSigner)
	sshConn, channels, requests, err := ssh.NewServerConn(raw, config)
	if err != nil {
		return nil, nil, fmt.Errorf("SSH server handshake: %w", err)
	}
	conn = &sshChannelConn{raw: raw, done: make(chan struct{})}
	conn.rejectRequests(requests)
	result := conn.acceptRPCChannel(channels)
	select {
	case accepted := <-result:
		if accepted.err != nil {
			return nil, nil, accepted.err
		}
		conn.channel = accepted.channel
	case <-guard.ctx.Done():
		return nil, nil, guard.ctx.Err()
	}
	if sshConn.Permissions == nil {
		return nil, nil, errors.New("missing authenticated SSH node identity")
	}
	extensions := sshConn.Permissions.Extensions
	auth := sshAuthInfo{
		CommonAuthInfo: credentials.CommonAuthInfo{SecurityLevel: credentials.PrivacyAndIntegrity},
		publicKey:      []byte(extensions[sshPublicKeyPermission]),
		nodeID:         extensions[sshNodeIDPermission],
	}
	if len(auth.publicKey) == 0 || auth.nodeID == "" {
		return nil, nil, errors.New("missing authenticated SSH node identity")
	}
	if err := guard.finish(); err != nil {
		return nil, nil, err
	}
	complete = true
	return conn, auth, nil
}

type sshHandshakeGuard struct {
	ctx      context.Context
	cancel   context.CancelFunc
	raw      net.Conn
	stopOnce sync.Once
	stopFunc func() bool
	stopped  chan struct{}
}

func startSSHHandshake(ctx context.Context, raw net.Conn) (*sshHandshakeGuard, error) {
	ctx, cancel := context.WithTimeout(ctx, authenticationTimeout)
	guard := &sshHandshakeGuard{ctx: ctx, cancel: cancel, raw: raw, stopped: make(chan struct{})}
	deadline, _ := ctx.Deadline()
	if err := raw.SetDeadline(deadline); err != nil {
		cancel()
		raw.Close()
		return nil, fmt.Errorf("set SSH handshake deadline: %w", err)
	}
	guard.stopFunc = context.AfterFunc(ctx, func() {
		raw.Close()
		close(guard.stopped)
	})
	return guard, nil
}

func (g *sshHandshakeGuard) stopCancellation() {
	g.stopOnce.Do(func() {
		if !g.stopFunc() {
			<-g.stopped
		}
	})
}

func (g *sshHandshakeGuard) stop() {
	g.stopCancellation()
	g.cancel()
}

func (g *sshHandshakeGuard) finish() error {
	g.stopCancellation()
	if err := g.ctx.Err(); err != nil {
		return err
	}
	return g.raw.SetDeadline(time.Time{})
}

type sshChannelResult struct {
	channel ssh.Channel
	err     error
}

type sshChannelConn struct {
	raw        net.Conn
	channel    ssh.Channel
	done       chan struct{}
	closeOnce  sync.Once
	closeErr   error
	cancelHost context.CancelFunc

	deadlineMu    sync.Mutex
	readDeadline  time.Time
	writeDeadline time.Time
	writeTimer    *time.Timer
	closed        bool
}

func (c *sshChannelConn) rejectRequests(requests <-chan *ssh.Request) {
	go func() {
		for {
			select {
			case <-c.done:
				return
			case request, ok := <-requests:
				if !ok {
					return
				}
				request.Reply(false, nil)
			}
		}
	}()
}

func (c *sshChannelConn) rejectChannels(channels <-chan ssh.NewChannel) {
	go func() {
		for {
			select {
			case <-c.done:
				return
			case channel, ok := <-channels:
				if !ok {
					return
				}
				channel.Reject(ssh.Prohibited, "SSH RPC transport does not accept incoming channels")
			}
		}
	}()
}

func (c *sshChannelConn) openRPCChannel(conn ssh.Conn) <-chan sshChannelResult {
	result := make(chan sshChannelResult, 1)
	go func() {
		channel, requests, err := conn.OpenChannel(sshRPCChannel, nil)
		if err == nil {
			c.rejectRequests(requests)
		}
		result <- sshChannelResult{channel: channel, err: err}
	}()
	return result
}

func (c *sshChannelConn) acceptRPCChannel(channels <-chan ssh.NewChannel) <-chan sshChannelResult {
	result := make(chan sshChannelResult, 1)
	go func() {
		accepted := false
		for {
			select {
			case <-c.done:
				return
			case request, ok := <-channels:
				if !ok {
					if !accepted {
						result <- sshChannelResult{err: io.EOF}
					}
					return
				}
				if accepted || request.ChannelType() != sshRPCChannel || len(request.ExtraData()) != 0 {
					request.Reject(ssh.Prohibited, "only one empty SSH RPC channel is permitted")
					continue
				}
				accepted = true
				channel, requests, err := request.Accept()
				if err == nil {
					c.rejectRequests(requests)
				}
				result <- sshChannelResult{channel: channel, err: err}
			}
		}
	}()
	return result
}

func (c *sshChannelConn) Read(data []byte) (int, error) {
	n, err := c.channel.Read(data)
	if err != nil {
		c.Close()
		err = c.deadlineError("read", err)
	}
	return n, err
}

func (c *sshChannelConn) Write(data []byte) (int, error) {
	n, err := c.channel.Write(data)
	if err != nil {
		c.Close()
		err = c.deadlineError("write", err)
	}
	return n, err
}

func (c *sshChannelConn) Close() error {
	c.closeOnce.Do(func() {
		c.deadlineMu.Lock()
		c.closed = true
		if c.writeTimer != nil {
			c.writeTimer.Stop()
		}
		c.deadlineMu.Unlock()
		close(c.done)
		// Do not join the SSH mux: malformed peers can strand its internal queues.
		c.closeErr = c.raw.Close()
		if c.cancelHost != nil {
			c.cancelHost()
		}
		// The dedicated SSH channel is owned by the raw connection and closes with it.
	})
	return c.closeErr
}

func (c *sshChannelConn) LocalAddr() net.Addr  { return c.raw.LocalAddr() }
func (c *sshChannelConn) RemoteAddr() net.Addr { return c.raw.RemoteAddr() }

func (c *sshChannelConn) SetDeadline(deadline time.Time) error {
	c.deadlineMu.Lock()
	defer c.deadlineMu.Unlock()
	if c.closed {
		return net.ErrClosed
	}
	if err := c.raw.SetDeadline(deadline); err != nil {
		return err
	}
	c.readDeadline = deadline
	c.setWriteTimer(deadline)
	return nil
}

func (c *sshChannelConn) SetReadDeadline(deadline time.Time) error {
	c.deadlineMu.Lock()
	defer c.deadlineMu.Unlock()
	if c.closed {
		return net.ErrClosed
	}
	if err := c.raw.SetReadDeadline(deadline); err != nil {
		return err
	}
	c.readDeadline = deadline
	return nil
}

func (c *sshChannelConn) SetWriteDeadline(deadline time.Time) error {
	c.deadlineMu.Lock()
	defer c.deadlineMu.Unlock()
	if c.closed {
		return net.ErrClosed
	}
	if err := c.raw.SetWriteDeadline(deadline); err != nil {
		return err
	}
	c.setWriteTimer(deadline)
	return nil
}

func (c *sshChannelConn) setWriteTimer(deadline time.Time) {
	c.writeDeadline = deadline
	if c.writeTimer != nil {
		c.writeTimer.Stop()
	}
	if !deadline.IsZero() {
		if c.writeTimer == nil {
			// A socket write deadline alone cannot interrupt an SSH window wait.
			c.writeTimer = time.AfterFunc(time.Until(deadline), c.expireWriteDeadline)
		} else {
			c.writeTimer.Reset(time.Until(deadline))
		}
	}
}

func (c *sshChannelConn) expireWriteDeadline() {
	c.deadlineMu.Lock()
	defer c.deadlineMu.Unlock()
	if !c.closed && !c.writeDeadline.IsZero() && !time.Now().Before(c.writeDeadline) {
		c.raw.Close()
	}
}

func (c *sshChannelConn) deadlineError(operation string, err error) error {
	c.deadlineMu.Lock()
	deadline := c.readDeadline
	if operation == "write" {
		deadline = c.writeDeadline
	}
	c.deadlineMu.Unlock()
	if !deadline.IsZero() && !time.Now().Before(deadline) {
		return &net.OpError{Op: operation, Net: c.raw.LocalAddr().Network(), Source: c.raw.LocalAddr(), Addr: c.raw.RemoteAddr(), Err: os.ErrDeadlineExceeded}
	}
	return err
}
