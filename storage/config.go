package storage

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	coretypes "github.com/drlinggg/gocurl/core/types"
)

//go:embed embed/presets.toml
var defaultPresets []byte

type presetColors struct {
	Status2xx string `toml:"status_2xx,omitempty"`
	Status4xx string `toml:"status_4xx,omitempty"`
	Status5xx string `toml:"status_5xx,omitempty"`
	Headers   string `toml:"headers,omitempty"`
	Body      string `toml:"body,omitempty"`
	Elapsed   string `toml:"elapsed,omitempty"`
}

type preset struct {
	Base        string            `toml:"base,omitempty"`
	Timeout     int               `toml:"timeout,omitempty"`
	HTTPVersion int               `toml:"http_version,omitempty"`
	Scheme      string            `toml:"scheme,omitempty"`
	Headers     map[string]string `toml:"headers,omitempty"`
	Query       map[string]string `toml:"query,omitempty"`
	Colors      presetColors      `toml:"colors,omitempty"`
}

type Config struct {
	presets map[string]preset
	path    string
}

func LoadConfig() (*Config, error) {
	path := os.Getenv("GOCURL_CONFIG")
	if path == "" {
		dir, err := defaultDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(dir, "presets.toml")
	}

	var presets map[string]preset
	if _, err := toml.DecodeFile(path, &presets); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if _, err := toml.Decode(string(defaultPresets), &presets); err != nil {
			return nil, err
		}
	}
	return &Config{presets: presets, path: path}, nil
}

// SetField записывает одно поле в пресет. Синтаксис:
//
//	@Key=Value  → headers[Key]
//	?key=value  → query[key]
//	base=url    → base URL для автоматчинга и разворачивания коротких путей
//	timeout=N   → таймаут запроса в секундах
//	scheme=http|https
//	http=1|2|3
func (c *Config) SetField(presetName, arg string) error {
	p := c.presets[presetName]

	switch {
	case strings.HasPrefix(arg, "@"):
		key, val, _ := strings.Cut(arg[1:], "=")
		if p.Headers == nil {
			p.Headers = make(map[string]string)
		}
		p.Headers[key] = val

	case strings.HasPrefix(arg, "?"):
		key, val, _ := strings.Cut(arg[1:], "=")
		if p.Query == nil {
			p.Query = make(map[string]string)
		}
		p.Query[key] = val

	default:
		key, val, found := strings.Cut(arg, "=")
		if !found {
			return fmt.Errorf("invalid argument %q: expected key=value", arg)
		}
		switch key {
		case "base":
			p.Base = val
		case "timeout":
			n, err := strconv.Atoi(val)
			if err != nil {
				return fmt.Errorf("timeout must be an integer, got %q", val)
			}
			p.Timeout = n
		case "scheme":
			if val != "http" && val != "https" {
				return fmt.Errorf("scheme must be http or https, got %q", val)
			}
			p.Scheme = val
		case "http":
			n, err := strconv.Atoi(val)
			if err != nil || (n != 1 && n != 2 && n != 3) {
				return fmt.Errorf("http must be 1, 2 or 3, got %q", val)
			}
			p.HTTPVersion = n
		default:
			if strings.HasPrefix(key, "color.") {
				if err := setColor(&p.Colors, strings.TrimPrefix(key, "color."), val); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("unknown field %q", key)
			}
		}
	}

	if c.presets == nil {
		c.presets = make(map[string]preset)
	}
	c.presets[presetName] = p
	return nil
}

func setColor(c *presetColors, field, val string) error {
	switch field {
	case "status_2xx":
		c.Status2xx = val
	case "status_4xx":
		c.Status4xx = val
	case "status_5xx":
		c.Status5xx = val
	case "headers":
		c.Headers = val
	case "body":
		c.Body = val
	case "elapsed":
		c.Elapsed = val
	default:
		return fmt.Errorf("unknown color field %q (valid: status_2xx, status_4xx, status_5xx, headers, body, elapsed)", field)
	}
	return nil
}

