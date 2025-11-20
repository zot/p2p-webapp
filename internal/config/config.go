// CRC: crc-ConfigLoader.md, Spec: main.md
package config

import "time"

// Config holds all server configuration
type Config struct {
	Server    ServerConfig    `toml:"server"`
	HTTP      HTTPConfig      `toml:"http"`
	WebSocket WebSocketConfig `toml:"websocket"`
	Behavior  BehaviorConfig  `toml:"behavior"`
	Files     FilesConfig     `toml:"files"`
	P2P       P2PConfig       `toml:"p2p"`
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Port          int             `toml:"port"`
	PortRange     int             `toml:"portRange"`
	Timeouts      TimeoutConfig   `toml:"timeouts"`
	MaxHeaderBytes int            `toml:"maxHeaderBytes"`
}

// TimeoutConfig holds timeout settings
type TimeoutConfig struct {
	Read       Duration `toml:"read"`
	Write      Duration `toml:"write"`
	Idle       Duration `toml:"idle"`
	ReadHeader Duration `toml:"readHeader"`
}

// HTTPConfig holds HTTP-specific settings
type HTTPConfig struct {
	CacheControl string         `toml:"cacheControl"`
	Security     SecurityConfig `toml:"security"`
	CORS         CORSConfig     `toml:"cors"`
}

// SecurityConfig holds security header settings
type SecurityConfig struct {
	XContentTypeOptions string `toml:"xContentTypeOptions"`
	XFrameOptions       string `toml:"xFrameOptions"`
	ContentSecurityPolicy string `toml:"contentSecurityPolicy"`
}

// CORSConfig holds CORS settings
type CORSConfig struct {
	Enabled      bool     `toml:"enabled"`
	AllowOrigin  string   `toml:"allowOrigin"`
	AllowMethods []string `toml:"allowMethods"`
	AllowHeaders []string `toml:"allowHeaders"`
}

// WebSocketConfig holds WebSocket settings
type WebSocketConfig struct {
	CheckOrigin     bool     `toml:"checkOrigin"`
	AllowedOrigins  []string `toml:"allowedOrigins"`
	ReadBufferSize  int      `toml:"readBufferSize"`
	WriteBufferSize int      `toml:"writeBufferSize"`
}

// BehaviorConfig holds application behavior settings
type BehaviorConfig struct {
	AutoExitTimeout Duration `toml:"autoExitTimeout"`
	AutoOpenBrowser bool     `toml:"autoOpenBrowser"`
	Linger          bool     `toml:"linger"`
	Verbosity       int      `toml:"verbosity"`
}

// FilesConfig holds file serving settings
type FilesConfig struct {
	IndexFile   string `toml:"indexFile"`
	SPAFallback bool   `toml:"spaFallback"`
}

// P2PConfig holds P2P protocol settings
type P2PConfig struct {
	ProtocolName            string `toml:"protocolName"`
	FileUpdateNotifyTopic   string `toml:"fileUpdateNotifyTopic"`
}

// Duration wraps time.Duration for TOML parsing
type Duration struct {
	time.Duration
}

// UnmarshalText implements encoding.TextUnmarshaler
func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

// MarshalText implements encoding.TextMarshaler
func (d Duration) MarshalText() ([]byte, error) {
	return []byte(d.Duration.String()), nil
}
