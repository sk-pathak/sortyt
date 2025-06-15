# sortyt- A Simple Youtube Playlist Sorter

A simple command-line tool to easily **sort any YouTube playlist by original upload date**. Creates a new **private** playlist with the videos in chronological order.


---

## Features

- 🔗 Accepts any YouTube playlist URL
- 🧹 Sorts videos from **oldest to newest**  
- 🛡 Creates a new **private playlist** with the sorted order  
- 🛠 Uses OAuth2 for secure authentication  

---

## Installation

```bash
go install https://github.com/sk-pathak/sortyt.git
```

---

## Build

Make sure you have Go installed (version 1.18+ recommended).

```bash
git clone https://github.com/sk-pathak/sortyt.git
cd sortyt
go mod tidy
go build -o sortyt
```

---

## Initial Setup

Before using the tool, set up your **YouTube API credentials**:

```bash
./sortyt setup
```

You'll be prompted to enter your **Client ID** and **Client Secret**.  
These are obtained by creating OAuth 2.0 credentials for a **Desktop App** in your [Google Cloud Console](https://console.cloud.google.com/).

The tool will save your credentials to `$XDG_CONFIG_HOME/sortyt/config.json` and fetch an OAuth token via a browser flow. Token is also stored in same directory.

---

## Usage

To sort a playlist:

```bash
./sortyt sort "https://www.youtube.com/playlist?list=YOUR_PLAYLIST_ID"
```

The tool will create a new private playlist with videos in chronological order.

## Next Steps

- Store client ID, client secret in ~/.config/sortyt, encrypted
- 
- Use secret key to decrypt (secret key to be inserted by user each time)
- close window automatically (maybe use xdg-open & not clickable link)

- use keyrings if possible, otherwise fallback to encryption (https://github.com/zalando/go-keyring)
