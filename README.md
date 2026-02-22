# CompressMaps

A Go utility to compress Team Fortress 2 BSP map files using bzip2 compression and automatically split them into chunks if they exceed the size threshold.

## Features

- Compresses `.bsp` files using bzip2 compression
- Automatically splits compressed files larger than 25 MiB into 20 MiB chunks
- Colored terminal output for better readability
- Interactive overwrite confirmation
- Cross-platform support (Linux, Windows, macOS)

## Installation

### Prerequisites

- Go 1.21 or higher

### Build from source

```bash
# Clone the repository
git clone https://github.com/UDL-TF/CompressMaps.git
cd CompressMaps

# Download dependencies
go mod download

# Build the executable
go build -o compressmaps

# Or build for specific platforms
# Linux
GOOS=linux GOARCH=amd64 go build -o compressmaps-linux

# Windows
GOOS=windows GOARCH=amd64 go build -o compressmaps.exe

# macOS
GOOS=darwin GOARCH=amd64 go build -o compressmaps-macos
```

## Usage

### Command Line

```bash
./compressmaps <map_file.bsp>
```

### Example

```bash
./compressmaps pl_upward.bsp
```

### Drag and Drop

You can also drag and drop a `.bsp` file onto the executable (works on Windows and some Linux desktop environments).

## Output

The script will:

1. Create a compressed `.bsp.bz2` file
2. If the compressed file exceeds 25 MiB:
   - Create a `.bsp.bz2.parts` directory
   - Split the compressed file into 20 MiB chunks named `<filename>.bsp.bz2.part.000`, `.part.001`, etc.
   - Remove the original large compressed file

## Configuration

You can modify these constants in `main.go`:

- `splitThresholdBytes`: Size threshold for splitting (default: 26214400 bytes = 25 MiB)
- `splitChunkSize`: Size of each chunk when splitting (default: 20 MiB)

## License

See [LICENSE](LICENSE) file for details.
Script to compress maps to bzip.
