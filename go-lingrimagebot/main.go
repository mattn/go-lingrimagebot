package lingrimagebot

import (
	"appengine"
	"appengine/urlfetch"
	"bytes"
	"code.google.com/p/draw2d/draw2d"
	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var reToken = regexp.MustCompile(`^!(image)\s((?:.|\n)*)`)
var reTokenP = regexp.MustCompile(`^!(image_p)\s((?:.|\n)*)`)
var reKomei = regexp.MustCompile(`^!(komei)\s((?:.|\n)*)`)
var reYuno = regexp.MustCompile(`^!(yuno)\s((?:.|\n)*)`)
var reDeris = regexp.MustCompile(`^!(d(?:eris)?|redis)\s((?:.|\n)*)`)
var reGolgo = regexp.MustCompile(`^!(golgo)\s((?:.|\n)*)`)
var reSeikai = regexp.MustCompile(`^!(seikai)\s((?:.|\n)*)`)

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

func upload(c appengine.Context, b []byte, ct string) (string, error) {
	u := urlfetch.Client(c)
	res, err := u.Post("http://gyazo.com/upload.cgi", ct, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode == 200 || res.StatusCode == 201 {
		content, err := ioutil.ReadAll(res.Body)
		c.Infof("debug %v", string(content))
		if err != nil {
			return "", err
		}
		gyazoUrl := string(content)
		if len(gyazoUrl) > 4 && gyazoUrl[:4] == "http" {
			return gyazoUrl + ".png\n", nil
		}
	}
	return "", nil
}

func imageNormal(lines []string) ([]byte, string, error) {
	maxWidth := 0
	for _, line := range lines {
		width := strWidth(line)
		if maxWidth < width {
			maxWidth = width
		}
	}
	rgba := image.NewRGBA(image.Rect(0, 0, maxWidth*11+70, len(lines)*20+20))
	draw.Draw(rgba, rgba.Bounds(), image.White, image.ZP, draw.Src)
	fc := freetype.NewContext()
	fc.SetDPI(72)
	fc.SetFont(font1)
	fc.SetFontSize(21)
	fc.SetClip(rgba.Bounds())
	fc.SetDst(rgba)
	fc.SetSrc(image.Black)

	pt := freetype.Pt(10, 10+int(fc.PointToFix32(21)>>8))
	for _, line := range lines {
		_, err := fc.DrawString(line, pt)
		if err != nil {
			return nil, "", err
		}
		pt.Y += fc.PointToFix32(11 * 1.8)
	}

	var b bytes.Buffer
	mp := multipart.NewWriter(&b)
	err := mp.WriteField("id", time.Now().Format("20060102030405"))
	if err != nil {
		return nil, "", err
	}
	part, err := mp.CreateFormFile("imagedata", "foo")
	if err != nil {
		return nil, "", err
	}
	err = png.Encode(part, rgba)
	if err != nil {
		return nil, "", err
	}
	err = mp.Close()
	if err != nil {
		return nil, "", err
	}
	return b.Bytes(), mp.FormDataContentType(), nil
}

func imageNormalP(lines []string) ([]byte, string, error) {
	maxWidth := 0
	for _, line := range lines {
		width := strWidth(line)
		if maxWidth < width {
			maxWidth = width
		}
	}
	rgba := image.NewRGBA(image.Rect(0, 0, maxWidth*11+70, len(lines)*20+20))
	draw.Draw(rgba, rgba.Bounds(), image.White, image.ZP, draw.Src)
	fc := freetype.NewContext()
	fc.SetDPI(72)
	fc.SetFont(font2)
	fc.SetFontSize(21)
	fc.SetClip(rgba.Bounds())
	fc.SetDst(rgba)
	fc.SetSrc(image.Black)

	pt := freetype.Pt(10, 10+int(fc.PointToFix32(21)>>8))
	for _, line := range lines {
		_, err := fc.DrawString(line, pt)
		if err != nil {
			return nil, "", err
		}
		pt.Y += fc.PointToFix32(11 * 1.8)
	}

	var b bytes.Buffer
	mp := multipart.NewWriter(&b)
	err := mp.WriteField("id", time.Now().Format("20060102030405"))
	if err != nil {
		return nil, "", err
	}
	part, err := mp.CreateFormFile("imagedata", "foo")
	if err != nil {
		return nil, "", err
	}
	err = png.Encode(part, rgba)
	if err != nil {
		return nil, "", err
	}
	err = mp.Close()
	if err != nil {
		return nil, "", err
	}
	return b.Bytes(), mp.FormDataContentType(), nil
}

func imageKomei(lines []string) ([]byte, string, error) {
	pngf, _ := os.Open("image/komei.png")
	pngi, _ := png.Decode(pngf)
	rgba := image.NewRGBA(image.Rect(0, 0, pngi.Bounds().Dx(), pngi.Bounds().Dy()))
	draw.Draw(rgba, rgba.Bounds(), pngi, image.ZP, draw.Src)
	fc := freetype.NewContext()
	fc.SetDPI(72)
	fc.SetFont(font1)
	fc.SetFontSize(18)
	fc.SetClip(rgba.Bounds())
	fc.SetDst(rgba)
	fc.SetSrc(image.Black)

	pt := freetype.Pt(pngi.Bounds().Dx()-25, 20)
	for _, line := range lines {
		for _, r := range []rune(line) {
			_, err := fc.DrawString(string(r), pt)
			if err != nil {
				return nil, "", err
			}
			pt.Y += fc.PointToFix32(11 * 1.8)
		}
		pt.Y = fc.PointToFix32(20)
		pt.X -= fc.PointToFix32(11 * 1.8)
	}

	var b bytes.Buffer
	mp := multipart.NewWriter(&b)
	err := mp.WriteField("id", time.Now().Format("20060102030405"))
	if err != nil {
		return nil, "", err
	}
	part, err := mp.CreateFormFile("imagedata", "foo")
	if err != nil {
		return nil, "", err
	}
	err = png.Encode(part, rgba)
	if err != nil {
		return nil, "", err
	}
	err = mp.Close()
	if err != nil {
		return nil, "", err
	}
	return b.Bytes(), mp.FormDataContentType(), nil
}

func imageYuno(lines []string) ([]byte, string, error) {
	pngf, _ := os.Open("image/yuno.png")
	pngi, _ := png.Decode(pngf)
	rgba := image.NewRGBA(image.Rect(0, 0, pngi.Bounds().Dx(), pngi.Bounds().Dy()))
	draw.Draw(rgba, rgba.Bounds(), pngi, image.ZP, draw.Src)
	fc := freetype.NewContext()
	fc.SetDPI(72)
	fc.SetFont(font1)
	fc.SetFontSize(22)
	fc.SetClip(rgba.Bounds())
	fc.SetDst(rgba)
	fc.SetSrc(image.White)

	pt := freetype.Pt(25, 25+int(fc.PointToFix32(21)>>8))
	for _, line := range lines {
		_, err := fc.DrawString(line, pt)
		if err != nil {
			return nil, "", err
		}
		pt.Y += fc.PointToFix32(22 * 1.8)
	}

	var b bytes.Buffer
	mp := multipart.NewWriter(&b)
	err := mp.WriteField("id", time.Now().Format("20060102030405"))
	if err != nil {
		return nil, "", err
	}
	part, err := mp.CreateFormFile("imagedata", "foo")
	if err != nil {
		return nil, "", err
	}
	err = png.Encode(part, rgba)
	if err != nil {
		return nil, "", err
	}
	err = mp.Close()
	if err != nil {
		return nil, "", err
	}
	return b.Bytes(), mp.FormDataContentType(), nil
}

func imageDeris(lines []string) ([]byte, string, error) {
	maxWidth := 0
	for _, line := range lines {
		width := strWidth(line)
		if maxWidth < width {
			maxWidth = width
		}
	}
	width := maxWidth*11 + 80
	if width < 200 {
		width = 200
	}
	pngf, _ := os.Open("image/deris.png")
	pngi, _ := png.Decode(pngf)
	rgba := image.NewRGBA(image.Rect(0, 0, width, len(lines)*21+50))
	gc := draw2d.NewGraphicContext(rgba)
	gc.SetFillColor(image.White)
	paths := &draw2d.PathStorage{}
	paths.MoveTo(0, 0)
	paths.LineTo(float64(rgba.Bounds().Dx())-1, 0)
	paths.LineTo(float64(rgba.Bounds().Dx())-1, float64(rgba.Bounds().Dy())-1)
	paths.LineTo(0, float64(rgba.Bounds().Dy())-1)
	paths.LineTo(0, 0)
	gc.Fill(paths.Close())
	draw.Draw(rgba, rgba.Bounds(), pngi, image.ZP, draw.Src)
	gc.SetStrokeColor(image.Black)
	gc.Stroke(paths.Close())
	fc := freetype.NewContext()
	fc.SetDPI(72)
	fc.SetFont(font1)
	fc.SetFontSize(21)
	fc.SetClip(rgba.Bounds())
	fc.SetDst(rgba)
	fc.SetSrc(image.Black)

	pt := freetype.Pt(70, 35+int(fc.PointToFix32(21)>>8))
	for _, line := range lines {
		_, err := fc.DrawString(line, pt)
		if err != nil {
			return nil, "", err
		}
		pt.Y += fc.PointToFix32(11 * 1.8)
	}

	var b bytes.Buffer
	mp := multipart.NewWriter(&b)
	err := mp.WriteField("id", time.Now().Format("20060102030405"))
	if err != nil {
		return nil, "", err
	}
	part, err := mp.CreateFormFile("imagedata", "foo")
	if err != nil {
		return nil, "", err
	}
	err = png.Encode(part, rgba)
	if err != nil {
		return nil, "", err
	}
	err = mp.Close()
	if err != nil {
		return nil, "", err
	}
	return b.Bytes(), mp.FormDataContentType(), nil
}

func imageGolgo(lines []string) ([]byte, string, error) {
	pngf, _ := os.Open("image/golgo.png")
	pngi, _ := png.Decode(pngf)
	rgba := image.NewRGBA(image.Rect(0, 0, pngi.Bounds().Dx(), pngi.Bounds().Dy()))
	draw.Draw(rgba, rgba.Bounds(), pngi, image.ZP, draw.Src)
	fc := freetype.NewContext()
	fc.SetDPI(72)
	fc.SetFont(font1)
	fc.SetFontSize(18)
	fc.SetClip(rgba.Bounds())
	fc.SetDst(rgba)
	fc.SetSrc(image.Black)

	pt := freetype.Pt(pngi.Bounds().Dx()-25, 25)
	for _, line := range lines {
		for _, r := range []rune(line) {
			_, err := fc.DrawString(string(r), pt)
			if err != nil {
				return nil, "", err
			}
			pt.Y += fc.PointToFix32(11 * 1.8)
		}
		pt.Y = fc.PointToFix32(25)
		pt.X -= fc.PointToFix32(11 * 1.8)
	}

	var b bytes.Buffer
	mp := multipart.NewWriter(&b)
	err := mp.WriteField("id", time.Now().Format("20060102030405"))
	if err != nil {
		return nil, "", err
	}
	part, err := mp.CreateFormFile("imagedata", "foo")
	if err != nil {
		return nil, "", err
	}
	err = png.Encode(part, rgba)
	if err != nil {
		return nil, "", err
	}
	err = mp.Close()
	if err != nil {
		return nil, "", err
	}
	return b.Bytes(), mp.FormDataContentType(), nil
}

func imageSeikai(lines []string) ([]byte, string, error) {
	pngf, _ := os.Open("image/seikai.png")
	pngi, _ := png.Decode(pngf)
	rgba := image.NewRGBA(image.Rect(0, 0, pngi.Bounds().Dx(), pngi.Bounds().Dy()))
	draw.Draw(rgba, rgba.Bounds(), pngi, image.ZP, draw.Src)
	fc := freetype.NewContext()
	fc.SetDPI(72)
	fc.SetFont(font1)
	fc.SetFontSize(18)
	fc.SetClip(rgba.Bounds())
	fc.SetDst(rgba)
	fc.SetSrc(image.Black)

	pt := freetype.Pt(50, pngi.Bounds().Dy()-30)
	for _, line := range lines {
		_, err := fc.DrawString(line, pt)
		if err != nil {
			return nil, "", err
		}
		pt.Y += fc.PointToFix32(11 * 1.8)
	}

	var b bytes.Buffer
	mp := multipart.NewWriter(&b)
	err := mp.WriteField("id", time.Now().Format("20060102030405"))
	if err != nil {
		return nil, "", err
	}
	part, err := mp.CreateFormFile("imagedata", "foo")
	if err != nil {
		return nil, "", err
	}
	err = png.Encode(part, rgba)
	if err != nil {
		return nil, "", err
	}
	err = mp.Close()
	if err != nil {
		return nil, "", err
	}
	return b.Bytes(), mp.FormDataContentType(), nil
}

var (
	font1 *truetype.Font
	font2 *truetype.Font
)

func init() {
	fontBytes1, err := ioutil.ReadFile("font/ipag-mona.ttf")
	if err != nil {
		log.Println(err)
		return
	}
	font1, err = freetype.ParseFont(fontBytes1)
	if err != nil {
		log.Println(err)
		return
	}
	fontBytes2, err := ioutil.ReadFile("font/ipagp-mona.ttf")
	if err != nil {
		log.Println(err)
		return
	}
	font2, err = freetype.ParseFont(fontBytes2)
	if err != nil {
		log.Println(err)
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			if r.Method == "POST" {
				var status Status

				c := appengine.NewContext(r)
				defer func() {
					if err := recover(); err != nil {
						c.Errorf("%s", fmt.Sprint(err))
					}
				}()
				e := json.NewDecoder(r.Body).Decode(&status)
				if e != nil {
					c.Errorf("%s", e.Error())
					return
				}
				results := ""

				tables := []struct {
					pat       *regexp.Regexp
					landscape bool
					f         func([]string) ([]byte, string, error)
				}{
					{reToken, false, imageNormal},
					{reTokenP, false, imageNormalP},
					{reKomei, true, imageKomei},
					{reYuno, true, imageYuno},
					{reDeris, true, imageDeris},
					{reGolgo, true, imageGolgo},
					{reSeikai, true, imageSeikai},
				}
				for _, event := range status.Events {
					for _, t := range tables {
						tokens := t.pat.FindStringSubmatch(event.Message.Text)
						if len(tokens) == 3 {
							c.Infof("debug %v", event.Message.Text)
							var lines []string
							if t.landscape {
								lines = strings.Split(strings.Replace(tokens[2], "ー", "｜", -1), "\n")
							} else {
								lines = strings.Split(tokens[2], "\n")
							}
							b, ct, err := t.f(lines)
							if err != nil {
								c.Errorf("%s", e.Error())
							}
							c.Infof("debug %v %v", ct, len(b))
							res, err := upload(c, b, ct)
							c.Infof("debug %v", res)
							if err != nil {
								c.Errorf("%s", e.Error())
							}
							results += res
						}
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
