package tui

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/psto/irw/internal/config"
	"golang.org/x/term"
)

func LaunchFile(cfg config.ConfigProvider, path string) error {
	launcher := cfg.GetLauncher()
	cmd := exec.Command(launcher, path)
	return cmd.Run()
}

func RunZk(args ...string) ([]byte, error) {
	cmd := exec.Command("zk", args...)
	return cmd.Output()
}

func ReadKey() (rune, error) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return 0, err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	buf := make([]byte, 1)
	_, err = os.Stdin.Read(buf)
	return rune(buf[0]), err
}

func GetHomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return home, nil
}

func AbsPath(input string) (string, error) {
	return filepath.Abs(input)
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func BaseName(path string) string {
	return filepath.Base(path)
}
