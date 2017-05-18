package sftp

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/ScaleFT/sshkeys"
	"github.com/pkg/sftp"
	"github.com/slavikmanukyan/itm/itmconfig"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type SaveFunc func(file string)

var client *sftp.Client
var connection *ssh.Client
var inited = false

var Client *sftp.Client

func PublicKeyFile(file string, pass string) ssh.AuthMethod {
	var key ssh.Signer
	var err error

	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	if pass != "" {
		key, err = sshkeys.ParseEncryptedPrivateKey(buffer, []byte(pass))
	} else {
		key, err = ssh.ParsePrivateKey(buffer)
	}
	if err != nil {
		panic(err)
	}
	return ssh.PublicKeys(key)
}

func InitClient(config itmconfig.ITMConfig) {
	var auths []ssh.AuthMethod

	if config.SSH_PRIVATE_KEY != "" {
		auths = append(auths, PublicKeyFile(config.SSH_PRIVATE_KEY, config.SSH_KEY_PASSPHRASE))
	}

	if config.SSH_AUTH_SOCK == true {
		if aconn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
			auths = append(auths, ssh.PublicKeysCallback(agent.NewClient(aconn).Signers))

		}
	}

	if config.SSH_PASSWORD != "" {
		auths = append(auths, ssh.Password(config.SSH_PASSWORD))
	}

	sshConfig := ssh.ClientConfig{
		User:            config.SSH_USER,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	var err error

	addr := fmt.Sprintf("%s:%d", config.SSH_HOST, config.SSH_PORT)
	connection, err = ssh.Dial("tcp", addr, &sshConfig)

	if err != nil {
		log.Fatalf("unable to connect to [%s]: %v", addr, err)
	}

	client, err = sftp.NewClient(connection, sftp.MaxPacket(1<<15))
	if err != nil {
		log.Fatalf("unable to start sftp subsytem: %v", err)
	}

	log.Printf("Connected to [%s]", addr)
	inited = true
	Client = client
}

func CopyFileFrom(src, dst string) (err error) {
	if inited != true {
		return
	}
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	in, err := client.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := client.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

func CopyFileTo(src, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := client.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = client.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

func CopyDirFrom(src, dst string) (err error) {
	if inited != true {
		return
	}
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := client.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := client.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		fmt.Println(entry.Name())

		if entry.IsDir() {
			err = CopyDirFrom(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = CopyFileFrom(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}

func CopyDirTo(src, dst string, config itmconfig.ITMConfig, saveCallback SaveFunc) (err error) {
	if inited != true {
		return
	}
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	client.Mkdir(dst)

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		a1, _ := filepath.Abs(srcPath)
		if config.IGNORE[a1] {
			continue
		}

		if entry.IsDir() {
			err = CopyDirTo(srcPath, dstPath, config, saveCallback)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = CopyFileTo(srcPath, dstPath)
			if err != nil {
				return
			}
			saveCallback(srcPath)
		}
	}

	return
}
