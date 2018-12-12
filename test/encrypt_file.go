package main

import (
	"crypto/rand"
	"fmt"
	"os"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
)

func failf(format string, args ...interface{}) {
	log.Errorf(format, args...)
	os.Exit(1)
}

const inputFile = "my_secret"
const outputFile = inputFile + ".gpg"

func main() {
	inputFileAbsPath, err := pathutil.AbsPath(inputFile)
	if err != nil {
		failf("Absolute file path expansion failed: %s", err)
	}
	outputFileAbsPath, err := pathutil.AbsPath(outputFile)
	if err != nil {
		failf("Absolute file path expansion failed: %s", err)
	}
	// Generate passphrase
	pwgenCmd := command.New("pwgen", "-s", "32", "1")
	out, err := pwgenCmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			failf("%s failed, %s", pwgenCmd.PrintableCommandArgs(), out)
		} else {
			failf("%s failed, %s", pwgenCmd.PrintableCommandArgs(), err)
		}
	}
	log.Printf("Generated passpharase.")
	passphrase := out
	// Create input file
	file, err := os.Create(inputFileAbsPath)
	if err != nil {
		failf("File create failed, error: %s", err)
	}
	log.Printf("Created file: %s", inputFileAbsPath)
	// Fill input file with random data
	data := make([]byte, 10^7)
	if _, err := rand.Read(data); err != nil {
		failf("Random data generation failed")
	}
	if _, err := file.Write(data); err != nil {
		failf("Writing encrypted file failed, error: %s", err)
	}
	log.Printf("Wrote random data to file: %s", inputFileAbsPath)
	// Remove output file
	if err := os.Remove(outputFile); err != nil && !os.IsNotExist(err) {
		failf("Output file removal failed, error: %s", err)
	}
	log.Printf("Output file removed/did not exist: %s", outputFileAbsPath)
	// Encrypt file
	gpgCmd := command.New("gpg", "--batch", "--passphrase", passphrase, "-c",
		inputFileAbsPath)
	if out, err := gpgCmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		if errorutil.IsExitStatusError(err) {
			failf("%s failed, %s", gpgCmd.PrintableCommandArgs(), out)
		} else {
			failf("%s failed, %s", gpgCmd.PrintableCommandArgs(), err)
		}
	}
	log.Printf("Encrypted file, output: %s", outputFileAbsPath)
	// Export env vars with envman
	envmanCmd := command.New("envman", "add", "--key", "ORIGINAL_FILE",
		"--value", inputFileAbsPath)
	fmt.Println()
	log.Donef("$ %s", envmanCmd.PrintableCommandArgs())
	fmt.Println()
	if out, err := envmanCmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		if errorutil.IsExitStatusError(err) {
			failf("%s failed, %s", envmanCmd.PrintableCommandArgs(), out)
		} else {
			failf("%s failed, %s", envmanCmd.PrintableCommandArgs(), err)
		}
	}
	envmanCmd = command.New("envman", "add", "--key", "ENCRYPTED_FILE",
		"--value", outputFileAbsPath)
	fmt.Println()
	log.Donef("$ %s", envmanCmd.PrintableCommandArgs())
	fmt.Println()
	if out, err := envmanCmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		if errorutil.IsExitStatusError(err) {
			failf("%s failed, %s", envmanCmd.PrintableCommandArgs(), out)
		} else {
			failf("%s failed, %s", envmanCmd.PrintableCommandArgs(), err)
		}
	}
	envmanCmd = command.New("envman", "add", "--key", "FILE_DECRYPT_PASSPHRASE",
		"--value", passphrase)
	fmt.Println()
	log.Donef("$ %s", envmanCmd.PrintableCommandArgs())
	fmt.Println()
	if out, err := envmanCmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		if errorutil.IsExitStatusError(err) {
			failf("%s failed, %s", envmanCmd.PrintableCommandArgs(), out)
		} else {
			failf("%s failed, %s", envmanCmd.PrintableCommandArgs(), err)
		}
	}
}
