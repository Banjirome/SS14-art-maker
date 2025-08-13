//go:build dnd
// +build dnd

package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/url"
	"strings"

	webview "github.com/webview/webview_go"
)

// rgb -> hex helper
func decHex(v uint32) string { return fmt.Sprintf("%02x", uint8(v)) }
func rgbToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return "#" + decHex(r>>8) + decHex(g>>8) + decHex(b>>8)
}

// Downscale if too large (simple box sampling)
func downscaleIfNeeded(img image.Image, maxDim int) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= maxDim && h <= maxDim {
		return img
	}
	// scale factor (ceiling)
	sfW := float64(w) / float64(maxDim)
	sfH := float64(h) / float64(maxDim)
	sf := sfW
	if sfH > sf {
		sf = sfH
	}
	newW := int(float64(w)/sf + 0.5)
	newH := int(float64(h)/sf + 0.5)
	// very naive average sampling
	type rgba struct{ R, G, B, A uint32 }
	out := image.NewNRGBA(image.Rect(0, 0, newW, newH))
	for ny := 0; ny < newH; ny++ {
		for nx := 0; nx < newW; nx++ {
			// source rectangle
			x0 := int(float64(nx) * sf)
			y0 := int(float64(ny) * sf)
			x1 := int(float64(nx+1) * sf)
			if x1 > w {
				x1 = w
			}
			y1 := int(float64(ny+1) * sf)
			if y1 > h {
				y1 = h
			}
			var rt, gt, bt, count uint32
			for sy := y0; sy < y1; sy++ {
				for sx := x0; sx < x1; sx++ {
					cr, cg, cb, _ := img.At(b.Min.X+sx, b.Min.Y+sy).RGBA()
					rt += cr >> 8
					gt += cg >> 8
					bt += cb >> 8
					count++
				}
			}
			if count == 0 {
				continue
			}
			out.Set(nx, ny, color.NRGBA{R: uint8(rt / count), G: uint8(gt / count), B: uint8(bt / count), A: 255})
		}
	}
	return out
}

func generateText(img image.Image, character string) string {
	if character == "" {
		character = "██"
	}
	maxSize := 8000.0
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	var sb strings.Builder
	for y := 0; y < h; y++ {
		last := ""
		for x := 0; x < w; x++ {
			hex := rgbToHex(img.At(b.Min.X+x, b.Min.Y+y))
			if hex != last && last != "" {
				sb.WriteString("[/color]")
			}
			if last == "" || hex != last {
				sb.WriteString("[color=" + hex + "]")
			}
			sb.WriteString(character)
			last = hex
		}
		sb.WriteString("[/color]\r\n")
	}
	_ = maxSize // оставлено на случай дальнейших метрик
	txt := sb.String()
	return txt
}

// convert base64 image (dataURL-safe) to text
func processBase64(b64 string, maxDim int, character string) (string, error) {
	// remove possible data URL prefix
	if i := strings.Index(b64, ","); i != -1 && strings.Contains(b64[:i], "base64") {
		b64 = b64[i+1:]
	}
	// URL decoding in case JS passed encoded
	if strings.Contains(b64, "%") {
		if uDec, err := url.QueryUnescape(b64); err == nil {
			b64 = uDec
		}
	}
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", errors.New("decode base64: " + err.Error())
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", errors.New("decode image: " + err.Error())
	}
	if maxDim <= 0 {
		maxDim = 100
	}
	if maxDim > 800 {
		maxDim = 800
	}
	img = downscaleIfNeeded(img, maxDim)
	return generateText(img, character), nil
}

