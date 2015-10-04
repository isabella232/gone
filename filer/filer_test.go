package filer

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/fxnn/gone/authenticator"
)

func TestOpenWriterSupportsCreatingFiles(t *testing.T) {
	tmpdir := createTempDirInCurrentwd(t, 0772)
	defer removeTempDirFromCurrentwd(t, tmpdir)

	sut := New(authenticator.NewNeverAuthenticated())

	writeCloser := sut.OpenWriter(requestGET("/" + tmpdir + "/newFile"))
	closed(writeCloser)
	if err := sut.Err(); err != nil {
		t.Fatalf("failed to open file for writing: %s", err)
	}
}

func TestOpenWriterDeniesWhenWorldPermissionIsMissing(t *testing.T) {
	tmpfile := createTempFileInCurrentwd(t, 0770)
	defer removeTempFileFromCurrentwd(t, tmpfile)

	sut := New(authenticator.NewNeverAuthenticated())

	writeCloser := sut.OpenWriter(requestGET("/" + tmpfile))
	closed(writeCloser)
	if err := sut.Err(); err == nil || !IsAccessDeniedError(err) {
		t.Fatalf("expected AccessDeniedError on %s, but got %s", tmpfile, err)
	}
}

func TestOpenReaderDeniesWhenWorldPermissionIsMissing(t *testing.T) {
	tmpfile := createTempFileInCurrentwd(t, 0770)
	defer removeTempFileFromCurrentwd(t, tmpfile)

	sut := New(authenticator.NewNeverAuthenticated())

	readCloser := sut.OpenReader(requestGET("/" + tmpfile))
	closed(readCloser)
	if err := sut.Err(); err == nil || !IsAccessDeniedError(err) {
		t.Fatalf("expected AccessDeniedError on %s, but got %s", tmpfile, err)
	}
}

func TestOpenReaderProceedsWhenAuthenticated(t *testing.T) {
	tmpfile := createTempFileInCurrentwd(t, 0770)
	defer removeTempFileFromCurrentwd(t, tmpfile)

	sut := New(authenticator.NewAlwaysAuthenticated())

	readCloser := sut.OpenReader(requestGET("/" + tmpfile))
	closed(readCloser)
	if err := sut.Err(); err != nil {
		t.Fatalf("failed to open %s for reading: %s", tmpfile, err)
	}
}

func requestGET(path string) (request *http.Request) {
	request, _ = http.NewRequest("GET", path, nil)
	return
}

func createTempDirInCurrentwd(t *testing.T, mode os.FileMode) string {
	wd := getwd(t)
	tmpdir, err := ioutil.TempDir(wd, "gone_test_")
	if err != nil {
		t.Fatalf("couldnt create tempdir in %s: %s", wd, err)
	}
	err = os.Chmod(tmpdir, mode)
	if err != nil {
		t.Fatalf("couldnt chmod tempdir %s: %s", tmpdir, err)
	}
	return path.Base(tmpdir)
}

func createTempFileInCurrentwd(t *testing.T, mode os.FileMode) string {
	wd := getwd(t)
	tmpfile, err := ioutil.TempFile(wd, "gone_test_")
	if err != nil {
		t.Fatalf("couldnt create tempfile in %s: %s", wd, err)
	}
	info, err := tmpfile.Stat()
	if err != nil {
		t.Fatalf("couldnt stat tmpfile %s: %s", tmpfile, err)
	}
	err = tmpfile.Chmod(mode)
	if err != nil {
		t.Fatalf("couldnt chmod tmpfile %s: %s", info.Name(), err)
	}
	err = tmpfile.Close()
	if err != nil {
		t.Fatalf("couldn close tmpfile %s: %s", info.Name(), err)
	}
	return info.Name()
}

func removeTempDirFromCurrentwd(t *testing.T, tmpdir string) {
	wd := getwd(t)
	tmpdirPath := path.Join(wd, tmpdir)
	err := os.RemoveAll(tmpdirPath)
	if err != nil {
		t.Fatalf("couldnt remove tmpdir %s: %s", tmpdirPath, err)
	}
}

func removeTempFileFromCurrentwd(t *testing.T, tmpfile string) {
	wd := getwd(t)
	tmpfilePath := path.Join(wd, tmpfile)
	err := os.RemoveAll(tmpfilePath)
	if err != nil {
		t.Fatalf("couldn remove tmpfile %s: %s", tmpfilePath, err)
	}
}

func getwd(t *testing.T) string {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("couldnt get working directory: %s", err)
	}
	return wd
}

func closed(c io.Closer) {
	if c != nil {
		c.Close()
	}
}