package config

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

const Version = "0.4.87"

// CheckLatestVersion fetches the latest version from the remote repository
func CheckLatestVersion() (string, error) {
	// Check the source code on GitHub for the version variable
	url := "https://raw.githubusercontent.com/dirmich/maru-bot/main/pkg/config/version.go"

	// Temporarily if the above doesn't exist yet, we might need a fallback.
	// But since we are creating it, we should ensure the URL is correct for when it's pushed.
	// For now, if we fail to fetch version.go, we try the old main.go path as fallback.

	latest, err := fetchFromUrl(url)
	if err != nil {
		oldUrl := "https://raw.githubusercontent.com/dirmich/maru-bot/main/cmd/marubot/main.go"
		return fetchFromUrl(oldUrl)
	}
	return latest, nil
}

func fetchFromUrl(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch version file: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Regex to find: const Version = "0.4.2" or var version = "0.4.2"
	re := regexp.MustCompile(`(const|var)\s+[Vv]ersion\s*=\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) > 2 {
		return matches[2], nil
	}
	return "", fmt.Errorf("version string not found in remote file at %s", url)
}

// IsNewVersionAvailable compares the current version with the latest version.
// It returns true only if the latest version is greater than the current version.
func IsNewVersionAvailable(latest string) bool {
	l := strings.TrimPrefix(latest, "v")
	v := strings.TrimPrefix(Version, "v")

	if l == v {
		return false
	}

	// Simple semver comparison
	lParts := strings.Split(l, ".")
	vParts := strings.Split(v, ".")

	for i := 0; i < len(lParts) && i < len(vParts); i++ {
		var lNum, vNum int
		fmt.Sscanf(lParts[i], "%d", &lNum)
		fmt.Sscanf(vParts[i], "%d", &vNum)

		if lNum > vNum {
			return true
		}
		if lNum < vNum {
			return false
		}
	}

	return len(lParts) > len(vParts)
}