func main() {
	html := `<!DOCTYPE html><html lang="ru"><head><meta charset="UTF-8" />
<title>Image2Text DnD</title>
<style>
body{font-family:Segoe UI,Arial,sans-serif;margin:0;padding:12px;background:#1e1e1e;color:#eee;}
#drop{border:2px dashed #555;border-radius:12px;padding:30px;text-align:center;font-size:15px;cursor:pointer;}
#drop.drag{border-color:#08f;background:#222;}
#layout{display:flex;flex-wrap:wrap;gap:18px;align-items:flex-start;margin-top:14px;}
#leftCol{max-width:280px;}
#preview{max-width:260px;max-height:260px;display:block;margin:14px auto 6px;border:1px solid #444;border-radius:8px;background:#000;}
#artPreview{flex:1;min-width:260px;max-width:480px;background:#111;color:#ddd;border:1px solid #444;border-radius:8px;padding:8px;font-family:Consolas,monospace;font-size:11px;line-height:11px;white-space:pre;overflow:auto;}
#out{width:100%;height:180px;background:#111;color:#ddd;border:1px solid #444;border-radius:6px;padding:8px;font-family:Consolas,monospace;font-size:11px;white-space:pre;}
button{background:#0b62d6;color:#fff;border:none;border-radius:6px;padding:8px 14px;font-size:14px;cursor:pointer;margin-right:8px;margin-top:6px;}
button:hover{background:#0d74ff;}
small{opacity:.6;display:block;margin-top:6px;}
label.block{display:block;margin-top:10px;font-size:13px;font-weight:600;}
.mono{font-family:Consolas,monospace;}
span.cc{display:inline;}
</style></head><body>
<h2>Перетащите изображение</h2>
<div id="drop">Drag & Drop или кликните для выбора</div>
<div id="layout">
	<div id="leftCol">
		<img id="preview" style="display:none"/>
				<div>
						<button id="genBtn" disabled>Генерировать</button>
						<button id="copyBtn" disabled>Копировать</button>
				</div>
				<div style="margin-top:8px;display:flex;flex-direction:column;gap:6px;">
					<label style="font-size:12px">Макс размер (px): <span id="maxDimVal">100</span></label>
					<input id="maxDim" type="range" min="20" max="400" value="100" />
					<label style="font-size:12px">Символ блока:
						<input id="charInput" type="text" value="██" maxlength="4" style="width:60px;margin-left:4px;" />
					</label>
				</div>
		<label class="block">BBCode результат:</label>
		<textarea id="out" placeholder="Результат..." readonly></textarea>
		<small id="status"></small>
	</div>
	<div style="flex:1;min-width:260px;">
		<label class="block">Превью:</label>
		<div id="artPreview" aria-label="preview"></div>
	</div>
</div>
<input type="file" id="fileInput" accept="image/*" style="display:none"/>
<script>
const drop = document.getElementById('drop');
const fileInput = document.getElementById('fileInput');
const preview = document.getElementById('preview');
const out = document.getElementById('out');
const genBtn = document.getElementById('genBtn');
const copyBtn = document.getElementById('copyBtn');
const statusEl = document.getElementById('status');
const artPreview = document.getElementById('artPreview');
const maxDimRange = document.getElementById('maxDim');
const maxDimVal = document.getElementById('maxDimVal');
const charInput = document.getElementById('charInput');
let currentBase64 = null;

function setStatus(t){ statusEl.textContent = t; }

function handleFile(f){
  if(!f) return;
  const reader = new FileReader();
  reader.onload = e => {
    currentBase64 = e.target.result;
    preview.src = currentBase64; preview.style.display='block';
    genBtn.disabled = false; setStatus('Файл загружен');
  };
  reader.readAsDataURL(f);
}

drop.addEventListener('click',()=>fileInput.click());
fileInput.addEventListener('change', e=>handleFile(e.target.files[0]));

drop.addEventListener('dragover', e=>{ e.preventDefault(); drop.classList.add('drag'); });
drop.addEventListener('dragleave', e=>{ drop.classList.remove('drag'); });
drop.addEventListener('drop', e=>{ e.preventDefault(); drop.classList.remove('drag'); handleFile(e.dataTransfer.files[0]); });

function escapeHtml(s){return s.replace(/[&<>]/g,c=>({"&":"&amp;","<":"&lt;",">":"&gt;"}[c]));}
function renderArt(bb){
	// Примитивный парсер: заменяем [color=#xxxxxx] на <span style="color:#xxxxxx"> и [/color] на </span>
	let safe = escapeHtml(bb);
	safe = safe.replace(/\[color=#[0-9a-fA-F]{6}\]/g, m=>{
		const col = m.match(/#[0-9a-fA-F]{6}/)[0];
		return '<span class="cc" style="color:'+col+'">';
	}).replace(/\[\/color\]/g,'</span>');
	// Удаляем служебную строку ;size=... если есть
	safe = safe.replace(/^;size=.*?%\r?\n/, '');
	artPreview.innerHTML = safe;
}

maxDimRange && maxDimRange.addEventListener('input',()=>{ maxDimVal.textContent = maxDimRange.value; });

genBtn.addEventListener('click', async ()=>{
  if(!currentBase64) return; setStatus('Обработка...'); genBtn.disabled=true;
  try {
		const dim = maxDimRange ? parseInt(maxDimRange.value,10) : 100;
		let ch = (charInput && charInput.value) ? charInput.value : '██';
		if(ch.length === 1) ch = ch.repeat(2); // для плотности
		const txt = await convertImage(currentBase64, dim, ch);
		out.value=txt; copyBtn.disabled=false; setStatus('Готово'); renderArt(txt);
  } catch(e){ setStatus('Ошибка: '+e); }
  genBtn.disabled=false;
});

copyBtn.addEventListener('click',()=>{ out.select(); document.execCommand('copy'); setStatus('Скопировано'); });
</script></body></html>`

	w := webview.New(true)
	defer w.Destroy()
	w.SetTitle("SS14 bbcode art maker")
	w.SetSize(500, 760, webview.HintNone)

	// Bind Go function with parameters: base64, maxDim, character
	w.Bind("convertImage", func(data string, maxDim int, character string) (string, error) {
		return processBase64(data, maxDim, character)
	})

	w.Navigate("data:text/html," + url.PathEscape(html))
	w.Run()
}
