package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	listenAddress = flag.String("web.listen-address", ":32142", "Address to listen on for web interface and telemetry")
	metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics")
	debugMode     = flag.Bool("debug", false, "Enable debug logging")
	utmpPath      = flag.String("utmp-path", "/var/run/utmp", "Path to the utmp file")
)

// utmp entry types
const (
	EMPTY         = 0
	RUN_LVL       = 1
	BOOT_TIME     = 2
	NEW_TIME      = 3
	OLD_TIME      = 4
	INIT_PROCESS  = 5
	LOGIN_PROCESS = 6
	USER_PROCESS  = 7
	DEAD_PROCESS  = 8
	ACCOUNTING    = 9
)

// Define metrics
var (
	// Gauge for total number of logged-in users
	totalUsers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "unix_users_logged_in_total",
		Help: "Total number of users currently logged in",
	})

	// Gauge vector for user sessions with labels
	userSessions = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "unix_user_session_info",
			Help: "Information about user sessions",
		},
		[]string{"username", "from", "tty", "login_time"},
	)

	// Gauge for user session count by username
	userSessionCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "unix_user_session_count",
			Help: "Number of sessions per user",
		},
		[]string{"username"},
	)

	// Gauge for user session count by origin IP
	userSessionByIP = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "unix_user_session_by_ip",
			Help: "Number of sessions per origin IP",
		},
		[]string{"ip"},
	)
)

func init() {
	// Register metrics with Prometheus
	prometheus.MustRegister(totalUsers)
	prometheus.MustRegister(userSessions)
	prometheus.MustRegister(userSessionCount)
	prometheus.MustRegister(userSessionByIP)
}

func parseUtmpEntry(data []byte) (int32, int32, string, string, string, time.Time) {
	if len(data) < 384 {
		if *debugMode {
			log.Printf("Data too short: got %d bytes, expected at least 384", len(data))
		}
		return 0, 0, "", "", "", time.Time{}
	}

	// Parse entry type (first 4 bytes, little endian)
	entryType := int32(data[0]) | int32(data[1])<<8 | int32(data[2])<<16 | int32(data[3])<<24
	
	// Parse PID (next 4 bytes)
	pid := int32(data[4]) | int32(data[5])<<8 | int32(data[6])<<16 | int32(data[7])<<24
	
	// Validate entry type
	if entryType < 0 || entryType > 9 {
		if *debugMode {
			log.Printf("Invalid entry type %d", entryType)
		}
		return 0, 0, "", "", "", time.Time{}
	}
	
	if *debugMode && (entryType == 7 || entryType == 8) {
		log.Printf("Raw entry data (first 100 bytes): %x", data[:100])
	}
	
	// Based on hexdump analysis of ARM64 utmp:
	// Line field starts at offset 8, 32 bytes
	line := strings.TrimRight(string(data[8:40]), "\x00")
	
	// ID field at offset 40, 4 bytes (skip)
	
	// User field starts at offset 44, 32 bytes
	user := strings.TrimRight(string(data[44:76]), "\x00")
	
	// Host field starts at offset 76, 256 bytes
	host := strings.TrimRight(string(data[76:332]), "\x00")
	
	// Special handling for DEAD_PROCESS entries - they might have different structure
	if entryType == 8 && user == "" {
		// For DEAD_PROCESS, try to extract user info from different locations
		// Look for user data in the area around offset 44-100
		for offset := 44; offset < 100; offset += 1 {
			if offset+8 > len(data) {
				break
			}
			testUser := strings.TrimRight(string(data[offset:offset+8]), "\x00")
			if len(testUser) > 0 && len(testUser) < 32 && isValidUsername(testUser) {
				user = testUser
				if *debugMode {
					log.Printf("Found DEAD_PROCESS user '%s' at offset %d", user, offset)
				}
				break
			}
		}
		
		// Look for host data in the expected area and beyond
		if host == "" {
			for offset := 76; offset < 400; offset += 4 {
				if offset+16 > len(data) {
					break
				}
				testHost := strings.TrimRight(string(data[offset:offset+16]), "\x00")
				if len(testHost) > 6 && (strings.Contains(testHost, ".") || strings.Contains(testHost, ":")) {
					host = testHost
					if *debugMode {
						log.Printf("Found DEAD_PROCESS host '%s' at offset %d", host, offset)
					}
					break
				}
			}
		}
		
		// If still no user found, try a broader search
		if user == "" {
			// Look for common usernames in the data
			dataStr := string(data)
			commonUsers := []string{"root", "admin", "user", "pi", "zerepl", "ubuntu", "debian"}
			for _, testUser := range commonUsers {
				if strings.Contains(dataStr, testUser) {
					user = testUser
					if *debugMode {
						log.Printf("Found DEAD_PROCESS user '%s' via string search", user)
					}
					break
				}
			}
		}
	}
	
	// Exit field at offset 332, 4 bytes (skip)
	
	// Session at offset 336, 4 bytes (skip)
	
	// Time at offset 340, 8 bytes (tv_sec + tv_usec)
	timeSec := int32(data[340]) | int32(data[341])<<8 | int32(data[342])<<16 | int32(data[343])<<24
	timeUsec := int32(data[344]) | int32(data[345])<<8 | int32(data[346])<<16 | int32(data[347])<<24
	
	var loginTime time.Time
	if timeSec > 0 && timeSec < 2147483647 {
		loginTime = time.Unix(int64(timeSec), int64(timeUsec)*1000)
	} else {
		loginTime = time.Now()
	}
	
	if *debugMode {
		log.Printf("ARM64 parsed entry: type=%d, pid=%d, user='%s', line='%s', host='%s', time=%v", 
			entryType, pid, user, line, host, loginTime)
	}
	
	return entryType, pid, user, line, host, loginTime
}

