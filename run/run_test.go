package run

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pdxjohnny/getfunky/getfunky"
)

const (
	TestName          = "testRunSetup"
	TestEndpoint      = "testRunSetup.service"
	TestEnvSetup      = "#!/bin/sh\necho testEnvSetup\necho -e \"\\x00\\x34\\x99\\xFF\\x00\\x02\\x01\\x2a\\x61\""
	TestPayload       = "#!/bin/sh\necho testPayload\necho -e \"\\x00\\x34\\x99\\xFF\\x00\\x02\\x01\\x2a\\x61\""
	TestPayloadOutput = "testPayload\n\x00\x34\x99\xFF\x00\x02\x01\x2a\x61\n"
)

func testFilePermissions(path string, perms os.FileMode) (os.FileInfo, error) {
	// Check that path was created
	i, err := os.Stat(path)
	if err != nil {
		return i, err
	}

	// Make sure its permisions are correct
	if i.Mode() != perms {
		return i, fmt.Errorf("%v permissions were %v should be %v", path, i.Mode(), perms)
	}

	return i, err
}

func testFileContents(path, contents string) error {
	// Make sure the file was created
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	// Make sure file content is correct
	if string(f) != contents {
		return fmt.Errorf(
			"EnvSetup file %v contents are incorrect, file: %v memory: %v",
			path,
			string(f),
			contents,
		)
	}

	return nil
}

func TestRunSetup(t *testing.T) {
	s := NewService(&getfunky.Service{
		Name:     TestName,
		Endpoint: TestEndpoint,
		EnvSetup: TestEnvSetup,
		Payload:  TestPayload,
	})

	err := s.RunSetup()
	if err != nil {
		t.Fatal(err)
	}

	// Make sure s.tempDir is only owner readable and writeable
	i, err := testFilePermissions(s.tempDir, 020000000700)
	if err != nil {
		t.Fatal(err)
	}
	// Make sure s.tempDir is a directory
	if !i.IsDir() {
		t.Fatal(fmt.Errorf("s.tempDir %v was not a directory", s.tempDir))
	}

	// Make sure EnvSetup permisions are only owner readable, writable, and exec
	_, err = testFilePermissions(filepath.Join(s.tempDir, s.envSetupFileName), 0700)
	if err != nil {
		t.Fatal(err)
	}
	// Make sure EnvSetup file contents are correct
	err = testFileContents(filepath.Join(s.tempDir, s.envSetupFileName), s.EnvSetup)
	if err != nil {
		t.Fatal(err)
	}

	// Make sure Payload permisions are only owner readable, writable, and exec
	_, err = testFilePermissions(filepath.Join(s.tempDir, s.payloadFileName), 0700)
	if err != nil {
		t.Fatal(err)
	}
	// Make sure Payload file contents are correct
	err = testFileContents(filepath.Join(s.tempDir, s.payloadFileName), s.Payload)
	if err != nil {
		t.Fatal(err)
	}

	err = s.RunTeardown()
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunTeardown(t *testing.T) {
	s := NewService(&getfunky.Service{
		Name:     TestName,
		Endpoint: TestEndpoint,
		EnvSetup: TestEnvSetup,
		Payload:  TestPayload,
	})

	err := s.RunSetup()
	if err != nil {
		t.Fatal(err)
	}

	err = s.RunTeardown()
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(s.tempDir)
	if err == nil {
		t.Fatal(fmt.Errorf("s.tempDir %v was not removed"))
	}
}

func TestRunEnvSetup(t *testing.T) {
	s := NewService(&getfunky.Service{
		Name:     TestName,
		Endpoint: TestEndpoint,
		EnvSetup: TestEnvSetup,
		Payload:  TestPayload,
	})

	err := s.RunSetup()
	if err != nil {
		t.Fatal(err)
	}

	err = s.RunEnvSetup()
	if err != nil {
		t.Fatal(err)
	}

	err = s.RunTeardown()
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunPayload(t *testing.T) {
	s := NewService(&getfunky.Service{
		Name:     TestName,
		Endpoint: TestEndpoint,
		EnvSetup: TestEnvSetup,
		Payload:  TestPayload,
	})

	err := s.RunSetup()
	if err != nil {
		t.Fatal(err)
	}

	err = s.RunEnvSetup()
	if err != nil {
		t.Fatal(err)
	}

	ob := new(bytes.Buffer)
	r := &getfunky.Request{
		Env:    []string{"TestRunPayload=42"},
		Body:   strings.NewReader("Endpoint = " + TestEndpoint),
		Output: ob,
	}
	err = s.RunPayload(r)
	if err != nil {
		t.Fatal(err)
	}

	o := ob.String()
	if o != TestPayloadOutput {
		t.Fatal(fmt.Errorf("Payload output was %v should be %v", o, TestPayloadOutput))
	}

	err = s.RunTeardown()
	if err != nil {
		t.Fatal(err)
	}
}
