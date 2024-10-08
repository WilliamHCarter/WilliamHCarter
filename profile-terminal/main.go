package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// func main() {
// 	http.HandleFunc("/", Handler)
// 	log.Println("Server starting on http://localhost:8080")
// 	log.Fatal(http.ListenAndServe(":8080", nil))
// }

const svgTemplate = `<svg width="400" height="280" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
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
		.fmt {
			font-family: monospace; 
			font-size: 9px;
			font-weight: bold;
		}
        .text { 
			font-family: monospace; 
			font-size: 9px;
			font-weight: bold;
            fill: #{{.TextColor}}; 
            filter: url(#glow);
            opacity: 0.8;
        }
        .highlight-group rect:hover {
            opacity: 0.2;
            transition: opacity 0.3s ease-in-out;
			fill: #F69525;
        }
		.highlight-group rect {
            transition: opacity 0.3s ease-in-out;
        }
		.highlight-group text {
            pointer-events: none;
        }
        .highlight-group a {
            cursor: pointer;
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
		<g class="highlight-group">
			{{.ProjectRects}}
			<text x="50%" y="210" text-anchor="middle" class="fmt" xml:space="preserve">{{.ProjectText}}</text>
		</g>

        <!-- CRT overlay -->
        <rect width="100%" height="100%" fill="url(#crtPattern)" style="mix-blend-mode: overlay; pointer-events: none;"/>

        <!-- Scanline effect -->
        <rect class="scanline" x="0" y="0" width="400" height="50" fill="url(#scanlineGradient)" style="mix-blend-mode: overlay; pointer-events: none;"/>
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
	ProjectRects    template.HTML
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

	languageLines, err := GetTopThreeLanguages("WilliamHCarter")
	if err != nil {
		logger.Printf("Error fetching top languages: %v\n", err)
		return
	}

	infoLines := append([]string{commitsLine, "Top Languages:"}, languageLines...)
	infoBox := createInfoBox("Info", infoLines)

	projectLinks := []string{"https://github.com/WilliamHCarter/zfetch", "https://github.com/WilliamHCarter/RattlesnakeRidge", "https://github.com/WilliamHCarter/LyreMusicPlayer"}
	projectBox, projectRects := createProjectBox("Projects", []string{"zfetch", "Rattlesnake Ridge", "Lyre Music Player"}, projectLinks)

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
		ProjectRects:    template.HTML(projectRects),
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
		if i == 0 {
			paddedLines[i] = fmt.Sprintf("│ %s%s │", line, strings.Repeat(" ", linePadding))
		} else {
			paddedLines[i] = fmt.Sprintf(" %s%s │", line, strings.Repeat(" ", linePadding))
		}
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

func createProjectBox(title string, lines []string, links []string) (string, string) {
	boxWidth := 50
	titlePadding := (boxWidth - len(title) - 2) / 2

	paddedLines := make([]string, len(lines))
	for i, line := range lines {
		linePadding := boxWidth - len(line) - 2
		if i == 0 {
			paddedLines[i] = fmt.Sprintf("│ %s%s │", line, strings.Repeat(" ", linePadding))
		} else {
			paddedLines[i] = fmt.Sprintf(" %s%s │", line, strings.Repeat(" ", linePadding))
		}
	}

	infoBox := fmt.Sprintf(`┌%s %s %s┐
%s
└%s┘`,
		strings.Repeat("─", titlePadding), title, strings.Repeat("─", boxWidth-titlePadding-len(title)-2),
		strings.Join(paddedLines, "\n│"),
		strings.Repeat("─", boxWidth))

	processedInfoBox := ""
	rects := ""

	infoBoxLines := strings.Split(infoBox, "\n")
	for i, line := range infoBoxLines {
		line = strings.ReplaceAll(line, " ", "&#160;")
		if i > 0 && i <= len(lines) {
			line = fmt.Sprintf("<a xlink:href=\"%s\" class=\"text\">%s</a>", links[i-1], line)
			processedInfoBox += "<tspan x=\"50%\" dy=\"1.2em\">" + line + "</tspan>"
			rects += fmt.Sprintf("<a href=\"%s\">"+"<rect x=\"73.5\" y=\"%d\" width=\"253\" height=\"11\" fill=\"#F69525\" opacity=\"0\"/>"+"</a>", links[i-1], 212+i*11)

		} else {
			processedInfoBox += "<tspan x=\"50%\" dy=\"1.2em\" class=\"text\">" + line + "</tspan>"
		}
	}
	return processedInfoBox, rects
}

// =========================== Queries ===========================
func GetTotalCommits(username string) (int, error) {
	logger.Printf("Fetching commit count for user: %s\n", username)

	url := fmt.Sprintf("https://api.github.com/search/commits?q=author:%s", username)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Printf("Error creating request: %v\n", err)
		return 0, err
	}

	req.Header.Set("User-Agent", "GetCommitsAgent")
	req.Header.Set("Accept", "application/vnd.github.cloak-preview")

	logger.Println("Sending request to GitHub API...")
	resp, err := client.Do(req)
	if err != nil {
		logger.Printf("Error sending request: %v\n", err)
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Printf("API request failed with status code: %d\n", resp.StatusCode)
		return 0, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var result struct {
		TotalCount int `json:"total_count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Printf("Error decoding response: %v\n", err)
		return 0, err
	}

	logger.Printf("Total commits counted: %d\n", result.TotalCount)
	return result.TotalCount, nil
}

