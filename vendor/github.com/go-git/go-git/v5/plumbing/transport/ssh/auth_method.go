package ssh

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing/transport"

	"github.com/skeema/knownhosts"
	sshagent "github.com/xanzy/ssh-agent"
	"golang.org/x/crypto/ssh"
)

const DefaultUsername = "git"

// AuthMethod is the interface all auth methods for the ssh client
// must implement. The clientConfig method returns the ssh client
// configuration needed to establish an ssh connection.
type AuthMethod interface {
	transport.AuthMethod
	// ClientConfig should return a valid ssh.ClientConfig to be used to create
	// a connection to the SSH server.
	ClientConfig() (*ssh.ClientConfig, error)
}

// The names of the AuthMethod implementations. To be returned by the
// Name() method. Most git servers only allow PublicKeysName and
// PublicKeysCallbackName.
const (
	KeyboardInteractiveName = "ssh-keyboard-interactive"
	PasswordName            = "ssh-password"
	PasswordCallbackName    = "ssh-password-callback"
	PublicKeysName          = "ssh-public-keys"
	PublicKeysCallbackName  = "ssh-public-key-callback"
)

// KeyboardInteractive implements AuthMethod by using a
// prompt/response sequence controlled by the server.
type KeyboardInteractive struct {
	User      string
	Challenge ssh.KeyboardInteractiveChallenge
	HostKeyCallbackHelper
}

func (a *KeyboardInteractive) Name() string {
	return KeyboardInteractiveName
}

func (a *KeyboardInteractive) String() string {
	return fmt.Sprintf("user: %s, name: %s", a.User, a.Name())
}

func (a *KeyboardInteractive) ClientConfig() (*ssh.ClientConfig, error) {
	return a.SetHostKeyCallbackAndAlgorithms(&ssh.ClientConfig{
		User: a.User,
		Auth: []ssh.AuthMethod{
			a.Challenge,
		},
	})
}

// Password implements AuthMethod by using the given password.
type Password struct {
	User     string
	Password string
	HostKeyCallbackHelper
}

func (a *Password) Name() string {
	return PasswordName
}

func (a *Password) String() string {
	return fmt.Sprintf("user: %s, name: %s", a.User, a.Name())
}

func (a *Password) ClientConfig() (*ssh.ClientConfig, error) {
	return a.SetHostKeyCallbackAndAlgorithms(&ssh.ClientConfig{
		User: a.User,
		Auth: []ssh.AuthMethod{ssh.Password(a.Password)},
	})
}

// PasswordCallback implements AuthMethod by using a callback
// to fetch the password.
type PasswordCallback struct {
	User     string
	Callback func() (pass string, err error)
	HostKeyCallbackHelper
}

func (a *PasswordCallback) Name() string {
	return PasswordCallbackName
}

func (a *PasswordCallback) String() string {
	return fmt.Sprintf("user: %s, name: %s", a.User, a.Name())
}

func (a *PasswordCallback) ClientConfig() (*ssh.ClientConfig, error) {
	return a.SetHostKeyCallbackAndAlgorithms(&ssh.ClientConfig{
		User: a.User,
		Auth: []ssh.AuthMethod{ssh.PasswordCallback(a.Callback)},
	})
}

// PublicKeys implements AuthMethod by using the given key pairs.
type PublicKeys struct {
	User   string
	Signer ssh.Signer
	HostKeyCallbackHelper
}

// NewPublicKeys returns a PublicKeys from a PEM encoded private key. An
// encryption password should be given if the pemBytes contains a password
// encrypted PEM block otherwise password should be empty. It supports RSA
// (PKCS#1), PKCS#8, DSA (OpenSSL), and ECDSA private keys.
func NewPublicKeys(user string, pemBytes []byte, password string) (*PublicKeys, error) {
	signer, err := ssh.ParsePrivateKey(pemBytes)
	if _, ok := err.(*ssh.PassphraseMissingError); ok {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(password))
	}
	if err != nil {
		return nil, err
	}
	return &PublicKeys{User: user, Signer: signer}, nil
}

// NewPublicKeysFromFile returns a PublicKeys from a file containing a PEM
// encoded private key. An encryption password should be given if the pemBytes
// contains a password encrypted PEM block otherwise password should be empty.
func NewPublicKeysFromFile(user, pemFile, password string) (*PublicKeys, error) {
	bytes, err := os.ReadFile(pemFile)
	if err != nil {
		return nil, err
	}

	return NewPublicKeys(user, bytes, password)
}

func (a *PublicKeys) Name() string {
	return PublicKeysName
}

func (a *PublicKeys) String() string {
	return fmt.Sprintf("user: %s, name: %s", a.User, a.Name())
}

func (a *PublicKeys) ClientConfig() (*ssh.ClientConfig, error) {
	return a.SetHostKeyCallbackAndAlgorithms(&ssh.ClientConfig{
		User: a.User,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(a.Signer)},
	})
}

func username() (string, error) {
	var username string
	if user, err := user.Current(); err == nil {
		username = user.Username
	} else {
		username = os.Getenv("USER")
	}

	if username == "" {
		return "", errors.New("failed to get username")
	}

	return username, nil
}

// PublicKeysCallback implements AuthMethod by asking a
// ssh.agent.Agent to act as a signer.
type PublicKeysCallback struct {
	User     string
	Callback func() (signers []ssh.Signer, err error)
	HostKeyCallbackHelper
}

