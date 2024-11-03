package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/color"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
)

type Data struct {
	Type          string       `json:"type" validate:"required"`
	Content       string       `json:"content" validate:"required"`
	SubContent    string       `json:"subContent,omitempty"`
	StatusColor   string       `json:"statusColor,omitempty"`
	ImageURL      string       `json:"imageUrl" validate:"required"`
	BackgroundURL string       `json:"backgroundUrl" validate:"required"`
	TextColor     string       `json:"textColor,omitempty"`
	AddOverlay    bool         `json:"addOverlay,omitempty"`
	OverlayColor  string       `json:"overlayColor,omitempty"`
	RankInfo      *RankInfo    `json:"rankInfo,omitempty"`
	SpotifyInfo   *SpotifyInfo `json:"spotifyInfo,omitempty"`
	Leaderboard   []UserEntry  `json:"leaderboard,omitempty"`
}

type RankInfo struct {
	TypeContent    string `json:"typeContent,omitempty"`
	RatioString    string `json:"ratioString" validate:"required"`
	RatioPercent   int    `json:"ratioPercent" validate:"required"`
	RatioPlacement string `json:"ratioPlacement,omitempty"`
	ProgressColor  string `json:"progressColor,omitempty"`
	EmptyColor     string `json:"emptyColor,omitempty"`
}

type SpotifyInfo struct {
	SongTitle  string `json:"songTitle" validate:"required"`
	ArtistName string `json:"artistName,omitempty"`
	AlbumCover string `json:"albumCover,omitempty"`
}

type UserEntry struct {
	DisplayName string `json:"displayName" validate:"required"`
	ImageURL    string `json:"imageUrl" validate:"required"`
	Rank        int    `json:"rankOrLevel" validate:"required"`
	XP          int    `json:"xpOrMessages" validate:"required"`
}

func handleCard(c *gin.Context, data Data, size int) []byte {
	svgTemplate := getSvgTemplate(data)
	if svgTemplate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "error": "Invalid card type."})
		return nil
	}

	defaultValues(&data)

	tmpl, err := template.New("Generated").Funcs(template.FuncMap{
		"add":          func(a, b int) int { return a + b },
		"mul":          func(a, b int) int { return a * b },
		"len":          func(slice []UserEntry) int { return len(slice) },
		"calcProgress": func(percentage int) int { return int(math.Round(float64(percentage) * 270 / 100)) },
		"sliceText": func(text string, length int) string {
			if len(text) <= length {
				return text
			}

			return text[:length] + ".."
		},
		"placement": func(condition string) string {
			if condition == "top" {
				return "60"
			}

			return "100"
		},
		"fetchImage": func(url string) string {
			cleanedURL := removeControlCharacters(url)

			if cacheItem, found := getCachedImage(cleanedURL); found {
				return fmt.Sprintf("data:%s;base64,%s", cacheItem.ContentType, base64.StdEncoding.EncodeToString(cacheItem.ImageData))
			}

			imageData, contentType, err := fetchImage(cleanedURL)
			if err != nil {
				return ""
			}

			cacheImage(cleanedURL, imageData, contentType)
			return fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(imageData))
		},
	}).Parse(svgTemplate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "error": "Error parsing SVG template.", "details": err.Error()})
		return nil
	}

	var svgBuffer bytes.Buffer
	if err := tmpl.Execute(&svgBuffer, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "error": "Error executing SVG template.", "details": err.Error()})
		return nil
	}

	webpData, err := convertSVGToWEBP(svgBuffer.String(), size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "error": "Error creating SVG.", "details": err.Error()})
		return nil
	}

	return webpData
}

func defaultValues(data *Data) {
	if data.TextColor == "" {
		data.TextColor = "#ffffff"
	}

	if data.RankInfo != nil {
		if data.RankInfo.ProgressColor == "" {
			data.RankInfo.ProgressColor = "#e03131"
		}

		if data.RankInfo.EmptyColor == "" {
			data.RankInfo.EmptyColor = "#6741d9"
		}
	}
}

func convertSVGToWEBP(svgData string, width int) ([]byte, error) {
	tmpSVGFile, err := os.CreateTemp("", "temp-*.svg")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpSVGFile.Name())

	if _, err := tmpSVGFile.Write([]byte(svgData)); err != nil {
		return nil, err
	}

	tmpSVGFile.Close()

	tmpPNGFile, err := os.CreateTemp("", "temp-*.png")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpPNGFile.Name())

	cmd := exec.Command("rsvg-convert", "-w", fmt.Sprintf("%d", width), tmpSVGFile.Name(), "-o", tmpPNGFile.Name())
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	var outBuffer bytes.Buffer
	cmd = exec.Command("cwebp", tmpPNGFile.Name(), "-o", "/dev/stdout", "-q", "100")
	cmd.Stdout = &outBuffer
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return outBuffer.Bytes(), nil
}

