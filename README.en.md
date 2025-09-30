<div align="center">

# SS14 Art Maker

Convert images into colorful BBCode art for Space Station 14 (and any BBCode target) via a simple drag & drop GUI.

**Languages:** [Русский](README.md) | **English**

![UI Screenshot](docs/images/preview.png)

</div>

## ✨ Features

- Drag & Drop or file picker (PNG / JPG / GIF – first frame only).
- Ctrl + V to paste images from the clipboard.
- Automatic downscale to selected max dimension (slider).
- Colored BBCode output using `[color=#RRGGBB]` tags.
- Live preview inside the window.
- Custom block character (default double `██` for density).
- Naive average box resampling for large images.
- One‑click copy to clipboard.

## 🖼 Sample

```
[color=#ff0000]██[/color][color=#00ff00]██[/color][color=#0000ff]██[/color]
```

## 🚀 Download

Grab the latest `art_maker.exe` from Releases (tags `v*`).

## 🏗 Build

Requires Go (see `go.mod`). Source guarded by build tag `dnd`:

```powershell
go build -tags dnd -o art_maker.exe .
./art_maker.exe
```

## ⚙️ Technical Notes

- Entirely local processing; no network.
- Downscale: simple averaged box sampling.
- Color extraction: direct RGBA -> hex.

## 🐞 Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| Blank window | WebView init blocked | Restart / check AV |
| Gray output | Image palette / transparency | Try another image |
| Huge output | Large image + high maxDim | Lower maxDim |

## 💡 Tips

- Single character looks sparse; double block denser.
- High maxDim grows output fast; balance size vs detail.

## 🤝 Contributing

PRs welcome. Fork → branch → commit → PR. Issues for ideas.

---

If this tool helps you, leave a ⭐.