// NewSSHAgentAuth returns a PublicKeysCallback based on a SSH agent, it opens
// a pipe with the SSH agent and uses the pipe as the implementer of the public
// key callback function.
func NewSSHAgentAuth(u string) (*PublicKeysCallback, error) {
	var err error
	if u == "" {
		u, err = username()
		if err != nil {
			return nil, err
		}
	}

	a, _, err := sshagent.New()
	if err != nil {
		return nil, fmt.Errorf("error creating SSH agent: %q", err)
	}

	return &PublicKeysCallback{
		User:     u,
		Callback: a.Signers,
	}, nil
}

func (a *PublicKeysCallback) Name() string {
	return PublicKeysCallbackName
}

func (a *PublicKeysCallback) String() string {
	return fmt.Sprintf("user: %s, name: %s", a.User, a.Name())
}

func (a *PublicKeysCallback) ClientConfig() (*ssh.ClientConfig, error) {
	return a.SetHostKeyCallbackAndAlgorithms(&ssh.ClientConfig{
		User: a.User,
		Auth: []ssh.AuthMethod{ssh.PublicKeysCallback(a.Callback)},
	})
}

// NewKnownHostsCallback returns ssh.HostKeyCallback based on a file based on a
// known_hosts file. http://man.openbsd.org/sshd#SSH_KNOWN_HOSTS_FILE_FORMAT
//
// If list of files is empty, then it will be read from the SSH_KNOWN_HOSTS
// environment variable, example:
//
//	/home/foo/custom_known_hosts_file:/etc/custom_known/hosts_file
//
// If SSH_KNOWN_HOSTS is not set the following file locations will be used:
//
//	~/.ssh/known_hosts
//	/etc/ssh/ssh_known_hosts
func NewKnownHostsCallback(files ...string) (ssh.HostKeyCallback, error) {
	kh, err := NewKnownHostsDb(files...)
	if err != nil {
		return nil, err
	}
	return kh.HostKeyCallback(), nil
}

// NewKnownHostsDb returns knownhosts.HostKeyDB based on a file based on a
// known_hosts file. http://man.openbsd.org/sshd#SSH_KNOWN_HOSTS_FILE_FORMAT
//
// If list of files is empty, then it will be read from the SSH_KNOWN_HOSTS
// environment variable, example:
//
//	/home/foo/custom_known_hosts_file:/etc/custom_known/hosts_file
//
// If SSH_KNOWN_HOSTS is not set the following file locations will be used:
//
//	~/.ssh/known_hosts
//	/etc/ssh/ssh_known_hosts
func NewKnownHostsDb(files ...string) (*knownhosts.HostKeyDB, error) {
	var err error

	if len(files) == 0 {
		if files, err = getDefaultKnownHostsFiles(); err != nil {
			return nil, err
		}
	}

	if files, err = filterKnownHostsFiles(files...); err != nil {
		return nil, err
	}

	return knownhosts.NewDB(files...)
}

func getDefaultKnownHostsFiles() ([]string, error) {
	files := filepath.SplitList(os.Getenv("SSH_KNOWN_HOSTS"))
	if len(files) != 0 {
		return files, nil
	}

	homeDirPath, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return []string{
		filepath.Join(homeDirPath, "/.ssh/known_hosts"),
		"/etc/ssh/ssh_known_hosts",
	}, nil
}

func filterKnownHostsFiles(files ...string) ([]string, error) {
	var out []string
	for _, file := range files {
		_, err := os.Stat(file)
		if err == nil {
			out = append(out, file)
			continue
		}

		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("unable to find any valid known_hosts file, set SSH_KNOWN_HOSTS env variable")
	}

	return out, nil
}

// HostKeyCallbackHelper is a helper that provides common functionality to
// configure HostKeyCallback and HostKeyAlgorithms into a ssh.ClientConfig.
type HostKeyCallbackHelper struct {
	// HostKeyCallback is the function type used for verifying server keys.
	// If nil, a default callback will be created using NewKnownHostsDb
	// without argument.
	HostKeyCallback ssh.HostKeyCallback

	// HostKeyAlgorithms is a list of supported host key algorithms that will
	// be used for host key verification.
	HostKeyAlgorithms []string

	// fallback allows for injecting the fallback call, which is called
	// when a HostKeyCallback is not set.
	fallback func(files ...string) (ssh.HostKeyCallback, error)
}

// SetHostKeyCallbackAndAlgorithms sets the field HostKeyCallback and HostKeyAlgorithms in the given cfg.
// If the host key callback or algorithms is empty it is left empty. It will be handled by the dial method,
// falling back to knownhosts.
func (m *HostKeyCallbackHelper) SetHostKeyCallbackAndAlgorithms(cfg *ssh.ClientConfig) (*ssh.ClientConfig, error) {
	if cfg == nil {
		cfg = &ssh.ClientConfig{}
	}

	if m.HostKeyCallback == nil {
		if m.fallback == nil {
			m.fallback = NewKnownHostsCallback
		}

		hkcb, err := m.fallback()
		if err != nil {
			return nil, fmt.Errorf("cannot create known hosts callback: %w", err)
		}

		cfg.HostKeyCallback = hkcb
		cfg.HostKeyAlgorithms = m.HostKeyAlgorithms
		return cfg, err
	}

	cfg.HostKeyCallback = m.HostKeyCallback
	cfg.HostKeyAlgorithms = m.HostKeyAlgorithms
	return cfg, nil
}

func (m *HostKeyCallbackHelper) SetHostKeyCallback(cfg *ssh.ClientConfig) (*ssh.ClientConfig, error) {
	return m.SetHostKeyCallbackAndAlgorithms(cfg)
}
