# Wallfetcher

Wallfetcher is an automatic wallpaper manager that fetches beautiful high-quality images from Pexels, ensures no duplicates, and sets them as your desktop background. The program uses image hashing to prevent downloading of duplicate wallpapers and also keeps track of which wallpaper was set last to provide a seamless flow of beautiful backgrounds for your desktop.

**NOTE:** This project is fully made using GPT-4 (AI model by OpenAI) including README. It is done with ~20 prompts.

## Features:

- Automatically download and store wallpapers from Pexels.
- Ensure no duplicate wallpapers using image hashing.
- Set wallpapers on your desktop based on timestamps.
- Easily extendable with more sources or functionalities.

## Prerequisites:

1. Pexels API Key
2. Configured paths for image storage and utilities.

## Getting Started:

### You can either build it from source or use the binary in the [Releases](https://github.com/unitythemaker/Wallfetcher/releases).

### From release

1. Download the binary and put it in a directory that is within your PATH. (e.g. /usr/bin/wallfetcher)
2. Add a keybinding in your favourite Desktop Environment/Window Manager to execute the wallfetcher binary.
3. Set ``PEXELS_API_KEY`` environment variable with your API key.
4. Put some topics to fetch wallpapers. (Default topic path: ``~/.local/bin/data/pexels.json``), Example topic file:
    ```json
    ["art", "painting", "neon city", "neon", "abstract", "aesthetic", "conceptual", "motivational quotes"]
    ```
5. Start the repo and have fun!

### From source

1. **Clone the Repository**:
    ```bash
    git clone https://github.com/unitythemaker/Wallfetcher.git
    cd Wallfetcher
    ```
2. **Setup**:
   Set ``PEXELS_API_KEY`` environment variable with your API key and replace the paths with your desired paths respectively in the code.

3. Put some topics to fetch wallpapers. (Default topic path: ``~/.local/bin/data/pexels.json``), Example topic file:
    ```json
    ["art", "painting", "neon city", "neon", "abstract", "aesthetic", "conceptual", "motivational quotes"]
    ```

4. **Run**:
    ```bash
    go build -o wallfetcher .
    ```

This will fetch new images if required, set the latest wallpaper, and manage older ones.

## (Maybe) Future Improvements:

- Integrate with other image sources.
- Add user preferences for image categories or themes.
- Provide GUI/TUI for easier interaction and customization. (especially the topics file)
- Break up the code into parts.

## Contributing:

Feel free to fork this repository and enhance whatever you want, send a pull request anytime.
