package main

import (
	"os"

	"github.com/bitrise-io/go-utils/command"
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
		failf("Could not create config: %v\n", err)
	}
	stepconf.Print(cfg)
	outputFileAbsPath, err := pathutil.AbsPath(cfg.OutputFilePath)
	if err != nil {
		failf("Converting to absolute path failed, error: %#v", err)
	}
	if err := os.Remove(outputFileAbsPath); err != nil && !os.IsNotExist(err) {
		failf("Output file already exists, removal failed: error: %#v", err)
	}

	log.Infof("Decrypting file...")
	cmdSlice := []string{"gpg", "--batch", "--passphrase", cfg.Passphrase.String(),
		"--output", outputFileAbsPath, "--decrypt", cfg.InputFilePath}
	log.Donef("=> %s", command.PrintableCommandArgs(true, cmdSlice))
	// Replace obfuscated passphrase with real one
	cmdSlice[3] = string(cfg.Passphrase)
	model, err := command.NewFromSlice(cmdSlice)
	if err != nil {
		failf("Creating command failed, error: %#v", err)
	}
	out, err := model.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		failf("Decryption with gpg failed, error: %#v | output: %s", err, out)
	}
	log.Donef("Decryption done, output file path: %s", outputFileAbsPath)
}
