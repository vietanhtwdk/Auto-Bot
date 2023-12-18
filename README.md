<!-- ABOUT THE PROJECT -->
## About The Project

An auto-click tool to automate tasks on windows

Main features:
* Image detection
* Simple script record
* Simple script editor

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- GETTING STARTED -->
## Getting Started
### Installation
#### Download
Check release page: https://github.com/vietanhtwdk/Auto-Bot/releases
#### Build locally
1. Install golang 1.20+ https://go.dev/
2. Install fyne2 https://fyne.io/
3. Install dependency
   ```sh
   go mod tidy
   ```
4. Build move mouse smooth
   ```sh
   cd move_mouse_smooth
   go build -ldflags -H=windowsgui -o ./MoveMouseSmooth.exe .
   ```
5. Build main
   ```sh
   cd move_mouse_smooth
   go build -ldflags -H=windowsgui -o ./AutoBot.exe .
   ```
<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- USAGE EXAMPLES -->
## Usage

Record simple: https://youtu.be/sHBBAGFQdrU

Script editor: https://youtu.be/RRqe50U_i8s

<p align="right">(<a href="#readme-top">back to top</a>)</p>