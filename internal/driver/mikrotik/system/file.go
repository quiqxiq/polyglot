package system

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// FileEntry represents a file in RouterOS storage.
type FileEntry struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	Size           int64  `json:"size"`
	CreationTime   string `json:"creation_time"`
	PackageVersion string `json:"package_version"`
}

// FileContent holds parsed text file content with masked secrets.
type FileContent struct {
	Name   string `json:"name"`
	Text   string `json:"text"`
	Size   int    `json:"size"`
	Masked int    `json:"masked"`
	Error  string `json:"error,omitempty"`
}

var secretAssign = regexp.MustCompile(`(?i)\b((?:[a-z0-9]+-)*(?:password|passphrase|secret|private-key|pre-shared-key|preshared-key|auth-key|psk))=("(?:[^"\\]|\\.)*"|[^\s"]+)`)

// MaskSecrets masks credential values in configuration exports or scripts.
func MaskSecrets(text string) (string, int) {
	maskedCount := 0
	masked := secretAssign.ReplaceAllStringFunc(text, func(m string) string {
		maskedCount++
		eqIdx := strings.Index(m, "=")
		return m[:eqIdx+1] + "«hidden»"
	})
	return masked, maskedCount
}

// ValidateTextFile checks that a file does not exceed maxSize, is valid UTF-8, and contains no private keys.
func ValidateTextFile(body []byte, maxSize int) error {
	if len(body) > maxSize {
		return fmt.Errorf("file size %d exceeds max allowed %d bytes", len(body), maxSize)
	}
	if bytes.IndexByte(body, 0) >= 0 || !utf8.Valid(body) {
		return errors.New("file is not valid text")
	}
	if bytes.Contains(body, []byte("PRIVATE KEY-----")) {
		return errors.New("file contains private key material")
	}
	return nil
}

// NewPrintFilesCommand builds a command to list files on the router with strict proplist.
func NewPrintFilesCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "type", "size", "creation-time", "package-version")
	return command.Command{
		Raw:  "/file/print",
		Args: args,
	}
}

// NewPrintFileCommand builds a query to find one file by name.
func NewPrintFileCommand(name string) command.Command {
	args := map[string]string{
		"?name": name,
	}
	rosutil.SetProplist(args, ".id", "name", "type", "size", "creation-time")
	return command.Command{
		Raw:  "/file/print",
		Args: args,
	}
}

// NewAddFileCommand builds a command to create an empty text file on RouterOS.
func NewAddFileCommand(name string) command.Command {
	return command.Command{
		Raw: "/file/add",
		Args: map[string]string{
			"name": name,
			"type": "file",
		},
	}
}

// NewSetFileCommand builds a command to set file content.
// Note: RouterOS API caps single contents set to ~60KB.
func NewSetFileCommand(id, contents string) command.Command {
	return command.Command{
		Raw: "/file/set",
		Args: map[string]string{
			".id":      id,
			"contents": contents,
		},
	}
}

// NewRemoveFileCommand builds a command to delete a file by ID or name.
func NewRemoveFileCommand(idOrNumber string) command.Command {
	return command.Command{
		Raw: "/file/remove",
		Args: map[string]string{
			"numbers": idOrNumber,
		},
	}
}

// NewReadFileChunkCommand builds a command to read a file chunk via /file/read.
func NewReadFileChunkCommand(fileName string, offset, chunkSize int) command.Command {
	return command.Command{
		Raw: "/file/read",
		Args: map[string]string{
			"file":       fileName,
			"offset":     strconv.Itoa(offset),
			"chunk-size": strconv.Itoa(chunkSize),
		},
	}
}

// NewFetchURLCommand builds a command to download a file onto the router via /tool/fetch.
func NewFetchURLCommand(url, dstPath string) command.Command {
	return command.Command{
		Raw: "/tool/fetch",
		Args: map[string]string{
			"url":      url,
			"dst-path": dstPath,
		},
	}
}

// ParseFileEntries parses /file/print reply rows.
func ParseFileEntries(rows []map[string]string) []FileEntry {
	out := make([]FileEntry, 0, len(rows))
	for _, r := range rows {
		size, _ := strconv.ParseInt(strings.ReplaceAll(r["size"], " ", ""), 10, 64)
		out = append(out, FileEntry{
			ID:             r[".id"],
			Name:           r["name"],
			Type:           r["type"],
			Size:           size,
			CreationTime:   r["creation-time"],
			PackageVersion: r["package-version"],
		})
	}
	return out
}
