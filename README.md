# 🎬 yt-playlist-sorter

A simple command-line tool to easily **sort any YouTube playlist by original upload date**. Creates a new **private** playlist with the videos in chronological order.


---

## 🚀 Features

- 🔗 Accepts any YouTube playlist URL
- 🧹 Sorts videos from **oldest to newest**  
- 🛡 Creates a new **private playlist** with the sorted order  
- 🛠 Uses OAuth2 for secure authentication  

---

## 📦 Installation

Make sure you have Go installed (version 1.18+ recommended).

```bash
git clone https://github.com/your-username/yt-playlist-sorter.git
cd yt-playlist-sorter
go mod tidy
go build -o yt-playlist-sorter
```

---

## 🛠 Initial Setup

Before using the tool, set up your **YouTube API credentials**:

```bash
./yt-playlist-sorter setup
```

You'll be prompted to enter your **Client ID** and **Client Secret**.  
These are obtained by creating OAuth 2.0 credentials for a **Desktop App** in your [Google Cloud Console](https://console.cloud.google.com/).

📝 The tool will save your credentials to `config.json` and fetch an OAuth token via a browser flow.

---

## ✅ Usage

To sort a playlist:

```bash
./yt-playlist-sorter sort "https://www.youtube.com/playlist?list=YOUR_PLAYLIST_ID"
```

The tool will create a new private playlist with videos in chronological order.
