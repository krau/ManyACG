package internalrpc

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/krau/ManyACG/internal/infra/internalrpc/internalpb"
	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const (
	authenticationTimeout   = 10 * time.Second
	authenticationNonceSize = 32
	maxAuthenticationBytes  = 16 << 10
	authenticationDomain    = "ManyACG internal node authentication v3\x00"
)

type connectionAuthKey struct{}

type connectionAuth struct {
	mu           sync.RWMutex
	attaching    bool
	nodeID       string
	connectionID string
}

// gRPC propagates TagConn's context to every RPC on the same physical connection.
type connectionAuthHandler struct{}

func (connectionAuthHandler) TagConn(ctx context.Context, _ *stats.ConnTagInfo) context.Context {
	return context.WithValue(ctx, connectionAuthKey{}, &connectionAuth{})
}
func (connectionAuthHandler) HandleConn(context.Context, stats.ConnStats) {}
func (connectionAuthHandler) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context {
	return ctx
}
func (connectionAuthHandler) HandleRPC(context.Context, stats.RPCStats) {}

func beginAuthentication(ctx context.Context) (*connectionAuth, error) {
	connection, ok := ctx.Value(connectionAuthKey{}).(*connectionAuth)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing internal connection identity")
	}
	connection.mu.Lock()
	defer connection.mu.Unlock()
	if connection.attaching {
		return nil, status.Error(codes.AlreadyExists, "this connection already has an active session")
	}
	connection.attaching = true
	return connection, nil
}

func (c *connectionAuth) authorize(nodeID, connectionID string) {
	c.mu.Lock()
	c.nodeID, c.connectionID = nodeID, connectionID
	c.mu.Unlock()
}

func (c *connectionAuth) clear() {
	c.mu.Lock()
	c.attaching = false
	c.nodeID, c.connectionID = "", ""
	c.mu.Unlock()
}

func authenticatedPeer(ctx context.Context) (nodeID, connectionID string, err error) {
	connection, ok := ctx.Value(connectionAuthKey{}).(*connectionAuth)
	if !ok {
		return "", "", status.Error(codes.Unauthenticated, "missing internal connection identity")
	}
	connection.mu.RLock()
	defer connection.mu.RUnlock()
	if connection.nodeID == "" || connection.connectionID == "" {
		return "", "", status.Error(codes.Unauthenticated, "internal connection is not authenticated")
	}
	return connection.nodeID, connection.connectionID, nil
}

func nodeKeyAlgorithm(key ssh.PublicKey) (string, error) {
	switch key.Type() {
	case ssh.KeyAlgoED25519, ssh.KeyAlgoECDSA256, ssh.KeyAlgoECDSA384, ssh.KeyAlgoECDSA521:
		return key.Type(), nil
	case ssh.KeyAlgoRSA:
		cryptoKey, ok := key.(ssh.CryptoPublicKey)
		if ok {
			publicKey, ok := cryptoKey.CryptoPublicKey().(*rsa.PublicKey)
			if ok && publicKey.N.BitLen() >= 2048 {
				return ssh.KeyAlgoRSASHA512, nil
			}
		}
	}
	return "", errors.New("node key must be Ed25519, ECDSA or RSA with at least 2048 bits")
}

func readNodeSigner(path string) (ssh.Signer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read internal agent private key: %w", err)
	}
	defer clear(data)
	signer, err := ssh.ParsePrivateKey(data)
	if err != nil {
		return nil, errors.New("internal agent private key must be valid and unencrypted")
	}
	if _, err := nodeKeyAlgorithm(signer.PublicKey()); err != nil {
		return nil, err
	}
	return signer, nil
}

type authorizedKeysCacheEntry struct {
	modTime time.Time
	size    int64
	keys    map[string]ssh.PublicKey
}

var (
	authorizedKeysMu    sync.Mutex
	authorizedKeysCache = make(map[string]authorizedKeysCacheEntry)
)

func readAuthorizedKeys(path string) (map[string]ssh.PublicKey, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("read internal authorized_keys: %w", err)
	}
	authorizedKeysMu.Lock()
	if entry, ok := authorizedKeysCache[path]; ok && entry.size == info.Size() && entry.modTime.Equal(info.ModTime()) {
		authorizedKeysMu.Unlock()
		return entry.keys, nil
	}
	authorizedKeysMu.Unlock()
	keys, err := parseAuthorizedKeys(path)
	if err != nil {
		return nil, err
	}
	authorizedKeysMu.Lock()
	authorizedKeysCache[path] = authorizedKeysCacheEntry{modTime: info.ModTime(), size: info.Size(), keys: keys}
	authorizedKeysMu.Unlock()
	return keys, nil
}

func parseAuthorizedKeys(path string) (map[string]ssh.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read internal authorized_keys: %w", err)
	}
	keys := make(map[string]ssh.PublicKey)
	lineNumber := 0
	for line := range bytes.Lines(data) {
		lineNumber++
		line = bytes.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		key, _, options, rest, err := ssh.ParseAuthorizedKey(line)
		if err != nil || len(rest) != 0 || bytes.ContainsAny(line, "\r\n") {
			return nil, fmt.Errorf("invalid authorized_keys line %d", lineNumber)
		}
		firstSpace := bytes.IndexAny(line, " \t")
		if len(options) != 0 || firstSpace < 0 || string(line[:firstSpace]) != key.Type() {
			return nil, fmt.Errorf("authorized_keys options are unsupported on line %d", lineNumber)
		}
		if _, err := nodeKeyAlgorithm(key); err != nil {
			return nil, fmt.Errorf("unsupported authorized_keys key on line %d", lineNumber)
		}
		keys[string(key.Marshal())] = key
	}
	return keys, nil
}

