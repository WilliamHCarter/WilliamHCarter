package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

const svgTemplate = `<svg width="400" height="200" xmlns="http://www.w3.org/2000/svg">
    <defs>
        <filter id="glow" x="-50%" y="-50%" width="200%" height="200%">
            <feFlood result="flood" flood-color="#{{.TextColor}}" flood-opacity=".4"/>
            <feComposite in="flood" result="mask" in2="SourceGraphic" operator="in"/>
            <feMorphology in="mask" result="dilated" operator="dilate" radius="1"/>
            <feGaussianBlur in="dilated" result="blurred" stdDeviation="1.5"/>
            <feMerge>
                <feMergeNode in="blurred"/>
                <feMergeNode in="SourceGraphic"/>
            </feMerge>
        </filter>
        
        <pattern id="crtPattern" x="0" y="0" width="3" height="2" patternUnits="userSpaceOnUse">
            <rect width="3" height="1" fill="rgba(18, 16, 16, 0)"/>
            <rect width="3" height="1" y="1" fill="rgba(0, 0, 0, 0.25)"/>
            <rect width="1" height="2" fill="rgba(255, 0, 0, 0.06)"/>
            <rect width="1" height="2" x="1" fill="rgba(0, 255, 0, 0.02)"/>
            <rect width="1" height="2" x="2" fill="rgba(0, 0, 255, 0.06)"/>
        </pattern>

        <radialGradient id="vignette" cx="50%" cy="50%" r="50%" fx="50%" fy="50%">
            <stop offset="0%" style="stop-color:rgba(59,36,13,1)" />
            <stop offset="80%" style="stop-color:rgba(36,22,6,1)" />
            <stop offset="100%" style="stop-color:rgba(20,12,4,1)" />
        </radialGradient>

		<linearGradient id="scanlineGradient" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" style="stop-color:rgba(0,0,0,0.025)" />
            <stop offset="90%" style="stop-color:rgba(255,255,255,0.05)" />
            <stop offset="100%" style="stop-color:rgba(0,0,0,0)" />
        </linearGradient>
    </defs>
    
    <style>
        .text { 
            fill: #{{.TextColor}}; 
            font-family: monospace; 
            font-size: 9px;
            font-weight: bold;
            filter: url(#glow);
            opacity: 0.8;
        }
        @keyframes scanline {
            0% {
                transform: translateY(-100%);
            }
            100% {
                transform: translateY(200%);
            }
        }
		.scanline {
			animation: scanline 6s cubic-bezier(0.4, 0.0, 0.2, 1) infinite;
		}
    </style>
    
    <!-- Background with vignette effect -->
    <rect width="100%" height="100%" fill="url(#vignette)" rx="6" ry="6"/>
    
    <clipPath id="rounded-corners">
        <rect width="100%" height="100%" rx="6" ry="6"/>
    </clipPath>
    
    <g clip-path="url(#rounded-corners)">
        <!-- ASCII Art Text layer -->
        <text x="10" y="60" class="text" xml:space="preserve">{{.Text}}</text>

        <!-- CRT overlay -->
        <rect width="100%" height="100%" fill="url(#crtPattern)" style="mix-blend-mode: overlay;"/>

        <!-- Scanline effect -->
        <rect class="scanline" x="0" y="0" width="400" height="50" fill="url(#scanlineGradient)" style="mix-blend-mode: overlay;"/>
    </g></svg>`

type SVGData struct {
	Text            template.HTML
	BackgroundColor string
	TextColor       string
}

func Handler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("svg").Parse(svgTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	text := r.URL.Query().Get("text")
	if text == "" {
		text = `
 _       ___ _____                    ______           __           
| |     / (_) / (_)___ _____ ___     / ____/___ ______/ /____  _____
| | /| / / / / / / __ ` + "`" + `/ __ ` + "`" + `__ \   / /   / __ ` + "`" + `/ ___/ __/ _ \/ ___/
| |/ |/ / / / / / /_/ / / / / / /  / /___/ /_/ / /  / /_/  __/ /    
|__/|__/_/_/_/_/\__,_/_/ /_/ /_/   \____/\__,_/_/   \__/\___/_/     
                                                                    
            `
	}

	// Process the text to preserve whitespace and line breaks
	lines := strings.Split(text, "\n")
	processedText := ""
	for i, line := range lines {
		line = strings.ReplaceAll(line, " ", "&#160;")
		if i > 0 {
			processedText += "<tspan x=\"10\" dy=\"1.2em\">" + line + "</tspan>"
		} else {
			processedText += line
		}
	}

	asciiBox := createASCIIBox("Info", "Lorem ipsum dolor sit amet", "Consectetur adipiscing elit", "Sed do eiusmod tempor incididunt")
	asciiBoxLines := strings.Split(asciiBox, "\n")
	for _, line := range asciiBoxLines {
		line = strings.ReplaceAll(line, " ", "&#160;")
		processedText += "<tspan x=\"10\" dy=\"1.2em\">" + line + "</tspan>"
	}

	backgroundColor := r.URL.Query().Get("background_color")
	if backgroundColor == "" {
		backgroundColor = "323f26"
	}

	textColor := r.URL.Query().Get("text_color")
	if textColor == "" {
		textColor = "F69525"
	}

	data := SVGData{
		Text:            template.HTML(processedText),
		BackgroundColor: backgroundColor,
		TextColor:       textColor,
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func createASCIIBox(title, line1, line2, line3 string) string {
	boxWidth := 50
	titlePadding := (boxWidth - len(title) - 2) / 2
	line1Padding := boxWidth - len(line1) - 4
	line2Padding := boxWidth - len(line2) - 4
	line3Padding := boxWidth - len(line3) - 4

	return fmt.Sprintf(`
    ┌%s %s %s┐
    │ %s%s │
    │ %s%s │
    │ %s%s │
    └%s┘`,
		strings.Repeat("─", titlePadding), title, strings.Repeat("─", boxWidth-titlePadding-len(title)-2),
		line1, strings.Repeat(" ", line1Padding),
		line2, strings.Repeat(" ", line2Padding),
		line3, strings.Repeat(" ", line3Padding),
		strings.Repeat("─", boxWidth-2))
}
