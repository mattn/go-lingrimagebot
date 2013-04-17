package lingrimagebot

import (
	"appengine"
	"appengine/urlfetch"
	"bytes"
	"code.google.com/p/freetype-go/freetype"
	"encoding/json"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

type Status struct {
	Events []Event `json:"events"`
}

type Event struct {
	Id      int      `json:"event_id"`
	Message *Message `json:"message"`
}

type Message struct {
	Id              string `json:"id"`
	Room            string `json:"room"`
	PublicSessionId string `json:"public_session_id"`
	IconUrl         string `json:"icon_url"`
	Type            string `json:"type"`
	SpeakerId       string `json:"speaker_id"`
	Nickname        string `json:"nickname"`
	Text            string `json:"text"`
}

func runeWidth(r rune) int {
	if r >= 0x1100 &&
		(r <= 0x115f || r == 0x2329 || r == 0x232a ||
			(r >= 0x2e80 && r <= 0xa4cf && r != 0x303f) ||
			(r >= 0xac00 && r <= 0xd7a3) ||
			(r >= 0xf900 && r <= 0xfaff) ||
			(r >= 0xfe30 && r <= 0xfe6f) ||
			(r >= 0xff00 && r <= 0xff60) ||
			(r >= 0xffe0 && r <= 0xffe6) ||
			(r >= 0x20000 && r <= 0x2fffd) ||
			(r >= 0x30000 && r <= 0x3fffd)) {
		return 2
	}
	return 1
}

func strWidth(str string) int {
	r := 0
	for _, c := range []rune(str) {
		r += runeWidth(c)
	}
	return r
}

func init() {
	fontBytes, err := ioutil.ReadFile("Ricty-Regular.ttf")
	if err != nil {
		log.Println(err)
		return
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println(err)
		return
	}

	fg, bg := image.Black, image.White

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			if r.Method == "POST" {
				var status Status

				c := appengine.NewContext(r)
				u := urlfetch.Client(c)
				e := json.NewDecoder(r.Body).Decode(&status)
				if e != nil {
					c.Errorf("%s", e.Error())
					return
				}
				results := ""
				for _, event := range status.Events {
					tokens := strings.SplitN(event.Message.Text, " ", 2)
					if tokens[0] == "!!image" {
						lines := strings.Split(tokens[1], "\n")
						maxWidth := 0
						for _, line := range lines {
							width := strWidth(line)
							if maxWidth < width {
								maxWidth = width
							}
						}
						rgba := image.NewRGBA(image.Rect(0, 0, maxWidth*11+10, len(lines)*20+20))
						draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
						fc := freetype.NewContext()
						fc.SetDPI(72)
						fc.SetFont(font)
						fc.SetFontSize(21)
						fc.SetClip(rgba.Bounds())
						fc.SetDst(rgba)
						fc.SetSrc(fg)

						pt := freetype.Pt(10, 10+int(fc.PointToFix32(21)>>8))
						for _, line := range lines {
							_, err = fc.DrawString(line, pt)
							if err != nil {
								c.Errorf("%s", e.Error())
								return
							}
							pt.Y += fc.PointToFix32(11 * 1.8)
						}

						var b bytes.Buffer
						mp := multipart.NewWriter(&b)
						err = mp.WriteField("id", time.Now().Format("20060102030405"))
						if err != nil {
							c.Errorf("%s", e.Error())
							return
						}
						part, err := mp.CreateFormFile("imagedata", "foo")
						if err != nil {
							c.Errorf("%s", e.Error())
							return
						}
						err = png.Encode(part, rgba)
						if err != nil {
							c.Errorf("%s", e.Error())
							return
						}
						err = mp.Close()
						if err != nil {
							c.Errorf("%s", e.Error())
							return
						}
						res, err := u.Post("http://gyazo.com/upload.cgi", mp.FormDataContentType(), bytes.NewReader(b.Bytes()))
						if err != nil {
							c.Errorf("%s", e.Error())
							return
						}
						defer res.Body.Close()
						content, err := ioutil.ReadAll(res.Body)
						if err != nil {
							c.Errorf("%s", e.Error())
							return
						}
						gyazoUrl := string(content)
						if len(gyazoUrl) > 4 && gyazoUrl[:5] == "http:" {
							gyazoUrl += ".png"
						}
						results += gyazoUrl + "\n"
					}
				}
				if len(results) > 0 {
					w.Header().Set("Content-Type", "text/plain; charset=utf8")
					results = strings.TrimRight(results, "\n")
					if runes := []rune(results); len(runes) > 1000 {
						results = string(runes[0:999])
					}
					w.Write([]byte(results))
				}
			} else {
				w.Header().Set("Content-Type", "text/html; charset=utf8")
				b, _ := ioutil.ReadFile("index.html")
				w.Write(b)
			}
		} else {
			http.NotFound(w, r)
		}
	})
}