func authenticationPayload(nonce []byte, register *internalpb.Register) ([]byte, error) {
	size := proto.Size(register)
	if len(nonce) != authenticationNonceSize || size > maxAuthenticationBytes {
		return nil, errors.New("invalid internal authentication challenge or registration")
	}
	data := make([]byte, 0, len(authenticationDomain)+len(nonce)+size)
	data = append(data, authenticationDomain...)
	data = append(data, nonce...)
	return proto.MarshalOptions{Deterministic: true}.MarshalAppend(data, register)
}

func authenticationProof(signer ssh.Signer, nonce []byte, register *internalpb.Register) (*internalpb.AuthenticationProof, error) {
	payload, err := authenticationPayload(nonce, register)
	if err != nil {
		return nil, err
	}
	algorithm, err := nodeKeyAlgorithm(signer.PublicKey())
	if err != nil {
		return nil, err
	}
	var signature *ssh.Signature
	if algorithmSigner, ok := signer.(ssh.AlgorithmSigner); ok {
		signature, err = algorithmSigner.SignWithAlgorithm(rand.Reader, payload, algorithm)
	} else {
		signature, err = signer.Sign(rand.Reader, payload)
	}
	if err != nil {
		return nil, errors.New("sign internal authentication challenge")
	}
	return &internalpb.AuthenticationProof{Algorithm: signature.Format, Signature: signature.Blob}, nil
}

type authenticatedRegistration struct {
	register     *internalpb.Register
	nodeID       string
	connectionID string
	err          error
}

func (g *Gateway) authenticate(stream internalpb.Gateway_AttachServer) authenticatedRegistration {
	ctx, cancel := context.WithTimeout(stream.Context(), authenticationTimeout)
	defer cancel()
	result := make(chan authenticatedRegistration, 1)
	// Returning the handler cancels a blocked Recv when authentication expires.
	go func() { result <- g.readAuthentication(stream) }()
	select {
	case registration := <-result:
		if ctx.Err() != nil {
			return authenticatedRegistration{err: status.FromContextError(ctx.Err()).Err()}
		}
		return registration
	case <-ctx.Done():
		return authenticatedRegistration{err: status.FromContextError(ctx.Err()).Err()}
	}
}

func (g *Gateway) readAuthentication(stream internalpb.Gateway_AttachServer) authenticatedRegistration {
	fail := func(code codes.Code, message string) authenticatedRegistration {
		return authenticatedRegistration{err: status.Error(code, message)}
	}
	first, err := stream.Recv()
	if err != nil {
		return authenticatedRegistration{err: err}
	}
	register := first.GetRegister()
	if register == nil || proto.Size(register) > maxAuthenticationBytes || len(register.GetPublicKey()) == 0 {
		return fail(codes.Unauthenticated, "first agent message must contain a public key registration")
	}
	if register.GetProtocolVersion() != protocolVersion {
		return fail(codes.FailedPrecondition, "unsupported internal protocol version")
	}
	keys, err := readAuthorizedKeys(g.authorizedKeysFile)
	if err != nil {
		return fail(codes.Unauthenticated, "internal authorized_keys is unavailable or invalid")
	}
	key := keys[string(register.GetPublicKey())]
	if key == nil {
		return fail(codes.Unauthenticated, "internal node public key is not authorized")
	}
	var nonce [authenticationNonceSize]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return fail(codes.Internal, "generate internal authentication challenge")
	}
	if publicKey, nodeID, ok := sshAuthenticatedKey(stream.Context()); ok {
		if !bytes.Equal(publicKey, register.GetPublicKey()) {
			return fail(codes.Unauthenticated, "registration public key does not match the SSH identity")
		}
		return authenticatedRegistration{register: register, nodeID: nodeID, connectionID: hex.EncodeToString(nonce[:])}
	}
	if err := stream.Send(&internalpb.Assignment{Payload: &internalpb.Assignment_Challenge{
		Challenge: &internalpb.AuthenticationChallenge{Nonce: nonce[:]},
	}}); err != nil {
		return authenticatedRegistration{err: err}
	}
	message, err := stream.Recv()
	if err != nil {
		return authenticatedRegistration{err: err}
	}
	proof := message.GetProof()
	if proof == nil || len(proof.GetSignature()) == 0 || len(proof.GetSignature()) > maxAuthenticationBytes {
		return fail(codes.Unauthenticated, "internal node private key proof is required")
	}
	algorithm, err := nodeKeyAlgorithm(key)
	if err != nil || proof.GetAlgorithm() != algorithm {
		return fail(codes.Unauthenticated, "invalid internal node signature algorithm")
	}
	payload, err := authenticationPayload(nonce[:], register)
	if err != nil || key.Verify(payload, &ssh.Signature{Format: proof.GetAlgorithm(), Blob: proof.GetSignature()}) != nil {
		return fail(codes.Unauthenticated, "invalid internal node private key proof")
	}
	return authenticatedRegistration{register: register, nodeID: ssh.FingerprintSHA256(key), connectionID: hex.EncodeToString(nonce[:])}
}
