/*
 *  Copyright IBM Corporation 2022
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package github

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
)

// Download downloads the given url and saves it at the given path.
func Download(url string, outputPath string, checkSum string) error {
	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create the output file at path %s . Error: %q", outputPath, err)
	}
	defer out.Close()
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to GET the url %s . Error: %q", url, err)
	}
	defer resp.Body.Close()
	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"downloading",
	)
	writers := []io.Writer{out, bar}
	hash := sha256.New()
	if checkSum != "" {
		writers = append(writers, hash)
	}
	n, err := io.Copy(io.MultiWriter(writers...), resp.Body)
	if err != nil {
		return fmt.Errorf("failed to GET the url %s . Error: %q", url, err)
	}
	logrus.Infof("Downloaded a file of size %d bytes from the url %s and saved it to %s", n, url, outputPath)
	if checkSum != "" {
		actualCheckSum := hex.EncodeToString(hash.Sum(nil))
		if actualCheckSum != checkSum {
			return fmt.Errorf("the checksum is incorrect. Expected: %s Actual: %s", checkSum, actualCheckSum)
		} else {
			logrus.Infof("Verified the checksum on the downloaded file!")
		}
	}
	return nil
}

// ExtractTarGz expands a gzip compressed tar archive.
func ExtractTarGz(path string) error {
	archiveDir := filepath.Dir(path)
	gzippedArchive, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open the archive at path %s . Error: %q", path, err)
	}
	defer gzippedArchive.Close()
	archive, err := gzip.NewReader(gzippedArchive)
	if err != nil {
		return fmt.Errorf("failed to decompress the archive at path %s using gzip. Error: %q", path, err)
	}
	tarReader := tar.NewReader(archive)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to parse the tar archive. Error: %q", err)
		}
		switch header.Typeflag {
		case tar.TypeDir:
			dirPath := filepath.Join(archiveDir, header.Name)
			if err := os.Mkdir(dirPath, header.FileInfo().Mode()); err != nil {
				return fmt.Errorf("failed to make the directory %s . Error: %q", dirPath, err)
			}
		case tar.TypeReg:
			filePath := filepath.Join(archiveDir, header.Name)
			outFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, header.FileInfo().Mode())
			if err != nil {
				return fmt.Errorf("failed to create the file at path %s . Error: %q", filePath, err)
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("failed to write to the file at path %s . Error: %q", filePath, err)
			}
		case tar.TypeSymlink:
			logrus.Warnf("found a symbolic link in the tar archive. Skipping.")
		default:
			return fmt.Errorf("failed to parse the tar archive. Found an unsupported header: %#v", header)
		}
	}
	return nil
}
