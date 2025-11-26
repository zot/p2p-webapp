// CRC: crc-ConfigLoader.md, Spec: main.md
package config

import "time"

// DefaultConfig returns the default configuration
// CRC: crc-ConfigLoader.md
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:      10000,
			PortRange: 100,
			Timeouts: TimeoutConfig{
				Read:       Duration{15 * time.Second},
				Write:      Duration{15 * time.Second},
				Idle:       Duration{60 * time.Second},
				ReadHeader: Duration{5 * time.Second},
			},
			MaxHeaderBytes: 1048576, // 1 MB
		},
		HTTP: HTTPConfig{
			CacheControl: "no-cache, no-store, must-revalidate",
			Security: SecurityConfig{
				XContentTypeOptions: "nosniff",
				XFrameOptions:       "DENY",
				ContentSecurityPolicy: "",
			},
			CORS: CORSConfig{
				Enabled:      false,
				AllowOrigin:  "",
				AllowMethods: []string{},
				AllowHeaders: []string{},
			},
		},
		WebSocket: WebSocketConfig{
			CheckOrigin:     false, // Allow all origins by default
			AllowedOrigins:  []string{},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		Behavior: BehaviorConfig{
			AutoExitTimeout: Duration{5 * time.Second},
			AutoOpenBrowser: true,
			Linger:          false,
			Verbosity:       0,
		},
		Files: FilesConfig{
			IndexFile:   "index.html",
			SPAFallback: true,
		},
		P2P: P2PConfig{
			IPFSGetTimeout: Duration{3 * time.Second},
		},
	}
}