// isValidUsername checks if a string looks like a valid username
func isValidUsername(s string) bool {
	if len(s) == 0 || len(s) > 32 {
		return false
	}
	// Check for printable ASCII characters typical in usernames
	for _, c := range s {
		if c < 32 || c > 126 {
			return false
		}
		// Usernames typically don't contain these characters
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			return false
		}
	}
	return true
}

func collectUserMetrics() {
	// Check if utmp file exists
	if _, err := os.Stat(*utmpPath); os.IsNotExist(err) {
		log.Printf("Error: utmp file %s does not exist", *utmpPath)
		return
	}

	// Open and read utmp file
	file, err := os.Open(*utmpPath)
	if err != nil {
		log.Printf("Error opening utmp file %s: %v", *utmpPath, err)
		return
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		log.Printf("Error getting file info: %v", err)
		return
	}

	entrySize := 384 // Fixed utmp entry size for Linux
	fileSize := fileInfo.Size()
	numEntries := int(fileSize) / entrySize

	if *debugMode {
		log.Printf("utmp file size: %d bytes, entry size: %d bytes, entries: %d", 
			fileSize, entrySize, numEntries)
	}

	// Reset metrics before collecting new data
	userSessions.Reset()
	userSessionCount.Reset()
	userSessionByIP.Reset()

	// Count valid sessions
	validSessions := 0
	userCounts := make(map[string]int)
	ipCounts := make(map[string]int)

	// Read all data at once
	data := make([]byte, fileSize)
	_, err = file.Read(data)
	if err != nil {
		log.Printf("Error reading utmp file: %v", err)
		return
	}

	// Parse each entry, but also scan for misaligned entries
	entryIndex := 0
	for offset := 0; offset < len(data)-384; offset += 4 {
		// Check if this looks like a valid entry start
		if offset%384 == 0 {
			entryIndex = offset / 384
		}
		
		entryType := int32(data[offset]) | int32(data[offset+1])<<8 | int32(data[offset+2])<<16 | int32(data[offset+3])<<24
		
		// Only process if this looks like a valid entry type
		if entryType < 0 || entryType > 9 {
			continue
		}
		
		// Skip if not at expected boundary and not a user-related entry
		if offset%384 != 0 && entryType != USER_PROCESS && entryType != DEAD_PROCESS {
			continue
		}
		
		entryData := data[offset : offset+384]
		if offset+384 > len(data) {
			entryData = data[offset:]
		}
		
		parsedType, pid, username, tty, host, loginTime := parseUtmpEntry(entryData)

		if *debugMode {
			if offset%384 == 0 {
				log.Printf("Entry %d (offset 0x%x): type=%d, pid=%d, user='%s', line='%s', host='%s'", 
					entryIndex, offset, parsedType, pid, username, tty, host)
			} else if parsedType == USER_PROCESS || parsedType == DEAD_PROCESS {
				log.Printf("Misaligned entry at offset 0x%x: type=%d, pid=%d, user='%s', line='%s', host='%s'", 
					offset, parsedType, pid, username, tty, host)
			}
		}

		// Only process USER_PROCESS entries (active user sessions) and DEAD_PROCESS (recently ended sessions)
		if parsedType != USER_PROCESS && parsedType != DEAD_PROCESS {
			continue
		}

		// Skip empty usernames
		if username == "" {
			continue
		}

		// For DEAD_PROCESS, only include if it's recent (within last hour)
		if parsedType == DEAD_PROCESS {
			if time.Since(loginTime) > time.Hour {
				if *debugMode {
					log.Printf("Skipping old DEAD_PROCESS entry: %s (age: %v)", username, time.Since(loginTime))
				}
				continue
			}
		}

		loginTimeStr := loginTime.Format("2006-01-02 15:04")

		if *debugMode {
			log.Printf("Found user session: user=%s, tty=%s, host=%s, time=%s, type=%d", 
				username, tty, host, loginTimeStr, parsedType)
		}

		validSessions++

		// Increment user count
		userCounts[username]++

		// Increment IP count (if it's not empty and not local)
		if host != "" && host != ":0" && host != ":0.0" && host != "console" {
			ipCounts[host]++
		}

		// Set user session info
		from := host
		if from == "" {
			from = "-"
		}
		userSessions.WithLabelValues(username, from, tty, loginTimeStr).Set(1)
		
		// Skip ahead to avoid processing the same entry multiple times
		if offset%384 != 0 {
			offset = ((offset / 384) + 1) * 384 - 4 // -4 because loop will add 4
		}
	}

	// Set total users metric
	totalUsers.Set(float64(validSessions))

	// Set user count metrics
	for username, count := range userCounts {
		userSessionCount.WithLabelValues(username).Set(float64(count))
	}

	// Set IP count metrics
	for ip, count := range ipCounts {
		userSessionByIP.WithLabelValues(ip).Set(float64(count))
	}

	if *debugMode {
		log.Printf("Collected metrics: %d total sessions, %d unique users, %d unique IPs", 
			validSessions, len(userCounts), len(ipCounts))
	}
}

func main() {
	flag.Parse()

	if *debugMode {
		log.Println("Debug mode enabled")
		log.Printf("Architecture: %s/%s", runtime.GOOS, runtime.GOARCH)
		log.Printf("Using utmp file: %s", *utmpPath)
		log.Printf("Expected utmp entry size: 384 bytes")
	}

	// Initial collection
	collectUserMetrics()

	// Start a goroutine to periodically collect metrics
	go func() {
		for {
			time.Sleep(15 * time.Second)
			collectUserMetrics()
		}
	}()

	// Set up HTTP server
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Unix User Exporter</title></head>
			<body>
			<h1>Unix User Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	log.Printf("Server listening on %s", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
