package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	API_KEY               = ""
	SEARCH_API_ENDPOINT   = "https://api.pexels.com/v1/search"
	PICTURES_DIR          = "{HOME}/Pictures/Wallpapers/Pexels"
	QUERIES_FILE          = "{HOME}/.local/bin/data/pexels.json"
	SWITCH_WALL_CMD       = "{HOME}/.config/eww/scripts/switchwall"
	LATEST_WALLPAPER_FILE = "{HOME}/.local/bin/data/latest_wallpaper.txt"
	HASHES_FILE           = "{HOME}/.local/bin/data/pexels_hashes.json"
	MAX_IMAGES            = 15
)

var (
	imageHashes = make(map[string]struct{})

	mutex = &sync.Mutex{}

	wg sync.WaitGroup
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home directory:", err)
		return
	}

	PICTURES_DIR = filepath.Join(strings.ReplaceAll(PICTURES_DIR, "{HOME}", homeDir))
	QUERIES_FILE = filepath.Join(strings.ReplaceAll(QUERIES_FILE, "{HOME}", homeDir))
	SWITCH_WALL_CMD = filepath.Join(strings.ReplaceAll(SWITCH_WALL_CMD, "{HOME}", homeDir))
	LATEST_WALLPAPER_FILE = filepath.Join(strings.ReplaceAll(LATEST_WALLPAPER_FILE, "{HOME}", homeDir))
	HASHES_FILE = filepath.Join(strings.ReplaceAll(HASHES_FILE, "{HOME}", homeDir))

	// Set API_KEY from environment variable
	API_KEY = os.Getenv("PEXELS_API_KEY")
	if API_KEY == "" {
		fmt.Println("PEXELS_API_KEY environment variable not set")
		return
	}
}

func getWallpaperCount() int {
	files, err := os.ReadDir(PICTURES_DIR)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return 0
	}

	count := 0
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".jpg" {
			count++
		}
	}
	return count
}

func getNextWallpaperCount() int {
	files, err := os.ReadDir(PICTURES_DIR)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return 0
	}

	latestWallpaper, err := getLatestWallpaper()
	if err != nil {
		fmt.Println("Error getting latest wallpaper:", err)
		return 0
	}

	// Find the latest wallpaper's index
	var latestIndex int
	for index, file := range files {
		if file.Name() == latestWallpaper {
			latestIndex = index
			break
		}
	}

	return len(files) - latestIndex - 1
}

func shouldSetAsNextWallpaper(latestWallpaper, newWallpaper string) bool {
	// If there's no latest wallpaper set, return true
	if latestWallpaper == "" {
		return true
	}

	// Parse the names into dates
	latestTime, err1 := time.Parse("20060102150405", latestWallpaper[:len("20060102150405")])
	newTime, err2 := time.Parse("20060102150405", newWallpaper[:len("20060102150405")])

	// If there's any error in parsing, don't set as wallpaper
	if err1 != nil || err2 != nil {
		return false
	}

	// If the new wallpaper's time is after the latest one's, it's the next in line
	return newTime.After(latestTime)
}

