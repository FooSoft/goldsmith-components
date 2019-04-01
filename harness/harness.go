package harness

import (
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/FooSoft/goldsmith"
)

type Stager func(gs *goldsmith.Goldsmith)

func Validate(t *testing.T, casePrefix string, stager Stager) {
	var (
		caseDir      = filepath.Join("testdata", casePrefix)
		sourceDir    = filepath.Join(caseDir, "source")
		targetDir    = filepath.Join(caseDir, "target")
		cacheDir     = filepath.Join(caseDir, "cache")
		referenceDir = filepath.Join(caseDir, "reference")
	)

	if errs := validate(sourceDir, targetDir, cacheDir, referenceDir, stager); len(errs) > 0 {
		for _, err := range errs {
			log.Println(err)
		}

		t.Fail()
	}
}

func validate(sourceDir, targetDir, cacheDir, referenceDir string, stager Stager) []error {
	if err := os.RemoveAll(targetDir); err != nil {
		return []error{err}
	}

	if err := os.RemoveAll(cacheDir); err != nil {
		return []error{err}
	}

	defer os.RemoveAll(cacheDir)

	for i := 0; i < 2; i++ {
		if errs := execute(sourceDir, targetDir, cacheDir, stager); errs != nil {
			return errs
		}

		match, err := compareDirs(targetDir, referenceDir)
		if err != nil {
			return []error{err}
		}

		if !match {
			return []error{errors.New("directory contents do not match")}
		}
	}

	if err := os.RemoveAll(targetDir); err != nil {
		return []error{err}
	}

	return nil
}

func execute(sourceDir, targetDir, cacheDir string, stager Stager) []error {
	gs := goldsmith.Begin(sourceDir).Cache(cacheDir)
	stager(gs)
	return gs.End(targetDir)
}

func hashDirState(dir string) (uint32, error) {
	hasher := crc32.NewIEEE()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		summary := fmt.Sprintf("%s %t", relPath, info.IsDir())
		if _, err := hasher.Write([]byte(summary)); err != nil {
			return err
		}

		if !info.IsDir() {
			fp, err := os.Open(path)
			if err != nil {
				return err
			}
			defer fp.Close()

			if _, err := io.Copy(hasher, fp); err != nil {
				return err
			}
		}

		return nil
	})

	return hasher.Sum32(), err
}

func compareDirs(sourceDir, targetDir string) (bool, error) {
	sourceHash, err := hashDirState(sourceDir)
	if err != nil {
		return false, err
	}

	targetHash, err := hashDirState(targetDir)
	if err != nil {
		return false, err
	}

	return sourceHash == targetHash, nil
}