func convertColorToRgba(input string) string {
	var c color.Color

	if strings.HasPrefix(input, "#") {
		c = parseHexColor(input)
	} else {
		c = parseRGBColor(input)
	}

	if c == nil {
		return "rgba(0,0,0,0.2)"
	}

	r, g, b, _ := c.RGBA()
	return "rgba(" + strconv.Itoa(int(r>>8)) + "," + strconv.Itoa(int(g>>8)) + "," + strconv.Itoa(int(b>>8)) + ",0.2)"
}

func getSvgTemplate(data Data) string {
	if data.OverlayColor == "" {
		data.OverlayColor = "#000000"
	}

	overlayColor := convertColorToRgba(data.OverlayColor)
	overlayStyle := `{{ if .AddOverlay }}<rect x="10" y="10" width="380" height="100" fill="` + overlayColor + `" rx="10" ry="10"/>{{ end }}`

	switch data.Type {
	case "welcome":
		return `
			<svg width="400" height="120" xmlns="http://www.w3.org/2000/svg">
				<defs>
					<pattern id="background" patternUnits="userSpaceOnUse" width="400" height="120">
						<image href="{{fetchImage .BackgroundURL}}" x="0" y="0" width="400" height="120" preserveAspectRatio="none"/>
					</pattern>
				</defs>
				<rect width="100%" height="100%" fill="url(#background)" />
				` + overlayStyle + `
				<defs>
					<clipPath id="avatarClip">
						<circle cx="60" cy="60" r="40" />
					</clipPath>
				</defs>
				<image href="{{fetchImage .ImageURL}}" x="20" y="20" width="80" height="80" clip-path="url(#avatarClip)" />
				<text x="110" y="{{if .SubContent}}55{{else}}60{{end}}" alignment-baseline="middle" font-size="24" fill="{{.TextColor}}" font-family="Arial, sans-serif">
					{{sliceText .Content 25}}
				</text>
				{{if .SubContent}}
					<text x="110" y="80" alignment-baseline="middle" font-size="14" fill="{{.TextColor}}" font-family="Arial, sans-serif" opacity="0.5">
						{{sliceText .SubContent 45}}
					</text>
				{{end}}
			</svg>
		`
	case "rank":
		return `
			<svg width="400" height="120" xmlns="http://www.w3.org/2000/svg">
				<defs>
					<pattern id="background" patternUnits="userSpaceOnUse" width="400" height="120">
						<image href="{{fetchImage .BackgroundURL}}" x="0" y="0" width="400" height="120" preserveAspectRatio="none"/>
					</pattern>
					<linearGradient id="progressGradient" x1="0%" y1="0%" x2="100%" y2="0%">
						<stop offset="0%" style="stop-color: {{.RankInfo.ProgressColor}}; stop-opacity: 1" />
						<stop offset="90%" style="stop-color: {{.RankInfo.ProgressColor}}; stop-opacity: 1" />
						<stop offset="100%" style="stop-color: {{.RankInfo.EmptyColor}}; stop-opacity: 1" />
					</linearGradient>
				</defs>
				<rect width="100%" height="100%" fill="url(#background)" />
				` + overlayStyle + `
				<defs>
					<clipPath id="avatarClip">
						<circle cx="60" cy="60" r="40" />
					</clipPath>
				</defs>
				{{if .RankInfo.TypeContent}}
					<text x="380" y="25" font-size="10" fill="{{.TextColor}}" opacity="0.7" font-family="Arial, sans-serif" text-anchor="end" dominant-baseline="hanging">
						{{.RankInfo.TypeContent}}
					</text>
				{{end}}
				<image href="{{fetchImage .ImageURL}}" x="20" y="20" width="80" height="80" clip-path="url(#avatarClip)" />
				<text x="110" y="55" alignment-baseline="middle" font-size="24" fill="{{.TextColor}}" font-family="Arial, sans-serif" dominant-baseline="hanging">
					{{sliceText .Content 25}}
				</text>
				<rect x="110" y="70" width="270" height="15" rx="8" ry="8" fill="{{.RankInfo.EmptyColor}}"/>
				<rect x="110" y="70" width="{{calcProgress .RankInfo.RatioPercent}}" height="15" rx="8" ry="8" fill="url(#progressGradient)"/>
				<text x="380" y="{{placement .RankInfo.RatioPlacement}}" font-size="10" fill="{{.TextColor}}" opacity="0.7" font-family="Arial, sans-serif" text-anchor="end" dominant-baseline="hanging">
					{{.RankInfo.RatioString}}
				</text>
			</svg>
		`
	default:
		return ""
	}
}