func getLatestWallpaper() (string, error) {
	data, err := os.ReadFile(LATEST_WALLPAPER_FILE)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func setLatestWallpaper(name string) error {
	return os.WriteFile(LATEST_WALLPAPER_FILE, []byte(name), 0644)
}

func saveHashes() error {
	hashesList := make([]string, 0, len(imageHashes))
	for hash := range imageHashes {
		hashesList = append(hashesList, hash)
	}

	data, err := json.Marshal(hashesList)
	if err != nil {
		return err
	}
	return os.WriteFile(HASHES_FILE, data, 0644)
}

func loadHashes() error {
	if _, err := os.Stat(HASHES_FILE); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(HASHES_FILE)
	if err != nil {
		return err
	}

	var hashesList []string
	if err := json.Unmarshal(data, &hashesList); err != nil {
		return err
	}

	for _, hash := range hashesList {
		imageHashes[hash] = struct{}{}
	}
	return nil
}

func computeHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func isDuplicateImage(data []byte) bool {
	hash := computeHash(data)
	_, exists := imageHashes[hash]
	return exists
}

func addImageHash(data []byte) {
	hash := computeHash(data)
	imageHashes[hash] = struct{}{}
}

func getRandomQuery() (string, error) {
	queriesBytes, err := os.ReadFile(QUERIES_FILE)
	if err != nil {
		return "", err
	}

	var queries []string
	if err := json.Unmarshal(queriesBytes, &queries); err != nil {
		return "", err
	}

	return queries[rand.Intn(len(queries))], nil
}

func downloadImage(url, filename string) error {
	fmt.Printf("Downloading image: %s\n", url)
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	// Check if the image is a duplicate
	if isDuplicateImage(data) {
		fmt.Printf("Duplicate image: %s\n", url)
		return fmt.Errorf("duplicate image")
	}

	// Add the hash of the image to the set
	addImageHash(data)

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

func setWallpaper(imagePath string) {
	fmt.Println("debug time", imagePath)
	start := time.Now()
	exec.Command(SWITCH_WALL_CMD, imagePath).Run()
	fmt.Println("debug time", time.Since(start))
}

func fetchSingleImage(index int) {
	defer wg.Done()

	query, err := getRandomQuery()
	if err != nil {
		fmt.Println("Error getting random query:", err)
		return
	}

	req, err := http.NewRequest("GET", SEARCH_API_ENDPOINT, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Add("Authorization", API_KEY)
	q := req.URL.Query()
	q.Add("query", query)
	q.Add("orientation", "landscape")
	q.Add("size", "large")
	q.Add("per_page", "1")
	q.Add("page", fmt.Sprintf("%d", rand.Intn(30)+1))
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}

	body, _ := io.ReadAll(response.Body)
	var data map[string]interface{}
	json.Unmarshal(body, &data)

	if photos, exists := data["photos"].([]interface{}); exists && len(photos) > 0 {
		photo := photos[0].(map[string]interface{})
		imageURL := photo["src"].(map[string]interface{})["original"].(string) + "?auto=compress&cs=tinysrgb&dpr=2&h=1600"
		imageName := fmt.Sprintf("%s-%v.jpg", time.Now().Format("20060102150405"), index)
		imagePath := filepath.Join(PICTURES_DIR, imageName)
		err := downloadImage(imageURL, imagePath)
		if err == nil && index == 0 {
			setWallpaper(imagePath)
			setLatestWallpaper(imageName)
		} else if err != nil {
			fmt.Println("Error downloading image:", err)
		}
	} else {
		fmt.Printf("No photos found for query: %s\n", query)
	}
}

func fetchImages() {
	// Handle the first image immediately
	wg.Add(1)
	fetchSingleImage(0)

	// Handle the rest asynchronously
	for i := 1; i < MAX_IMAGES; i++ {
		wg.Add(1)
		go fetchSingleImage(i)
	}

	wg.Wait()
}

func main() {
	err := loadHashes()
	if err != nil {
		fmt.Println("Error loading image hashes:", err)
		return
	}
	err = saveHashes()
	if err != nil {
		fmt.Println("Error saving image hashes:", err)
	}

	rand.Seed(time.Now().UnixNano())
	os.MkdirAll(PICTURES_DIR, os.ModePerm)

	if getNextWallpaperCount() < 2 {
		fmt.Println("Insufficient images in directory, fetching more...")
		fetchImages()
	} else {
		fmt.Println("Sufficient images in directory, skipping fetch...")
		// Set the next wallpaper
		files, err := os.ReadDir(PICTURES_DIR)
		if err != nil {
			fmt.Println("Error reading directory:", err)
			return
		}

		latestWallpaper, err := getLatestWallpaper()
		if err != nil {
			fmt.Println("Error getting latest wallpaper:", err)
			return
		}

		// Find the next wallpaper in line based on timestamps
		var nextWallpaper string
		// Similar logic with getNextWallpaperCount, sort files first
		var sortedFiles []string
		for _, file := range files {
			if filepath.Ext(file.Name()) == ".jpg" {
				sortedFiles = append(sortedFiles, file.Name())
			}
		}
		sort.Strings(sortedFiles)
		sort.Slice(sortedFiles, func(i, j int) bool {
			return sortedFiles[i] > sortedFiles[j]
		})

		for _, file := range sortedFiles {
			if file == latestWallpaper {
				break
			}
			nextWallpaper = file
		}

		if nextWallpaper != "" {
			fmt.Printf("Setting wallpaper: %s\n", nextWallpaper)
			setWallpaper(filepath.Join(PICTURES_DIR, nextWallpaper))
			setLatestWallpaper(nextWallpaper)
		}
	}
}
