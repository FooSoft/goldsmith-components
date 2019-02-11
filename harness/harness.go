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

func Validate(t *testing.T, casePrefix string, plugins ...goldsmith.Plugin) {
	var (
		sourceDir    = filepath.Join("testdata", casePrefix, "source")
		targetDir    = filepath.Join("testdata", casePrefix, "target")
		cacheDir     = filepath.Join("testdata", casePrefix, "cache")
		referenceDir = filepath.Join("testdata", casePrefix, "reference")
	)

	if err := validate(sourceDir, targetDir, cacheDir, referenceDir, plugins); err != nil {
		log.Println(err)
		t.Fail()
	}
}

func validate(sourceDir, targetDir, cacheDir, referenceDir string, plugins []goldsmith.Plugin) error {
	if err := os.RemoveAll(targetDir); err != nil {
		return err
	}

	if err := os.RemoveAll(cacheDir); err != nil {
		return err
	}

	defer os.RemoveAll(cacheDir)

	for i := 0; i < 2; i++ {
		if err := execute(sourceDir, targetDir, cacheDir, plugins); err != nil {
			return err
		}

		match, err := compareDirs(targetDir, referenceDir)
		if err != nil {
			return err
		}

		if !match {
			return errors.New("directory contents do not match")
		}
	}

	if err := os.RemoveAll(targetDir); err != nil {
		return err
	}

	return nil
}

func execute(sourceDir, targetDir, cacheDir string, plugins []goldsmith.Plugin) error {
	gs := goldsmith.Begin(sourceDir).Cache(cacheDir)
	for _, plugin := range plugins {
		gs = gs.Chain(plugin)
	}

	if errs := gs.End(targetDir); len(errs) > 0 {
		return errors.New("errors detected in chain")
	}

	return nil
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