// Save записывает всю конфигурацию обратно в TOML-файл.
func (c *Config) Save() error {
	if err := os.MkdirAll(filepath.Dir(c.path), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(c.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(c.presets)
}

func (c *Config) FillRequest(req *coretypes.Request) (coretypes.HTTPRequest, error) {
	h := c.presets["default"]

	// Если URL начинается с имени пресета (например "dev/users"),
	// разворачиваем в base + path и берём настройки этого пресета.
	matched, resolvedURL := c.resolvePresetURL(req.URL)
	if resolvedURL == "" {
		resolvedURL = req.URL
		matched = c.match(req.URL)
	}

	p := c.merge(h, matched)
	out := req.HTTPRequest
	out.URL = resolvedURL

	if out.Timeout == 0 && p.Timeout > 0 {
		out.Timeout = p.Timeout
	}
	if out.Scheme == nil {
		out.Scheme = parseScheme(p.Scheme)
	}
	if out.HTTP == nil {
		out.HTTP = parseHTTPVersion(p.HTTPVersion)
	}
	for k, v := range p.Headers {
		out.Headers = append(out.Headers, coretypes.Field{
			Key:   k,
			Value: coretypes.StringValue{Val: os.ExpandEnv(v)},
		})
	}
	for k, v := range p.Query {
		out.Query = append(out.Query, coretypes.Field{
			Key:   k,
			Value: coretypes.StringValue{Val: v},
		})
	}

	return out, nil
}

// resolvePresetURL проверяет, начинается ли url с имени пресета.
// "dev"       → base пресета
// "dev/users" → base + "/users"
// Возвращает (пресет, итоговый URL) или (пусто, "") если не совпало.
func (c *Config) resolvePresetURL(url string) (preset, string) {
	name, path, _ := strings.Cut(url, "/")
	p, ok := c.presets[name]
	if !ok || name == "default" || p.Base == "" {
		return preset{}, ""
	}
	base := strings.TrimRight(p.Base, "/")
	if path == "" {
		return p, base
	}
	return p, base + "/" + path
}

var defaultColors = coretypes.Colors{
	Status2xx: coretypes.Color{0, 200, 83},
	Status4xx: coretypes.Color{255, 171, 0},
	Status5xx: coretypes.Color{255, 23, 68},
	Headers:   coretypes.Color{136, 136, 136},
	Body:      coretypes.Color{97, 218, 251},
	Elapsed:   coretypes.Color{85, 85, 85},
}

func colorOrDefault(s string, fallback coretypes.Color) coretypes.Color {
	c, err := coretypes.ParseColor(s)
	if err != nil {
		return fallback
	}
	return c
}

func (c *Config) Colors() coretypes.Colors {
	d := c.presets["default"]
	return coretypes.Colors{
		Status2xx: colorOrDefault(d.Colors.Status2xx, defaultColors.Status2xx),
		Status4xx: colorOrDefault(d.Colors.Status4xx, defaultColors.Status4xx),
		Status5xx: colorOrDefault(d.Colors.Status5xx, defaultColors.Status5xx),
		Headers:   colorOrDefault(d.Colors.Headers, defaultColors.Headers),
		Body:      colorOrDefault(d.Colors.Body, defaultColors.Body),
		Elapsed:   colorOrDefault(d.Colors.Elapsed, defaultColors.Elapsed),
	}
}

func (c *Config) match(url string) preset {
	for name, p := range c.presets {
		if name == "default" {
			continue
		}
		if p.Base != "" && strings.HasPrefix(url, p.Base) {
			return p
		}
	}
	return preset{}
}

func (c *Config) merge(base, override preset) preset {
	if override.Timeout > 0 {
		base.Timeout = override.Timeout
	}
	if override.HTTPVersion > 0 {
		base.HTTPVersion = override.HTTPVersion
	}
	if override.Scheme != "" {
		base.Scheme = override.Scheme
	}
	for k, v := range override.Headers {
		if base.Headers == nil {
			base.Headers = map[string]string{}
		}
		base.Headers[k] = v
	}
	for k, v := range override.Query {
		if base.Query == nil {
			base.Query = map[string]string{}
		}
		base.Query[k] = v
	}
	return base
}

func parseScheme(s string) coretypes.Scheme {
	if s == "http" {
		return coretypes.SchemeHTTP{}
	}
	return coretypes.SchemeHTTPS{}
}

func parseHTTPVersion(v int) coretypes.Http {
	switch v {
	case 2:
		return coretypes.Http2{}
	case 3:
		return coretypes.Http3{}
	default:
		return coretypes.Http1{}
	}
}
