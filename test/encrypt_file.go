package main

import (
	"crypto/rand"
	"os"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-tools/go-steputils/stepconf"
)

func failf(format string, args ...interface{}) {
	log.Errorf(format, args...)
	os.Exit(1)
}

func runCommand(cmdSlice []string) (string, error) {
	model, err := command.NewFromSlice(cmdSlice)
	if err != nil {
		failf("Command creation failed: %#v", err)
	}
	log.Donef("=> %s", model.PrintableCommandArgs())
	out, err := model.RunAndReturnTrimmedCombinedOutput()
	return out, err
}

const inputFile = "my_secret"
const outputFile = inputFile + ".gpg"

func main() {
	log.SetEnableDebugLog(true)
	inputFileAbsPath, err := pathutil.AbsPath(inputFile)
	if err != nil {
		failf("File path: %#v", err)
	}
	// Generate passphrase
	out, err := runCommand([]string{"pwgen", "-s", "22", "1"})
	if err != nil {
		failf("Generating passphrase failed, error: %#v, out: %s", err, out)
	}
	passphrase := stepconf.Secret(out)
	// Create input file
	file, err := os.Create(inputFileAbsPath)
	if err != nil {
		failf("File create failed, error: %#v", err)
	}
	data := make([]byte, 10^7)
	_, err = rand.Read(data)
	if err != nil {
		failf("Random data generation failed")
	}
	_, err = file.Write(data)
	if err != nil {
		failf("Writing encrypted file failed, error: %#v", err)
	}
	// Remove output file
	err = os.Remove(outputFile)
	if err != nil && !os.IsNotExist(err) {
		failf("Output file removal failed, error: %#v", err)
	}
	// Encrypt file
	outputFileAbsPath, err := pathutil.AbsPath(outputFile)
	if err != nil {
		failf("Absolute file path expansion failed: %#v", err)
	}
	out, err = runCommand([]string{"gpg", "--batch", "--passphrase", string(passphrase), "-c", inputFileAbsPath})
	if err != nil {
		failf("Encryption failed, error: %#v, out: %s", err, out)
	}
	// Set env vars in envman
	out, err = runCommand([]string{"envman", "add", "--key", "ORIGINAL_FILE", "--value", inputFileAbsPath})
	if err != nil {
		failf("Envman add failed, error: %#v, out: %s", err, out)
	}
	out, err = runCommand([]string{"envman", "add", "--key", "ENCRYPTED_FILE", "--value", outputFileAbsPath})
	if err != nil {
		failf("Envman add failed, error: %#v, out: %s", err, out)
	}
	out, err = runCommand([]string{"envman", "add", "--key", "FILE_DECRYPT_PASSPHRASE", "--value", string(passphrase)})
	if err != nil {
		failf("Envman add failed, error: %#v, out: %s", err, out)
	}
}
