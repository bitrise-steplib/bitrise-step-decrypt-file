package main

import (
	"fmt"
	"os"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-tools/go-steputils/stepconf"
)

type config struct {
	InputFilePath  string          `env:"encrypted_file_path,file"`
	OutputFilePath string          `env:"output_file_path,required"`
	Passphrase     stepconf.Secret `env:"decrypt_passphrase,required"`
}

func failf(format string, args ...interface{}) {
	log.Errorf(format, args...)
	os.Exit(1)
}

func main() {
	var cfg config
	if err := stepconf.Parse(&cfg); err != nil {
		failf("Could not create config: %s", err)
	}
	stepconf.Print(cfg)
	outputFileAbsPath, err := pathutil.AbsPath(cfg.OutputFilePath)
	if err != nil {
		failf("Converting to absolute path failed, error: %s", err)
	}
	if err := os.Remove(outputFileAbsPath); err != nil && !os.IsNotExist(err) {
		failf("Output file already exists, removal failed: error: %s", err)
	}

	log.Infof("Decrypting file...")
	cmdSlice := []string{"gpg", "--batch", "--passphrase", cfg.Passphrase.String(),
		"--output", outputFileAbsPath, "--decrypt", cfg.InputFilePath}
	fmt.Println()
	log.Donef("$ %s", command.PrintableCommandArgs(true, cmdSlice))
	fmt.Println()
	// Replace obfuscated passphrase with real one
	cmdSlice[3] = string(cfg.Passphrase)
	model, err := command.NewFromSlice(cmdSlice)
	if err != nil {
		failf("Creating command failed, error: %s", err)
	}
	if out, err := model.RunAndReturnTrimmedCombinedOutput(); err != nil {
		if errorutil.IsExitStatusError(err) {
			failf("gpg decryption failed: %s", out)
		} else {
			failf("gpg decryption failed: %s", err)
		}
	}
	log.Donef("Decryption done, output file path: %s", outputFileAbsPath)
}