type Language struct {
	Name       string
	Percentage float64
}

type LanguageEdge struct {
	Size int `json:"size"`
	Node struct {
		Name string `json:"name"`
	} `json:"node"`
}

type Repository struct {
	Languages struct {
		Edges []LanguageEdge `json:"edges"`
	} `json:"languages"`
}

type GraphQLResponse struct {
	Data struct {
		User struct {
			Repositories struct {
				Nodes []Repository `json:"nodes"`
			} `json:"repositories"`
		} `json:"user"`
	} `json:"data"`
}

func GetTopThreeLanguages(username string) ([]string, error) {
	excludeLanguages := []string{"C#", "ShaderLab"}
	url := "https://api.github.com/graphql"
	query := fmt.Sprintf(`
	{
		user(login: "%s") {
			repositories(ownerAffiliations: OWNER, isFork: false, first: 100) {
				nodes {
					languages(first: 10, orderBy: {field: SIZE, direction: DESC}) {
						edges {
							size
							node {
								name
							}
						}
					}
				}
			}
		}
	}
	`, username)

	client := &http.Client{Timeout: 10 * time.Second}
	reqBody := map[string]string{"query": query}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	token := os.Getenv("GITHUB_TOKEN") //Set through vercel
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable is not set")
	}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var result GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	languageCounts := make(map[string]int)
	totalSize := 0

	for _, repo := range result.Data.User.Repositories.Nodes {
		for _, edge := range repo.Languages.Edges {
			if !contains(excludeLanguages, edge.Node.Name) {
				languageCounts[edge.Node.Name] += edge.Size
				totalSize += edge.Size
			}
		}
	}
	languages := make([]Language, 0, len(languageCounts))
	for name, size := range languageCounts {
		percentage := float64(size) / float64(totalSize) * 100
		languages = append(languages, Language{Name: name, Percentage: percentage})
	}

	sort.Slice(languages, func(i, j int) bool {
		return languages[i].Percentage > languages[j].Percentage
	})

	if len(languages) > 3 {
		languages = languages[:3]
	}

	languageLines := make([]string, len(languages))
	for i, lang := range languages {
		languageLines[i] = fmt.Sprintf("%s: %.2f%%", lang.Name, lang.Percentage)
	}

	return addBarChart(languages), nil
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func addBarChart(languages []Language) []string {
	const totalBarLength = 20
	const fullBlock = "#"
	const emptyBlock = "-"

	maxNameLength := 0
	for _, lang := range languages {
		if len(lang.Name) > maxNameLength {
			maxNameLength = len(lang.Name)
		}
	}
	log.Printf("Max name length: %d", maxNameLength)

	var chartLines []string
	for _, lang := range languages {
		solidBlocks := int(math.Max(0, math.Floor(lang.Percentage/100*float64(totalBarLength))))
		log.Printf("Language: %s, Percentage: %.2f%%, Solid blocks: %d", lang.Name, lang.Percentage, solidBlocks)

		bar := make([]byte, totalBarLength)
		for i := 0; i < totalBarLength; i++ {
			if i < solidBlocks {
				bar[i] = fullBlock[0]
			} else {
				bar[i] = emptyBlock[0]
			}
		}

		line := fmt.Sprintf("  %-*s [%s] %.2f%%", maxNameLength, lang.Name, string(bar), lang.Percentage)
		log.Printf("Formatted line: %s", line)
		chartLines = append(chartLines, line)
	}

	return chartLines
}
