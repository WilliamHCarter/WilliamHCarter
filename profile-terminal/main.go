package handler

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const svgTemplate = `<svg width="400" height="280" xmlns="http://www.w3.org/2000/svg">
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
        <text x="50%" y="60" class="text" text-anchor="middle" xml:space="preserve">{{.Text}}</text>

        <!-- Info Box Text layer -->
        <text x="50%" y="140" class="text" text-anchor="middle" xml:space="preserve">{{.InfoText}}</text>

		<!-- Project Box Text layer -->]
		<text x="50%" y="210" class="text" text-anchor="middle" xml:space="preserve">{{.ProjectText}}</text>

        <!-- CRT overlay -->
        <rect width="100%" height="100%" fill="url(#crtPattern)" style="mix-blend-mode: overlay;"/>

        <!-- Scanline effect -->
        <rect class="scanline" x="0" y="0" width="400" height="50" fill="url(#scanlineGradient)" style="mix-blend-mode: overlay;"/>
    </g>
	<script type="text/javascript">
		<![CDATA[
			function visit(url) {
				window.top.location.href = url;
			}
		]]>
	</script>
</svg>`

type SVGData struct {
	Text            template.HTML
	InfoText        template.HTML
	ProjectText     template.HTML
	BackgroundColor string
	TextColor       string
}

type LanguageStats struct {
	Name       string
	Percentage float64
}

var (
	cachedCommitCount int
	lastUpdateTime    time.Time
	mutex             sync.Mutex
	logger            *log.Logger
)

func init() {
	logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	logger.Println("Handler function called")

	mutex.Lock()
	defer mutex.Unlock()

	if time.Since(lastUpdateTime) > 24*time.Hour {
		logger.Println("Updating commit count...")
		count, err := GetTotalCommits("WilliamHCarter")
		if err == nil {
			cachedCommitCount = count
			lastUpdateTime = time.Now()
			logger.Printf("Commit count updated successfully: %d\n", count)
		} else {
			logger.Printf("Error updating commit count: %v\n", err)
		}
	} else {
		logger.Println("Using cached commit count")
	}

	logger.Printf("Current commit count: %d\n", cachedCommitCount)

	tmpl, err := template.New("svg").Parse(svgTemplate)
	if err != nil {
		logger.Printf("Error parsing SVG template: %v\n", err)
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
			processedText += "<tspan x=\"50%\" dy=\"1.2em\">" + line + "</tspan>"
		} else {
			processedText += line
		}
	}

	commitsLine := fmt.Sprintf("Total Commits: %d", cachedCommitCount)
	logger.Printf("Current cached commit count: %d\n", cachedCommitCount)

	infoLines := []string{commitsLine, "Lorem ipsum dolor sit amet", "Consectetur adipiscing elit", "Sed do eiusmod tempor incididunt"}
	infoBox := createInfoBox("Info", infoLines)

	projectLinks := []string{"https://github.com/WilliamHCarter/zfetch", "https://github.com/WilliamHCarter/RattlesnakeRidge", "https://github.com/WilliamHCarter/LyreMusicPlayer"}
	projectBox := createProjectBox("Projects", []string{"zfetch", "Rattlesnake Ridge", "Lyre Music Player"}, projectLinks)

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
		InfoText:        template.HTML(infoBox),
		ProjectText:     template.HTML(projectBox),
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

func createInfoBox(title string, lines []string) string {
	boxWidth := 50
	titlePadding := (boxWidth - len(title) - 2) / 2

	paddedLines := make([]string, len(lines))
	for i, line := range lines {
		linePadding := boxWidth - len(line) - 2
		paddedLines[i] = fmt.Sprintf(" %s%s │", line, strings.Repeat(" ", linePadding))
	}

	infoBox := fmt.Sprintf(`┌%s %s %s┐
%s
└%s┘`,
		strings.Repeat("─", titlePadding), title, strings.Repeat("─", boxWidth-titlePadding-len(title)-2),
		strings.Join(paddedLines, "\n│"),
		strings.Repeat("─", boxWidth))

	processedInfoBox := ""

	infoBoxLines := strings.Split(infoBox, "\n")
	for _, line := range infoBoxLines {
		line = strings.ReplaceAll(line, " ", "&#160;")
		processedInfoBox += "<tspan x=\"50%\" dy=\"1.2em\">" + line + "</tspan>"
	}
	return processedInfoBox
}

func createProjectBox(title string, lines []string, links []string) string {
	boxWidth := 50
	titlePadding := (boxWidth - len(title) - 2) / 2

	paddedLines := make([]string, len(lines))
	for i, line := range lines {
		linePadding := boxWidth - len(line) - 2
		paddedLines[i] = fmt.Sprintf(" %s%s │", line, strings.Repeat(" ", linePadding))
	}

	infoBox := fmt.Sprintf(`┌%s %s %s┐
%s
└%s┘`,
		strings.Repeat("─", titlePadding), title, strings.Repeat("─", boxWidth-titlePadding-len(title)-2),
		strings.Join(paddedLines, "\n│"),
		strings.Repeat("─", boxWidth))

	processedInfoBox := ""

	infoBoxLines := strings.Split(infoBox, "\n")
	for i, line := range infoBoxLines {
		line = strings.ReplaceAll(line, " ", "&#160;")
		if i > 0 && i <= len(lines) {
			line = fmt.Sprintf("<a xlink:href=\"%s\">%s</a>", links[i-1], line)
		}
		processedInfoBox += "<tspan x=\"50%\" dy=\"1.2em\">" + line + "</tspan>"
	}
	return processedInfoBox
}
