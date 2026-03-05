package draw

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/golang/freetype/truetype"
)

func getFontsDir() string {
	execPath, err := os.Executable()
	if err != nil {
		return "fonts"
	}

	projectRoot := filepath.Dir(filepath.Dir(execPath))
	fontsDir := filepath.Join(projectRoot, "fonts")

	if _, err := os.Stat(fontsDir); os.IsNotExist(err) {
		return "fonts"
	}

	return fontsDir
}

type FontStyle byte

const (
	FontStyleNormal FontStyle = iota
	FontStyleBold
	FontStyleItalic
)

type FontFamily byte

const (
	FontFamilySans FontFamily = iota
	FontFamilySerif
	FontFamilyMono
)

type FontData struct {
	Name   string
	Family FontFamily
	Style  FontStyle
}

type FontFileNamer func(fontData FontData) string

func FontFileName(fontData FontData) string {
	fontFileName := fontData.Name
	switch fontData.Family {
	case FontFamilySans:
		fontFileName += "s"
	case FontFamilySerif:
		fontFileName += "r"
	case FontFamilyMono:
		fontFileName += "m"
	}
	if fontData.Style&FontStyleBold != 0 {
		fontFileName += "b"
	} else {
		fontFileName += "r"
	}

	if fontData.Style&FontStyleItalic != 0 {
		fontFileName += "i"
	}
	fontFileName += ".ttf"
	return fontFileName
}

type FontCache interface {
	Load(FontData) (*truetype.Font, error)
	Store(FontData, *truetype.Font)
}

func RegisterFont(fontData FontData, font *truetype.Font) {
	fontCache.Store(fontData, font)
}

func GetFont(fontData FontData) *truetype.Font {
	font, err := fontCache.Load(fontData)
	if err != nil {
		log.Println(err)
	}
	return font
}

func GetFontFolder() string {
	return defaultFonts.folder
}

func SetFontFolder(folder string) {
	defaultFonts.setFolder(filepath.Clean(folder))
}

func GetGlobalFontCache() FontCache {
	return fontCache
}

func SetFontNamer(fn FontFileNamer) {
	defaultFonts.setNamer(fn)
}

func SetFontCache(cache FontCache) {
	if cache == nil {
		fontCache = defaultFonts
	} else {
		fontCache = cache
	}
}

type FolderFontCache struct {
	fonts  map[string]*truetype.Font
	folder string
	namer  FontFileNamer
}

func NewFolderFontCache(folder string) *FolderFontCache {
	return &FolderFontCache{
		fonts:  make(map[string]*truetype.Font),
		folder: folder,
		namer:  FontFileName,
	}
}

func (cache *FolderFontCache) Load(fontData FontData) (*truetype.Font, error) {
	font := cache.fonts[cache.namer(fontData)]
	if font != nil {
		return font, nil
	}

	var data []byte
	var file = cache.namer(fontData)

	data, err := os.ReadFile(filepath.Join(cache.folder, file))
	if err != nil {
		return nil, err
	}

	font, err = truetype.Parse(data)
	if err != nil {
		return nil, err
	}

	cache.fonts[file] = font
	return font, nil
}

func (cache *FolderFontCache) Store(fontData FontData, font *truetype.Font) {
	cache.fonts[cache.namer(fontData)] = font
}

type SyncFolderFontCache struct {
	sync.RWMutex
	fonts  map[string]*truetype.Font
	folder string
	namer  FontFileNamer
}

func NewSyncFolderFontCache(folder string) *SyncFolderFontCache {
	return &SyncFolderFontCache{
		fonts:  make(map[string]*truetype.Font),
		folder: folder,
		namer:  FontFileName,
	}
}

func (cache *SyncFolderFontCache) setFolder(folder string) {
	cache.Lock()
	cache.folder = folder
	cache.Unlock()
}

func (cache *SyncFolderFontCache) setNamer(namer FontFileNamer) {
	cache.Lock()
	cache.namer = namer
	cache.Unlock()
}

func (cache *SyncFolderFontCache) Load(fontData FontData) (*truetype.Font, error) {
	cache.RLock()
	font := cache.fonts[cache.namer(fontData)]
	cache.RUnlock()

	if font != nil {
		return font, nil
	}

	var data []byte
	var file = cache.namer(fontData)

	data, err := os.ReadFile(filepath.Join(cache.folder, file))
	if err != nil {
		return nil, err
	}

	font, err = truetype.Parse(data)
	if err != nil {
		return nil, err
	}

	cache.Lock()
	cache.fonts[file] = font
	cache.Unlock()

	return font, nil
}

func (cache *SyncFolderFontCache) Store(fontData FontData, font *truetype.Font) {
	cache.Lock()
	cache.fonts[cache.namer(fontData)] = font
	cache.Unlock()
}

var (
	defaultFonts           = NewSyncFolderFontCache(getFontsDir())
	fontCache    FontCache = defaultFonts
)
